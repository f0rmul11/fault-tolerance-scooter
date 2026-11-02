package load

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type reportSummary struct {
	Sent           *int64           `json:"sent,omitempty"`
	Address        string           `json:"address"`
	Method         string           `json:"method"`
	TargetRPS      int              `json:"target_rps"`
	Duration       string           `json:"duration"`
	RequestTimeout string           `json:"request_timeout"`
	Partial        bool             `json:"partial"`
	Completed      int64            `json:"completed"`
	OK             int64            `json:"ok"`
	Errors         int64            `json:"errors"`
	ActualRPS      float64          `json:"completed_per_load_second"`
	P95MS          int64            `json:"p95_ms"`
	P99MS          int64            `json:"p99_ms"`
	Over300MS      int64            `json:"over_300ms"`
	Codes          map[string]int64 `json:"codes"`
}

// Millisecond buckets bound memory independently of the number of requests.
// Percentiles include successes AND errors; durations are rounded up by <1ms.
func readReport(path string, rows io.Writer) (reportSummary, error) {
	result := reportSummary{Codes: map[string]int64{}}
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer file.Close()
	fmt.Fprintln(rows, "number\tstarted_unix\ttag\tduration_ms\trequest_bytes\tcode\terrno")
	histogram := map[int64]int64{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 12 {
			return result, fmt.Errorf("некорректная строка phout %d", result.Completed+1)
		}
		micros, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || micros < 0 {
			return result, fmt.Errorf("некорректная длительность в phout")
		}
		ms := (micros + 999) / 1000
		histogram[ms]++
		result.Completed++
		if fields[11] == "grpc_0" && fields[10] == "0" {
			result.OK++
		} else {
			result.Errors++
		}
		if micros > 300000 {
			result.Over300MS++
		}
		code := fields[11]
		if fields[10] != "0" {
			code += " (errno=" + fields[10] + ")"
		}
		result.Codes[code]++
		if _, err := fmt.Fprintf(rows, "%d\t%s\t%s\t%.3f\t%s\t%s\t%s\n", result.Completed, fields[0], fields[1], float64(micros)/1000, fields[8], fields[11], fields[10]); err != nil {
			return result, err
		}
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	keys := make([]int64, 0, len(histogram))
	for ms := range histogram {
		keys = append(keys, ms)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	percentile := func(fraction float64) int64 {
		rank := int64(math.Ceil(float64(result.Completed) * fraction))
		var count int64
		for _, ms := range keys {
			count += histogram[ms]
			if count >= rank {
				return ms
			}
		}
		return 0
	}
	result.P95MS, result.P99MS = percentile(.95), percentile(.99)
	return result, nil
}

func saveReport(path, dir string, options *options, elapsed time.Duration, partial bool, sentCount int64, out io.Writer) error {
	rows, err := os.Create(filepath.Join(dir, "requests.tsv"))
	if err != nil {
		return err
	}
	summary, err := readReport(path, rows)
	closeErr := rows.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	summary.Address, summary.Method = options.addr, options.method
	summary.TargetRPS, summary.Duration = options.rps, options.duration.String()
	if sentCount >= 0 {
		summary.Sent = &sentCount
	}
	incomplete := sentCount < 0 || sentCount != summary.Completed
	summary.RequestTimeout, summary.Partial = options.requestTimeout.String(), partial || incomplete
	if elapsed > 0 {
		summary.ActualRPS = float64(summary.Completed) / elapsed.Seconds()
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), append(data, '\n'), 0600); err != nil {
		return err
	}
	var exceeded float64
	if summary.Completed > 0 {
		exceeded = 100 * float64(summary.Over300MS) / float64(summary.Completed)
	}
	text := fmt.Sprintf("Завершено: %d · OK: %d · ошибки: %d\nRPS: задано %d · по завершённым запросам %.2f\np95: %dms · p99: %dms · дольше 300ms: %.2f%%\nПроцентили включают успехи и ошибки. Неполный отчёт: %t\n", summary.Completed, summary.OK, summary.Errors, summary.TargetRPS, summary.ActualRPS, summary.P95MS, summary.P99MS, exceeded, summary.Partial)
	if summary.Sent != nil {
		text = fmt.Sprintf("Отправлено: %d\n", *summary.Sent) + text
	}
	codes := make([]string, 0, len(summary.Codes))
	for code := range summary.Codes {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		text += fmt.Sprintf("  %s: %d\n", code, summary.Codes[code])
	}
	if _, err := io.WriteString(out, text); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "summary.txt"), []byte(text), 0600); err != nil {
		return err
	}
	if incomplete && !partial {
		return fmt.Errorf("не все отправленные запросы попали в отчёт: отправлено %d, завершено %d", sentCount, summary.Completed)
	}
	return nil
}

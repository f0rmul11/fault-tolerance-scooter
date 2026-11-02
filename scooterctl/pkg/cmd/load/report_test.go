package load

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportIncludesFailuresInLatency(t *testing.T) {
	var data strings.Builder
	for i := 0; i < 100; i++ {
		duration, code := 1000, "grpc_0"
		if i >= 94 {
			duration, code = 500001, "grpc_14"
		}
		fmt.Fprintf(&data, "123.004\trequest\t%d\t0\t0\t0\t0\t0\t42\t0\t0\t%s\n", duration, code)
	}
	path := filepath.Join(t.TempDir(), "report.phout")
	if err := os.WriteFile(path, []byte(data.String()), 0600); err != nil {
		t.Fatal(err)
	}
	var rows bytes.Buffer
	result, err := readReport(path, &rows)
	if err != nil {
		t.Fatal(err)
	}
	if result.Completed != 100 || result.OK != 94 || result.Errors != 6 || result.Over300MS != 6 || result.P95MS != 501 || result.P99MS != 501 {
		t.Fatalf("incorrect report: %+v", result)
	}
	if !strings.Contains(rows.String(), "123.004\trequest\t500.001") {
		t.Fatal("readable rows lost timestamp or duration")
	}
}

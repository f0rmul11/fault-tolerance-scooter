package fault

import (
	"fmt"
	"time"
)

const (
	ControlPort = 9901
	MaxSlowdown = time.Minute
	MaxDuration = time.Hour
	GetScooter  = "/scooter.scooter.v1.ScooterService/GetScooter"
)

type Edge struct {
	From string
	To   string
	Port int
}

func Edges() []Edge {
	return []Edge{
		{
			From: "api-gateway",
			To:   "scooter",
			Port: 10001,
		},
	}
}

func FindEdge(from, to string) (Edge, error) {
	for _, edge := range Edges() {
		if edge.From == from && edge.To == to {
			return edge, nil
		}
	}
	return Edge{}, fmt.Errorf("доступна RPC-связь api-gateway → scooter")
}

type Options struct {
	From        string
	To          string
	Method      string
	Percent     int
	Slowdown    time.Duration
	Duration    time.Duration
	Unavailable bool
	Clear       bool
}

func (o Options) Validate() error {
	if _, err := FindEdge(o.From, o.To); err != nil {
		return err
	}
	if o.Clear {
		if o.Slowdown != 0 || o.Duration != 0 || o.Unavailable || o.Method != "" || o.Percent != 100 {
			return fmt.Errorf("--clear не совмещается с настройками сбоя")
		}
		return nil
	}
	if o.Method != "GetScooter" && o.Method != GetScooter {
		return fmt.Errorf("укажите --method GetScooter (или %s)", GetScooter)
	}
	if o.Percent < 1 || o.Percent > 100 {
		return fmt.Errorf("--percent должен быть от 1 до 100")
	}
	if o.Unavailable {
		if o.Slowdown != 0 {
			return fmt.Errorf("--unavailable нельзя совмещать с --slowdown")
		}
	} else if o.Slowdown < time.Millisecond || o.Slowdown > MaxSlowdown || o.Slowdown%time.Millisecond != 0 {
		return fmt.Errorf("укажите --unavailable или --slowdown от 1ms до 1m с точностью до миллисекунды")
	}
	if o.Duration < time.Millisecond || o.Duration > MaxDuration || o.Duration%time.Millisecond != 0 {
		return fmt.Errorf("--duration должен быть от 1ms до 1h с точностью до миллисекунды")
	}
	return nil
}

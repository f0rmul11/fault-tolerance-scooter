package rtconfig_test

import (
	"context"
	"fmt"

	config "api-gateway/internal/config"
	"api-gateway/internal/pkg/rtconfig"
)

func ExampleNew() {
	schema, err := rtconfig.NewSchema(rtconfig.Definition{
		Name:     "scooters_list_enabled",
		Type:     rtconfig.BoolType,
		Default:  false,
		Writable: true,
	})
	if err != nil {
		panic(err)
	}
	manager, err := rtconfig.New(schema, rtconfig.Options{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Prefix:    "/scooterflow/dev/services/api-gateway/rtconfig/v1/values/",
	})
	if err != nil {
		panic(err)
	}

	// Defaults are available before Start, without contacting etcd.
	cfg := config.NewClient(manager)
	fmt.Println(cfg.GetValue(context.Background(), config.ScootersListEnabled).Bool())
	if err := manager.Close(context.Background()); err != nil {
		panic(err)
	}
	// Output: false
}

package main

import (
	"context"
	"log/slog"
	"os"

	"pricing/internal"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx := context.Background()
	internal.New(ctx).Run(ctx)
}

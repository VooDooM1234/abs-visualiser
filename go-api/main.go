package main

import (
	"context"
	"fmt"
	"os"

	"github.com/VooDooM1234/abs-visualiser/server"
)

func main() {
	ctx := context.Background()
	if err := server.Run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

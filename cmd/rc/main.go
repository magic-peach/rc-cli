package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	rc "github.com/runtimeconditions/go-rc-cli"
)

var version = "dev"

func main() {
	err := rc.Run(context.Background(), version, os.Args[1:], os.Stdout, os.Stderr)
	var exit *rc.ExitError
	if errors.As(err, &exit) {
		os.Exit(exit.Code)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

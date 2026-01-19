package main

import (
	"fmt"
	"os"

	"github.com/febritecno/sqdesk-cli/internal/tui"
)

func main() {
	app, err := tui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SQDesk: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running SQDesk: %v\n", err)
		os.Exit(1)
	}
}

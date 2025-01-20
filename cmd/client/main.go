package main

import (
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/cli"
	"github.com/Renal37/goph-keeper/internal/zlog"
)

func main() {
	defer func() {
		_ = zlog.Logger().Sync()
	}()

	if err := run(); err != nil {
		fmt.Println("error:", err)
	}
}

func run() error {
	app, err := cli.NewApplication()
	if err != nil {
		return err
	}

	err = app.Run()
	if err != nil {
		fmt.Println("failed: ", err)
	}

	return nil
}

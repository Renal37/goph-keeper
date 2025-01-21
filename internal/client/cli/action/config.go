package action

import (
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/config"
	"github.com/urfave/cli/v2"
)

func ConfigAction(configFileName string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		params := map[string]string{}
		for _, flag := range ctx.FlagNames() {
			value := ctx.String(flag)
			if len(value) != 0 {
				params[flag] = value
			}
		}

		if len(params) == 0 {
			cli.ShowAppHelpAndExit(ctx, 1)
		}

		err := config.UpdateConfig(configFileName, params)
		if err != nil {
			fmt.Printf("update config error: %s\n", err)
		}

		return err
	}

}

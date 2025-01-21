package cli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Renal37/goph-keeper/internal/client/cli/action"
	"github.com/Renal37/goph-keeper/internal/client/cli/args"
	"github.com/Renal37/goph-keeper/internal/client/config"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
	"github.com/urfave/cli/v2"
)

const configFileName = "client_config.yaml"

type Application struct {
	cli     cli.App
	client  *transport.Client
	user    *storage.User
	storage storage.Storage
	config  *config.Config
}

func NewApplication() (*Application, error) {
	app := &Application{}

	conf, err := config.ReadConfig(configFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Use config command for authentification. See gophkeep-client config --help")

			return nil, err
		} else {
			fmt.Printf("Unknown error: %s\n", err)

			return nil, err
		}
	}

	app.initCLI()

	if conf != nil {
		app.config = conf

		dbStorage, err := sqlstorage.StartNewDbStorage(conf.Database)
		if err != nil {
			return nil, err
		}

		app.storage = dbStorage
		app.client = transport.NewClient(conf)

		user, err := app.storage.GetActive(context.Background())
		if err != nil {
			if !errors.Is(err, sqlstorage.ErrNotActiveOrRegistredUsers) {
				return nil, err
			}
		}

		// it's check need for case when we don't have active or registred users in client storage (e.g. first app start)
		if user != nil {
			user.CryptoKey, err = base64.RawStdEncoding.DecodeString(string(user.CryptoKey))
			if err != nil {
				return nil, err
			}

			app.user = user
		}
	}

	return app, nil
}

func (a *Application) initCLI() {
	a.cli = cli.App{
		Name:         "gophkeep",
		Version:      "v1.0.0",
		Compiled:     time.Now(),
		BashComplete: cli.DefaultAppComplete,
		Usage:        "Use for send your data to gokeep server",
		Commands: []*cli.Command{
			a.makeConfigCmd(),
			a.makeRegisterCmd(),
			a.makeDataCmd(),
			a.makeWalletCmd(),
			a.makeSecretCmd(),
		},
	}
}

func (a *Application) makeSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "secret",
		Usage:        "Operations with bank's cards",
		Before:       a.checkConfig,
		BashComplete: cli.DefaultAppComplete,
		Subcommands: []*cli.Command{
			a.makeCreateSecretCmd(),
			a.makeDeleteSecretCmd(),
			a.makeGetSecretCmd(),
		},
	}
}

func (a *Application) makeCreateSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
			&cli.StringFlag{Name: "key"},
			&cli.StringFlag{Name: "value"},
		},
		Action: func(ctx *cli.Context) error {
			secret, err := args.GetSecret(ctx)
			if err != nil {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}

			return action.CreateSecretAction(ctx.Context, a.user, a.storage, a.client, secret)
		},
	}
}

func (a *Application) makeDeleteSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
		},
		Action: func(ctx *cli.Context) error {
			secretName, err := args.GetSecretName(ctx)
			if err != nil {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}

			return action.DeleteSecretAction(ctx.Context, a.user, a.storage, a.client, secretName)
		},
	}
}

func (a *Application) makeGetSecretCmd() *cli.Command {
	return &cli.Command{
		Name:         "get",
		Usage:        "Create new secret",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name"},
		},
		Action: func(ctx *cli.Context) error {
			secretName, err := args.GetSecretName(ctx)
			if err != nil {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}

			return action.GetSecretAction(ctx.Context, a.user, a.storage, a.client, secretName)
		},
	}
}

func (a *Application) makeWalletCmd() *cli.Command {
	return &cli.Command{
		Name:         "wallet",
		Usage:        "Operations with bank's cards",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateCardCmd(),
			a.makeDeleteCardCmd(),
			a.makeListCardCmd(),
		},
	}
}

func (a *Application) makeCreateCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Create new card",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
			&cli.StringFlag{Name: "expiration", Usage: fmt.Sprintf("Data in format: %s", storage.ExpirationFormat)},
			&cli.StringFlag{Name: "cvv"},
			&cli.StringFlag{Name: "owner"},
		},
		Action: func(ctx *cli.Context) error {
			card, err := args.GetBankCard(ctx)
			if err != nil {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}

			return action.CreateCardActionHandler(ctx.Context, a.user, a.storage, a.client, card)
		},
	}
}

func (a *Application) makeDeleteCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Usage:        "Delete card",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "number"},
		},
		Action: func(ctx *cli.Context) error {
			card, err := args.GetCardNumber(ctx)
			if err != nil {
				cli.ShowSubcommandHelpAndExit(ctx, 1)
			}

			return action.DeleteCardActionHandler(ctx.Context, a.user, a.storage, a.client, card)
		},
	}
}

func (a *Application) makeListCardCmd() *cli.Command {
	return &cli.Command{
		Name:         "list",
		Usage:        "List with all user cards",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Action: func(ctx *cli.Context) error {
			return action.ListCardActionHandler(ctx.Context, a.user, a.storage, a.client)
		},
	}
}

func (a *Application) makeDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "data",
		Usage:        "Operations with text or binary data",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Subcommands: []*cli.Command{
			a.makeCreateDataCmd(),
			a.makeGetDataCmd(),
			a.makeListDataCmd(),
			a.makeUpdateDataCmd(),
			a.makeDeleteDataCmd(),
		},
	}
}

func (a *Application) makeConfigCmd() *cli.Command {
	return &cli.Command{
		Name:        "config",
		Usage:       "Client configuration",
		Description: "Add configuration to ~/.goph-keeper/client_config.yaml",
		Before:      a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server-url",
				Usage: "Pair of server's IP:port",
			},
			&cli.StringFlag{
				Name:  "database-name",
				Usage: "Database's file's name",
			},
		},
		Action: action.ConfigAction(configFileName),
	}
}

func (a *Application) makeRegisterCmd() *cli.Command {
	return &cli.Command{
		Name:         "register",
		Usage:        "Registrates user in system",
		Description:  "You shoud register before using application",
		BashComplete: cli.DefaultAppComplete,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "login",
				Usage: "User's login",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "User's password",
			},
		},
		Action: func(ctx *cli.Context) error {
			login := args.GetLogin(ctx)
			pass := args.GetPassword(ctx)

			return action.RegisterAction(ctx.Context, a.storage, a.client, login, pass)
		},
	}
}

func (a *Application) makeCreateDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Send new file to server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path of creating file",
			},
		},
		Description: "Read file and send it to server",
		Action: func(ctx *cli.Context) error {
			filename := args.GetFileArg(ctx)

			return action.CreateDataAction(ctx.Context, a.user, a.storage, a.client, filename)
		},
	}
}

func (a *Application) makeGetDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "get",
		Usage:        "Download file from server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Output directory",
			},
		},
		Action: func(ctx *cli.Context) error {
			filename := args.GetFileArg(ctx)

			return action.GetDataAction(ctx.Context, a.user, a.storage, a.client, filename)
		},
	}
}

func (a *Application) makeListDataCmd() *cli.Command {
	return &cli.Command{
		Name:   "list",
		Usage:  "Print local data names and revisions",
		Before: a.checkConfig,
		Action: func(ctx *cli.Context) error {
			return action.ListDataAction(ctx.Context, a.user, a.storage, a.client)
		},
	}
}

func (a *Application) makeUpdateDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "update",
		Usage:        "Update existed data on server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "update date",
			},
		},
		Action: func(ctx *cli.Context) error {
			filename := args.GetFileArg(ctx)

			return action.UpdateAction(ctx.Context, a.user, a.storage, a.client, filename)
		},
	}
}

func (a *Application) makeDeleteDataCmd() *cli.Command {
	return &cli.Command{
		Name:         "delete",
		Aliases:      []string{"d"},
		Usage:        "Delete data on server",
		BashComplete: cli.DefaultAppComplete,
		Before:       a.checkConfig,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "path to file",
			},
		},
		Action: func(ctx *cli.Context) error {
			filename := args.GetFileArg(ctx)

			return action.DeleteBinaryDataAction(ctx.Context, a.user, a.storage, a.client, filename)
		},
	}
}

func (a *Application) checkConfig(ctx *cli.Context) error {
	if a.config == nil {
		fmt.Println("client isn't configured")

		cli.ShowAppHelpAndExit(ctx, 1)
	}

	if a.user == nil {
		fmt.Println("User isn't registred")
		cli.ShowCommandHelpAndExit(ctx, "register", 1)
	}

	if a.client == nil {
		fmt.Println("Need register before using goph-keeper client. Use --help for more information.")
		cli.ShowCommandHelpAndExit(ctx, "register", 1)
	}

	return nil
}

func (a *Application) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}

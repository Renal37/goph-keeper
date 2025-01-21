package args

import (
	"errors"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/urfave/cli/v2"
)

func GetFileArg(ctx *cli.Context) string {
	filename := ctx.String("file")
	if len(filename) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return filename
}

func GetSecretName(ctx *cli.Context) (string, error) {
	name := ctx.String("name")
	if len(name) == 0 {
		return "", errors.New("bad secret's name")
	}

	return name, nil
}

func GetSecret(ctx *cli.Context) (*storage.Secret, error) {
	name := ctx.String("name")
	if len(name) == 0 {
		return nil, errors.New("bad secret's name")
	}

	key := ctx.String("key")
	if len(key) == 0 {
		return nil, errors.New("bad secret's key")
	}

	value := ctx.String("value")
	if len(value) == 0 {
		return nil, errors.New("bad secret's value")
	}

	return &storage.Secret{
		Name:  name,
		Key:   key,
		Value: value,
	}, nil
}

func GetBankCard(ctx *cli.Context) (*storage.BankCard, error) {
	number, ok := validateCardNumber(ctx.String("number"))
	if !ok {
		return nil, errors.New("bad card's number")
	}

	exp, ok := validateExpDate(ctx.String("expiration"))
	if !ok {
		return nil, errors.New("bad expiration date")
	}

	cvv, ok := validateCvvCode(ctx.String("cvv"))
	if !ok {
		return nil, fmt.Errorf("bad card's cvv=%s", ctx.String("cvv"))
	}

	owner, ok := validateCardOwner(ctx.String("owner"))
	if !ok {
		return nil, errors.New("bad card owner name")
	}

	return &storage.BankCard{
		Number:     number,
		ExpiryDate: exp,
		Owner:      owner,
		CvvCode:    cvv,
	}, nil
}

func GetCardNumber(ctx *cli.Context) (string, error) {
	number, ok := validateCardNumber(ctx.String("number"))
	if !ok {
		return "", errors.New("bad card number")
	}

	return number, nil
}

func GetLogin(ctx *cli.Context) string {
	value := ctx.String("login")
	if len(value) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return value
}

func GetPassword(ctx *cli.Context) string {
	value := ctx.String("password")
	if len(value) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	return value
}

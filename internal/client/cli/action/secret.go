package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
)

func CreateSecretAction(
	ctx context.Context,
	user *storage.User,
	storage storage.SecretStorage,
	client transport.SecretDataClient,
	secret *storage.Secret,
) error {
	cryptedSecret, err := storage.CreateSecret(ctx, user, secret)
	if err != nil && !errors.Is(err, sqlstorage.ErrAlreadyExist) {
		return err
	}

	if err = client.CreateSecret(ctx, user.Token, secret.Name, cryptedSecret); err != nil {
		return err
	}

	return nil
}

func GetSecretAction(
	ctx context.Context,
	user *storage.User,
	storage storage.SecretStorage,
	client transport.SecretDataClient,
	key string,
) error {
	secret, err := storage.GetSecret(ctx, user, key)
	if err != nil {
		if errors.Is(err, sqlstorage.ErrDataNotExist) {
			secret, err = client.GetSecret(ctx, user.Token, key)
			if err != nil {
				return err
			}

			if _, err = storage.CreateSecret(ctx, user, secret); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fmt.Printf("\tname: %s; key: %s; value: %s\n", secret.Name, secret.Key, secret.Value)

	return nil
}

func DeleteSecretAction(
	ctx context.Context,
	user *storage.User,
	storage storage.SecretStorage,
	client transport.SecretDataClient,
	key string,
) error {
	// we are firstly deleting data on the server
	if err := client.DeleteSecret(ctx, user.Token, key); err != nil {
		return err
	}

	if err := storage.DeleteSecret(ctx, user, key); err != nil {
		return err
	}

	return nil
}

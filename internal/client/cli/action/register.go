package action

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
)

func RegisterAction(
	ctx context.Context,
	s storage.UserStorage,
	client transport.RegisterClient,
	login string,
	password string,
) error {
	cryptoKey, err := gophcrypto.GenerateCryptoKey()
	if err != nil {
		return err
	}

	user := &storage.User{Login: login, IsActive: false, CryptoKey: cryptoKey}

	crypto, err := gophcrypto.New(user.CryptoKey)
	if err != nil {
		return err
	}

	encryptedPassword := crypto.Encrypt([]byte(password))
	user.Password = encryptedPassword

	token, err := client.RegisterUser(ctx, login, encryptedPassword)
	if err != nil {
		return fmt.Errorf("registration on server failed with error: %w", err)
	}

	err = s.Register(ctx, login, encryptedPassword, token, base64.RawStdEncoding.EncodeToString(cryptoKey))
	if err != nil {
		return err
	}

	return nil
}

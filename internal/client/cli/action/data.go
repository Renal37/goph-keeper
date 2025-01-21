package action

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/storage/sqlstorage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
)

func CreateDataAction(
	ctx context.Context,
	user *storage.User,
	s storage.DataStorage,
	client transport.BinaryDataClient,
	filename string,
) error {
	r, err := readDataFromFile(filename, user)
	if err != nil {
		return fmt.Errorf("read data from file, err=%w", err)
	}

	err = s.CreateData(ctx, user, r)
	if err != nil && !errors.Is(err, sqlstorage.ErrAlreadyExist) {
		return err
	}

	r.Revision = 1

	err = client.UploadBinaryData(ctx, user, r)
	if err != nil {
		return err
	}

	fmt.Printf("Data from file=%s is saved\ns", r.Name)

	return nil
}

func GetDataAction(
	ctx context.Context,
	user *storage.User,
	s storage.DataStorage,
	client transport.BinaryDataClient,
	filename string,
) error {
	var r *storage.Record

	r, err := s.LoadData(ctx, user, filename)
	if err != nil {
		if errors.Is(err, sqlstorage.ErrDataNotExist) {
			r, err = client.DownloadBinaryData(ctx, user, filename)
			if err != nil {
				return err
			}

			if err = s.CreateData(ctx, user, r); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if r == nil {
		return sqlstorage.ErrDataNotExist
	}

	decryptedData, err := decryptUserData(user, []byte(r.Data))
	if err != nil {
		return err
	}

	fmt.Println(string(decryptedData))

	return nil
}

func ListDataAction(
	ctx context.Context,
	user *storage.User,
	s storage.DataStorage,
	_ transport.BinaryDataClient,
) error {
	records, err := s.ListData(ctx, user)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("Records isn't exist")
	}

	for _, r := range records {
		fmt.Printf("\t%s (%d)\n", r.Name, r.Revision)
	}

	return nil
}

func UpdateAction(
	ctx context.Context,
	user *storage.User,
	s storage.DataStorage,
	client transport.BinaryDataClient,
	filename string,
) error {
	r, err := readDataFromFile(filename, user)
	if err != nil {
		return fmt.Errorf("read data from file, err=%w", err)
	}

	rev, needUpload, err := s.UpdateData(ctx, user, r)
	if err != nil {
		return err
	}

	if needUpload {
		r.Revision = rev

		err = client.UpdateBinaryData(ctx, user, r)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Nothing for updating")
	}

	return nil
}

func DeleteBinaryDataAction(
	ctx context.Context,
	user *storage.User,
	s storage.DataStorage,
	client transport.BinaryDataClient,
	name string,
) error {
	if err := s.DeleteData(ctx, user, name); err != nil {
		return err
	}

	if err := client.DeleteBinaryData(ctx, user, name); err != nil {
		return err
	}

	return nil
}

func readDataFromFile(filename string, user *storage.User) (*storage.Record, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	encryptedData, err := encryptUserData(user, data)
	if err != nil {
		return nil, err
	}

	r := &storage.Record{Name: filename, Data: string(encryptedData), Revision: 1}

	return r, nil
}

func decryptUserData(
	user *storage.User,
	data []byte,
) ([]byte, error) {
	crypto, err := gophcrypto.New(user.CryptoKey)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(data)
}

func encryptUserData(
	user *storage.User,
	data []byte,
) ([]byte, error) {
	crypto, err := gophcrypto.New([]byte(user.CryptoKey))
	if err != nil {
		return nil, err
	}

	return []byte(crypto.Encrypt(data)), nil
}

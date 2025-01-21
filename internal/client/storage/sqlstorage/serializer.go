package sqlstorage

import (
	"encoding/json"

	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
)

type DbSerializer struct {
	crypt *gophcrypto.Cryptographer
}

func NewSerializer(crypt *gophcrypto.Cryptographer) *DbSerializer {
	return &DbSerializer{crypt: crypt}
}

func (s *DbSerializer) SerializeBankCard(card *storage.BankCard) (string, error) {
	return doSerializarion(s, card)
}

func (s *DbSerializer) DeserializeBankCard(base64data string) (*storage.BankCard, error) {
	return doDeserializarion[storage.BankCard](s, base64data)
}

func (s *DbSerializer) SerializeSecret(card *storage.Secret) (string, error) {
	return doSerializarion(s, card)
}

func (s *DbSerializer) DeserializeSecret(base64data string) (*storage.Secret, error) {
	return doDeserializarion[storage.Secret](s, base64data)
}

func doSerializarion[T any](serializer *DbSerializer, obj *T) (string, error) {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	encryptedData := serializer.crypt.Encrypt(marshaled)

	return encryptedData, nil
}

func doDeserializarion[T any](serializer *DbSerializer, base64data string) (*T, error) {
	data, err := serializer.crypt.Decrypt([]byte(base64data))
	if err != nil {
		return nil, err
	}

	obj := new(T)
	err = json.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

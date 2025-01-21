package action

import (
	"context"
	"fmt"

	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/Renal37/goph-keeper/internal/client/transport"
)

func CreateCardActionHandler(
	ctx context.Context,
	user *storage.User,
	storage storage.WalletStorage,
	client transport.WalletClient,
	card *storage.BankCard,
) error {
	data, err := storage.CreateCard(ctx, user, card)
	if err != nil {
		return err
	}

	if err = client.CreateCardData(ctx, user.Token, card.Number, data); err != nil {
		return err
	}

	return nil
}

func DeleteCardActionHandler(
	ctx context.Context,
	user *storage.User,
	storage storage.WalletStorage,
	client transport.WalletClient,
	cardNumber string,
) error {
	if err := storage.DeleteCard(ctx, user, cardNumber); err != nil {
		return err
	}

	if err := client.DeleteCardData(ctx, user.Token, cardNumber); err != nil {
		return err
	}

	return nil
}

func ListCardActionHandler(
	ctx context.Context,
	user *storage.User,
	s storage.WalletStorage,
	client transport.WalletClient,
) error {
	list, err := s.ListCard(ctx, user)
	if err != nil {
		return err
	}

	for _, card := range list {
		fmt.Printf("holder: %s; num: %s; expiration: %v; cvv: %s\n", card.Owner, card.Number, card.ExpiryDate.Format(storage.ExpirationFormat), card.CvvCode)
	}

	return nil
}

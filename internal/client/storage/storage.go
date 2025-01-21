package storage

import "context"

//go:generate mockgen -source=storage.go -destination=./mock_storage.go -package=storage
type Storage interface {
	UserStorage
	DataStorage
	WalletStorage
	SecretStorage
	Stop() error
}

type DataStorage interface {
	CreateData(ctx context.Context, u *User, r *Record) error
	UpdateData(ctx context.Context, u *User, r *Record) (uint64, bool, error)
	LoadData(ctx context.Context, u *User, name string) (*Record, error)
	ListData(ctx context.Context, u *User) ([]*Record, error)
	DeleteData(ctx context.Context, u *User, name string) error
}

type UserStorage interface {
	Register(ctx context.Context, login string, password string, token string, cryptokey string) error
	GetActive(ctx context.Context) (*User, error)
}

type WalletStorage interface {
	CreateCard(ctx context.Context, u *User, c *BankCard) (string, error)
	ListCard(ctx context.Context, u *User) ([]*BankCard, error)
	DeleteCard(ctx context.Context, u *User, cardNumber string) error
}

type SecretStorage interface {
	CreateSecret(ctx context.Context, u *User, s *Secret) (string, error)
	GetSecret(ctx context.Context, u *User, secretKey string) (*Secret, error)
	DeleteSecret(ctx context.Context, u *User, secretKey string) error
}

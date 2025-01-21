package sqlstorage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Renal37/goph-keeper/internal/client/config"
	"github.com/Renal37/goph-keeper/internal/client/gophcrypto"
	"github.com/Renal37/goph-keeper/internal/client/storage"
	"github.com/mattn/go-sqlite3"
)

var (
	ErrNotExist         = errors.New("doesn't exist")
	ErrUpdateConflict   = errors.New("updae conflict")
	ErrUserNotRegistred = errors.New("user isn't registred")
)

var _ storage.DataStorage = &DbStorage{}
var _ storage.UserStorage = &DbStorage{}
var _ storage.WalletStorage = &DbStorage{}

type DbStorage struct {
	db *sql.DB
}

func StartNewDbStorage(dbName string) (*DbStorage, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(homedir, config.DefaultAppDirName, dbName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &DbStorage{db: db}

	if err = storage.initTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *DbStorage) initTables() error {
	var err error

	for _, t := range []string{
		createDataTableQuery,
		createUserTableQuery,
		createCardTableQuery,
		createSecretTableQuery,
	} {
		if _, err = s.db.Exec(t); err != nil {
			fmt.Println("init table", t)

			return err
		}
	}

	return nil
}

func (s *DbStorage) Register(
	ctx context.Context,
	login string,
	password string,
	token,
	cryptoKey string,
) error {
	cryptoKeyBase64 := base64.RawStdEncoding.EncodeToString([]byte(cryptoKey))

	if err := s.changeCurrentUserStatus(ctx); err != nil {
		return err
	}

	fmt.Println("User was registred. Context was switched to a new user.")

	return s.addNewUser(ctx, login, password, token, cryptoKeyBase64)
}

func (s *DbStorage) addNewUser(
	ctx context.Context,
	login string,
	password string,
	token,
	cryptoKey string,
) error {
	query := prepareInsertUserQuery(login, password, token, cryptoKey)

	_, err := s.db.ExecContext(ctx, query.request, query.args...)
	if err != nil {
		if isUniqueConstraint(err) {
			fmt.Println("User alreay registred in client")

			return nil
		}

		return fmt.Errorf("add new user error: %w", err)
	}

	return nil
}

func (s *DbStorage) changeCurrentUserStatus(ctx context.Context) error {
	var err error

	q := prepareGetUserQuery()
	rows, err := s.db.QueryContext(ctx, q.request, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		// don't have active user
		return nil
	}

	u := storage.User{IsActive: true}
	if err = rows.Scan(&u.Login, &u.Password, &u.Token, &u.CryptoKey); err != nil {
		return err
	}

	if err = rows.Err(); err != nil {
		return err
	}

	q = prepareChangeActiveQuery(u.Login)
	_, err = s.db.ExecContext(ctx, q.request, q.args...)
	if err != nil {
		return err
	}

	return nil
}

func isUniqueConstraint(err error) bool {
	sqliteErr, ok := err.(sqlite3.Error)
	if !ok {
		return false
	}

	return sqliteErr.Code.Error() == sqlite3.ErrConstraintUnique.Error()
}

var ErrNotActiveOrRegistredUsers = errors.New("don't have active or registred users")

func (s *DbStorage) GetActive(
	ctx context.Context,
) (*storage.User, error) {
	query := prepareGetUserQuery()

	rows, err := s.db.QueryContext(ctx, query.request)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotActiveOrRegistredUsers
	}

	user := &storage.User{IsActive: true}

	cryptoKeyBase64 := ""

	err = rows.Scan(&user.Login, &user.Password, &user.Token, &cryptoKeyBase64)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	cryptoKey, err := base64.RawStdEncoding.DecodeString(cryptoKeyBase64)
	if err != nil {
		return nil, err
	}

	user.CryptoKey = cryptoKey

	return user, nil
}

func (s *DbStorage) DeleteData(
	ctx context.Context,
	u *storage.User,
	name string,
) error {
	q := prepareDeleteDataQuery(u.Login, name)

	_, err := s.db.ExecContext(ctx, q.request, q.args...)

	return err
}

var ErrAlreadyExist = errors.New("already exist")

func (s *DbStorage) CreateData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record,
) error {
	err := s.saveNewData(ctx, u, r)
	if err != nil {
		if isUniqueConstraint(err) {
			return ErrAlreadyExist
		}
	}

	return nil
}

func (s *DbStorage) UpdateData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record,
) (uint64, bool, error) {
	storedData, err := s.LoadData(ctx, u, r.Name)
	if err != nil {
		return 0, false, fmt.Errorf("load data, err=%w", err)
	}

	if storedData.Data == r.Data {
		return storedData.Revision, false, nil
	}

	if err = s.updateData(ctx, u, r); err != nil {
		return 0, false, err
	}

	return storedData.Revision, true, nil
}

func (s *DbStorage) saveNewData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record) error {
	q := prepareAddDataQuery(u.Login, r.Name, r.Data)

	res, err := s.db.ExecContext(ctx, q.request, q.args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *DbStorage) updateData(
	ctx context.Context,
	u *storage.User,
	r *storage.Record,
) error {
	q := prepareUpdateDataQuery(u.Login, r.Name, r.Data)

	res, err := s.db.ExecContext(ctx, q.request, q.args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

var ErrDataNotExist = errors.New("data not exist")

func (s *DbStorage) LoadData(
	ctx context.Context,
	u *storage.User,
	name string,
) (*storage.Record, error) {
	query := prepareGetDataQuery(u.Login, name)

	rows, err := s.db.QueryContext(ctx, query.request, query.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrDataNotExist
	}

	record := &storage.Record{Name: name}
	err = rows.Scan(&record.Data, &record.Revision)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *DbStorage) ListData(
	ctx context.Context,
	u *storage.User,
) ([]*storage.Record, error) {
	query := prepareListDataQuery(u.Login)

	rows, err := s.db.QueryContext(ctx, query.request, query.args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	records := []*storage.Record{}

	for rows.Next() {
		r := storage.Record{}
		err = rows.Scan(&r.Name, &r.Data, &r.Revision)
		if err != nil {
			return nil, err
		}

		records = append(records, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (s *DbStorage) CreateCard(
	ctx context.Context,
	u *storage.User,
	c *storage.BankCard,
) (string, error) {
	data, err := serializeUserBankCard(u, c)
	if err != nil {
		return "", fmt.Errorf("serialize card err=%w", err)
	}

	query := prepareAddCard(u.Login, c.Number, data)

	_, err = s.db.ExecContext(ctx, query.request, query.args...)
	if err != nil {
		if isUniqueConstraint(err) {
			return "", ErrAlreadyExist
		}

		return "", err
	}

	return data, nil
}

func (s *DbStorage) DeleteCard(
	ctx context.Context,
	u *storage.User,
	number string,
) error {
	query := prepareDeleteCard(u.Login, number)

	_, err := s.db.ExecContext(ctx, query.request, query.args...)

	return err
}

func (s *DbStorage) ListCard(
	ctx context.Context,
	u *storage.User,
) ([]*storage.BankCard, error) {
	query := prepareListCard(u.Login)

	rows, err := s.db.QueryContext(ctx, query.request, query.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := make([]*storage.BankCard, 0, 10)

	for rows.Next() {
		data := ""
		err := rows.Scan(&data)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			return nil, errors.New("empty card data")
		}

		card, err := desirializeUserBankCard(u, data)
		if err != nil {
			return nil, fmt.Errorf("desirialize user's card's data err=%w", err)
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func serializeUserBankCard(u *storage.User, c *storage.BankCard) (string, error) {
	crypt, err := gophcrypto.New(u.CryptoKey)
	if err != nil {
		return "", err
	}

	serializer := NewSerializer(crypt)

	data, err := serializer.SerializeBankCard(c)
	if err != nil {
		return "", nil
	}

	return data, nil
}

func desirializeUserBankCard(u *storage.User, data string) (*storage.BankCard, error) {
	crypt, err := gophcrypto.New(u.CryptoKey)
	if err != nil {
		return nil, err
	}

	serializer := NewSerializer(crypt)

	card, err := serializer.DeserializeBankCard(data)
	if err != nil {
		return nil, nil
	}

	return card, nil
}

func (s *DbStorage) CreateSecret(
	ctx context.Context,
	u *storage.User,
	secret *storage.Secret,
) (string, error) {
	cryptedSecret, err := serializeSecret(u, secret)
	if err != nil {
		return "", err
	}

	q := prepareAddSecretQuery(u.Login, secret.Name, cryptedSecret)
	_, err = s.db.Exec(q.request, q.args...)
	if err != nil {
		if isUniqueConstraint(err) {
			return cryptedSecret, ErrAlreadyExist
		}

		return "", err
	}

	return cryptedSecret, nil
}

func (s *DbStorage) GetSecret(
	ctx context.Context,
	u *storage.User,
	secretName string,
) (*storage.Secret, error) {
	q := preareGetSecretQuery(u.Login, secretName)
	rows, err := s.db.QueryContext(ctx, q.request, q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("secret %s isn't exist", secretName)
	}

	cryptedData := ""
	if err = rows.Scan(&cryptedData); err != nil {
		return nil, err
	}

	return deserializeSecret(u, cryptedData)
}

func deserializeSecret(u *storage.User, cryptedData string) (*storage.Secret, error) {
	crypt, err := gophcrypto.New(u.CryptoKey)
	if err != nil {
		return nil, err
	}

	serializer := NewSerializer(crypt)

	secret, err := serializer.DeserializeSecret(cryptedData)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func serializeSecret(u *storage.User, s *storage.Secret) (string, error) {
	crypt, err := gophcrypto.New(u.CryptoKey)
	if err != nil {
		return "", err
	}

	serializer := NewSerializer(crypt)

	cryptedSecret, err := serializer.SerializeSecret(s)
	if err != nil {
		return "", err
	}

	return cryptedSecret, nil
}

func (s *DbStorage) DeleteSecret(
	ctx context.Context,
	u *storage.User,
	secretKey string,
) error {
	q := prepareDeleteSecretQuery(u.Login, secretKey)

	_, err := s.db.ExecContext(ctx, q.request, q.args...)

	return err
}

func (s *DbStorage) Stop() error {
	return s.db.Close()
}

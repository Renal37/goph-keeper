package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Renal37/goph-keeper/internal/server/handler"
	"github.com/Renal37/goph-keeper/internal/zlog"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var _ handler.DataStorage = &Storage{}
var _ handler.WalletStorage = &Storage{}
var _ handler.Registrator = &Storage{}
var _ handler.SecretStorage = &Storage{}

type Storage struct {
	db *sql.DB
}

func StartNewStorage(dataSourceName string) (*Storage, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("sql open dataSourceName=%s, err=%w", dataSourceName, err)
	}

	ctrl := &Storage{db: db}
	if err := ctrl.init(); err != nil {
		return nil, fmt.Errorf("init, err=%w", err)
	}

	return ctrl, nil
}

func (c *Storage) Stop() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("db close err=%w", err)
	}

	return nil
}

func (c *Storage) init() error {
	createTableQueries := []string{
		createUsersTableQuery,
		createBinaryDataTableQuery,
		createWalletTableQuery,
		createSecretTableQuery,
	}

	for _, q := range createTableQueries {
		if err := c.exec(q); err != nil {
			return fmt.Errorf("exec query=%s, err=%w", q, err)
		}
	}

	return nil
}

func (c *Storage) exec(query string) error {
	const createTablesTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("exec create table query, err=%w", err)
	}

	return nil
}

func (c *Storage) Check(ctx context.Context, token string) error {
	userQuery := prepareGetUserQuery(token)

	rows, err := doQuery(ctx, func(ctx context.Context) (*sql.Rows, error) {
		return c.db.QueryContext(ctx, userQuery.request, userQuery.args...)
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	storedUser := &handler.User{}

	if !rows.Next() {
		return handler.ErrUnknownUser
	}

	if err := rows.Scan(&storedUser.Login, &storedUser.Password, &storedUser.Token); err != nil {
		return err
	}

	if storedUser.Token != token {
		return handler.ErrUnknownUser
	}

	return nil
}

func (c *Storage) CreateData(ctx context.Context, userToken string, d *handler.Record) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	err = doTransactionExec(ctx, tx, prepareNewDataQuery(u.Login, d.Name, d.Data))
	if err != nil {
		if isNotUniqueError(err) {
			return handler.ErrDataAlreadyExist
		}

		return fmt.Errorf("add user=%s data=%s, err=%w", u.Login, d.Data, err)
	}

	return tx.Commit()
}

func (c *Storage) UpdateData(baseCtx context.Context, userToken string, d *handler.Record) error {
	ctx, cancel := context.WithTimeout(baseCtx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	storedData, err := getDataInTransaction(ctx, tx, u, d.Name)
	if err != nil {
		return err
	}

	if storedData.Revision != d.Revision {
		return fmt.Errorf("user=%s data=%s err=%w", u.Login, d.Name, handler.ErrBadRevision)
	}

	if storedData.Data == d.Data {
		// doesn't need in changes
		return nil
	}

	err = doTransactionExec(ctx, tx, prepareUpdateDataQuery(u.Login, d.Name, d.Data))
	if err != nil {
		return fmt.Errorf("do update user=%s, data=%s, rev=%d, err=%w", u.Login, d.Name, d.Revision, err)
	}

	return tx.Commit()
}

func (c *Storage) LoadData(baseCtx context.Context, userToken string, dataKey string) (*handler.Record, error) {
	ctx, cancel := context.WithTimeout(baseCtx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return nil, err
	}

	d, err := getDataInTransaction(ctx, tx, u, dataKey)
	if err != nil {
		return nil, err
	}

	return d, tx.Commit()
}

func (c *Storage) ListData(baseCtx context.Context, userToken string) ([]*handler.Record, error) {
	ctx, cancel := context.WithTimeout(baseCtx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return nil, err
	}

	q := prepareListBinaryData(u.Login)

	rows, err := tx.QueryContext(ctx, q.request, q.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*handler.Record, 0, 10)
	for rows.Next() {
		r := &handler.Record{}
		if err = rows.Scan(&r.Name, &r.Data, &r.Revision); err != nil {
			return nil, err
		}

		records = append(records, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil

}

func (c *Storage) DeleteData(ctx context.Context, userToken string, d *handler.Record) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	err = deleteDataInTransaction(ctx, tx, u, d.Name)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Storage) Register(ctx context.Context, u *handler.User) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	query := prepareGetUserQuery(u.Login)

	user, err := doTransactionQuery(ctxWithTimeout, tx, query, func(rows *sql.Rows) (*handler.User, error) {
		if !rows.Next() {
			return nil, handler.ErrUnknownUser
		}

		user := &handler.User{}
		if err := rows.Scan(&user.Login, &user.Password); err != nil {
			return nil, err
		}

		return user, nil
	})
	if err == nil {
		if user.Password != u.Password {
			return handler.ErrBadPassword
		}

		return nil
	}

	if !errors.Is(err, handler.ErrUnknownUser) {
		return err
	}

	query = prepareCreateUserQuery(u.Login, u.Password, u.Token)

	err = doTransactionExec(ctxWithTimeout, tx, query)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Storage) CreateCard(ctx context.Context, userToken string, card *handler.CardData) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	if err = doTransactionExec(ctx, tx, prepareAddCard(u.Login, card.Number, card.Data)); err != nil {
		return nil
	}

	return tx.Commit()
}

func (c *Storage) ListCard(ctx context.Context, userToken string) ([]*handler.CardData, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return nil, err
	}

	cards, err := doTransactionQuery(ctx, tx, prepareListCard(u.Login), func(rows *sql.Rows) ([]*handler.CardData, error) {
		list := make([]*handler.CardData, 0, 10)

		for rows.Next() {
			card := &handler.CardData{}
			if err = rows.Scan(&card.Data, &card.Number); err != nil {
				return nil, err
			}

			list = append(list, card)
		}

		return list, nil
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (c *Storage) DeleteCard(ctx context.Context, userToken string, card *handler.CardData) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	u, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	return doTransactionExec(ctx, tx, prepareDeleteCard(u.Login, card.Number))
}

func recoverAndRollBack(tx *sql.Tx) {
	_ = tx.Rollback()

	if err := recover(); err != nil {
		zlog.Logger().Errorf("tx rollback failed, err=%w", err)
	}
}

func (c *Storage) CreateSecret(ctx context.Context, userToken string, secret *handler.Secret) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	user, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	query := prepareAddSecret(user.Login, secret.Key, secret.Value)

	err = doTransactionExec(ctx, tx, query)
	if err != nil {
		return fmt.Errorf("add secret user=%s data=%s, err=%w", user.Login, secret.Key, err)
	}

	return tx.Commit()
}

func (c *Storage) GetSecret(ctx context.Context, userToken, secretKey string) (*handler.Secret, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	user, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return nil, err
	}

	secret, err := doTransactionQuery(ctx, tx, prepareGetSecret(user.Login, secretKey), func(rows *sql.Rows) (*handler.Secret, error) {
		if !rows.Next() {
			return nil, fmt.Errorf("secretKey=%s, err=%w", secretKey, handler.ErrDataNotFound)
		}

		storedData := &handler.Secret{Key: secretKey}
		if err := rows.Scan(&storedData.Key); err != nil {
			return nil, err
		}

		return storedData, nil
	})

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return secret, nil
}

func (c *Storage) DeleteSecret(ctx context.Context, userToken, secretKey string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return err
	}
	defer recoverAndRollBack(tx)

	user, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return err
	}

	err = doTransactionExec(ctx, tx, prepareDeleteSecret(user.Login, secretKey))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *Storage) ListSecret(ctx context.Context, userToken string) ([]*handler.Secret, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := c.db.BeginTx(ctxWithTimeout, nil)
	if err != nil {
		return nil, err
	}
	defer recoverAndRollBack(tx)

	user, err := getUserInTransaction(ctx, tx, userToken)
	if err != nil {
		return nil, err
	}

	list, err := doTransactionQuery(ctx, tx, prepareListSecret(user.Login), func(rows *sql.Rows) ([]*handler.Secret, error) {
		secrets := make([]*handler.Secret, 0, 10)
		for rows.Next() {
			s := &handler.Secret{}
			if err = rows.Scan(&s.Key, &s.Value); err != nil {
				return nil, err
			}

			secrets = append(secrets, s)
		}

		return secrets, nil
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return list, nil
}

func getDataInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	u *handler.User,
	dataKey string,
) (*handler.Record, error) {
	return doTransactionQuery(ctx, tx, prepareGetDataQuery(u.Login, dataKey), func(rows *sql.Rows) (*handler.Record, error) {
		if !rows.Next() {
			return nil, fmt.Errorf("key=%s, err=%w", dataKey, handler.ErrDataNotFound)
		}

		storedData := &handler.Record{Name: dataKey}
		if err := rows.Scan(&storedData.Data, &storedData.Revision); err != nil {
			return nil, err
		}

		return storedData, nil
	})
}

func deleteDataInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	u *handler.User,
	dataKey string,
) error {
	return doTransactionExec(ctx, tx, prepareDeleteDataQuery(u.Login, dataKey))
}

func getUserInTransaction(
	ctx context.Context,
	tx *sql.Tx,
	userToken string,
) (*handler.User, error) {
	userQuery := prepareGetUserQuery(userToken)

	storedUser, err := doTransactionQuery(ctx, tx, userQuery, func(rows *sql.Rows) (*handler.User, error) {
		storedUser := &handler.User{}

		if !rows.Next() {
			return nil, fmt.Errorf("user isn't registred err=%w", handler.ErrUnknownUser)
		}

		if err := rows.Scan(&storedUser.Login, &storedUser.Password, &storedUser.Token); err != nil {
			return nil, err
		}

		return storedUser, nil
	})
	if err != nil {
		return nil, err
	}

	return storedUser, nil
}

// ----------------------------------------------------------------------------------------------
// -------------------------------------- Internal Methods --------------------------------------
// ----------------------------------------------------------------------------------------------

var tryingIntervals = []time.Duration{
	time.Millisecond * 100,
	time.Millisecond * 300,
	time.Millisecond * 500,
}

func doQuery[T any](ctx context.Context, queryFunc func(context.Context) (T, error)) (T, error) {
	var commonErr error
	max := len(tryingIntervals)

	var result T
	var err error

	for trying := 0; trying <= max; trying++ {
		result, err = queryFunc(ctx)
		if err != nil {
			commonErr = errors.Join(commonErr, err)

			if trying < max && isRetriableError(err) {
				time.Sleep(tryingIntervals[trying])
				continue
			}

			return result, commonErr
		}

		return result, nil
	}

	return result, commonErr
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}

func isNotUniqueError(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func doTransactionExec(ctx context.Context, tx *sql.Tx, query *query) error {
	_, err := doQuery(ctx, func(ctx context.Context) (sql.Result, error) {
		return tx.ExecContext(ctx, query.request, query.args...)
	})

	return err
}

func doTransactionQuery[T any](ctx context.Context, tx *sql.Tx, query *query, parseFun func(rows *sql.Rows) (T, error)) (T, error) {
	var result T

	rows, err := doQuery(ctx, func(ctx context.Context) (*sql.Rows, error) {
		return tx.QueryContext(ctx, query.request, query.args...)
	})
	if err != nil {
		return result, nil
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			zlog.Logger().Errorf("rows close err=%s", err)
		}
	}()

	result, err = parseFun(rows)
	if err != nil {
		return result, err
	}

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

package sqlstorage

const (
	createUserTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login" 		text NOT NULL,
		"password"		text NOT NULL,
		"token"         text NOT NULL,
		"crypto_key" 	text NOT NULL,
		"active"		integer NOT NULL,
		PRIMARY KEY "login"
	);`

	insertUser   = `INSERT INTO users ("login", "password", "token",  "crypto_key", "active") VALUES ($1, $2, $3, $4, 1);`
	getUser      = `SELECT "login", "password", "token", "crypto_key" FROM users WHERE "active" == 1;`
	changeActive = `UPDATE users SET "active" = 0 WHERE "user" = $1`
)

func prepareInsertUserQuery(login, password, token, crypto_key string) *query {
	return &query{request: insertUser, args: []interface{}{login, password, token, crypto_key}}
}

func prepareGetUserQuery() *query {
	return &query{request: getUser}
}

func prepareChangeActiveQuery(login string) *query {
	return &query{request: changeActive, args: []any{login}}
}

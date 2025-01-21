package sql

const (
	createSecretTableQuery = `CREATE TABLE IF NOT EXISTS secrets (
		"user"				text	NOT NULL,
		"secret_key"		text	NOT NULL,
		"secret_value"		text	NOT NULL,
		PRIMARY KEY ( "user", "secret_key" )
	);`

	addSecret    = `INSERT INTO secrets ("user", "secret_key", "secret_value") VALUES ($1, $2, $3);`
	getSecret    = `SELECT "secret_value" FROM secrets WHERE "user" = $1 AND "secret_key" = $2;`
	deleteSecret = `DELETE FROM secrets WHERE "user" = $1 AND "card_number" = $2;`
	listSecret   = `SELECT "secret_key", "secret_value" FROM secrets WHERE "user" = $1`
)

func prepareAddSecret(user, secretKey, secretValue string) *query {
	return &query{request: addSecret, args: []any{user, secretKey, secretValue}}
}

func prepareGetSecret(user, secretKey string) *query {
	return &query{request: getSecret, args: []any{user, secretKey}}
}

func prepareDeleteSecret(user, secretKey string) *query {
	return &query{request: deleteSecret, args: []any{user, secretKey}}
}

func prepareListSecret(user string) *query {
	return &query{request: listSecret, args: []any{user}}
}

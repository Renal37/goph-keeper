package sqlstorage

const (
	createSecretTableQuery = `CREATE TABLE IF NOT EXISTS secrets (
		"user"			text		NOT_NULL,
		"name"			text		NOT_NULL,
		"secret"		text		NOT NULL,
		PRIMARY KEY ( "user", "name" )
	);`

	addSecretQuery    = `INSERT INTO secrets ("user", "name", "secret") VALUES ($1, $2, $3);`
	getSecretQuery    = `SELECT "secret" FROM secrets WHERE "user" = $1 AND "name" = $2;`
	deleteSecretQuery = `DELETE FROM secrets WHERE "user" = $1 AND "name" = $2;`
)

func prepareAddSecretQuery(user string, name string, cryptedSecret string) *query {
	return &query{request: addSecretQuery, args: []any{user, name, cryptedSecret}}
}

func preareGetSecretQuery(user, name string) *query {
	return &query{request: getSecretQuery, args: []any{user, name}}
}

func prepareDeleteSecretQuery(user, name string) *query {
	return &query{request: deleteSecretQuery, args: []any{user, name}}
}

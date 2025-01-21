package sql

const (
	createUsersTableQuery = `CREATE TABLE IF NOT EXISTS users (
		"login"		text NOT NULL,
		"password"  text NOT NULL,
		"token"		text NOT NULL,
		PRIMARY KEY ( "token" )
	);`

	createUserQuery = `INSERT INTO users (login, password, token) VALUES ($1, $2, $3);`
	getUserByToken  = `SELECT * FROM users WHERE token = $1;`
)

func prepareCreateUserQuery(login, password, token string) *query {
	return &query{request: createUserQuery, args: []any{login, password, token}}
}

func prepareGetUserQuery(token string) *query {
	return &query{request: getUserByToken, args: []any{token}}
}

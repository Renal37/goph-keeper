package sqlstorage

const (
	createCardTableQuery = `CREATE TABLE IF NOT EXISTS cards (
		"user"			text		NOT_NULL,
		"number"		text		NOT NULL,
		"data"			text		NOT NULL,
		PRIMARY KEY ( "user", "number" )
	);`

	addCard    = `INSERT INTO cards ("user", "number", "data") VALUES ($1, $2, $3);`
	listCard   = `SELECT "data" FROM cards WHERE "user" = $1;`
	deleteCard = `DELETE FROM cards WHERE "user" = $1 AND "number" = $2;`
)

func prepareAddCard(user, number, data string) *query {
	return &query{request: addCard, args: []any{user, number, data}}
}

func prepareListCard(user string) *query {
	return &query{request: listCard, args: []any{user}}
}

func prepareDeleteCard(user, number string) *query {
	return &query{request: deleteCard, args: []any{user, number}}
}

package sql

const (
	createWalletTableQuery = `CREATE TABLE IF NOT EXISTS wallet (
		"user"				text	NOT NULL,
		"card_number"		text	NOT NULL,
		"card_data"			text	NOT NULL,
		PRIMARY KEY ( "user", "card_number" )
	);`

	addCard    = `INSERT INTO wallet ("user", "card_number", "card_data") VALUES ($1, $2, $3);`
	listCard   = `SELECT "card_number", "card_data" FROM wallet WHERE "user" = $1;`
	deleteCard = `DELETE FROM wallet WHERE "user" = $1 AND "card_number" = $2;`
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

package sql

const (
	createBinaryDataTableQuery = `CREATE TABLE IF NOT EXISTS binary_data (
		"user"			text	NOT NULL,
		"key"			text	NOT NULL,
		"value"			text	NOT NULL,
		"revision"		bigint 	NOT NULL,
		"metainfo"      text,
		PRIMARY KEY ( "user", "key" )
	);`

	addNewBinaryDataQuery = `INSERT INTO binary_data ("user", "key", "value", "revision") VALUES ($1, $2, $3, 1);`
	updateBinaryDataQuery = `UPDATE binary_data SET "value" = $3, "revision" = "revision" + 1 WHERE "user" = $1 AND "key" = $2;`
	getBinaryData         = `SELECT "value", "revision" FROM binary_data WHERE "user" = $1 AND "key" = $2;`
	deleteBinaryData      = `DELETE FROM binary_data WHERE "user" = $1 AND "key" = $2;`
	listBinaryData        = `SELECT "key", "value", "revision" FROM binary_data WHERE "user" = $1;`
)

func prepareNewDataQuery(user, key, value string) *query {
	return &query{request: addNewBinaryDataQuery, args: []any{user, key, value}}
}

func prepareGetDataQuery(user, key string) *query {
	return &query{request: getBinaryData, args: []any{user, key}}
}

func prepareUpdateDataQuery(user, key, data string) *query {
	return &query{request: updateBinaryDataQuery, args: []any{user, key, data}}
}

func prepareDeleteDataQuery(user, key string) *query {
	return &query{request: deleteBinaryData, args: []any{user, key}}
}

func prepareListBinaryData(user string) *query {
	return &query{request: listBinaryData, args: []any{user}}
}

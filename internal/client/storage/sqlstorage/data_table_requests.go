package sqlstorage

const (
	createDataTableQuery = `CREATE TABLE IF NOT EXISTS data (
		"user"			text		NOT_NULL,
		"key"			text		NOT NULL,
		"value"			text		NOT NULL,
		"revision"		integer 	NOT NULL,
		PRIMARY KEY ( "user", "key" )
	);`

	addNewDataQuery  = `INSERT INTO data ("user", "key", "value", "revision") VALUES ($1, $2, $3, 1);`
	updateDataQuery  = `UPDATE data SET "value" = $1, "revision" = "revision" + 1 WHERE "user" = $2 AND "key" = $3;`
	getRevisionQuery = `SELECT "revision" FROM data WHERE "user" = $1 AND "key" = $2;`
	getData          = `SELECT "value", "revision" FROM data WHERE "user" = $1 AND "key" = $2;`
	listData         = `SELECT "key", "value", "revision" FROM data WHERE "user" = $1;`
	deleteBinaryData = `DELETE FROM data WHERE "user" = $1 AND "key" = $2;`
)

func prepareAddDataQuery(user string, key string, value string) *query {
	return &query{request: addNewDataQuery, args: []any{user, key, value}}
}

func prepareUpdateDataQuery(user, key, value string) *query {
	return &query{request: updateDataQuery, args: []any{value, user, key}}
}

func prepareGetDataQuery(user, key string) *query {
	return &query{request: getData, args: []any{user, key}}
}

func prepareListDataQuery(user string) *query {
	return &query{request: listData, args: []any{user}}
}

func prepareDeleteDataQuery(user string, dataKey string) *query {
	return &query{request: deleteBinaryData, args: []any{user, dataKey}}
}

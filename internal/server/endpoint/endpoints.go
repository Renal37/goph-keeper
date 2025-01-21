package endpoint

// endpoints
const (
	// POST - registrer new user or auth existing user
	RegisterEndpoint = "/api/user/register"
	// PUT - load new data to storage
	// POST - update binary data to storage
	// GET - get binary data from storage
	// DELETE - delete item from storage
	BinaryDataEndpoint = "/api/data/binary"

	// GET - get all binary data
	BinariesDataEndpoint = "/api/data/binaries"

	// PUT, GET, DELETE
	// Key = Value secret
	SecretEndpoint = "/api/data/secret"

	// GET
	SecretsEndpoint = "/api/data/secrets"

	// PUT, GET, DELETE
	// card data
	WalletEndpoint = "/api/data/wallet"

	// GET
	WalletsEndpoint = "/api/data/wallets"
)

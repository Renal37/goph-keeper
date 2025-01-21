package storage

import "time"

type Record struct {
	Name     string
	Data     string
	Revision uint64
}

type User struct {
	Login     string
	Password  string
	Token     string
	IsActive  bool
	CryptoKey []byte
}

type Secret struct {
	Name  string
	Key   string
	Value string
}

const ExpirationFormat = "2006-01-02"

type BankCard struct {
	Number     string
	ExpiryDate time.Time
	Owner      string
	CvvCode    string
}

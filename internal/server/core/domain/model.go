// Package domain contains the core domain models for the application,
// such as the `User` and `Storage` structures, which represent users
// and data storage entries, respectively. These structures are used in
// the business logic of the application and are stored in the database
// using an ORM (such as GORM).
package domain

// User represents a user in the system. It includes an ID,
// login, hashed password, and additional data for working with
// the database. The `Password` field has the tag `gorm:"-:all"`
// to exclude it from all ORM operations (create, read, etc.).
type User struct {
	ID       int    `json:"id"    gorm:"type:serial;autoIncrement;primaryKey;unique;not null"`
	Login    string `json:"login" gorm:"type:string;size:256;unique;not null"`
	Password string `json:"password" gorm:"-:all"`
	Hash     string `gorm:"type:string;size:1000;not null"`
}

// Storage represents a data storage entry in the system.
// It includes an ID, name, type, value, key, and owner (user ID).
// This structure is used to represent various types of data stored
// in the system. All fields have corresponding tags for JSON and
// ORM GORM, ensuring proper data storage and serialization.
type Storage struct {
	ID    int    `json:"id"    gorm:"type:serial;autoIncrement;primaryKey;unique;not null"`
	Name  string `json:"name"  gorm:"type:string;size:256;not null"`
	Type  string `json:"type"  gorm:"type:string;size:256;not null"`
	Value string `json:"text"  gorm:"type:string;not null"`
	Key   string `gorm:"type:string;size:1000;not null"`
	Owner int    `json:"owner" gorm:"type:int;not null"`
}

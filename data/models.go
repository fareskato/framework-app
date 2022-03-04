package data

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

///////////////////////////
// Models errors variables
///////////////////////////
var (
	ErrorDuplicateEmailMessage = "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"
	ErrorDuplicateEmail        = errors.New("duplicate email, please use another email")
	ErrorRecordNotFound        = errors.New("record not found")
	ErrorPasswordMatch         = errors.New("password matches error")
)

////////////////
// DB variables
////////////////
var db *sql.DB
var upper db2.Session

// Models type: wraps all models, and it's filed of application type,
// and it's accessible though out the entire application
type Models struct {
	Users  User
	Tokens Token
}

// New create Models type
func New(dbPool *sql.DB) Models {
	db = dbPool
	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(dbPool)
	case "postgres", "postgresql":
		upper, _ = postgresql.New(dbPool)
	default:
		// nothing to do
	}
	return Models{
		Users:  User{},
		Tokens: Token{},
	}
}

// getInsertedId returns the id as an int
// because postgres default is int64 and mysql(mariadb) default is int
func getInsertedId(id db2.ID) int {
	idType := fmt.Sprintf("%T", id)
	// postgres returns int65
	if idType == "int64" {
		return int(id.(int64))
	}
	// mysql,mariadb return int
	return id.(int)
}

package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	ID 				int
	Name 			string
	Email 			string 
	HashedPassword 	[]byte
	Created			time.Time
}

type UserModel struct {
	DB 					*sql.DB
	InsertStmt 			*sql.Stmt
	AuthenticateStmt 	*sql.Stmt
	ExistsStmt 			*sql.Stmt
}

func NewUserModel(db *sql.DB) (*UserModel, error) {
	insertStmt, err :=
		db.Prepare(`INSERT INTO users (name, email, hashed_password, created)
			 		VALUES (?, ?, ?, UTC_TIMESTAMP())`)
	if err != nil { return nil, err }

	existsStmt, err :=
		db.Prepare("SELECT EXISTS(SELECT true FROM users WHERE id = ?)")
	if err != nil { return nil, err }

	model := &UserModel{ DB: db,
						 InsertStmt: insertStmt,
						 ExistsStmt: existsStmt, }

	return model, nil
}

/*	Insert creates a new record within the 'users' table	*/
func (m *UserModel) Insert(name, email, password string) (int, error) {

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil { return 0, err }
	
	result, err := m.InsertStmt.Exec(name, email, hashedPass)

	// If there is an error, we can check what kind of SQL error it is
	if err != nil { 
	
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return 0, ErrDuplicateEmail
			}
		}

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil { return 0, err }

	return int(id), nil
}
/*	Authenticate verifies whether a users with the given email and password exists.
	If it does, return its ID.	*/
func (m *UserModel) Authenticate(email, password string) (int, error) {

	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)

	if err != nil {
		// There is no matching email in the DB
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}

		// Any other error
		return 0, err
	}

	// Check if the given password matches the stored one
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	// Otherwise, the credentials are correct; return the user's id
	return id, nil

}


/*	Exists checks whether a user with a given id exists in the database	*/
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	err := m.ExistsStmt.QueryRow(id).Scan(&exists)
	return exists, err
}
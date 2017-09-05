package dbmodels

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	// "github.com/zemirco/uid"
	// "time"
	"errors"
	"fmt"
)

type Page struct {
	Title string
	Body  []byte
	Timestamp string
	PostURL string
	Creator string
}
type User struct {
	Username string
	Email string
	UserURL string
	Passhash []byte

}
func CheckLogin(db *sql.DB, input *User) error {
	var savedhash []byte
	err := db.QueryRow("SELECT Users.Password FROM Users WHERE Users.Email = ?", input.Email).Scan(&savedhash)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}
	err = bcrypt.CompareHashAndPassword(savedhash, []byte(input.Passhash))
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}
	return nil
}

func UserSignup(db *sql.DB, input *User) (int64, error) {
	var boolint int
	err := db.QueryRow(`SELECT EXISTS(SELECT 1
							  FROM Users
							  WHERE Users.Username = ?
							  OR Users.Email = ?)`, input.Username, input.Email).Scan(&boolint)
	if err != nil {
		fmt.Printf("%s", err)
		return -1, err
	}
	if boolint == 1 {
		return -1, errors.New("account already exists")
	} else {
		prep, err := db.Prepare(
			`INSERT INTO Users (Users.Email, Users.Username, Users.UserURL, Users.Password)
				  VALUES (?, ?, ?, ?)`)
		if err != nil {
			fmt.Printf("%s", err)
			return -1, err
		}
		passhash, err := bcrypt.GenerateFromPassword([]byte(input.Passhash), bcrypt.DefaultCost)
		if err != nil {

			fmt.Printf("%s", err)
			return -1, err
		}

		updt, err := prep.Exec(input.Email, input.Username, input.UserURL, passhash)
		if err != nil {
			fmt.Printf("%s", err)
			return -1, err
		}
		id, _ := updt.LastInsertId()
		return id, nil
	}
}

func SingleApp(id string, db *sql.DB) (*Page, error) {
	p := new(Page)
	err := db.QueryRow(
		`SELECT Posts.Title, Posts.Content, Posts.CreatedOn, Posts.PostURL, Users.Username
			  FROM Posts
			  INNER JOIN Users ON Posts.CreatorID = Users.UserID
			  WHERE Posts.PostURL = ?`, id).Scan(&p.Title, &p.Body, &p.Timestamp, &p.PostURL, &p.Creator)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func AllApps(db *sql.DB) ([]*Page, error) {
	rows, err := db.Query(
		`SELECT Posts.Title, Posts.Content, Posts.CreatedOn, Posts.PostURL, Users.Username
			  FROM Posts
			  INNER JOIN Users ON Posts.CreatorID = Users.UserID`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bks := make([]*Page, 0)
	for rows.Next() {
		bk := new(Page)
		err := rows.Scan(&bk.Title, &bk.Body, &bk.Timestamp, &bk.PostURL, &bk.Creator)
		if err != nil {
			return nil, err
		}
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return bks, nil
}
func NewApp(db *sql.DB, input *Page) (int64, error) {
	prep, err := db.Prepare(
		`INSERT INTO Applications (AppName, Description, AppURL)
			  VALUES (?, ?, ?)`)
	if err != nil {
		err.Error()
	}
	updt, err := prep.Exec(input.Title, input.Body, input.PostURL)
	if err != nil {
		err.Error()
	}
	insert, err := updt.LastInsertId()
	return insert, err

}
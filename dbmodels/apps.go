package dbmodels

import (
	"database/sql"
	"html/template"

	"golang.org/x/crypto/bcrypt"
	// "github.com/zemirco/uid"
	"errors"
	"fmt"
	"time"
)

type Page struct {
	Title      string
	Body       template.HTML
	Timestamp  string
	PostURL    string
	Creator    string
	CreatorURL string
}
type User struct {
	Username string
	Email    string
	UserURL  string
	Userpass string
	Joindate string
}

func CheckLogin(db *sql.DB, input *User) (string, error) {
	var savedhash []byte
	var userurl string
	err := db.QueryRow("SELECT Users.Password, Users.UserURL FROM Users WHERE Users.Email = ?", input.Email).Scan(&savedhash, &userurl)
	if err != nil {
		fmt.Printf(time.Now().String()+"%s\n", err)
		return "", err
	}
	err = bcrypt.CompareHashAndPassword(savedhash, []byte(input.Userpass))
	if err != nil {
		fmt.Printf(time.Now().String()+"%s\n", err)
		return "", err
	}
	return userurl, nil
}

func UserSignup(db *sql.DB, input *User) (int64, error) {
	var boolint int
	err := db.QueryRow(`SELECT EXISTS(SELECT 1
							  FROM Users
							  WHERE Users.Username = ?
							  OR Users.Email = ?)`, input.Username, input.Email).Scan(&boolint)
	if err != nil {
		fmt.Printf(time.Now().String()+"%s\n", err)
		return 0, errors.New("Error checking database")
	}
	if boolint == 1 {
		return 0, errors.New("account already exists")
	} else {
		prep, err := db.Prepare(
			`INSERT INTO Users (Users.Email, Users.Username, Users.UserURL, Users.Password)
				  VALUES (?, ?, ?, ?)`)
		if err != nil {
			fmt.Printf(time.Now().String()+"%s\n", err)
			return 0, errors.New("failed db prep")
		}
		passhash, err := bcrypt.GenerateFromPassword([]byte(input.Userpass), bcrypt.DefaultCost)
		if err != nil {

			fmt.Printf(time.Now().String()+"%s\n", err)
			return 0, errors.New("cant generate pwhash")
		}

		updt, err := prep.Exec(input.Email, input.Username, input.UserURL, passhash)
		if err != nil {
			fmt.Printf(time.Now().String()+"%s\n", err)
			return 0, errors.New("failed to insert user")
		}
		id, _ := updt.LastInsertId()
		return id, nil
	}
}

func SingleApp(id string, db *sql.DB) (*Page, error) {
	p := new(Page)
	var placeholder []byte
	err := db.QueryRow(
		`SELECT Posts.Title, Posts.Content, Posts.CreatedOn, Posts.PostURL, Posts.CreatorURL, Users.Username
			  FROM Posts
			  INNER JOIN Users ON Posts.CreatorURL = Users.UserURL
			  WHERE Posts.PostURL = ?`, id).Scan(&p.Title, &placeholder, &p.Timestamp, &p.PostURL, &p.CreatorURL, &p.Creator)
	if err != nil {
		return nil, err
	}
	p.Body = template.HTML(placeholder)
	return p, nil
}
func AllApps(db *sql.DB) ([]*Page, error) {
	rows, err := db.Query(
		`SELECT Posts.Title, Posts.Content, Posts.CreatedOn, Posts.PostURL, Posts.CreatorURL, Users.Username
			  FROM Posts
			  INNER JOIN Users ON Posts.CreatorURL = Users.UserURL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bks := make([]*Page, 0)
	var placeholder []byte
	for rows.Next() {
		bk := new(Page)
		err = rows.Scan(&bk.Title, &placeholder, &bk.Timestamp, &bk.PostURL, &bk.CreatorURL, &bk.Creator)
		if err != nil {
			return nil, err
		}
		bk.Body = template.HTML(placeholder)
		bks = append(bks, bk)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return bks, nil
}
func AllUsers(db *sql.DB) ([]*User, error) {
	rows, err := db.Query(
		`SELECT Users.Username, Users.UserURL, Users.CreatedOn
		FROM Users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*User, 0)
	for rows.Next() {
		user := new(User)
		err = rows.Scan(&user.Username, &user.UserURL, &user.Joindate)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func NewApp(db *sql.DB, input *Page) (int64, error) {
	prep, err := db.Prepare(
		`INSERT INTO Posts (Title, Content, PostURL, CreatorURL)
			  VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	articlebyte := []byte(input.Body)
	updt, err := prep.Exec(input.Title, articlebyte, input.PostURL, input.CreatorURL)
	if err != nil {
		return 0, err
	}
	insert, err := updt.LastInsertId()
	return insert, err

}

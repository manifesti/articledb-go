package dbmodels

import (
	"database/sql"
	// "github.com/zemirco/uid"
	// "time"
)

type Page struct {
	Id int
	Title string
	Body  []byte
	Timestamp string
	AppURL string
}

func SingleApp(id string, db *sql.DB) (*Page, error) {
	p := new(Page)
	err := db.QueryRow(
		`SELECT AppID, AppName, Description, CreatedOn, AppURL
			  FROM Applications
			  WHERE AppURL = ?`, id).Scan(&p.Id, &p.Title, &p.Body, &p.Timestamp, &p.AppURL)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func AllApps(db *sql.DB) ([]*Page, error) {
	rows, err := db.Query(
		`SELECT AppID, AppName, Description, CreatedOn, AppURL
			  FROM Applications`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bks := make([]*Page, 0)
	for rows.Next() {
		bk := new(Page)
		err := rows.Scan(&bk.Id, &bk.Title, &bk.Body, &bk.Timestamp, &bk.AppURL)
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
	updt, err := prep.Exec(input.Title, input.Body, input.AppURL)
	if err != nil {
		err.Error()
	}
	insert, err := updt.LastInsertId()
	return insert, err

}
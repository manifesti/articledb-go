package main

import (
	"os"
	"structapp/dbmodels"
	// "strconv"
	// "time"
	// "fmt"
	"html/template"
	"net/http"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"log"
	"github.com/zemirco/uid"
	"fmt"
)

type App struct {
	db *sql.DB
	templates *template.Template
}

func main() {
	// open db connection
	if len(os.Args) < 3 {
		fmt.Printf("structapp 0.0.2\n\nstart with \"./structapp {mysql user} {mysql password} {database name}\"\n")
		return
	}
	db, err := sql.Open("mysql", os.Args[1]+":"+os.Args[2]+"@/"+os.Args[3])
	if err != nil {
		log.Panic(err)
	}
	// cache templates
	templates := template.Must(template.ParseFiles("templates/new.html", "templates/view.html", "templates/home.html"))
	// init app struct
	env := &App{db: db, templates: templates}

	// open http server
	http.HandleFunc("/view/", env.viewApp)
	http.HandleFunc("/new/", env.createApp)
	http.ListenAndServe(":8080", nil)
}

func (env *App) viewApp(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	title := req.URL.Path[len("/view/"):]
	if len(title) == 0 {
		bks, err := dbmodels.AllApps(env.db)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		renderList(w, "home", bks, env.templates)
		return
	}
	if len(title) != 10 {
		err := "AppID on väärän pituinen."
		http.Error(w, err, 404)
		return
	}
		bk, err := dbmodels.SingleApp(title, env.db)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	renderApplication(w, "view", bk, env.templates)
	return
}

func (env *App) createApp(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderApplication(w, "new", new(dbmodels.Page), env.templates)
		return
	}
	if req.Method == "POST" {
		p := new(dbmodels.Page)
		p.Title = req.FormValue("appname")
		p.Body = []byte(req.FormValue("body"))
		p.AppURL = uid.New(10)
		last, err := dbmodels.NewApp(env.db, p)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		fmt.Printf("application %s added into database as id %d", p.Title, last)

		http.Redirect(w, req, "/view/"+p.AppURL, http.StatusFound)
		return
	}
}
func renderApplication(w http.ResponseWriter, tmpl string, p *dbmodels.Page, templates *template.Template) {

	err := templates.ExecuteTemplate(w, tmpl +".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderList(w http.ResponseWriter, tmpl string, p []*dbmodels.Page, templates *template.Template) {
	err := templates.ExecuteTemplate(w, tmpl +".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
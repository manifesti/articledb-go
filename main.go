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
	"github.com/gorilla/sessions"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/zemirco/uid"
	"fmt")

type App struct {
	db *sql.DB
	templates *template.Template
	sesStorage *sessions.CookieStore
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
	templates := template.Must(template.ParseFiles("templates/new.html", "templates/view.html",
		"templates/login.html", "templates/signup.html", "templates/home.html"))
	// session storage
	sesStorage := sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32)))
	// init app struct
	env := &App{db: db, templates: templates, sesStorage: sesStorage}

	// static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// open http server
	http.HandleFunc("/signup/", env.signupRoute)
	http.HandleFunc("/view/", env.viewApp)
	http.HandleFunc("/new/", env.createApp)
	http.HandleFunc("/login/", env.loginRoute)
	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func (env *App) signupRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderStatic(w, env.templates, "signup")
		return
	}
	if req.Method == "POST" {
		info := new(dbmodels.User)
		info.Username = req.FormValue("username")
		info.Email = req.FormValue("email")
		info.Passhash = []byte(req.FormValue("p"))
		info.UserURL = uid.New(10)
		out, err := dbmodels.UserSignup(env.db, info)
		if err != nil {
			fmt.Printf("oops i did it again in signup")
			http.Redirect(w, req, "/signup/", http.StatusFound)
			return
		}
		fmt.Printf("user %s added into database as id %d", info.Email, out)
		http.Redirect(w, req, "/view/", http.StatusFound)
		return
	}
}

func (env *App) loginRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderStatic(w, env.templates, "login")
		return
	}
	if req.Method == "POST" {
		info := new(dbmodels.User)
		info.Email = req.FormValue("email")
		info.Passhash = []byte(req.FormValue("p"))
		err := dbmodels.CheckLogin(env.db, info)
		if err != nil {
			fmt.Printf("oops i did it again in login")
			http.Redirect(w, req, "/login/", http.StatusFound)
			return
		}
		session, _ := env.sesStorage.Get(req, "golangcookie")
		session.Values["loggedin"] = true
		session.Save(req, w)
		http.Redirect(w, req, "/view/", http.StatusFound)
		return
	}
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
	check, _ := env.sesStorage.Get(req, "golangcookie")
	if check.Values["loggedin"] != true {
		http.Redirect(w, req, "/login/", http.StatusFound)
		return
	}
	if req.Method == "GET" {
		renderApplication(w, "new", new(dbmodels.Page), env.templates)
		return
	}
	if req.Method == "POST" {
		p := new(dbmodels.Page)
		p.Title = req.FormValue("appname")
		p.Body = []byte(req.FormValue("body"))
		p.PostURL = uid.New(10)
		last, err := dbmodels.NewApp(env.db, p)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		fmt.Printf("article %s added into database as id %d", p.Title, last)
		http.Redirect(w, req, "/view/"+p.PostURL, http.StatusFound)
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
func renderStatic(w http.ResponseWriter, templates *template.Template, choice string) {
	err := templates.ExecuteTemplate(w, choice + ".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
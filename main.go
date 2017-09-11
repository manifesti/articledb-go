package main

import (
	"os"
	"structapp/dbmodels"
	// "strconv"
	// "time"
	// "fmt"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zemirco/uid"
)

type App struct {
	db         *sql.DB
	templates  *template.Template
	sesStorage *sessions.CookieStore
}
type listData struct {
	Loginstatus bool
	Pagesdata   []*dbmodels.Page
}
type singleData struct {
	Loginstatus bool
	Pagedata    *dbmodels.Page
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
	templates := template.Must(template.ParseFiles("templates/new.gohtml", "templates/view.gohtml",
		"templates/login.gohtml", "templates/signup.gohtml", "templates/home.gohtml", "templates/header.gohtml",
		"templates/headmenu.gohtml"))
	// session storage
	sesStorage := sessions.NewCookieStore([]byte(securecookie.GenerateRandomKey(32)))
	// init app struct
	env := &App{db: db, templates: templates, sesStorage: sesStorage}

	// static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// open http server
	http.HandleFunc("/logout/", env.logoutRoute)
	http.HandleFunc("/signup/", env.signupRoute)
	http.HandleFunc("/view/", env.viewPage)
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
		fmt.Printf("%s %s", info.Email, info.Passhash)
		userurl, err := dbmodels.CheckLogin(env.db, info)
		if err != nil {
			fmt.Printf("oops i did it again in login")
			http.Redirect(w, req, "/login/", http.StatusFound)
			return
		}
		session, _ := env.sesStorage.Get(req, "golangcookie")
		session.Values["loggedin"] = true
		session.Values["userurl"] = userurl
		session.Save(req, w)
		http.Redirect(w, req, "/view/", http.StatusFound)
		return
	}
}
func (env *App) logoutRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "405 method not allowed", 405)
		return
	}
	session, _ := env.sesStorage.Get(req, "golangcookie")
	session.Values["loggedin"] = false
	session.Values["userurl"] = ""
	http.Redirect(w, req, "/view/", http.StatusFound)
}

func (env *App) viewPage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	title := req.URL.Path[len("/view/"):]
	if len(title) == 0 {
		totmpl := new(listData)
		session, _ := env.sesStorage.Get(req, "golangcookie")
		if session.Values["loggedin"] != true {
			totmpl.Loginstatus = false
		} else {
			totmpl.Loginstatus = true
		}
		bks, err := dbmodels.AllApps(env.db)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		totmpl.Pagesdata = bks
		renderList(w, "home", totmpl, env.templates)
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
		return
	}
	totmpl := new(singleData)
	totmpl.Pagedata = bk
	session, _ := env.sesStorage.Get(req, "golangcookie")
	if session.Values["loggedin"] != true {
		totmpl.Loginstatus = false
	} else {
		totmpl.Loginstatus = true
	}
	renderApplication(w, "view", totmpl, env.templates)
	return
}

func (env *App) createApp(w http.ResponseWriter, req *http.Request) {
	session, _ := env.sesStorage.Get(req, "golangcookie")
	if session.Values["loggedin"] != true {
		http.Redirect(w, req, "/login/", http.StatusFound)
		return
	}
	if req.Method == "GET" {
		input := new(singleData)
		input.Loginstatus = true
		renderApplication(w, "new", input, env.templates)
		return
	}
	if req.Method == "POST" {
		cookiedata := session.Values["userurl"]
		var cookiestring string
		cookiestring, ok := cookiedata.(string)
		if !ok {
			http.Error(w, "ongelma keksin kanssa", 500)
			fmt.Printf("ongelma keksin kanssa")
			return
		}

		p := new(dbmodels.Page)
		p.Title = req.FormValue("appname")
		p.Body = []byte(req.FormValue("body"))
		p.PostURL = uid.New(10)
		p.CreatorURL = cookiestring
		_, err := dbmodels.NewApp(env.db, p)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Printf("article %s added into database as id %s", p.Title, p.PostURL)
		http.Redirect(w, req, "/view/"+p.PostURL, http.StatusFound)
		return
	}
}
func renderApplication(w http.ResponseWriter, tmpl string, p *singleData, templates *template.Template) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderList(w http.ResponseWriter, tmpl string, p *listData, templates *template.Template) {
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderStatic(w http.ResponseWriter, templates *template.Template, choice string) {
	err := templates.ExecuteTemplate(w, choice, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

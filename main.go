package main

import (
	"articledb-go/dbmodels"
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/zemirco/uid"
	"golang.org/x/crypto/acme/autocert"
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

// global regex for email validation
var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func main() {
	// open db connection
	if len(os.Args) < 3 {
		fmt.Printf("articledb-go 0.1.2\n\nstart with \"./articledb-go {mysql user} {mysql password} {database name}\"\n")
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
	// routes
	http.HandleFunc("/logout/", env.logoutRoute)
	http.HandleFunc("/signup/", env.signupRoute)
	http.HandleFunc("/view/", env.viewPage)
	http.HandleFunc("/new/", env.createApp)
	http.HandleFunc("/login/", env.loginRoute)
	http.HandleFunc("/", env.homeRoute)
	// LetsEncrypt setup
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("toreni.us"), // your domain here
		Cache:      autocert.DirCache("certs"),          // folder for storing certificates
	}
	server := &http.Server{
		Addr:      ":1443",
		Handler:   context.ClearHandler(http.DefaultServeMux),
		TLSConfig: &tls.Config{GetCertificate: certManager.GetCertificate},
	}
	// open https server
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Printf("ListenAndServe: %s", err)
	}
}

func (env *App) signupRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderStatic(w, env.templates, "signup")
		return
	}
	if req.Method == "POST" {
		info := new(dbmodels.User)
		_ = req.ParseForm()
		info.Username = req.FormValue("username")
		info.Email = req.FormValue("email")
		info.Userpass = req.FormValue("password")
		passcheck := req.FormValue("conf")
		if info.Userpass != passcheck {
			w.Write([]byte("1"))
			return
		}
		if len(info.Username) > 32 {
			w.Write([]byte("2"))
			return
		}
		if len(info.Email) > 255 || !emailRegexp.MatchString(info.Email) {
			w.Write([]byte("3"))
			return
		}
		info.UserURL = uid.New(10)
		_, err := dbmodels.UserSignup(env.db, info)
		if err != nil {
			fmt.Printf(time.Now().String() + "oops i did it again in signup\n")
			w.Write([]byte("4"))
			return
		}
		fmt.Printf(time.Now().String()+"user %s added into database as id %s\n", info.Email, info.UserURL)
		w.Write([]byte("5"))
		return
	}
}

func (env *App) loginRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		renderStatic(w, env.templates, "login")
		return
	}
	if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			fmt.Printf(time.Now().String() + "oops i did it again in login\n")
			http.Redirect(w, req, "/login/", http.StatusFound)
			return
		}
		info := new(dbmodels.User)
		info.Email = req.FormValue("email")
		info.Userpass = req.FormValue("password")
		fmt.Printf(time.Now().String()+"%s logged in\n", info.Email)
		userurl, err := dbmodels.CheckLogin(env.db, info)
		if err != nil {
			w.Write([]byte("1"))
			return
		}
		session, _ := env.sesStorage.Get(req, "golangcookie")
		session.Values["loggedin"] = true
		session.Values["userurl"] = userurl
		session.Save(req, w)
		w.Write([]byte("2"))
		// http.Redirect(w, req, "/view/", http.StatusFound)
		return
	}
}
func (env *App) logoutRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "405 method not allowed", 405)
		return
	}
	session, _ := env.sesStorage.Get(req, "golangcookie")
	session.Options.MaxAge = -1
	session.Save(req, w)
	http.Redirect(w, req, "/", http.StatusFound)
}

func (env *App) homeRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
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

func (env *App) viewPage(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	title := req.URL.Path[len("/view/"):]
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
			fmt.Printf(time.Now().String() + "ongelma keksin kanssa\n")
			return
		}

		p := new(dbmodels.Page)
		p.Title = req.FormValue("appname")
		unsafe := blackfriday.MarkdownBasic([]byte(req.FormValue("body")))
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		p.Body = template.HTML(html) // []byte(req.FormValue("body"))
		p.PostURL = uid.New(10)
		p.CreatorURL = cookiestring
		_, err := dbmodels.NewApp(env.db, p)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Printf(time.Now().String()+"article %s added into database as id %s\n", p.Title, p.PostURL)
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

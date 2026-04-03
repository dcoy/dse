package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/workos/workos-go/v3/pkg/directorysync"
	"github.com/workos/workos-go/v3/pkg/usermanagement"
)

var (
	key    = []byte("super-secret-key")
	store  = sessions.NewCookieStore(key)
	router = http.NewServeMux()
)

type Profile struct {
	First_name  string
	Last_name   string
	Raw_profile string
}

var conf struct {
	Addr        string
	APIKey      string
	ClientID    string
	RedirectURI string
	Connection  string
	Provider    string
	DirectoryID string
}

func loadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	flag.StringVar(&conf.Addr, "addr", ":8000", "The server addr.")
	flag.StringVar(&conf.APIKey, "api-key", os.Getenv("WORKOS_API_KEY"), "The WorkOS API key.")
	flag.StringVar(&conf.ClientID, "client-id", os.Getenv("WORKOS_CLIENT_ID"), "The WorkOS client id.")
	flag.StringVar(&conf.RedirectURI, "redirect-uri", os.Getenv("WORKOS_REDIRECT_URI"), "The redirect uri.")
	flag.StringVar(&conf.Connection, "connection", os.Getenv("WORKOS_CONNECTION"), "Use the Connection ID associated with your SSO Connection.")
	flag.StringVar(&conf.Provider, "provider", "", "The OAuth provider used for the SSO connection.")
	flag.StringVar(&conf.DirectoryID, "directory-id", os.Getenv("WORKOS_DIRECTORY_ID"), "Use the Directory ID associated with your Directory Sync connection.")
	flag.Parse()

	log.Printf("launching DSE demo with configuration: %+v", conf)

	usermanagement.SetAPIKey(conf.APIKey)
	directorysync.SetAPIKey(conf.APIKey)
}

func init() {
	loadEnvVariables()
}

func directoryUsers(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("./static/directory_users.html"))

	list, err := directorysync.ListUsers(
		context.Background(),
		directorysync.ListUsersOpts{
			Directory: conf.DirectoryID,
		},
	)
	if err != nil {
		log.Printf("get users failed: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		FirstName string
		LastName  string
		Users     interface{}
	}{
		FirstName: session.Values["first_name"].(string),
		LastName:  session.Values["last_name"].(string),
		Users:     list.Data,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Panic(err)
	}
}

func signin(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "cookie-name")
	if err != nil {
		log.Println(err)
	}

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		http.Redirect(w, r, "/logged_in", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/signin/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = false
	session.Values["first_name"] = ""
	session.Values["last_name"] = ""
	session.Values["raw_profile"] = ""

	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getAuthorizationURL(loginType string) (*url.URL, error) {
	opts := usermanagement.GetAuthorizationURLOpts{
		ClientID:    conf.ClientID,
		RedirectURI: conf.RedirectURI,
		ConnectionID: conf.Connection,
		Provider: "authkit",
	}
	return usermanagement.GetAuthorizationURL(opts)
}

func login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Panic(err)
	}

	loginType := r.Form.Get("login_method")

	authURL, err := getAuthorizationURL(loginType)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authURL.String(), http.StatusSeeOther)
}

func callback(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	log.Printf("callback is called with %s", r.URL)

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	authResp, err := usermanagement.AuthenticateWithCode(
		context.Background(),
		usermanagement.AuthenticateWithCodeOpts{
			ClientID:  conf.ClientID,
			Code:      code,
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		},
	)
	if err != nil {
		log.Printf("authenticate with code failed: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userJSON, err := json.MarshalIndent(authResp, "", "    ")
	if err != nil {
		log.Printf("marshal auth response failed: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = true
	session.Values["first_name"] = authResp.User.FirstName
	session.Values["last_name"] = authResp.User.LastName
	session.Values["raw_profile"] = userJSON

	if err := session.Save(r, w); err != nil {
		log.Panic(err)
	}

	thisProfile := Profile{
		First_name:  authResp.User.FirstName,
		Last_name:   authResp.User.LastName,
		Raw_profile: string(userJSON),
	}

	if err := tmpl.Execute(w, thisProfile); err != nil {
		log.Panic(err)
	}
}

func loggedin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/logged_in.html"))
	session, _ := store.Get(r, "cookie-name")

	thisProfile := Profile{
		First_name:  session.Values["first_name"].(string),
		Last_name:   session.Values["last_name"].(string),
		Raw_profile: string(session.Values["raw_profile"].([]byte)),
	}

	if err := tmpl.Execute(w, thisProfile); err != nil {
		log.Panic(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router.HandleFunc("/login", login)
	router.HandleFunc("/callback", callback)
	router.HandleFunc("/logged_in", loggedin)
	router.HandleFunc("/directory-users", directoryUsers)
	router.HandleFunc("/", signin)
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/signin/", http.StripPrefix("/signin/", http.FileServer(http.Dir("static"))))
	router.HandleFunc("/logout", logout)

	if err := http.ListenAndServe(conf.Addr, router); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

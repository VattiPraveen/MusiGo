package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type FileInfo struct {
	Name  string      `json:"Name"`
	IsDir bool        `json:"IsDir"`
	Mode  os.FileMode `json:"Mode"`
}

const (
	filePrefix = "/music/"
	root       = "./music"
)

var templates *template.Template
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
var users = map[string]string{
	"praveen": "praveen",
	"admin":   "admin",
}

func main() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
	RunServer()
}

func RunServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", getLogin).Methods("GET")
	router.HandleFunc("/", login).Methods("POST")
	router.HandleFunc(filePrefix, fileHandler)

	fs := http.FileServer(http.Dir("./music"))
	router.PathPrefix("/music/").Handler(http.StripPrefix("/music/", fs))

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalln("There's an error with the server", err)
	}

}

func getLogin(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.PostForm.Get("uname")
	password := r.PostForm.Get("psw")

	if originalPassword, ok := users[username]; ok {
		session, _ := store.Get(r, "session.id")
		if password == originalPassword {
			session.Values["authenticated"] = true
			session.Save(r, w)
		} else {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
	} else {
		http.Error(w, "User is not found", http.StatusNotFound)
		return
	}
	templates.ExecuteTemplate(w, "player.html", nil)
	// w.Write([]byte("Logged In successfully"))
	// http.Redirect(w, r, "/music/", 302)
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(root, r.URL.Path[len(filePrefix):])
	stat, err := os.Stat(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if stat.IsDir() {
		serveDir(w, r, path)
		return
	}

	http.ServeFile(w, r, path)

}

func serveDir(w http.ResponseWriter, r *http.Request, path string) {
	defer func() {
		if err, ok := recover().(error); ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	files, err := file.Readdir(-1)
	// fmt.Println((files[0].Name()))
	if err != nil {
		panic(err)
	}

	fileinfos := make([]FileInfo, len(files), len(files))

	for i := range files {
		fileinfos[i].Name = files[i].Name()
		fileinfos[i].IsDir = files[i].IsDir()
		fileinfos[i].Mode = files[i].Mode()
		// fmt.Println(files[i].Mode())
	}

	err = json.NewEncoder(w).Encode(fileinfos)
	// j := json.NewEncoder(w)

	if err != nil {
		panic(err)
	}

	// fmt.Println(fileinfos)
}

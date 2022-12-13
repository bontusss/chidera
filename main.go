package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/bontusss/go-chatapp/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
)

// set the active Avatar implementation
var avatars Avatar = TryAvatars{
	UseFileSystemAvatar,
	UseAuthAvatar,
	UseGravatar}

// templ represents a single template
type templateHandler struct {
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.templ == nil {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	}

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

var host = flag.String("host", ":8080", "The host of the application.")

func main() {

	flag.Parse() // parse the flags

	// setup gomniauth
	gomniauth.SetSecurityKey("AIzaSyB8dvFpceTGIFHqAY4p8bOVAUFVlWA-GlA")
	gomniauth.WithProviders(google.New("931859203863-ju5hte1pdmo17et0nm9bg0jmc9aatprj.apps.googleusercontent.com", "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", "http://localhost:8080/auth/callback/google"))

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
	http.Handle("/upload", &templateHandler{filename: "upload.html"})
	// http.HandleFunc("/uploader", uploaderHandler)

	http.Handle("/avatars/",
		http.StripPrefix("/avatars/",
			http.FileServer(http.Dir("./avatars"))))

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
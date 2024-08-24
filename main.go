package main

import (
	"html/template"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("index.html").ParseFiles("templates/layout.html", "templates/index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = t.ExecuteTemplate(w, "layout", "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("POST /generate", func(w http.ResponseWriter, r *http.Request) {
		text := r.FormValue("string")
		CreateDataSegment(text)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.ListenAndServe(":9876", mux)
}

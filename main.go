package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

const IMAGE_NAME = "imgName"
const STATIC_DIR = "static"

func main() {
	mux := http.NewServeMux()

	fileServerDir := fmt.Sprintf("./%s", STATIC_DIR)
	fileServerPattern := fmt.Sprintf("GET /%s/", STATIC_DIR)
	fileServerStrip := fmt.Sprintf("/%s/", STATIC_DIR)

	fs := http.FileServer(http.Dir(fileServerDir))
	mux.Handle(fileServerPattern, http.StripPrefix(fileServerStrip, fs))

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

	mux.HandleFunc("GET /{imgName}", func(w http.ResponseWriter, r *http.Request) {
		q := r.PathValue(IMAGE_NAME)
		url := fmt.Sprintf("%s/%s.png", STATIC_DIR, q)
		t, err := template.New("index.html").ParseFiles("templates/layout.html", "templates/index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = t.ExecuteTemplate(w, "layout", url)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("POST /generate", func(w http.ResponseWriter, r *http.Request) {
		text := r.FormValue("string")
		qrc, err := qrcode.New(text)
		if err != nil {
			fmt.Printf("could not generate QRCode: %v", err)
			return
		}
		id := uuid.New()
		filePath := fmt.Sprintf("%s/%s.png", STATIC_DIR, id.String())
		wr, err := standard.New(filePath)
		if err != nil {
			fmt.Printf("standard.New failed: %v", err)
			return
		}
		if err = qrc.Save(wr); err != nil {
			fmt.Printf("could not save image: %v", err)
		}
		redirectUrl := fmt.Sprintf("/%s", id.String())
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	})

	http.ListenAndServe(":9876", mux)
}

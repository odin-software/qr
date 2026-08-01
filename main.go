package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

// Assets fijos y plantillas viajan dentro del binario: el deploy es un solo
// ELF, sin directorios que sincronizar al lado.
//
//go:embed templates static
var assetsFS embed.FS

const IMAGE_NAME = "imgName"
const MAX_INPUT_LENGTH = 2048
const MAX_BODY_SIZE = 4096

// imageDir es el directorio en disco donde se escriben los QR generados.
// En produccion apunta a un directorio de estado escribible (systemd usa
// StateDirectory); en desarrollo cae en ./generated.
var imageDir = func() string {
	if d := os.Getenv("QR_IMAGE_DIR"); d != "" {
		return d
	}
	return "generated"
}()

// listenAddr permite mover el puerto sin recompilar.
var listenAddr = func() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":7003"
}()

type PageData struct {
	ImageURL string
	Error    string
}

// parseTemplates lee las plantillas del FS embebido. Antes eran tres llamadas
// identicas a ParseFiles repartidas por el archivo.
func parseTemplates() (*template.Template, error) {
	return template.New("index.html").ParseFS(assetsFS, "templates/layout.html", "templates/index.html")
}

var ImageOptions []standard.ImageOption = []standard.ImageOption{
	standard.WithBgColorRGBHex("#0F1822"),
	standard.WithFgColorRGBHex("#DDE61F"),
}

func main() {
	mux := http.NewServeMux()
	cleanupTick := time.NewTicker(5 * time.Minute)

	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		log.Fatalf("no se pudo crear el directorio de imagenes %q: %v", imageDir, err)
	}

	// /static/ sirve los assets embebidos (css, icons).
	staticSub, err := fs.Sub(assetsFS, "static")
	if err != nil {
		log.Fatalf("no se pudo abrir el FS embebido: %v", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// /img/ sirve los QR generados desde el directorio de estado.
	mux.Handle("GET /img/", http.StripPrefix("/img/", http.FileServer(http.Dir(imageDir))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		t, err := parseTemplates()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data := PageData{}
		err = t.ExecuteTemplate(w, "layout", data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("GET /{imgName}", func(w http.ResponseWriter, r *http.Request) {
		q := r.PathValue(IMAGE_NAME)

		// Validate imgName is a valid UUID to prevent path traversal.
		if _, err := uuid.Parse(q); err != nil {
			renderWithError(w, "QR code not found or has expired")
			return
		}

		filePath := filepath.Join(imageDir, q+".png")

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			renderWithError(w, "QR code not found or has expired")
			return
		}

		t, err := parseTemplates()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Ruta web, no ruta de disco: el directorio en disco es configurable
		// pero la URL publica siempre es /img/.
		data := PageData{ImageURL: "/img/" + q + ".png"}
		err = t.ExecuteTemplate(w, "layout", data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("POST /generate", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MAX_BODY_SIZE)
		text := r.FormValue("string")

		if text == "" {
			renderWithError(w, "Please enter text to generate QR code")
			return
		}

		if len(text) > MAX_INPUT_LENGTH {
			renderWithError(w, "Input text is too long")
			return
		}

		qrc, err := qrcode.New(text)
		if err != nil {
			log.Printf("QR generation error: %v", err)
			renderWithError(w, "Could not generate QR code")
			return
		}
		id := uuid.New()
		filePath := filepath.Join(imageDir, id.String()+".png")
		wr, err := standard.New(filePath, ImageOptions...)
		if err != nil {
			log.Printf("Image file creation error: %v", err)
			renderWithError(w, "Could not create image file")
			return
		}
		if err = qrc.Save(wr); err != nil {
			log.Printf("Image save error: %v", err)
			renderWithError(w, "Could not save image")
			return
		}
		redirectUrl := fmt.Sprintf("/%s", id.String())
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Server starting on %s (imagenes en %s)", listenAddr, imageDir)
		if err := http.ListenAndServe(listenAddr, mux); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go func() {
		for range cleanupTick.C {
			cleanupImages()
		}
	}()

	wg.Wait()
}

func renderWithError(w http.ResponseWriter, errorMsg string) {
	t, err := parseTemplates()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := PageData{Error: errorMsg}
	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func cleanupImages() {
	dir, err := os.ReadDir(imageDir)
	if err != nil {
		log.Printf("error reading image dir for cleanup: %v", err)
		return
	}
	var count int
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		// Only clean up .png files generated by us.
		name := entry.Name()
		if len(name) < 4 || name[len(name)-4:] != ".png" {
			continue
		}
		fileName := filepath.Join(imageDir, name)
		if err := os.Remove(fileName); err != nil {
			log.Printf("error cleaning up file: %s", fileName)
		} else {
			count++
		}
	}
	if count > 0 {
		log.Printf("Cleanup: removed %d QR image(s)", count)
	}
}

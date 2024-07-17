package routes

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type configdata struct {
	Templates *template.Template
}

func (cfg *configdata) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing index:", err)
	}
}

func newBrowserRouter() *http.ServeMux {
	pattern := filepath.Join("internal/routes/templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	config := configdata{
		Templates: templates,
	}

	browserRouter := http.NewServeMux()
	browserRouter.Handle("GET /css/styles.css", http.FileServer(http.FS(css)))
	browserRouter.HandleFunc("GET /", config.handleIndex)

	return browserRouter
}

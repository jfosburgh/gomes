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

type game struct {
	Name        string
	Description string
}

type gamedata struct {
	Games []game
}

type configdata struct {
	Templates *template.Template
	GameData  gamedata
}

func (cfg *configdata) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "index.html", &cfg.GameData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing index:", err)
	}
}

func newBrowserRouter() *http.ServeMux {

	pattern := filepath.Join("internal/routes/templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	gameData := gamedata{
		[]game{
			{Name: "tic-tac-toe", Description: "Be the first to get three in a row!"},
			{Name: "chess", Description: "A game of tactical prowess. Be the first to capture the enemy's king!"},
		},
	}

	config := configdata{
		Templates: templates,
		GameData:  gameData,
	}

	browserRouter := http.NewServeMux()
	browserRouter.Handle("/css/styles.css", http.FileServer(http.FS(css)))
	browserRouter.HandleFunc("/", config.handleIndex)

	return browserRouter
}

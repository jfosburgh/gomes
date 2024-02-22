package routes

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	// "strconv"

	"github.com/jfosburgh/gomes/pkg/game"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type gamedata struct {
	Handlers map[string]game.Handler
	Games    map[string]game.Game
}

type gameinfo struct {
	Name        string
	Description string
}

type configdata struct {
	Templates *template.Template
	GameData  gamedata
}

func (cfg *configdata) gamelist() map[string]gameinfo {
	games := map[string]gameinfo{}
	for id, g := range cfg.GameData.Handlers {
		name, description := g.Info()
		games[id] = gameinfo{
			Name:        name,
			Description: description,
		}
	}

	return games
}

func (cfg *configdata) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "index.html", cfg.gamelist())
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing index:", err)
	}
}

func (cfg *configdata) handleGetGamelist(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "comp_gamelist.html", cfg.gamelist())
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing comp_gamelist:", err)
	}
}

func (cfg *configdata) handleNewGame(w http.ResponseWriter, r *http.Request) {
	gamekey := r.PathValue("game")
	handler, exists := cfg.GameData.Handlers[gamekey]
	if !exists {
		w.WriteHeader(404)
		return
	}

	currentGame, id := handler.NewGame()
	gameOptions := handler.GameOptions()
	templateID, gameData := currentGame.TemplateData()

	cfg.GameData.Games[id] = currentGame

	template_name := fmt.Sprintf("%s.html", templateID)
	type templatedata struct {
		GameData    interface{}
		GameOptions interface{}
		Started     bool
		ID          string
	}

	templateData := templatedata{
		GameData:    gameData,
		GameOptions: gameOptions,
		Started:     false,
		ID:          id,
	}

	err := cfg.Templates.ExecuteTemplate(w, template_name, templateData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", template_name, err)
	}
}

func (cfg *configdata) handleGameStart(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("game")
	currentGame, exists := cfg.GameData.Games[gameID]
	if !exists {
		w.WriteHeader(404)
		return
	}

	currentGame.Start()
	templateID, gameData := currentGame.TemplateData()

	handler := cfg.GameData.Handlers[templateID]
	gameOptions := handler.GameOptions()
	gameOptions.SelectedMode = r.FormValue("modeselect")
	gameOptions.SelectedDifficulty = r.FormValue("difficultyselect")

	template_name := fmt.Sprintf("%s.html", templateID)
	type templatedata struct {
		GameData    interface{}
		GameOptions interface{}
		Started     bool
		ID          string
	}

	templateData := templatedata{
		GameData:    gameData,
		GameOptions: gameOptions,
		Started:     true,
		ID:          gameID,
	}

	err := cfg.Templates.ExecuteTemplate(w, template_name, templateData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", template_name, err)
	}
}

func (cfg *configdata) handleGameTurn(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("game")
	currentGame, exists := cfg.GameData.Games[gameID]
	if !exists {
		w.WriteHeader(404)
		return
	}

	params := r.URL.Query()
	move := params.Get("move")

	err := currentGame.ProcessTurn(move)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error processing turn:", err)
		return
	}

	templateID, templateData := currentGame.TemplateData()
	template_name := fmt.Sprintf("comp_%s_gameboard.html", templateID)

	err = cfg.Templates.ExecuteTemplate(w, template_name, templateData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", template_name, err)
	}
}

func (cfg *configdata) handleModeChange(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	gamekey := params.Get("game")
	handler, exists := cfg.GameData.Handlers[gamekey]
	if !exists {
		w.WriteHeader(404)
		return
	}

	// modeIndex, err := strconv.Atoi(r.FormValue("modeselect"))
	mode := r.FormValue("modeselect")
	difficulty := r.FormValue("difficultyselect")
	gameOptions := handler.GameOptions()
	// gameOptions.SelectedMode = gameOptions.Modes[modeIndex].Value
	gameOptions.SelectedMode = mode
	if difficulty != "" {
		gameOptions.SelectedDifficulty = difficulty
	}

	params = r.URL.Query()
	id := params.Get("id")
	type templatedata struct {
		GameOptions game.GameOptions
		Started     bool
		ID          string
	}
	templateData := templatedata{
		GameOptions: gameOptions,
		Started:     false,
		ID:          id,
	}

	err := cfg.Templates.ExecuteTemplate(w, "comp_mode_select.html", templateData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", "comp_mode_select.html", err)
	}
}

func newBrowserRouter() *http.ServeMux {
	pattern := filepath.Join("internal/routes/templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	gameData := gamedata{
		Handlers: map[string]game.Handler{
			"tictactoe": &tictactoe.Game{},
		},
		Games: map[string]game.Game{},
	}

	config := configdata{
		Templates: templates,
		GameData:  gameData,
	}

	browserRouter := http.NewServeMux()
	browserRouter.Handle("GET /css/styles.css", http.FileServer(http.FS(css)))
	browserRouter.HandleFunc("GET /", config.handleIndex)
	browserRouter.HandleFunc("GET /gamelist", config.handleGetGamelist)
	browserRouter.HandleFunc("GET /games/{game}", config.handleNewGame)
	browserRouter.HandleFunc("POST /games/{game}", config.handleGameTurn)
	browserRouter.HandleFunc("POST /mode", config.handleModeChange)
	browserRouter.HandleFunc("POST /games/{game}/start", config.handleGameStart)

	return browserRouter
}

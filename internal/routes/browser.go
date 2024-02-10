package routes

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type game struct {
	Name        string
	Description string
	NewGame     func() ([]string, string, string)
	ProcessTurn func([]string, string, string) ([]string, string, string, bool, []int, error)
}

type gamedata struct {
	Games map[string]game
}

type configdata struct {
	Templates *template.Template
	GameData  gamedata
}

type gamecell struct {
	Content   string
	Classes   string
	Clickable bool
}

type gamestate struct {
	State        []gamecell
	StateString  string
	StatusText   string
	ActivePlayer string
	GameOver     bool
}

func (cfg *configdata) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "index.html", &cfg.GameData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing index:", err)
	}
}

func (cfg *configdata) handleGetGamelist(w http.ResponseWriter, r *http.Request) {
	err := cfg.Templates.ExecuteTemplate(w, "comp_gamelist.html", &cfg.GameData)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing comp_gamelist:", err)
	}
}

func fillBoard(state []string, empty_state string, winning []int, gameOver bool) []gamecell {
	gamecells := []gamecell{}
	for index, content := range state {
		classes := "game-cell"

		if !(content != empty_state || gameOver) {
			classes += " enabled"
		}
		if slices.Contains(winning, index) {
			classes += " correct"
		}
		gamecells = append(gamecells, gamecell{Content: content, Classes: classes, Clickable: content == empty_state && !gameOver})
	}

	return gamecells
}

func (cfg *configdata) handleNewGame(w http.ResponseWriter, r *http.Request) {
	game_key := r.PathValue("game")
	game, exists := cfg.GameData.Games[game_key]
	if !exists {
		w.WriteHeader(404)
		return
	}

	state, status, player := game.NewGame()
	stateString := strings.Join(state, "")
	template_name := fmt.Sprintf("%s.html", game_key)

	err := cfg.Templates.ExecuteTemplate(w, template_name, gamestate{State: fillBoard(state, "_", []int{}, false), StateString: stateString, StatusText: status, ActivePlayer: player, GameOver: false})
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", template_name, err)
	}
}

func (cfg *configdata) handleGameTurn(w http.ResponseWriter, r *http.Request) {
	game_key := r.PathValue("game")
	game, exists := cfg.GameData.Games[game_key]
	if !exists {
		w.WriteHeader(404)
		return
	}

	params := r.URL.Query()
	stateString := params.Get("state")
	state := strings.Split(stateString, "")
	player := params.Get("player")
	id := params.Get("id")

	state, status, player, gameOver, winningCells, err := game.ProcessTurn(state, player, id)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error processing turn:", err)
		return
	}
	stateString = strings.Join(state, "")
	template_name := fmt.Sprintf("comp_%s_gameboard.html", game_key)

	err = cfg.Templates.ExecuteTemplate(w, template_name, gamestate{State: fillBoard(state, "_", winningCells, gameOver), StateString: stateString, StatusText: status, ActivePlayer: player, GameOver: gameOver})
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("Error parsing %s, %v\n", template_name, err)
	}
}

func newBrowserRouter() *http.ServeMux {
	pattern := filepath.Join("internal/routes/templates", "*.html")
	templates := template.Must(template.ParseGlob(pattern))

	gameData := gamedata{
		make(map[string]game),
	}
	gameData.Games["tictactoe"] = game{
		Name:        "Tic-Tac-Toe",
		Description: "Be the first to get three in a row!",
		NewGame:     tictactoe.NewGame,
		ProcessTurn: tictactoe.ProcessTurn,
	}
	gameData.Games["chess"] = game{
		Name:        "Chess",
		Description: "A game of tactical prowess. Be the first to capture the enemy's king!",
		NewGame:     func() ([]string, string, string) { return []string{}, "", "" },
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

	return browserRouter
}

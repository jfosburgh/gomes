package routes

import (
	"crypto/rand"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jfosburgh/gomes/pkg/chess"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type configdata struct {
	Components *template.Template
	Pages      map[string]*template.Template
	Games      map[string]interface{}
	GameData   map[string]*twoplayergame
}

type twoplayergame struct {
	ID     string
	Player string
	Active string

	Started bool
	Ended   bool

	State string
	Cells []cell

	Status string
}

type chessdata struct {
}

type cell struct {
	Clickable bool
	Content   string
	Classes   string
}

func (cfg *configdata) handleIndex(w http.ResponseWriter, r *http.Request) {
	err := cfg.Pages["index"].Execute(w, nil)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println("Error parsing index:", err)
	}
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)

	return fmt.Sprintf("%x", bytes)
}

var chessPieces = map[int]string{
	chess.BLACK | chess.PAWN:   "p",
	chess.BLACK | chess.KNIGHT: "n",
	chess.BLACK | chess.BISHOP: "b",
	chess.BLACK | chess.ROOK:   "r",
	chess.BLACK | chess.QUEEN:  "q",
	chess.BLACK | chess.KING:   "k",
	chess.WHITE | chess.PAWN:   "P",
	chess.WHITE | chess.KNIGHT: "N",
	chess.WHITE | chess.BISHOP: "B",
	chess.WHITE | chess.ROOK:   "R",
	chess.WHITE | chess.QUEEN:  "Q",
	chess.WHITE | chess.KING:   "K",
	chess.EMPTY:                " ",
}
var tttPieces = map[int]string{
	1:  "X",
	0:  " ",
	-1: "O",
}

func fillTTTCells(game *tictactoe.TicTacToeGame, gameState *twoplayergame) []cell {
	cells := make([]cell, 9)
	currentTurn := tttPieces[game.State.Active] == gameState.Player || gameState.Player == ""

	for i := range 9 {
		cells[i].Content = tttPieces[game.State.Board[i]]

		empty := cells[i].Content == " "
		running := gameState.Started && !gameState.Ended
		cells[i].Clickable = currentTurn && running && empty

		classes := "ttt-game-cell"
		cells[i].Classes = classes
	}

	return cells
}

func fillChessCells(game *chess.ChessGame, gameState *twoplayergame) []cell {
	cells := make([]cell, 64)

	for i := range 64 {
		cells[i].Content = chessPieces[game.EBE.Board[i]]

		clickable := gameState.Started && !gameState.Ended
		cells[i].Clickable = clickable

		classes := "chess-game-cell"
		if (i/8+i%8)%2 == 1 {
			classes += " black"
		}

		cells[i].Classes = classes
	}

	return cells
}

func (cfg *configdata) handleNewPage(w http.ResponseWriter, r *http.Request) {
	gameName := r.PathValue("game")

	game := twoplayergame{}
	game.ID = generateID()
	game.Player = ""

	game.Started = false
	game.Ended = false

	switch gameName {
	case "chess":
		chessGame := chess.NewGame()
		cfg.Games[game.ID] = chessGame

		game.Active = "white"

		game.State = chessGame.EBE.ToFEN()
		game.Cells = fillChessCells(chessGame, &game)
	case "tictactoe":
		ticTacToeGame := tictactoe.NewGame()
		cfg.Games[game.ID] = ticTacToeGame

		game.Active = "X"

		game.State = ticTacToeGame.ToGameString()
		game.Cells = fillTTTCells(ticTacToeGame, &game)
	default:
		fmt.Printf("unhandled game: %s\n", gameName)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.GameData[game.ID] = &game

	err := cfg.Pages[gameName].Execute(w, game)
	if err != nil {
		fmt.Printf("error executing template:\n%s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func newBrowserRouter() *http.ServeMux {
	pattern := filepath.Join("internal/routes/templates/components", "*.html")
	components := template.Must(template.ParseGlob(pattern))

	pattern = filepath.Join("internal/routes/templates", "*.html")
	pages := make(map[string]*template.Template)
	base := "internal/routes/templates/base.html"

	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		if match == base {
			continue
		}

		game := strings.Split(filepath.Base(match), ".")[0]
		var t *template.Template

		if game != "index" {
			t, _ = template.ParseFiles(base, match, fmt.Sprintf("internal/routes/templates/components/%s_gameboard.html", game))
			fmt.Printf("created game template for %s\n", game)
		} else {
			t, _ = template.ParseFiles(base, match)
		}
		pages[game] = t
	}

	config := configdata{
		Components: components,
		Pages:      pages,
		Games:      make(map[string]interface{}),
		GameData:   make(map[string]*twoplayergame),
	}

	browserRouter := http.NewServeMux()
	browserRouter.Handle("GET /css/styles.css", http.FileServer(http.FS(css)))
	browserRouter.HandleFunc("GET /", config.handleIndex)
	browserRouter.HandleFunc("GET /games/{game}", config.handleNewPage)

	return browserRouter
}

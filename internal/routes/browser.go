package routes

import (
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
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
		if cells[i].Clickable {
			classes += " enabled"
		}
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

func (cfg *configdata) respondWithComponent(w http.ResponseWriter, t string, data twoplayergame) {
	fmt.Printf("%+v\n", data)
	err := cfg.Components.ExecuteTemplate(w, t, data)
	if err != nil {
		fmt.Printf("error executing template:\n%s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cfg *configdata) getGameFromRequest(r *http.Request) (interface{}, *twoplayergame, error) {
	gameID := r.PathValue("id")

	data, ok := cfg.GameData[gameID]
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("no game data for %s\n", gameID))
	}
	data.Started = true

	game, ok := cfg.Games[gameID]
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("no game for %s\n", gameID))
	}

	return game, data, nil
}

func (cfg *configdata) handleStartGame(w http.ResponseWriter, r *http.Request) {
	gameInterface, data, err := cfg.getGameFromRequest(r)
	if err != nil {
		fmt.Printf("error retrieving game from id:\n%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	mode := r.FormValue("gamemode")

	var compName string
	switch gameInterface.(type) {
	case *tictactoe.TicTacToeGame:
		game := gameInterface.(*tictactoe.TicTacToeGame)
		if mode == "pvb" {
			data.Player = r.FormValue("playerID")
			game.SearchDepth, _ = strconv.Atoi("depth")
		}
		data.Cells = fillTTTCells(game, data)
		data.Status = "X makes the first move!"
		compName = "tictactoe_gameboard.html"
	case *chess.ChessGame:
		game := gameInterface.(*chess.ChessGame)
		data.Cells = fillChessCells(game, data)
		data.Status = "White makes the first move!"
		compName = "chess_gameboard.html"
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
}

func (cfg *configdata) handleMove(w http.ResponseWriter, r *http.Request) {
	gameInterface, data, err := cfg.getGameFromRequest(r)
	if err != nil {
		fmt.Printf("error retrieving game from id:\n%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	queries := r.URL.Query()
	moveStr := queries.Get("move")
	if moveStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	move, err := strconv.Atoi(moveStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var compName string
	switch gameInterface.(type) {
	case *tictactoe.TicTacToeGame:
		game := gameInterface.(*tictactoe.TicTacToeGame)
		game.MakeMove(move)

		var winner int
		data.Ended, winner = game.GameOver()

		data.Active = tttPieces[game.State.Active]
		data.Cells = fillTTTCells(game, data)

		if data.Ended {
			if winner == 0 {
				data.Status = "It's a tie!"
			} else {
				data.Status = fmt.Sprintf("%s Wins!", tttPieces[winner])
			}

			delete(cfg.GameData, data.ID)
			delete(cfg.Games, data.ID)
		} else {
			data.Status = fmt.Sprintf("%s's Turn!", data.Active)
		}
		compName = "tictactoe_gameboard.html"
	case *chess.ChessGame:
		game := gameInterface.(*chess.ChessGame)
		data.Cells = fillChessCells(game, data)
		data.Status = "White makes the first move!"
		compName = "chess_gameboard.html"
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
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
	browserRouter.HandleFunc("POST /games/{id}/start", config.handleStartGame)
	browserRouter.HandleFunc("POST /games/{id}", config.handleMove)
	browserRouter.HandleFunc("POST /games/{id}/bot", config.handleBotTurn)

	return browserRouter
}

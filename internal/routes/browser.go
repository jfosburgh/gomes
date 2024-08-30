package routes

import (
	"crypto/rand"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jfosburgh/gomes/pkg/chess"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

var (
	//go:embed css/styles.css
	css embed.FS
)

type configdata struct {
	Components map[string]*template.Template
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
	chess.BLACK | chess.PAWN:   "♟",
	chess.BLACK | chess.KNIGHT: "♞",
	chess.BLACK | chess.BISHOP: "♝",
	chess.BLACK | chess.ROOK:   "♜",
	chess.BLACK | chess.QUEEN:  "♛",
	chess.BLACK | chess.KING:   "♚",
	chess.WHITE | chess.PAWN:   "♙",
	chess.WHITE | chess.KNIGHT: "♘",
	chess.WHITE | chess.BISHOP: "♗",
	chess.WHITE | chess.ROOK:   "♖",
	chess.WHITE | chess.QUEEN:  "♕",
	chess.WHITE | chess.KING:   "♔",
	chess.EMPTY:                " ",
}

var chessPlayers = map[string]int{
	"White": chess.WHITE,
	"Black": chess.BLACK,
}

var chessNames = map[int]string{
	0: "White",
	1: "Black",
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

func fillChessCells(game *chess.ChessGame, gameState *twoplayergame, selected int, promoting bool) []cell {
	cells := make([]cell, 64)
	gameActive := gameState.Started && !gameState.Ended
	playerTurn := gameState.Player == "" || chessPlayers[gameState.Active] == game.EBE.Active<<3

	validTargets := []int{}
	if selected != -1 {
		validTargets = game.GetMoveTargets(selected)
		fmt.Printf("valid moves for %d: %+v\n", selected, validTargets)
	}

	side := game.EBE.Active << 3

	cellCount := 0
	for rank := 7; rank >= 0; rank-- {
		for file := range 8 {
			i := 8*rank + file
			cells[cellCount].Content = chessPieces[game.EBE.Board[i]]

			classes := "chess-game-cell"
			if (i/8+i%8)%2 == 1 {
				classes += " black"
			}

			if i == selected {
				classes += " selected"
			}

			validTarget := slices.Contains(validTargets, i)
			if validTarget {
				classes += " target"

				fmt.Printf("checking for promotion: %04b, %04b, %d\n", game.EBE.Board[selected]&0b0111, chess.PAWN, rank)
				if game.EBE.Board[selected]&0b0111 == chess.PAWN && (rank == 7 || rank == 0) {
					fmt.Printf("this was a valid promotion")
					classes += " promote"
				}
			}

			activeSide := (game.EBE.Board[i]&0b1000 == side) && game.EBE.Board[i] != chess.EMPTY

			moveable := playerTurn && activeSide
			cells[cellCount].Clickable = gameActive && (moveable || validTarget) && !promoting
			if cells[cellCount].Clickable {
				classes += " enabled"
			}

			cells[cellCount].Classes = classes
			cellCount += 1
		}
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

		game.Active = "White"

		game.State = chessGame.EBE.ToFEN()
		game.Cells = fillChessCells(chessGame, &game, -1, false)
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

	err := cfg.Pages[gameName].ExecuteTemplate(w, "base.html", game)
	if err != nil {
		fmt.Printf("error executing template:\n%s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cfg *configdata) respondWithComponent(w http.ResponseWriter, t string, data any) {
	// fmt.Printf("generating template %s with \n%+v\n", t, data)
	err := cfg.Components[t].Execute(w, data)
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
			game.SearchDepth, _ = strconv.Atoi(r.FormValue("depth"))
			// fmt.Printf("setting game parameters:\nPlayer: %s\nSearch Depth: %d\n", data.Player, game.SearchDepth)
		}
		data.Cells = fillTTTCells(game, data)
		data.Status = "X makes the first move!"
		compName = "tictactoe_gameboard.html"
	case *chess.ChessGame:
		game := gameInterface.(*chess.ChessGame)
		if mode == "pvb" {
			data.Player = r.FormValue("playerID")
			depth, _ := strconv.Atoi(r.FormValue("depth"))
			searchTime, _ := strconv.Atoi(r.FormValue("time"))
			game.MaxSearchDepth = depth * 2
			game.SearchTime = time.Duration(searchTime) * time.Second
		}
		data.Cells = fillChessCells(game, data, -1, false)
		data.Status = "White makes the first move!"
		compName = "chess_gameboard.html"
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
}

func flipRank(input int) int {
	return 8*(7-input/8) + input%8
}

func (cfg *configdata) handleSelect(w http.ResponseWriter, r *http.Request) {
	gameInterface, data, err := cfg.getGameFromRequest(r)
	if err != nil {
		fmt.Printf("error retrieving game from id:\n%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	queries := r.URL.Query()
	locationStr := queries.Get("location")
	if locationStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	location, err := strconv.Atoi(locationStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	location = flipRank(location)

	var compName string
	switch gameInterface.(type) {
	case *tictactoe.TicTacToeGame:
		fmt.Println("select not implemented for tictactoe")
		w.WriteHeader(http.StatusBadRequest)
		return
	case *chess.ChessGame:
		game := gameInterface.(*chess.ChessGame)
		fmt.Printf("returning chess board with piece %d selected\n", location)
		data.Cells = fillChessCells(game, data, location, false)
		compName = "chess_gameboard.html"
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
}

func (cfg *configdata) handlePromotion(w http.ResponseWriter, r *http.Request) {
	gameInterface, data, err := cfg.getGameFromRequest(r)
	if err != nil {
		fmt.Printf("error retrieving game from id:\n%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	queries := r.URL.Query()
	start, err := strconv.Atoi(queries.Get("start"))
	if err != nil {
		fmt.Println("couldn't parse start")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	end, err := strconv.Atoi(queries.Get("end"))
	if err != nil {
		fmt.Println("couldn't parse end")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	promote, err := strconv.Atoi(queries.Get("promote"))
	if err != nil {
		fmt.Println("couldn't parse promote")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game := gameInterface.(*chess.ChessGame)

	gameMove, valid := game.MoveFromLocations(start, end)
	if !valid {
		fmt.Printf("requested move is invalid: %d->%d\n", start, end)
		w.WriteHeader(http.StatusBadRequest)
	}
	gameMove.Promotion = promote
	game.MakeMove(gameMove)
	data.Active = chessNames[game.EBE.Active]

	data.Cells = fillChessCells(game, data, -1, false)
	data.Ended = len(game.GetLegalMoves()) == 0
	if data.Ended {
		data.Status = fmt.Sprintf("%s Wins!", chessNames[^(game.EBE.Active)&0b1])

		delete(cfg.GameData, data.ID)
		delete(cfg.Games, data.ID)
	} else {
		data.Status = fmt.Sprintf("%s played %s, %s's Turn!", chessNames[^game.EBE.Active&0b1], gameMove, data.Active)
	}

	cfg.respondWithComponent(w, "chess_gameboard.html", *data)
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

		srcStr := queries.Get("piece")
		if srcStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		src, err := strconv.Atoi(srcStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		promote := queries.Get("promote") == "true"

		move = flipRank(move)
		src = flipRank(src)

		gameMove, valid := game.MoveFromLocations(src, move)
		if !valid {
			fmt.Printf("requested move is invalid: %d->%d\n", src, move)
			w.WriteHeader(http.StatusBadRequest)
		}
		if promote {
			gameMove.Promotion = gameMove.Piece
		}
		game.MakeMove(gameMove)
		if !promote {
			data.Active = chessNames[game.EBE.Active]
		}

		data.Cells = fillChessCells(game, data, -1, promote)
		checkmate := len(game.GetLegalMoves()) == 0
		draw := !checkmate && game.EBE.Halfmoves >= 100
		data.Ended = checkmate || draw
		if data.Ended {
			data.Status = fmt.Sprintf("%s Wins!", chessNames[^(game.EBE.Active)&0b1])
			if draw {
				data.Status = "It's a tie!"
			}

			delete(cfg.GameData, data.ID)
			delete(cfg.Games, data.ID)
		} else {
			data.Status = fmt.Sprintf("%s played %s, %s's Turn!", chessNames[^game.EBE.Active&0b1], gameMove, data.Active)
		}
		if promote {
			compName = "promotion.html"
			data.Status = fmt.Sprintf("%s, choose your promotion!", data.Active)
			game.UnmakeMove(gameMove)
			type promoteoption struct {
				Piece   int
				Picture string
			}
			type promotedata struct {
				GameData twoplayergame
				Start    int
				End      int
				Options  [4]promoteoption
			}
			side := game.EBE.Active << 3
			promoteData := promotedata{
				*data,
				src,
				move,
				[4]promoteoption{
					{side | chess.KNIGHT, chessPieces[side|chess.KNIGHT]},
					{side | chess.ROOK, chessPieces[side|chess.ROOK]},
					{side | chess.BISHOP, chessPieces[side|chess.BISHOP]},
					{side | chess.QUEEN, chessPieces[side|chess.QUEEN]},
				},
			}

			cfg.respondWithComponent(w, compName, promoteData)
			return
		} else {
			compName = "chess_gameboard.html"
		}
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
}

func (cfg *configdata) handleBotTurn(w http.ResponseWriter, r *http.Request) {
	gameInterface, data, err := cfg.getGameFromRequest(r)
	if err != nil {
		fmt.Printf("error retrieving game from id:\n%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}

	var compName string
	switch gameInterface.(type) {
	case *tictactoe.TicTacToeGame:
		game := gameInterface.(*tictactoe.TicTacToeGame)
		move := game.BestMove()
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

		move := game.BestMove()
		game.MakeMove(move)
		data.Active = chessNames[game.EBE.Active]

		data.Cells = fillChessCells(game, data, -1, false)
		data.Ended = len(game.GetLegalMoves()) == 0
		if data.Ended {
			data.Status = fmt.Sprintf("%s Wins!", chessNames[^(game.EBE.Active)&0b1])

			delete(cfg.GameData, data.ID)
			delete(cfg.Games, data.ID)
		} else {
			data.Status = fmt.Sprintf("%s played %s, %s's Turn!", chessNames[^game.EBE.Active&0b1], move, data.Active)
		}
		compName = "chess_gameboard.html"
	default:
		fmt.Printf("unhandled game type: %t\n", gameInterface)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg.respondWithComponent(w, compName, *data)
}

func join(sep string, s ...string) string {
	return strings.Join(s, sep)
}

var tempFuncs = map[string]any{
	"contains": strings.Contains,
	"join":     join,
	"toString": fmt.Sprint,
}

func newBrowserRouter() *http.ServeMux {
	pattern := filepath.Join("internal/routes/templates/components", "*.html")
	components := make(map[string]*template.Template)

	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {

		name := filepath.Base(match)
		var t *template.Template
		if name == "promotion.html" {
			t = template.Must(template.New(name).Funcs(tempFuncs).ParseFiles(match, "internal/routes/templates/components/chess_gameboard.html"))
		} else {
			t = template.Must(template.New(name).Funcs(tempFuncs).ParseFiles(match))
		}

		components[name] = t
		fmt.Printf("added %s to components dict\n", name)
	}

	pattern = filepath.Join("internal/routes/templates", "*.html")
	pages := make(map[string]*template.Template)
	base := "internal/routes/templates/base.html"

	matches, _ = filepath.Glob(pattern)
	for _, match := range matches {
		if match == base {
			continue
		}

		game := strings.Split(filepath.Base(match), ".")[0]

		if game != "index" {
			t := template.Must(template.New(base).Funcs(tempFuncs).ParseFiles(base, match, fmt.Sprintf("internal/routes/templates/components/%s_gameboard.html", game)))
			fmt.Printf("created game template for %s\n", game)
			pages[game] = t
		} else {
			t, _ := template.ParseFiles(base, match)
			pages[game] = t
		}
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
	browserRouter.HandleFunc("POST /games/{id}/select", config.handleSelect)
	browserRouter.HandleFunc("POST /games/{id}/promote", config.handlePromotion)

	return browserRouter
}

package utils

import (
	"fmt"
	"slices"

	"github.com/jfosburgh/gomes/pkg/chess"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

type TwoPlayerGame struct {
	ID     string
	Player string
	Active string

	Started bool
	Ended   bool

	State string
	Cells []Cell

	Status string
}

type Cell struct {
	Clickable bool
	Content   string
	Classes   string
}

var ChessPieces = map[int]string{
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

var ChessPlayers = map[string]int{
	"White": chess.WHITE,
	"Black": chess.BLACK,
}

var ChessNames = map[int]string{
	0: "White",
	1: "Black",
}

var TTTPieces = map[int]string{
	1:  "X",
	0:  " ",
	-1: "O",
}

func FlipRank(input int) int {
	return 8*(7-input/8) + input%8
}

func FillTTTCells(game *tictactoe.TicTacToeGame, gameState *TwoPlayerGame) []Cell {
	cells := make([]Cell, 9)
	currentTurn := TTTPieces[game.State.Active] == gameState.Player || gameState.Player == ""

	for i := range 9 {
		cells[i].Content = TTTPieces[game.State.Board[i]]

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

func FillChessCells(game *chess.ChessGame, gameState *TwoPlayerGame, selected int, promoting bool) []Cell {
	cells := make([]Cell, 64)
	gameActive := gameState.Started && !gameState.Ended
	playerTurn := gameState.Player == "" || ChessPlayers[gameState.Active] == game.EBE.Active<<3

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
			cells[cellCount].Content = ChessPieces[game.EBE.Board[i]]

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

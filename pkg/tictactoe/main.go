package tictactoe

import (
	"crypto/rand"
	"errors"
	"fmt"
	mrand "math/rand"
	"slices"
	"strconv"
	"strings"

	"github.com/jfosburgh/gomes/pkg/game"
)

var (
	Modes = []game.SelectOption{
		{
			Value:   "0",
			Content: "Player vs. Player",
		},
		{
			Value:   "1",
			Content: "Player vs. Bot",
		},
	}

	Difficulties = []game.SelectOption{
		{Value: "0", Content: "Easy"},
		{Value: "1", Content: "Medium"},
		{Value: "2", Content: "Hard"},
		{Value: "3", Content: "Unbeatable"},
	}

	Depths = map[string]int{
		"0": 1,
		"1": 2,
		"2": 4,
		"3": 9,
	}
)

type Game struct {
	State         []string
	Text          string
	CurrentPlayer string
	PlayerID      string
	Active        bool
	Correct       []int
	ID            string
	Name          string
	Depth         int
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
	PlayerID     string
	Started      bool
	GameOver     bool
	ID           string
}

func (g *Game) Info() (string, string) {
	return "Tic-Tac-Toe", "Be the first to get three in a row!"
}

func (g *Game) GameOptions() game.GameOptions {
	return game.GameOptions{
		Difficulties:       Difficulties,
		Modes:              Modes,
		SelectedMode:       Modes[0].Value,
		SelectedDifficulty: Difficulties[0].Value,
		Name:               "tictactoe",
		Bot:                false,
		PlayerID:           "",
	}
}

func (g *Game) fillBoard() []gamecell {
	gamecells := []gamecell{}
	botTurn := g.PlayerID != "" && g.PlayerID != g.CurrentPlayer
	for index, content := range g.State {
		classes := "game-cell"
		clickable := false
		if content == "_" && g.Active && !botTurn {
			classes += " enabled"
			clickable = true
		}
		if slices.Contains(g.Correct, index) {
			classes += " correct"
		}
		gamecells = append(gamecells, gamecell{Content: content, Classes: classes, Clickable: clickable})
	}

	return gamecells
}

func (g *Game) TemplateData() (string, interface{}) {
	gameState := gamestate{
		State:        g.fillBoard(),
		StateString:  strings.Join(g.State, ","),
		StatusText:   g.Text,
		ActivePlayer: g.CurrentPlayer,
		PlayerID:     g.PlayerID,
		Started:      g.Active && len(g.Correct) == 0,
		GameOver:     !g.Active && len(g.Correct) != 0,
		ID:           g.ID,
	}
	return "tictactoe", gameState
}

func (g *Game) NewGame() (game.Game, string) {
	id := make([]byte, 16)
	rand.Read(id)
	stringID := fmt.Sprintf("%x", id)
	g = &Game{
		State:         []string{"_", "_", "_", "_", "_", "_", "_", "_", "_"},
		Text:          "X's Turn!",
		CurrentPlayer: "X",
		PlayerID:      "",
		Active:        false,
		Correct:       []int{},
		ID:            stringID,
		Name:          "tictactoe",
		Depth:         0,
	}
	return g, stringID
}

func (g *Game) Start(opts game.GameOptions) {
	g.Active = true
	g.PlayerID = opts.PlayerID
	g.Depth = Depths[opts.SelectedDifficulty]
}

func status(state []string) (bool, []int) {
	for i := 0; i < 3; i++ {
		if state[i] != "_" && (state[i] == state[i+3] && state[i] == state[i+6]) {
			return true, []int{i, i + 3, i + 6}
		}
		if state[3*i] != "_" && (state[3*i] == state[3*i+1] && state[3*i] == state[3*i+2]) {
			return true, []int{3 * i, 3*i + 1, 3*i + 2}
		}
	}
	if state[0] != "_" && (state[0] == state[4] && state[0] == state[8]) {
		return true, []int{0, 4, 8}
	}
	if state[2] != "_" && (state[2] == state[4] && state[2] == state[6]) {
		return true, []int{2, 4, 6}
	}
	if !slices.Contains(state, "_") {
		return true, []int{-1}
	}
	return false, []int{}
}

func moves(state []string) []int {
	empty := []int{}
	for i, val := range state {
		if val == "_" {
			empty = append(empty, i)
		}
	}

	return empty
}

func (g *Game) BotTurn() int {
	empty := moves(g.State)
	bestVal := -100
	vals := []int{}
	for _, move := range empty {
		newState := make([]string, len(g.State))
		copy(newState, g.State)
		newState[move] = g.CurrentPlayer
		nextPlayer := "X"
		if g.CurrentPlayer == "X" {
			nextPlayer = "O"
		}
		value := minimax(newState, nextPlayer, g.Depth-1)
		if g.CurrentPlayer == "O" {
			value *= -1
		}

		vals = append(vals, value)

		if value > bestVal {
			bestVal = value
		}
	}

	moveOptions := []int{}
	for i, val := range vals {
		if val == bestVal {
			moveOptions = append(moveOptions, empty[i])
		}
	}

	bestMove := moveOptions[mrand.Intn(len(moveOptions))]
	return bestMove
}

func boardString(state []string) string {
	return fmt.Sprintf("%s|%s|%s\n-----\n%s|%s|%s\n-----\n%s|%s|%s",
		state[0],
		state[1],
		state[2],
		state[3],
		state[4],
		state[5],
		state[6],
		state[7],
		state[8],
	)
}

func boardValue(state []string) int {
	gameOver, winningCells := status(state)
	if gameOver && len(winningCells) == 1 {
		return 0
	} else if gameOver && state[winningCells[0]] == "X" {
		return 10
	} else if gameOver {
		return -10
	}

	tris := [][]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 5},
	}

	value := 0

	for _, tri := range tris {
		states := []string{
			state[tri[0]],
			state[tri[1]],
			state[tri[2]],
		}

		xCount := 0
		oCount := 0
		for _, p := range states {
			if p == "X" {
				xCount++
			}
			if p == "O" {
				oCount++
			}
		}
		if xCount > 1 && oCount == 0 {
			rowVal := 1 + 2*(xCount-1)
			value += rowVal
		}
		if oCount > 1 && xCount == 0 {
			rowVal := 1 + 2*(oCount-1)
			value -= rowVal
		}
	}

	return value
}

func minimax(state []string, player string, depth int) int {
	if depth == 0 {
		return boardValue(state)
	}
	empty := moves(state)
	if len(empty) == 0 {
		return boardValue(state)
	}
	value := []int{}
	for _, move := range empty {
		newState := make([]string, len(state))
		copy(newState, state)
		newState[move] = player
		gameOver, winningCells := status(newState)
		if gameOver && len(winningCells) == 1 {
			value = append(value, 0)
		} else if gameOver && newState[winningCells[0]] == "X" {
			value = append(value, 10)
		} else if gameOver {
			value = append(value, -10)
		} else {
			if player == "X" {
				value = append(value, minimax(newState, "O", depth-1))
			} else {
				value = append(value, minimax(newState, "X", depth-1))
			}
		}
	}

	if player == "O" {
		return slices.Min(value)
	}
	return slices.Max(value)
}

func (g *Game) ProcessTurn(id string) error {
	index := -1
	if id == "" {
		index = g.BotTurn()
	} else {
		index, _ = strconv.Atoi(id)
	}

	if g.State[index] != "_" {
		return errors.New(fmt.Sprintf("Index %d already filled", index))
	}

	g.State[index] = g.CurrentPlayer
	if g.CurrentPlayer == "X" {
		g.CurrentPlayer = "O"
	} else {
		g.CurrentPlayer = "X"
	}

	g.Text = fmt.Sprintf("%s's Turn!", g.CurrentPlayer)
	gameOver, winningCells := status(g.State)
	g.Active = !gameOver
	g.Correct = winningCells
	if !g.Active && len(g.Correct) == 3 {
		g.Text = fmt.Sprintf("Game Over, %s won!", g.State[g.Correct[0]])
	}
	if !g.Active && len(g.Correct) == 1 {
		g.Text = "It's a tie!"
	}

	return nil
}

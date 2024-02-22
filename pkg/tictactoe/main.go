package tictactoe

import (
	"crypto/rand"
	"errors"
	"fmt"
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
)

type Game struct {
	State         []string
	Text          string
	CurrentPlayer string
	Active        bool
	Correct       []int
	ID            string
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
	}
}

func (g *Game) fillBoard() []gamecell {
	gamecells := []gamecell{}
	for index, content := range g.State {
		classes := "game-cell"
		clickable := false
		if content == "_" && g.Active {
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
		Started:      g.Active && len(g.Correct) == 0,
		GameOver:     !g.Active && len(g.Correct) > 0,
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
		Active:        false,
		Correct:       []int{},
		ID:            stringID,
	}
	return g, stringID
}

func (g *Game) Start() {
	g.Active = true
}

func (g *Game) status() {
	state := g.State
	for i := 0; i < 3; i++ {
		if state[i] != "_" && (state[i] == state[i+3] && state[i] == state[i+6]) {
			g.Active = false
			g.Correct = []int{i, i + 3, i + 6}
			return
		}
		if state[3*i] != "_" && (state[3*i] == state[3*i+1] && state[3*i] == state[3*i+2]) {
			g.Active = false
			g.Correct = []int{3 * i, 3*i + 1, 3*i + 2}
			return
		}
	}
	if state[0] != "_" && (state[0] == state[4] && state[0] == state[8]) {
		g.Active = false
		g.Correct = []int{0, 4, 8}
		return
	}
	if state[2] != "_" && (state[2] == state[4] && state[2] == state[6]) {
		g.Active = false
		g.Correct = []int{2, 4, 6}
		return
	}
	if !slices.Contains(state, "_") {
		g.Active = false
		return
	}
	return
}

func (g *Game) ProcessTurn(id string) error {
	index, err := strconv.Atoi(id)
	if err != nil {
		return err
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
	g.status()
	if !g.Active && len(g.Correct) == 3 {
		g.Text = fmt.Sprintf("Game Over, %s won!", g.State[g.Correct[0]])
	}
	if !g.Active && len(g.Correct) == 0 {
		g.Text = "It's a tie!"
	}

	return nil
}

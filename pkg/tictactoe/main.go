package tictactoe

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	playerToInt = map[string]int{
		"X": 1,
		"O": -1,
		" ": 0,
	}
	intToPlayer = map[int]string{
		1:  "X",
		0:  " ",
		-1: "O",
	}

	tris = [][]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 5},
	}
)

type TicTacToeGame struct {
	State       TBT
	SearchDepth int
	TopK        int
}

type TBT struct {
	Board  TBTBoard
	Active int
}

type TBTBoard [9]int

func (t TBTBoard) String() string {
	s := fmt.Sprintf("\n %s | %s | %s \n", intToPlayer[t[0]], intToPlayer[t[1]], intToPlayer[t[2]])
	s += "-----------\n"
	s += fmt.Sprintf(" %s | %s | %s \n", intToPlayer[t[3]], intToPlayer[t[4]], intToPlayer[t[5]])
	s += "-----------\n"
	s += fmt.Sprintf(" %s | %s | %s \n", intToPlayer[t[6]], intToPlayer[t[7]], intToPlayer[t[8]])

	return s
}

func NewGame() *TicTacToeGame {
	return &TicTacToeGame{
		State: TBT{
			Active: 1,
		},
		SearchDepth: 9,
	}
}

func (t *TicTacToeGame) String() string {
	return fmt.Sprintf("Board:\n%s\nNext Player: %s", t.State.Board, intToPlayer[t.State.Active])
}

func (t *TicTacToeGame) FromString(state string) error {
	parts := strings.Split(state, ",")
	board := parts[0]
	if len(board) != 9 {
		return errors.New(fmt.Sprintf("TTT FromString: could not parse '%s' as board state", board))
	}

	xCount := 0
	oCount := 0
	for i, char := range strings.Split(board, "") {
		player, ok := playerToInt[char]
		if !ok {
			return errors.New(fmt.Sprintf("TTT FromString: '%s' is not a valid active player key", char))
		}
		t.State.Board[i] = player
		if player == 1 {
			xCount += 1
		} else if player == -1 {
			oCount += 1
		}
	}

	if oCount > xCount {
		return errors.New(fmt.Sprintf("TTT FromString: invalid board '%s', can't have more O's than X's", state))
	}

	if xCount-oCount > 1 {
		return errors.New(fmt.Sprintf("TTT FromString: invalid board '%s', unbalanced", state))
	}

	nextPlayer := parts[1]
	player, ok := playerToInt[nextPlayer]
	if !ok {
		return errors.New(fmt.Sprintf("TTT FromString: '%s' is not a valid active player key", nextPlayer))
	}

	if xCount > oCount && player != -1 {
		return errors.New(fmt.Sprintf("TTT FromString: invalid board '%s', O should be next", state))
	}

	t.State.Active = player

	return nil
}

func (t *TicTacToeGame) ToGameString() string {
	s := ""
	for _, square := range t.State.Board {
		s += intToPlayer[square]
	}

	s += ","
	return s + intToPlayer[t.State.Active]
}

func (t *TicTacToeGame) GenerateMoves() []int {
	moves := []int{}

	for i := range 9 {
		if t.State.Board[i] == 0 {
			moves = append(moves, i)
		}
	}

	return moves
}

func (t *TicTacToeGame) MakeMove(index int) {
	t.State.Board[index] = t.State.Active
	t.State.Active *= -1
}

func (t *TicTacToeGame) UnmakeMove(index int) {
	t.State.Board[index] = 0
	t.State.Active *= -1
}

func (t *TicTacToeGame) Search() ([]int, []int) {
	options := t.GenerateMoves()
	if len(options) == 0 {
		return []int{}, []int{}
	}

	vals := []int{}
	for _, option := range options {
		t.MakeMove(option)
		vals = append(vals, t.Minimax(t.SearchDepth-1))
		t.UnmakeMove(option)
	}

	for i := range len(options) - 1 {
		for j := 0; j < len(options)-i-1; j++ {
			if vals[j] > vals[j+1] {
				temp := vals[j]
				vals[j] = vals[j+1]
				vals[j+1] = temp

				temp = options[j]
				options[j] = options[j+1]
				options[j+1] = temp
			}
		}
	}

	return options, vals
}

func (t *TicTacToeGame) GameOver() (bool, int) {
	for _, tri := range tris {
		if t.State.Board[tri[0]] != 0 && (t.State.Board[tri[0]] == t.State.Board[tri[1]] && t.State.Board[tri[0]] == t.State.Board[tri[2]]) {
			return true, t.State.Board[tri[0]]
		}
	}

	if len(t.GenerateMoves()) == 0 {
		return true, 0
	}

	return false, 0
}

func (t *TicTacToeGame) Evaluate() int {
	score := 0
	depth := 0
	for i := range 9 {
		if t.State.Board[i] != 0 {
			depth += 1
		}
	}
	if t.State.Active == 1 {
		depth *= -1
	}

	for _, tri := range tris {
		triVal := t.State.Board[tri[0]] + t.State.Board[tri[1]] + t.State.Board[tri[2]]
		triValSquared := t.State.Board[tri[0]]*t.State.Board[tri[0]] + t.State.Board[tri[1]]*t.State.Board[tri[1]] + t.State.Board[tri[2]]*t.State.Board[tri[2]]
		if triVal*triVal == 9 {
			// three in a row
			score += 10*triVal - depth
		} else if triValSquared == 3 {
			// row filled
			continue
		} else if triVal*triVal == 4 {
			// two in a row w/ third empty
			score += 5 * triVal
		} else {
			// opposing two in row or one in row
			score += triVal
		}
	}

	return score
}

func (t *TicTacToeGame) Minimax(depth int) int {
	if depth <= 0 {
		return t.Evaluate()
	}

	moves := t.GenerateMoves()
	if len(moves) == 0 {
		return t.Evaluate()
	}

	vals := []int{}
	for _, move := range moves {
		t.MakeMove(move)
		vals = append(vals, t.Minimax(depth-1))
		t.UnmakeMove(move)
	}

	if t.State.Active == 1 {
		return slices.Max(vals)
	}
	return slices.Min(vals)
}

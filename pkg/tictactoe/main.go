package tictactoe

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

func NewGame() ([]string, string, string) {
	return []string{"_", "_", "_", "_", "_", "_", "_", "_", "_"}, "X's Turn!", "X"
}

func isOver(state []string) (bool, []int) {
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
		return true, []int{}
	}
	return false, []int{}
}

func ProcessTurn(state []string, player, id string) ([]string, string, string, bool, []int, error) {
	index, err := strconv.Atoi(id)
	if err != nil {
		return []string{}, "", "", false, []int{}, err
	}

	if state[index] != "_" {
		return []string{}, "", "", false, []int{}, errors.New(fmt.Sprintf("Index %d already filled", index))
	}

	state[index] = player
	if player == "X" {
		player = "O"
	} else {
		player = "X"
	}

	gameText := fmt.Sprintf("%s's Turn!", player)
	gameOver, winningCells := isOver(state)
	if gameOver && len(winningCells) == 3 {
		gameText = fmt.Sprintf("Game Over, %s won!", state[winningCells[0]])
	}
	if gameOver && len(winningCells) == 0 {
		gameText = "It's a tie!"
	}

	return state, gameText, player, gameOver, winningCells, nil
}

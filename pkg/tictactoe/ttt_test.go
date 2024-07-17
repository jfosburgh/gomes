package tictactoe

import "testing"

func GameEquals(t *testing.T, expected, actual *TicTacToeGame) {
	equal := true
	for i := range 9 {
		if expected.State.Board[i] != actual.State.Board[i] {
			equal = false
		}
	}

	if expected.State.Active != actual.State.Active {
		equal = false
	}

	if !equal {
		t.Errorf("Actual board does not match expected board:\nExpected\n%s\nActual:\n%s\n", expected, actual)
	}
}

func GameValueEquals(t *testing.T, g *TicTacToeGame, expected int) {
	actual := g.Evaluate()
	if actual != expected {
		t.Errorf("Expected value %d but got value %d for board\n%s", expected, actual, g)
	}
}

func MoveSearchEquals(t *testing.T, g *TicTacToeGame, expectedMoves, expectedVals []int) {
	actualMoves, actualVals := g.Search()

	equal := true
	if len(actualMoves) != len(expectedMoves) {
		equal = false
	}

	for i := range len(expectedMoves) {
		if expectedMoves[i] != actualMoves[i] || expectedVals[i] != actualVals[i] {
			equal = false
			break
		}
	}

	if !equal {
		t.Errorf("Move search for following board state was not as expected\n%s\nExpected Moves: %+v\nActual Moves:   %+v\n\nExpected Vals: %+v\nActual Vals:   %+v", g, expectedMoves, actualMoves, expectedVals, actualVals)
	}
}

func TestFromString(t *testing.T) {
	gameString := "         ,X"
	expected := NewGame()

	actual := NewGame()
	actual.FromString(gameString)
	GameEquals(t, expected, actual)

	expected.State.Board[0] = 1
	expected.State.Active = -1
	actual.FromString("X        ,O")
	GameEquals(t, expected, actual)

	expected.State.Board[8] = -1
	expected.State.Active = 1
	actual.FromString("X       O,X")
	GameEquals(t, expected, actual)
}

func TestMakeMove(t *testing.T) {
	gameString := "X        ,O"
	expected := NewGame()
	expected.FromString(gameString)

	actual := NewGame()
	actual.MakeMove(0)
	GameEquals(t, expected, actual)

	expected = NewGame()
	actual.UnmakeMove(0)
	GameEquals(t, expected, actual)
}

//
// func TestEvaluate(t *testing.T) {
// 	game := NewGame()
// 	GameValueEquals(t, game, 0)
//
// 	game.FromString("X        ,O")
// 	GameValueEquals(t, game, 3)
//
// 	game.FromString("XO       ,X")
// 	GameValueEquals(t, game, 1)
//
// 	game.FromString("XXX      ,O")
// 	GameValueEquals(t, game, 302)
//
// 	game.FromString("OOO      ,X")
// 	GameValueEquals(t, game, 302)
//
// 	game.FromString("XX O     ,O")
// 	GameValueEquals(t, game, 11)
//
// 	game.FromString("XX OO    ,X")
// 	GameValueEquals(t, game, -1)
// }
//
// func TestMinimax(t *testing.T) {
// 	game := NewGame()
// 	expected := 0
// 	actual := game.Minimax(game.SearchDepth)
// 	if actual != expected {
// 		t.Errorf("Expected minimax value of %d, got %d for game state \n%s", expected, actual, game)
// 	}
//
// 	game.State.Active = -1
// 	expected = 0
// 	actual = game.Minimax(game.SearchDepth)
// 	if actual != expected {
// 		t.Errorf("Expected minimax value of %d, got %d for game state \n%s", expected, actual, game)
// 	}
// }
//
// func TestSearch(t *testing.T) {
// 	game := NewGame()
//
// 	expectedMoves := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
// 	expectedVals := []int{0, 0, 0, 0, 0, 0, 0, 0, 0}
// 	MoveSearchEquals(t, game, expectedMoves, expectedVals)
//
// 	game.FromString("XXOOXOXO ,X")
// 	expectedMoves = []int{8}
// 	expectedVals = []int{291}
// 	MoveSearchEquals(t, game, expectedMoves, expectedVals)
//
// 	game.FromString("XXOOXOXO ,O")
// 	expectedMoves = []int{8}
// 	expectedVals = []int{-291}
// 	MoveSearchEquals(t, game, expectedMoves, expectedVals)
//
// 	game.FromString("XXOOX XO ,O")
// 	expectedMoves = []int{8, 5}
// 	expectedVals = []int{0, 291}
// 	MoveSearchEquals(t, game, expectedMoves, expectedVals)
//
// 	game.FromString("X        ,O")
// 	game.SearchDepth = 4
// 	expectedMoves = []int{4, 8, 5, 2, 1, 3, 6, 7}
// 	expectedVals = []int{9, 11, 12, 19, 20, 21, 21, 22}
// 	MoveSearchEquals(t, game, expectedMoves, expectedVals)
// }

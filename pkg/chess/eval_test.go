package chess

import "testing"

func TestMaterial(t *testing.T) {
	game := NewGame()

	whiteMaterial := game.Material(WHITE)
	blackMaterial := game.Material(BLACK)
	if whiteMaterial != blackMaterial {
		t.Errorf("White material (%d) should equal black material (%d)", whiteMaterial, blackMaterial)
	}
}

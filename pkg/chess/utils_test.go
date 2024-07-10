package chess

import "testing"

func TestAlgebraicLocation(t *testing.T) {
	expectedAlgebraic1 := "a1"
	expectedInt1 := 0

	actualInt1 := algebraic2Int(expectedAlgebraic1)
	actualAlgebraic1 := int2algebraic(actualInt1)

	if expectedInt1 != actualInt1 {
		t.Errorf("Expected integer representation (%d) != actual integer representation (%d)", expectedInt1, actualInt1)
	}

	if expectedAlgebraic1 != actualAlgebraic1 {
		t.Errorf("Expected integer representation (%s) != actual integer representation (%s)", expectedAlgebraic1, actualAlgebraic1)
	}
}

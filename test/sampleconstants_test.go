package testing

/*
A Sample test harness.
*/

import (
	"github.com/AaronSaikovski/gostarter/constants"
	"testing"
)

// A testing function.
func TestConstant(t *testing.T) {

	expected := 10
	if constants.LoopConstant != expected {
		t.Errorf("const expected '%d' but got '%d'", expected, constants.LoopConstant)
	}
}

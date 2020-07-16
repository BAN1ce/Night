package utils

import "testing"

func TestHasWildcard(t *testing.T) {

	if HasWildcard(nil) == true {
		t.Error("nil has wildcard error")
	}

	if HasWildcard([]string{}) == true {
		t.Error("empty slice has wildcard error")
	}

	if HasWildcard([]string{"product", "device", "number"}) == true {
		t.Error(" normal slice has wildcard error")
	}
	if HasWildcard([]string{"#"}) == false {

		t.Error("# does not have wildcard error")
	}
	if HasWildcard([]string{"+", "device", "num"}) == false {
		t.Error("+ does not have wildcard error")
	}

}

package utils

import (
	"strings"
	"testing"
)

func TestConcatStringSlice(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"d", "e", "f"}

	res := ConcatStringSlice(a, b)
	if strings.Join(res, "") != "abcdef" {
		t.Error("Unexpected result")
	}
}

func TestConcatStringSliceEmpty(t *testing.T) {
	a := []string{}
	b := []string{}

	res := ConcatStringSlice(a, b)
	if strings.Join(res, "") != "" {
		t.Error("Unexpected result")
	}
}

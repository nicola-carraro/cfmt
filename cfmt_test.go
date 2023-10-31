package main

import (
	"fmt"
	"testing"
)

func TestTokenizeString(t *testing.T) {

	text := fmt.Sprint("\"toto\"")

	tokens := tokenize(text)

	if len(tokens) != 1 {
		t.Error("Tokens should have length 1")
	}

	token := tokens[0]

	if token.Type != String {
		t.Error("Token should be String")
	}

	if token.Content != text {
		t.Errorf("Token content should be %s", text)
	}
}

func TestTokenizeFloat(t *testing.T) {

	text := fmt.Sprint("55.0f")

	tokens := tokenize(text)

	if len(tokens) != 1 {
		t.Error("Tokens should have length 1")
	}

	token := tokens[0]

	if token.Type != Float {
		t.Error("Token should be Float")
	}

	if token.Content != text {
		t.Errorf("Token content should be %s", text)
	}
}

func TestTokenizeDoubleWithExponent(t *testing.T) {

	text := fmt.Sprint("123.456e-67")

	tokens := tokenize(text)

	if len(tokens) != 1 {
		t.Error("Tokens should have length 1")
	}

	token := tokens[0]

	if token.Type != Float {
		t.Error("Token should be Float")
	}

	if token.Content != text {
		t.Errorf("Token content should be %s", text)
	}
}

package main

import (
	"testing"
)

func _testTokenizeSingleToken(t *testing.T, text string, tType TokenType, typeName string) {

	tokens := tokenize(text)

	if len(tokens) != 1 {
		t.Error("Tokens should have length 1")
	}

	token := tokens[0]

	if token.Type != tType {
		t.Errorf("Token should be %s", typeName)
	}

	if token.Content != text {
		t.Errorf("Token content should be %s", text)
	}
}

func TestTokenizeString(t *testing.T) {
	_testTokenizeSingleToken(t, "\"toto\"", String, "String")
}

func TestTokenizeFloat(t *testing.T) {
	_testTokenizeSingleToken(t, "55.0f", Float, "Float")
	_testTokenizeSingleToken(t, "123.456e-67", Float, "Float")
	_testTokenizeSingleToken(t, "123e+86", Float, "Float")
}

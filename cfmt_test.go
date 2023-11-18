package main

import (
	"testing"
)

func _testTokenizeSingleToken(t *testing.T, text string, tType TokenType) {

	token := parseToken(text)

	if token.Type != tType {
		t.Errorf("Token should be %s, found %s", tType, token.Type)
	}

	if token.Content != text {
		t.Errorf("Token content should be %s, found %s", text, token.Content)
	}
}

func TestTokenizeString(t *testing.T) {
	_testTokenizeSingleToken(t, "\"toto\"", String)
	_testTokenizeSingleToken(t, "\"to\\\"o\"", String)
}

func TestTokenizeFloat(t *testing.T) {
	_testTokenizeSingleToken(t, "55.0f", Float)
	_testTokenizeSingleToken(t, "123.456e-67", Float)
	_testTokenizeSingleToken(t, "123e+86", Float)
}

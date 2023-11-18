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

func _testFormat(t *testing.T, input string, expected string) {
	output := format(input)

	for i, r := range []byte(expected) {
		if r != output[i] {
			t.Errorf("Index %d, expected %d, output %d", i, r, output[i])
		}
	}

	if output != expected {
		t.Errorf("Output should be:\n%s\n, found:\n%s\n", expected, output)
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

func TestFormatStructDecl(t *testing.T) {
	input :=
		"typedef struct {\r\n" +
			"    int bar;     char *baz;}Foo;"

	expected :=
		"typedef struct {\r\n" +
			"  int bar;\r\n" +
			"  char *baz;\r\n" +
			"} Foo;\r\n"

	_testFormat(t, input, expected)
}

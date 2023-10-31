package main

import (
	"fmt"
	"testing"
)

func TestTokenizeString(t *testing.T) {
     
     text := fmt.Sprint("\"toto\"")

	 tokens:= tokenize(text)

	 if(len(tokens) != 1){
		t.Error("Tokens should have length 1")
	 }

	 token := tokens[0]

	 if(token.Type != String){
		t.Error("Token should be string")
	 }

	 if(token.Content != text){
		t.Errorf("Token content should be %s", text)
	 }
}

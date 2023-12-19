package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path>\n", os.Args[0])
	}

	for i := 1; i < len(os.Args); i++ {
		path := os.Args[i]

		data, err := os.ReadFile(path)

		if err != nil {
			log.Fatalf("Error reading %s: %s", path, err)
		}

		text := string(data)

		formattedText := format(text)

		fmt.Print(formattedText)

		// os.WriteFile(path, []byte(formattedText), 0600)

		// if err != nil {
		// 	log.Fatalf("Error writing %s: %s", path, err)
		// }

	}

}

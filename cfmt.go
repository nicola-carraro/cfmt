package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {

	var stdout bool = false

	flag.BoolVar(&stdout, "stdout", false, "print to std out")

	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatalf("Usage: %s [-stdout] <path>[...]>\n", os.Args[0])
	}

	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)

		if err != nil {
			log.Fatalf("Error reading %s: %s", path, err)
		}

		text := string(data)

		formattedText := format(text)

		if stdout {
			fmt.Print(formattedText)
		} else {
			os.WriteFile(path, []byte(formattedText), 0600)

			if err != nil {
				log.Fatalf("Error writing %s: %s", path, err)
			}
		}
	}
}

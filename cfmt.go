package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-stdout] path1 [path2 ...]\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func main() {
	var stdout bool = false
	flag.BoolVar(&stdout, "stdout", false, "print to standard output instead of overwriting files")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	for _, path := range flag.Args() {
		data, err := os.ReadFile(path)

		if err != nil {
			usage()
			flag.PrintDefaults()
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

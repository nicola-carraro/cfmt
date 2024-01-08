package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-stdout] path1 [path2 ...]\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func printError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
}

func main() {

	timer := time.Now()
	var stdout bool = false
	flag.BoolVar(&stdout, "stdout", false, "print to standard output instead of overwriting files")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	for _, path := range flag.Args() {
		fmt.Println(path)
		data, err := os.ReadFile(path)

		if err != nil {
			printError(err)
			continue
		}

		text := string(data)

		formattedText := Format(text)

		if stdout {
			fmt.Print(formattedText)
		} else {
			os.WriteFile(path, []byte(formattedText), 0600)

			if err != nil {
				printError(err)
			}
		}
	}

	elapsed := time.Since(timer)
	fmt.Println(elapsed)
}

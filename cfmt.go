package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-stdout] path1 [path2 ...]\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

func printError(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
}

func formatFile(path string, stdout bool) {

	data, err := os.ReadFile(path)

	if err != nil {
		printError(err)
		return
	}

	text := string(data)

	formattedText, err := Format(text)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:%s\n", path, err)
		return
	}

	fmt.Println(path)
	if stdout {
		fmt.Print(formattedText)
	} else {
		os.WriteFile(path, []byte(formattedText), 0600)

		if err != nil {
			printError(err)
		}
	}
}

func main() {

	var stdout bool = false
	flag.BoolVar(&stdout, "stdout", false, "print to standard output instead of overwriting files")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	paths := []string{}

	for _, path := range flag.Args() {
		matches, err := filepath.Glob(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			return
		}

		paths = append(paths, matches...)
	}

	wg := sync.WaitGroup{}


	for _, path := range paths {
		path := path

		wg.Add(1)

		go func() {
			defer wg.Done()
			formatFile(path, stdout)
		}()

	}

	wg.Wait()
}

package main

import "fmt"
import "os"

func main() {

	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		fmt.Println("Error reading ", path)
	}

	fmt.Printf("%s", data)
}

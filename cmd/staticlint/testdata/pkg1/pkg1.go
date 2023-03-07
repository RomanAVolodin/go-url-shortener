package main

import (
	"fmt"
	"os"
)

func mulfunc(i int) int {
	return i * 2
}

func main() {
	fmt.Println("some string", mulfunc(4))
	os.Exit(0) // want "found os.Exit in main func of package main"
}

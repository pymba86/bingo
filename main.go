package main

import (
	"fmt"
	"github.com/pymba86/bingo/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

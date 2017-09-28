package main

import (
	"fmt"
	"os"
)

func main() {
	rootcmd.AddCommand(startcmd)

	if err := rootcmd.Execute(); err != nil {
		fmt.Println(err)
		rootcmd.Help()
		os.Exit(1)
	}
}

func start() {
}

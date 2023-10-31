package main

import "os"

func main() {
	err := newRootCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}

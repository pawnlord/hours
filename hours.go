package main

import (
	"fmt"
	"os"
)

func main() {

	files, err := os.ReadDir(".")
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fmt.Println(file.Name())
			stat, err := os.Stat(file.Name())
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(stat.Name(), stat.Size(), stat.ModTime())
		}
	}
}

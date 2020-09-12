package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

func main() {

	file, err := os.Open("input.txt")
	if err != nil {
		log.Fatalf("Cannot read file: %s", err)
	}
	reader := bufio.NewReader(file)
	for {
		in, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Cannot read file: %s", err)
		}
	}

}

package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	Linecounts := make(map[string]int)

	files := os.Args[1:]

	if len(files) == 0 {
		countLine(os.Stdin, Linecounts)
	} else {

		for _, filename := range files {
			file_pt, err := os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Mdup2: %v\n", err)
				continue
			}
			countLine(file_pt, Linecounts)
			file_pt.Close()
		}
	}

	// output
	for line, times := range Linecounts {
		if times > 1 {
			fmt.Printf("%d\t%s\n", times, line)
		}
	}

}

func countLine(file_point *os.File, counter map[string]int) {
	input := bufio.NewScanner(file_point)
	for input.Scan() {
		counter[input.Text()]++
	}

}

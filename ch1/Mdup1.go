package main

import (
	"os"
	"fmt"
	"bufio"
)

func main(){
	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan(){
		counts[input.Text()]++
		// word := input.Text()
		// counts[word] += 1
	}

	for word, times := range counts {
		if times > 1 {
			fmt.Printf("%d\t%s\n", times, word)
		}
	}

}
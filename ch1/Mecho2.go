package main

import (
	"fmt"
	"os"
)


func main(){
	//var s,sep string
	
	s, sep := "", ""
	//use range 
	for _, arg := range os.Args[1:] {
		s += sep + arg
		sep = " "
	}
	fmt.Printf(s)
}
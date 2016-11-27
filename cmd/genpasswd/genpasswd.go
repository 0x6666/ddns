package main

import (
	"crypto/sha1"
	"fmt"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: genpasswd password")
		return
	}

	fmt.Printf("%x\n", sha1.Sum([]byte(os.Args[1])))
	return
}

package main

import (
	"fmt"
	"net/http"
)

func main() {

	resp, err := http.Head("http://github.com/")
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("%v", resp)
	}

}

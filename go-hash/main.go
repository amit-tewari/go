package main

import (
	xxh "bitbucket.org/StephaneBunel/xxhash-go"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		// `Text` returns the current token, here the next line,
		// from the input.
		ucl := strings.ToLower(scanner.Text())
		h32 := xxh.Checksum32(([]byte)(ucl))
		fmt.Println(h32, ucl)

		// Write out the uppercased line.
		//        fmt.Println(ucl)
	}

	// Check for errors during `Scan`. End of file is
	// expected and not reported by `Scan` as an error.
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

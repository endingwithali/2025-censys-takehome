package main

import (
	"fmt"
	"os"

	"github.com/nsf/jsondiff"
)

func main() {
	file1 := "/Users/mommy/Documents/coding/interviews/2025censys/custom/host_snapshots/host_125.199.235.74_2025-09-10T03-00-00Z.json"
	file2 := "/Users/mommy/Documents/coding/interviews/2025censys/custom/host_snapshots/host_125.199.235.74_2025-09-20T12-00-00Z.json"

	// Read both files
	b1, err := os.ReadFile(file1)
	if err != nil {
		panic(err)
	}
	b2, err := os.ReadFile(file2)
	if err != nil {
		panic(err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	diff, explanation := jsondiff.Compare(b1, b2, &opts)
	fmt.Println(diff)        // → jsondiff.FullMatch / SupersetMatch / NoMatch
	fmt.Println(explanation) // → pretty printed diff output
}

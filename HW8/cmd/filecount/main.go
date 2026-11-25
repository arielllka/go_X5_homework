package main

import (
	"fmt"
	"os"

	"example.com/lesson11/files/processor"
)

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = []string{"testdata/orders-a.txt", "testdata/orders-b.txt", "testdata/empty.txt"}
	}
	count := 0
	failed := false
	for result := range processor.ProcessFiles(paths, 3) {
		count++
		if result.Err != nil {
			fmt.Fprintln(os.Stderr, result.Err)
			failed = true
			continue
		}
		fmt.Printf("%s: строк %d, слов %d\n", result.Path, result.Lines, result.Words)
	}
	fmt.Println("Обработано файлов:", count)
	if failed {
		os.Exit(1)
	}
}

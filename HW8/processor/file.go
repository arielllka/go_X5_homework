package processor

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

type Result struct {
	Path  string
	Lines int
	Words int
	Err   error
}

// analyzeFile обрабатывает один файл. При ошибке возвращает путь и причину.
func analyzeFile(path string) Result {
	file, err := os.Open(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}
	defer file.Close()

	result := Result{Path: path}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			result.Lines++
			result.Words += len(strings.Fields(line))
		}
		if errors.Is(err, io.EOF) {
			return result
		}
		if err != nil {
			return Result{Path: path, Err: err}
		}
	}
}

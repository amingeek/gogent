package plugins

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file %s: %v", path, err)
	}
	return string(data), nil
}

func ReadFileLines(path string, startLine, endLine int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %v", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && (endLine == -1 || lineNum <= endLine) {
			lines = append(lines, fmt.Sprintf("%4d: %s", lineNum, scanner.Text()))
		}
		if endLine != -1 && lineNum > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	if len(lines) == 0 {
		return fmt.Sprintf("No lines found in range %d-%d", startLine, endLine), nil
	}

	return strings.Join(lines, "\n"), nil
}

func SearchFileContent(path, pattern string, contextLines int, caseSensitive bool) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %v", path, err)
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	searchPattern := pattern
	if !caseSensitive {
		searchPattern = strings.ToLower(pattern)
	}

	var matches []int
	for i, line := range allLines {
		checkLine := line
		if !caseSensitive {
			checkLine = strings.ToLower(line)
		}
		if strings.Contains(checkLine, searchPattern) {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for pattern: %s", pattern), nil
	}

	var results []string
	for _, matchIdx := range matches {
		start := matchIdx - contextLines
		if start < 0 {
			start = 0
		}
		end := matchIdx + contextLines
		if end >= len(allLines) {
			end = len(allLines) - 1
		}

		var contextBlock []string
		for i := start; i <= end; i++ {
			prefix := "  "
			if i == matchIdx {
				prefix = ">> "
			}
			contextBlock = append(contextBlock, fmt.Sprintf("%s%4d: %s", prefix, i+1, allLines[i]))
		}
		results = append(results, strings.Join(contextBlock, "\n"))
	}

	return strings.Join(results, "\n\n---\n\n"), nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

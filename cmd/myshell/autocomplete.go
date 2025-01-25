package main

import (
	"os"
	"slices"
	"strings"
)

type AutocompleteResult int

const (
	AutocompleteNone AutocompleteResult = iota
	AutocompleteFound
	AutocompleteMore
)

func autocompletePrint(line *string, candidate string, hasMore bool) {
	os.Stdout.WriteString(candidate)
	*line += candidate

	if !hasMore {
		os.Stdout.WriteString(" ")
		*line += " "
	}
}

func findSharedPrefix(candidates []string) string {
	first := candidates[0]
	candidates = candidates[1:]

	firstLength := len(first)
	if firstLength == 0 {
		return ""
	}

	end := 1
	for ; end < firstLength; end += 1 {
		oneIsNotMatching := false

		for index, candidate := range candidates {
			if index == 0 {
				continue
			}

			if first[:end] != candidate[:end] {
				oneIsNotMatching = true
				break
			}
		}

		if oneIsNotMatching {
			end -= 1
			break
		}
	}

	return first[:end]
}

func autocomplete(line *string, bell_rang bool) AutocompleteResult {
	var candidates []string

	for name := range builtins {
		if strings.HasPrefix(name, *line) {
			candidate := name[len(*line):]
			candidates = append(candidates, candidate)
		}
	}

	PATH := os.Getenv("PATH")
	for _, directory := range strings.Split(PATH, ":") {
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, *line) {
				continue
			}

			path := strings.Join([]string{directory, name}, "/")

			stat, err := os.Stat(path)
			if err != nil || !stat.Mode().IsRegular() || stat.Mode().Perm()&0111 == 0 {
				continue
			}

			candidate := name[len(*line):]
			if !slices.Contains(candidates, candidate) {
				candidates = append(candidates, candidate)
			}
		}
	}

	if len(candidates) == 0 {
		return AutocompleteNone
	}

	if len(candidates) == 1 {
		candidate := candidates[0]
		autocompletePrint(line, candidate, false)

		return AutocompleteFound
	}

	slices.SortFunc(candidates, func(left string, right string) int {
		leftLength := len(left)
		rightLength := len(right)

		if leftLength != rightLength {
			return leftLength - rightLength
		}

		return strings.Compare(left, right)
	})

	prefix := findSharedPrefix(candidates)
	if len(prefix) != 0 {
		autocompletePrint(line, prefix, true)

		return AutocompleteFound
	}

	if bell_rang {
		os.Stdout.WriteString("\n")

		for index, candidate := range candidates {
			if index != 0 {
				os.Stdout.WriteString("  ")
			}

			os.Stdout.WriteString(*line)
			os.Stdout.WriteString(candidate)
		}

		os.Stdout.WriteString("\n")
		prompt()
		os.Stdout.WriteString(*line)
	}

	return AutocompleteMore
}

func bell() {
	os.Stdout.Write([]byte{'\a'})
}

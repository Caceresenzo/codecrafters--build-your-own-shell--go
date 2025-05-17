package main

import (
	"unicode"
)

type standardStreamName int

const (
	standardOutput  standardStreamName = 1
	standardError   standardStreamName = 2
	standardUnknown standardStreamName = -1
)

type redirect struct {
	streamName standardStreamName
	path       string
	append     bool
}

type lineParser struct {
	chars     []rune
	index     int
	arguments []string
	redirects []redirect
}

type parsedLine struct {
	arguments []string
	redirects []redirect
}

const (
	end         = '\000'
	space       = ' '
	single      = '\''
	double      = '"'
	backslash   = '\\'
	greaterThan = '>'
)

type nullString struct {
	value string
	valid bool
}

func toStreamName(character rune) standardStreamName {
	switch character {
	case '1':
		return standardOutput
	case '2':
		return standardError
	default:
		return standardUnknown
	}
}

func (parser *lineParser) next() rune {
	parser.index++

	return parser.charAt(parser.index)
}

func (parser *lineParser) peek() rune {
	return parser.charAt(parser.index + 1)
}

func (parser *lineParser) charAt(index int) rune {
	if index >= len(parser.chars) {
		return end
	}

	return parser.chars[index]
}

func (parser *lineParser) handleBackslash(builder *[]rune, inQuote bool) {
	character := parser.next()
	if character == end {
		return
	}

	if inQuote {
		mapped := mapBackslashCharacter(character)
		if mapped != end {
			character = mapped
		} else {
			*builder = append(*builder, backslash)
		}
	}

	*builder = append(*builder, character)
}

func (parser *lineParser) handleRedirect(streamName standardStreamName) {
	append_ := parser.peek() == greaterThan
	if append_ {
		parser.next()
	}

	path := parser.nextArgument().value

	parser.redirects = append(parser.redirects, redirect{
		streamName: streamName,
		path:       path,
		append:     append_,
	})
}

func (parser *lineParser) nextArgument() nullString {
	builder := make([]rune, 0)

	character := end
	for {
		character = parser.next()
		if character == end {
			break
		}

		switch character {
		case space:
			if len(builder) != 0 {
				return nullString{
					string(builder),
					true,
				}
			}
		case single:
			for {
				character = parser.next()
				if character == end || character == single {
					break
				}

				builder = append(builder, character)
			}
		case double:
			for {
				character = parser.next()
				if character == end || character == double {
					break
				}

				switch character {
				case backslash:
					parser.handleBackslash(&builder, true)
				default:
					builder = append(builder, character)
				}
			}
		case backslash:
			parser.handleBackslash(&builder, false)
		case greaterThan:
			parser.handleRedirect(standardOutput)
		default:
			if unicode.IsDigit(character) && parser.peek() == greaterThan {
				parser.next()
				parser.handleRedirect(toStreamName(character))
			} else {
				builder = append(builder, character)
			}
		}
	}

	if len(builder) != 0 {
		return nullString{
			string(builder),
			true,
		}
	}

	return nullString{
		"",
		false,
	}
}

func mapBackslashCharacter(character rune) rune {
	switch character {
	case backslash:
		fallthrough
	case double:
		return character
	default:
		return end
	}
}

func parseArgv(line string) parsedLine {
	parser := lineParser{
		chars:     []rune(line),
		index:     -1,
		arguments: make([]string, 0),
		redirects: make([]redirect, 0),
	}

	for {
		argument := parser.nextArgument()
		if !argument.valid {
			break
		}

		parser.arguments = append(parser.arguments, argument.value)
	}

	return parsedLine{
		arguments: parser.arguments,
		redirects: parser.redirects,
	}
}

package main

type lineParser struct {
	chars   []rune
	index   int
	builder []rune
}

const (
	end       = '\000'
	space     = ' '
	single    = '\''
	double    = '"'
	backslash = '\\'
)

func next(parser *lineParser) rune {
	length := len(parser.chars)

	if parser.index < length {
		index := parser.index
		parser.index++

		return parser.chars[index]
	}

	return end
}

func handleBackslash(state *lineParser, inQuote bool) {
	character := next(state)
	if character == end {
		return
	}

	if inQuote {
		mapped := mapBackslashCharacter(character)
		if mapped != end {
			character = mapped
		} else {
			state.builder = append(state.builder, backslash)
		}
	}

	state.builder = append(state.builder, character)
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

func parseArgv(line string) []string {
	argv := make([]string, 0)

	state := lineParser{
		chars:   []rune(line),
		index:   0,
		builder: make([]rune, 0),
	}

	character := end
	for {
		character = next(&state)
		if character == end {
			break
		}

		switch character {
		case space:
			if len(state.builder) != 0 {
				argv = append(argv, string(state.builder))
				state.builder = state.builder[:0]
			}
		case single:
			for {
				character = next(&state)
				if character == end || character == single {
					break
				}

				state.builder = append(state.builder, character)
			}
		case double:
			for {
				character = next(&state)
				if character == end || character == double {
					break
				}

				switch character {
				case backslash:
					handleBackslash(&state, true)
				default:
					state.builder = append(state.builder, character)
				}
			}
		case backslash:
			handleBackslash(&state, false)
		default:
			state.builder = append(state.builder, character)
		}
	}

	if len(state.builder) != 0 {
		argv = append(argv, string(state.builder))
	}

	return argv
}

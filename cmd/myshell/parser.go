package main

type line_parser struct {
	chars   []rune
	index   int
	builder []rune
}

const (
	END    = '\000'
	SPACE  = ' '
	SINGLE = '\''
)

func next(parser *line_parser) rune {
	length := len(parser.chars)

	if parser.index < length {
		index := parser.index
		parser.index++

		return parser.chars[index]
	}

	return END
}

func parse_argv(line string) []string {
	argv := make([]string, 0)

	state := line_parser{
		chars:   []rune(line),
		index:   0,
		builder: make([]rune, 0),
	}

	character := END
	for {
		character = next(&state)
		if character == END {
			break
		}

		switch character {
		case SPACE:
			if len(state.builder) != 0 {
				argv = append(argv, string(state.builder))
				state.builder = state.builder[:0]
			}
		case SINGLE:
			for {
				character = next(&state)
				if character == END || character == SINGLE {
					break
				}

				state.builder = append(state.builder, character)
			}
		default:
			state.builder = append(state.builder, character)
		}
	}

	if len(state.builder) != 0 {
		argv = append(argv, string(state.builder))
	}

	return argv
}

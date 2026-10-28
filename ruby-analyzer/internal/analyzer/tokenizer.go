package analyzer

import (
	"strings"
	"unicode"
)

func isIdentifierStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentifierPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func startsWith(runes []rune, position int, value string) bool {
	valueRunes := []rune(value)
	if position+len(valueRunes) > len(runes) {
		return false
	}
	for i := 0; i < len(valueRunes); i++ {
		if runes[position+i] != valueRunes[i] {
			return false
		}
	}
	return true
}

func Tokenize(source string) []Token {
	runes := []rune(source)
	tokens := make([]Token, 0)
	i := 0

	for i < len(runes) {
		hasNL := false
		for i < len(runes) && unicode.IsSpace(runes[i]) {
			if runes[i] == '\n' {
				hasNL = true
			}
			i++
		}
		if i >= len(runes) {
			break
		}

		if runes[i] == '#' {
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			continue
		}

		if runes[i] == '"' || runes[i] == '\'' {
			quote := runes[i]
			i++
			var builder strings.Builder
			builder.WriteRune(quote)
			for i < len(runes) {
				if runes[i] == '\\' && i+1 < len(runes) {
					builder.WriteRune(runes[i])
					builder.WriteRune(runes[i+1])
					i += 2
					continue
				}
				builder.WriteRune(runes[i])
				if runes[i] == quote {
					i++
					break
				}
				i++
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand, AfterNewline: hasNL})
			continue
		}

		if runes[i] == '@' || runes[i] == '$' {
			var builder strings.Builder
			builder.WriteRune(runes[i])
			i++
			if i < len(runes) && runes[i] == '@' {
				builder.WriteRune(runes[i])
				i++
			}
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand, AfterNewline: hasNL})
			continue
		}

		if unicode.IsDigit(runes[i]) {
			var builder strings.Builder
			hasDot := false
			for i < len(runes) {
				if unicode.IsDigit(runes[i]) || runes[i] == '_' {
					builder.WriteRune(runes[i])
					i++
					continue
				}
				if runes[i] == '.' && !hasDot && i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
					hasDot = true
					builder.WriteRune(runes[i])
					i++
					continue
				}
				break
			}
			tokens = append(tokens, Token{Value: builder.String(), Type: TokenOperand, AfterNewline: hasNL})
			continue
		}

		if isIdentifierStart(runes[i]) {
			var builder strings.Builder
			for i < len(runes) && isIdentifierPart(runes[i]) {
				builder.WriteRune(runes[i])
				i++
			}
			if i < len(runes) && (runes[i] == '?' || runes[i] == '!') {
				builder.WriteRune(runes[i])
				i++
			}
			word := builder.String()

			if blockOpeners[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenOpenBlock, AfterNewline: hasNL})
				continue
			}
			if word == "end" {
				tokens = append(tokens, Token{Value: word, Type: TokenCloseBlock, AfterNewline: hasNL})
				continue
			}
			if blockParts[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenBlockPart, AfterNewline: hasNL})
				continue
			}
			if keywordOperators[word] {
				tokens = append(tokens, Token{Value: word, Type: TokenOperator, AfterNewline: hasNL})
				continue
			}
			tokens = append(tokens, Token{Value: word, Type: TokenOperand, AfterNewline: hasNL})
			continue
		}

		if runes[i] == '(' || runes[i] == ')' || runes[i] == '[' || runes[i] == ']' || runes[i] == '{' || runes[i] == '}' {
			tokens = append(tokens, Token{Value: string(runes[i]), Type: TokenOperator, AfterNewline: hasNL})
			i++
			continue
		}

		found := false
		for _, operator := range multiCharOperators {
			if startsWith(runes, i, operator) {
				tokens = append(tokens, Token{Value: operator, Type: TokenOperator, AfterNewline: hasNL})
				i += len([]rune(operator))
				found = true
				break
			}
		}
		if found {
			continue
		}

		current := string(runes[i])
		if symbolOperators[current] {
			tokens = append(tokens, Token{Value: current, Type: TokenOperator, AfterNewline: hasNL})
			i++
			continue
		}

		i++
	}

	return tokens
}

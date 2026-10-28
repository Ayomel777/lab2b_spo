package analyzer

type TokenType int

const (
	TokenOperand TokenType = iota
	TokenOperator
	TokenOpenBlock
	TokenCloseBlock
	TokenBlockPart
)

type Token struct {
	Value        string
	Type         TokenType
	AfterNewline bool
}

var blockOpeners = map[string]bool{
	"if": true, "unless": true, "while": true, "until": true,
	"for": true, "case": true, "def": true, "class": true,
	"module": true, "begin": true, "do": true,
}

var blockParts = map[string]bool{
	"else": true, "elsif": true, "when": true,
	"rescue": true, "ensure": true, "then": true, "in": true,
}

var keywordOperators = map[string]bool{
	"return": true, "break": true, "next": true, "redo": true,
	"retry": true, "yield": true, "super": true, "alias": true,
	"undef": true, "and": true, "or": true, "not": true,
}

var symbolOperators = map[string]bool{
	"=": true, "+": true, "-": true, "*": true, "/": true, "%": true, "**": true,
	"+=": true, "-=": true, "*=": true, "/=": true, "%=": true, "**=": true,
	"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"<=>": true, "===": true, "=~": true, "!~": true,
	"&&": true, "||": true, "!": true,
	"&": true, "|": true, "^": true, "~": true,
	"&=": true, "|=": true, "^=": true,
	"<<": true, ">>": true, "<<=": true, ">>=": true,
	"&&=": true, "||=": true,
	"..": true, "...": true, "=>": true, "->": true,
	"&.": true, "::": true, ".": true, "?": true, ":": true,
	",": true, ";": true,
}

var multiCharOperators = []string{
	"<<=", ">>=", "**=", "&&=", "||=",
	"<=>", "===", "...",
	"==", "!=", ">=", "<=", "=~", "!~",
	"<<", ">>", "**",
	"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=",
	"&&", "||",
	"..", "=>", "->", "&.", "::",
}

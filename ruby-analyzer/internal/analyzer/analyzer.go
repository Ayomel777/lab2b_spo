package analyzer

type Result struct {
	Boundary BoundaryResult `json:"boundary"`
	CFG      *CFG           `json:"cfg"`
}

func Analyze(source string) Result {
	tokens := Tokenize(source)
	cfg := BuildCFG(tokens)

	return Result{
		Boundary: AnalyzeBoundary(cfg),
		CFG:      cfg,
	}
}

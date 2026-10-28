package analyzer

import (
	"fmt"
	"strings"
)

type CFGNode struct {
	ID    int        `json:"id"`
	Kind  string     `json:"kind"`
	Label string     `json:"label"`
	Out   []*CFGNode `json:"-"`
	In    []*CFGNode `json:"-"`
}

type CFG struct {
	Nodes []*CFGNode `json:"nodes"`
	Start *CFGNode   `json:"-"`
	End   *CFGNode   `json:"-"`
}

func (cfg *CFG) addNode(kind, label string) *CFGNode {
	n := &CFGNode{ID: len(cfg.Nodes), Kind: kind, Label: label}
	cfg.Nodes = append(cfg.Nodes, n)
	return n
}

func (cfg *CFG) addEdge(from, to *CFGNode) {
	if from == nil || to == nil {
		return
	}
	for _, existing := range from.Out {
		if existing == to {
			return
		}
	}
	from.Out = append(from.Out, to)
	to.In = append(to.In, from)
}

func BuildCFG(tokens []Token) *CFG {
	cfg := &CFG{}
	start := cfg.addNode("start", "Начало")
	p := &cfgBuilder{tokens: tokens, cfg: cfg}
	end := cfg.addNode("end", "Конец")

	prev := p.parseSeq(start, stopNever)
	if prev != nil {
		cfg.addEdge(prev, end)
	} else {
		cfg.addEdge(start, end)
	}

	cfg.Start = start
	cfg.End = end
	return cfg
}

type cfgBuilder struct {
	tokens []Token
	pos    int
	cfg    *CFG
}

func stopNever(Token) bool { return false }

func stopAtBranchOrEnd(t Token) bool {
	if t.Type == TokenCloseBlock && t.Value == "end" {
		return true
	}
	if t.Type == TokenBlockPart && (t.Value == "else" || t.Value == "elsif" || t.Value == "rescue" || t.Value == "ensure") {
		return true
	}
	return false
}

func stopAtEnd(t Token) bool {
	return t.Type == TokenCloseBlock && t.Value == "end"
}

func (b *cfgBuilder) parseSeq(prev *CFGNode, stopFn func(Token) bool) *CFGNode {
	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if stopFn(t) {
			return prev
		}

		if t.Type == TokenCloseBlock {
			b.pos++
			continue
		}
		if t.Type == TokenBlockPart {
			return prev
		}

		if t.Type == TokenOpenBlock {
			switch t.Value {
			case "if", "unless":
				prev = b.parseIf(prev)
				continue
			case "while", "until", "for":
				prev = b.parseWhile(prev)
				continue
			case "case":
				prev = b.parseCase(prev)
				continue
			case "def", "class", "module", "begin":
				// НЕ ПРОПУСКАЕМ блок, а заходим внутрь!
				b.pos++ // пропускаем ключевое слово
				// Пропускаем имя функции/класса и аргументы до конца строки
				for b.pos < len(b.tokens) && !b.tokens[b.pos].AfterNewline {
					b.pos++
				}
				// Парсим тело
				prev = b.parseSeq(prev, stopAtEnd)
				if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
					b.pos++ // пропускаем end
				}
				continue
			case "do":
				b.pos++
				// Пропускаем параметры блока (|x|) до конца строки
				for b.pos < len(b.tokens) && !b.tokens[b.pos].AfterNewline {
					b.pos++
				}
				prev = b.parseSeq(prev, stopAtEnd)
				if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
					b.pos++
				}
				continue
			}
		}

		prev = b.parseStatement(prev)
	}
	return prev
}

func (b *cfgBuilder) parseStatement(prev *CFGNode) *CFGNode {
	startPos := b.pos
	depth := 0

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenOperator {
			switch t.Value {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				if depth > 0 {
					depth--
				}
			case ";":
				if depth == 0 {
					b.pos++
					goto emit
				}
			}
		}
		if depth == 0 && (t.Type == TokenOpenBlock || t.Type == TokenCloseBlock || t.Type == TokenBlockPart) {
			break
		}
		if depth == 0 && b.pos > startPos && t.AfterNewline {
			break
		}
		b.pos++
	}

emit:
	if b.pos == startPos {
		b.pos++
		return prev
	}
	label := b.labelBetween(startPos, b.pos)
	node := b.cfg.addNode("op", label)
	b.cfg.addEdge(prev, node)
	return node
}

func (b *cfgBuilder) parseIf(prev *CFGNode) *CFGNode {
	b.pos++
	condLabel := b.readCondition()
	cond := b.cfg.addNode("cond", condLabel)
	b.cfg.addEdge(prev, cond)

	lastCond := cond
	branchEnds := []*CFGNode{}
	hasElse := false

	thenEnd := b.parseSeq(cond, stopAtBranchOrEnd)
	if thenEnd != cond {
		branchEnds = append(branchEnds, thenEnd)
	}

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenBlockPart && t.Value == "elsif" {
			b.pos++
			label := b.readCondition()
			elsifCond := b.cfg.addNode("cond", label)
			b.cfg.addEdge(lastCond, elsifCond)

			elsifEnd := b.parseSeq(elsifCond, stopAtBranchOrEnd)
			if elsifEnd != elsifCond {
				branchEnds = append(branchEnds, elsifEnd)
			}
			lastCond = elsifCond
			continue
		}
		if t.Type == TokenBlockPart && t.Value == "else" {
			b.pos++
			elseEnd := b.parseSeq(lastCond, stopAtBranchOrEnd)
			if elseEnd != lastCond {
				branchEnds = append(branchEnds, elseEnd)
			}
			hasElse = true
			continue
		}
		break
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	merge := b.cfg.addNode("merge", "Слияние if")
	for _, be := range branchEnds {
		b.cfg.addEdge(be, merge)
	}
	if !hasElse {
		b.cfg.addEdge(lastCond, merge)
	}
	return merge
}

func (b *cfgBuilder) parseWhile(prev *CFGNode) *CFGNode {
	b.pos++
	condLabel := b.readCondition()
	cond := b.cfg.addNode("cond", condLabel)
	b.cfg.addEdge(prev, cond)

	bodyEnd := b.parseSeq(cond, stopAtEnd)
	if bodyEnd != cond {
		b.cfg.addEdge(bodyEnd, cond)
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	merge := b.cfg.addNode("merge", "Выход из цикла")
	b.cfg.addEdge(cond, merge)
	return merge
}

func (b *cfgBuilder) parseCase(prev *CFGNode) *CFGNode {
	b.pos++
	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenBlockPart && t.Value == "when" {
			break
		}
		if t.Type == TokenCloseBlock {
			break
		}
		b.pos++
	}

	first := true
	var lastCond *CFGNode
	branchEnds := []*CFGNode{}
	hasElse := false

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenBlockPart && t.Value == "when" {
			b.pos++
			label := b.readCondition()
			whenCond := b.cfg.addNode("cond", label)

			if first {
				b.cfg.addEdge(prev, whenCond)
				first = false
			} else {
				b.cfg.addEdge(lastCond, whenCond)
			}

			bodyEnd := b.parseSeq(whenCond, func(tk Token) bool {
				if tk.Type == TokenBlockPart && (tk.Value == "when" || tk.Value == "else") {
					return true
				}
				return tk.Type == TokenCloseBlock && tk.Value == "end"
			})
			if bodyEnd != whenCond {
				branchEnds = append(branchEnds, bodyEnd)
			}
			lastCond = whenCond
			continue
		}
		if t.Type == TokenBlockPart && t.Value == "else" {
			b.pos++
			elseEnd := b.parseSeq(lastCond, stopAtEnd)
			if elseEnd != lastCond {
				branchEnds = append(branchEnds, elseEnd)
			}
			hasElse = true
			continue
		}
		break
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	merge := b.cfg.addNode("merge", "Слияние case")
	for _, be := range branchEnds {
		b.cfg.addEdge(be, merge)
	}
	if !hasElse && lastCond != nil {
		b.cfg.addEdge(lastCond, merge)
	}
	if first {
		b.cfg.addEdge(prev, merge)
	}
	return merge
}

func (b *cfgBuilder) readCondition() string {
	startPos := b.pos
	depth := 0

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenOperator {
			switch t.Value {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				if depth > 0 {
					depth--
				}
			}
		}
		if depth == 0 && (t.Type == TokenBlockPart || t.Type == TokenCloseBlock) {
			break
		}
		if depth == 0 && b.pos > startPos && t.AfterNewline {
			break
		}
		b.pos++
	}
	return b.labelBetween(startPos, b.pos)
}

func (b *cfgBuilder) labelBetween(from, to int) string {
	if from >= to {
		return "?"
	}
	var sb strings.Builder
	for i := from; i < to; i++ {
		if sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(b.tokens[i].Value)
	}
	s := sb.String()
	if len([]rune(s)) > 36 {
		r := []rune(s)
		return string(r[:33]) + "..."
	}
	return s
}

func reachableSet(src *CFGNode) map[int]*CFGNode {
	visited := map[int]*CFGNode{}
	queue := []*CFGNode{}
	for _, s := range src.Out {
		queue = append(queue, s)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if _, ok := visited[n.ID]; ok {
			continue
		}
		visited[n.ID] = n
		for _, s := range n.Out {
			queue = append(queue, s)
		}
	}
	return visited
}

func canReach(src, target *CFGNode) bool {
	if src == target {
		return true
	}
	visited := map[int]bool{src.ID: true}
	queue := []*CFGNode{src}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, s := range n.Out {
			if s == target {
				return true
			}
			if !visited[s.ID] {
				visited[s.ID] = true
				queue = append(queue, s)
			}
		}
	}
	return false
}

func findMergePoint(branches []*CFGNode) *CFGNode {
	if len(branches) == 0 {
		return nil
	}
	if len(branches) == 1 {
		return branches[0]
	}

	type item struct {
		n    *CFGNode
		from int
	}
	reachCount := map[int]int{}
	visited := map[[2]int]bool{}
	queue := []item{}
	for i, s := range branches {
		queue = append(queue, item{s, i})
	}

	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		key := [2]int{it.n.ID, it.from}
		if visited[key] {
			continue
		}
		visited[key] = true
		reachCount[it.n.ID]++
		if reachCount[it.n.ID] == len(branches) {
			return it.n
		}
		for _, s := range it.n.Out {
			queue = append(queue, item{s, it.from})
		}
	}
	return nil
}

func (cfg *CFG) String() string {
	var sb strings.Builder
	for _, n := range cfg.Nodes {
		fmt.Fprintf(&sb, "[%d] %s (%s) ->", n.ID, n.Label, n.Kind)
		for _, o := range n.Out {
			fmt.Fprintf(&sb, " %d", o.ID)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

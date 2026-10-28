package analyzer

import (
	"fmt"
	"strings"
)

type CFGNode struct {
	ID    int        `json:"id"`
	Kind  string     `json:"kind"` // "start", "end", "op", "cond"
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
	n := &CFGNode{ID: len(cfg.Nodes) + 1, Kind: kind, Label: label}
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

	var exits []*CFGNode

	// Ищем функцию def, чтобы парсить только её тело
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == TokenOpenBlock && t.Value == "def" {
			p.pos++ // пропускаем def
			for p.pos < len(p.tokens) && !p.tokens[p.pos].AfterNewline {
				p.pos++
			}
			exits = p.parseSeq([]*CFGNode{start})
			if p.pos < len(p.tokens) && p.tokens[p.pos].Type == TokenCloseBlock {
				p.pos++
			}
			break
		}
		p.pos++
	}

	end := cfg.addNode("end", "Конец")
	for _, e := range exits {
		cfg.addEdge(e, end)
	}
	if len(cfg.Nodes) == 2 {
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

// parseSeq обрабатывает последовательность операторов
func (b *cfgBuilder) parseSeq(prev []*CFGNode) []*CFGNode {
	current := prev
	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]

		if t.Type == TokenCloseBlock || t.Type == TokenBlockPart {
			return current
		}

		if t.Type == TokenOpenBlock {
			switch t.Value {
			case "if", "unless":
				current = b.parseIf(current)
				continue
			case "while", "until", "for":
				current = b.parseWhile(current)
				continue
			case "case":
				current = b.parseCase(current)
				continue
			}
		}

		current = b.parseStmt(current)
	}
	return current
}

func (b *cfgBuilder) parseStmt(prev []*CFGNode) []*CFGNode {
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
	for _, p := range prev {
		b.cfg.addEdge(p, node)
	}
	return []*CFGNode{node}
}

func (b *cfgBuilder) parseIf(prev []*CFGNode) []*CFGNode {
	b.pos++
	condLabel := b.readCondition()
	cond := b.cfg.addNode("cond", condLabel)
	for _, p := range prev {
		b.cfg.addEdge(p, cond)
	}

	exits := []*CFGNode{}
	thenExits := b.parseSeq([]*CFGNode{cond})
	exits = append(exits, thenExits...)

	lastCond := cond
	hasElse := false

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenBlockPart && t.Value == "elsif" {
			b.pos++
			label := b.readCondition()
			elsifCond := b.cfg.addNode("cond", label)
			b.cfg.addEdge(lastCond, elsifCond)
			elsifExits := b.parseSeq([]*CFGNode{elsifCond})
			exits = append(exits, elsifExits...)
			lastCond = elsifCond
			continue
		}
		if t.Type == TokenBlockPart && t.Value == "else" {
			b.pos++
			elseExits := b.parseSeq([]*CFGNode{lastCond})
			exits = append(exits, elseExits...)
			hasElse = true
			lastCond = nil
			continue
		}
		break
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	if !hasElse && lastCond != nil {
		exits = append(exits, lastCond)
	}
	return exits
}

func (b *cfgBuilder) parseWhile(prev []*CFGNode) []*CFGNode {
	b.pos++
	condLabel := b.readCondition()
	cond := b.cfg.addNode("cond", condLabel)
	for _, p := range prev {
		b.cfg.addEdge(p, cond)
	}

	bodyExits := b.parseSeq([]*CFGNode{cond})
	for _, p := range bodyExits {
		b.cfg.addEdge(p, cond) // обратное ребро
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	return []*CFGNode{cond} // выход из цикла (ложная ветка)
}

func (b *cfgBuilder) parseCase(prev []*CFGNode) []*CFGNode {
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

	exits := []*CFGNode{}
	current := prev
	hasElse := false

	for b.pos < len(b.tokens) {
		t := b.tokens[b.pos]
		if t.Type == TokenBlockPart && t.Value == "when" {
			b.pos++
			label := b.readCondition()
			cond := b.cfg.addNode("cond", label)
			for _, p := range current {
				b.cfg.addEdge(p, cond)
			}

			whenExits := b.parseSeq([]*CFGNode{cond})
			exits = append(exits, whenExits...)
			current = []*CFGNode{cond} // ложная ветка ведет к следующему when
			continue
		}
		if t.Type == TokenBlockPart && t.Value == "else" {
			b.pos++
			elseExits := b.parseSeq(current)
			exits = append(exits, elseExits...)
			hasElse = true
			current = nil
			break
		}
		break
	}

	if b.pos < len(b.tokens) && b.tokens[b.pos].Type == TokenCloseBlock {
		b.pos++
	}

	if !hasElse {
		exits = append(exits, current...)
	}
	return exits
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

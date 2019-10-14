package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleroot
	ruleEOT
	ruleexpression
	ruleempty_line
	rulecomment_line
	ruletitle_info
	ruletable_info
	ruletable_title
	ruletable_column
	rulecolumn_name
	rulerelation_info
	rulerelation_left
	rulecardinality_left
	rulerelation_right
	rulecardinality_right
	ruletitle_attribute
	ruletable_attribute
	rulecolumn_attribute
	rulerelation_attribute
	ruleattribute_key
	ruleattribute_value
	rulebare_value
	rulequoted_value
	ruleattribute_sep
	rulecomment_string
	rulews
	rulenewline
	rulenewline_or_eot
	rulespace
	rulestring
	rulestring_in_quote
	rulecardinality
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"root",
	"EOT",
	"expression",
	"empty_line",
	"comment_line",
	"title_info",
	"table_info",
	"table_title",
	"table_column",
	"column_name",
	"relation_info",
	"relation_left",
	"cardinality_left",
	"relation_right",
	"cardinality_right",
	"title_attribute",
	"table_attribute",
	"column_attribute",
	"relation_attribute",
	"attribute_key",
	"attribute_value",
	"bare_value",
	"quoted_value",
	"attribute_sep",
	"comment_string",
	"ws",
	"newline",
	"newline_or_eot",
	"space",
	"string",
	"string_in_quote",
	"cardinality",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type Parser struct {
	Erd

	Buffer string
	buffer []rune
	rules  [51]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Parser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Parser) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *Parser) Highlighter() {
	p.PrintSyntax()
}

func (p *Parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.Err(begin, buffer)
		case ruleAction1:
			p.Err(begin, buffer)
		case ruleAction2:
			p.ClearTableAndColumn()
		case ruleAction3:
			p.AddTable(text)
		case ruleAction4:
			p.AddColumn(text)
		case ruleAction5:
			p.AddRelation()
		case ruleAction6:
			p.SetRelationLeft(text)
		case ruleAction7:
			p.SetCardinalityLeft(text)
		case ruleAction8:
			p.SetRelationRight(text)
		case ruleAction9:
			p.SetCardinalityRight(text)
		case ruleAction10:
			p.AddTitleKeyValue()
		case ruleAction11:
			p.AddTableKeyValue()
		case ruleAction12:
			p.AddColumnKeyValue()
		case ruleAction13:
			p.AddRelationKeyValue()
		case ruleAction14:
			p.SetKey(text)
		case ruleAction15:
			p.SetValue(text)
		case ruleAction16:
			p.SetValue(text)

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Parser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 root <- <((expression EOT) / (expression <.+> Action0 EOT) / (<.+> Action1 EOT))> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[ruleexpression]() {
						goto l3
					}
					if !_rules[ruleEOT]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleexpression]() {
						goto l4
					}
					{
						position5 := position
						depth++
						if !matchDot() {
							goto l4
						}
					l6:
						{
							position7, tokenIndex7, depth7 := position, tokenIndex, depth
							if !matchDot() {
								goto l7
							}
							goto l6
						l7:
							position, tokenIndex, depth = position7, tokenIndex7, depth7
						}
						depth--
						add(rulePegText, position5)
					}
					if !_rules[ruleAction0]() {
						goto l4
					}
					if !_rules[ruleEOT]() {
						goto l4
					}
					goto l2
				l4:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position8 := position
						depth++
						if !matchDot() {
							goto l0
						}
					l9:
						{
							position10, tokenIndex10, depth10 := position, tokenIndex, depth
							if !matchDot() {
								goto l10
							}
							goto l9
						l10:
							position, tokenIndex, depth = position10, tokenIndex10, depth10
						}
						depth--
						add(rulePegText, position8)
					}
					if !_rules[ruleAction1]() {
						goto l0
					}
					if !_rules[ruleEOT]() {
						goto l0
					}
				}
			l2:
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 EOT <- <!.> */
		func() bool {
			position11, tokenIndex11, depth11 := position, tokenIndex, depth
			{
				position12 := position
				depth++
				{
					position13, tokenIndex13, depth13 := position, tokenIndex, depth
					if !matchDot() {
						goto l13
					}
					goto l11
				l13:
					position, tokenIndex, depth = position13, tokenIndex13, depth13
				}
				depth--
				add(ruleEOT, position12)
			}
			return true
		l11:
			position, tokenIndex, depth = position11, tokenIndex11, depth11
			return false
		},
		/* 2 expression <- <(title_info / relation_info / table_info / comment_line / empty_line)*> */
		func() bool {
			{
				position15 := position
				depth++
			l16:
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					{
						position18, tokenIndex18, depth18 := position, tokenIndex, depth
						if !_rules[ruletitle_info]() {
							goto l19
						}
						goto l18
					l19:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[rulerelation_info]() {
							goto l20
						}
						goto l18
					l20:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[ruletable_info]() {
							goto l21
						}
						goto l18
					l21:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[rulecomment_line]() {
							goto l22
						}
						goto l18
					l22:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[ruleempty_line]() {
							goto l17
						}
					}
				l18:
					goto l16
				l17:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
				}
				depth--
				add(ruleexpression, position15)
			}
			return true
		},
		/* 3 empty_line <- <(ws Action2)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[rulews]() {
					goto l23
				}
				if !_rules[ruleAction2]() {
					goto l23
				}
				depth--
				add(ruleempty_line, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 4 comment_line <- <(space* '#' comment_string newline)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
			l27:
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
				}
				if buffer[position] != rune('#') {
					goto l25
				}
				position++
				if !_rules[rulecomment_string]() {
					goto l25
				}
				if !_rules[rulenewline]() {
					goto l25
				}
				depth--
				add(rulecomment_line, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 5 title_info <- <('t' 'i' 't' 'l' 'e' ws* '{' ws* (title_attribute ws* attribute_sep? ws*)* ws* '}' newline)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if buffer[position] != rune('t') {
					goto l29
				}
				position++
				if buffer[position] != rune('i') {
					goto l29
				}
				position++
				if buffer[position] != rune('t') {
					goto l29
				}
				position++
				if buffer[position] != rune('l') {
					goto l29
				}
				position++
				if buffer[position] != rune('e') {
					goto l29
				}
				position++
			l31:
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l32
					}
					goto l31
				l32:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
				}
				if buffer[position] != rune('{') {
					goto l29
				}
				position++
			l33:
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l34
					}
					goto l33
				l34:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
				}
			l35:
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if !_rules[ruletitle_attribute]() {
						goto l36
					}
				l37:
					{
						position38, tokenIndex38, depth38 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l38
						}
						goto l37
					l38:
						position, tokenIndex, depth = position38, tokenIndex38, depth38
					}
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !_rules[ruleattribute_sep]() {
							goto l39
						}
						goto l40
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
				l40:
				l41:
					{
						position42, tokenIndex42, depth42 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l42
						}
						goto l41
					l42:
						position, tokenIndex, depth = position42, tokenIndex42, depth42
					}
					goto l35
				l36:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
				}
			l43:
				{
					position44, tokenIndex44, depth44 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex, depth = position44, tokenIndex44, depth44
				}
				if buffer[position] != rune('}') {
					goto l29
				}
				position++
				if !_rules[rulenewline]() {
					goto l29
				}
				depth--
				add(ruletitle_info, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 6 table_info <- <('[' table_title ']' (space* '{' ws* (table_attribute ws* attribute_sep?)* ws* '}' space*)? newline_or_eot (table_column / empty_line)*)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if buffer[position] != rune('[') {
					goto l45
				}
				position++
				if !_rules[ruletable_title]() {
					goto l45
				}
				if buffer[position] != rune(']') {
					goto l45
				}
				position++
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
				l49:
					{
						position50, tokenIndex50, depth50 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l50
						}
						goto l49
					l50:
						position, tokenIndex, depth = position50, tokenIndex50, depth50
					}
					if buffer[position] != rune('{') {
						goto l47
					}
					position++
				l51:
					{
						position52, tokenIndex52, depth52 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l52
						}
						goto l51
					l52:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
					}
				l53:
					{
						position54, tokenIndex54, depth54 := position, tokenIndex, depth
						if !_rules[ruletable_attribute]() {
							goto l54
						}
					l55:
						{
							position56, tokenIndex56, depth56 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l56
							}
							goto l55
						l56:
							position, tokenIndex, depth = position56, tokenIndex56, depth56
						}
						{
							position57, tokenIndex57, depth57 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l57
							}
							goto l58
						l57:
							position, tokenIndex, depth = position57, tokenIndex57, depth57
						}
					l58:
						goto l53
					l54:
						position, tokenIndex, depth = position54, tokenIndex54, depth54
					}
				l59:
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l60
						}
						goto l59
					l60:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
					}
					if buffer[position] != rune('}') {
						goto l47
					}
					position++
				l61:
					{
						position62, tokenIndex62, depth62 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l62
						}
						goto l61
					l62:
						position, tokenIndex, depth = position62, tokenIndex62, depth62
					}
					goto l48
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
			l48:
				if !_rules[rulenewline_or_eot]() {
					goto l45
				}
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					{
						position65, tokenIndex65, depth65 := position, tokenIndex, depth
						if !_rules[ruletable_column]() {
							goto l66
						}
						goto l65
					l66:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
						if !_rules[ruleempty_line]() {
							goto l64
						}
					}
				l65:
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				depth--
				add(ruletable_info, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 7 table_title <- <(<string> Action3)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				{
					position69 := position
					depth++
					if !_rules[rulestring]() {
						goto l67
					}
					depth--
					add(rulePegText, position69)
				}
				if !_rules[ruleAction3]() {
					goto l67
				}
				depth--
				add(ruletable_title, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 8 table_column <- <(space* column_name (space* '{' ws* (column_attribute ws* attribute_sep?)* ws* '}' space*)? newline_or_eot)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l73
					}
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				if !_rules[rulecolumn_name]() {
					goto l70
				}
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
				l76:
					{
						position77, tokenIndex77, depth77 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l77
						}
						goto l76
					l77:
						position, tokenIndex, depth = position77, tokenIndex77, depth77
					}
					if buffer[position] != rune('{') {
						goto l74
					}
					position++
				l78:
					{
						position79, tokenIndex79, depth79 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l79
						}
						goto l78
					l79:
						position, tokenIndex, depth = position79, tokenIndex79, depth79
					}
				l80:
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if !_rules[rulecolumn_attribute]() {
							goto l81
						}
					l82:
						{
							position83, tokenIndex83, depth83 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l83
							}
							goto l82
						l83:
							position, tokenIndex, depth = position83, tokenIndex83, depth83
						}
						{
							position84, tokenIndex84, depth84 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l84
							}
							goto l85
						l84:
							position, tokenIndex, depth = position84, tokenIndex84, depth84
						}
					l85:
						goto l80
					l81:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
					}
				l86:
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l87
						}
						goto l86
					l87:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
					}
					if buffer[position] != rune('}') {
						goto l74
					}
					position++
				l88:
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l89
						}
						goto l88
					l89:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
					}
					goto l75
				l74:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
				}
			l75:
				if !_rules[rulenewline_or_eot]() {
					goto l70
				}
				depth--
				add(ruletable_column, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 9 column_name <- <(<string> Action4)> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				{
					position92 := position
					depth++
					if !_rules[rulestring]() {
						goto l90
					}
					depth--
					add(rulePegText, position92)
				}
				if !_rules[ruleAction4]() {
					goto l90
				}
				depth--
				add(rulecolumn_name, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 10 relation_info <- <(space* relation_left space* cardinality_left ('-' '-') cardinality_right space* relation_right (ws* '{' ws* (relation_attribute ws* attribute_sep? ws*)* ws* '}')? newline_or_eot Action5)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
			l95:
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				if !_rules[rulerelation_left]() {
					goto l93
				}
			l97:
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l98
					}
					goto l97
				l98:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
				}
				if !_rules[rulecardinality_left]() {
					goto l93
				}
				if buffer[position] != rune('-') {
					goto l93
				}
				position++
				if buffer[position] != rune('-') {
					goto l93
				}
				position++
				if !_rules[rulecardinality_right]() {
					goto l93
				}
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l100
					}
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				if !_rules[rulerelation_right]() {
					goto l93
				}
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
				l103:
					{
						position104, tokenIndex104, depth104 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l104
						}
						goto l103
					l104:
						position, tokenIndex, depth = position104, tokenIndex104, depth104
					}
					if buffer[position] != rune('{') {
						goto l101
					}
					position++
				l105:
					{
						position106, tokenIndex106, depth106 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l106
						}
						goto l105
					l106:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
					}
				l107:
					{
						position108, tokenIndex108, depth108 := position, tokenIndex, depth
						if !_rules[rulerelation_attribute]() {
							goto l108
						}
					l109:
						{
							position110, tokenIndex110, depth110 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l110
							}
							goto l109
						l110:
							position, tokenIndex, depth = position110, tokenIndex110, depth110
						}
						{
							position111, tokenIndex111, depth111 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l111
							}
							goto l112
						l111:
							position, tokenIndex, depth = position111, tokenIndex111, depth111
						}
					l112:
					l113:
						{
							position114, tokenIndex114, depth114 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l114
							}
							goto l113
						l114:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
						}
						goto l107
					l108:
						position, tokenIndex, depth = position108, tokenIndex108, depth108
					}
				l115:
					{
						position116, tokenIndex116, depth116 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l116
						}
						goto l115
					l116:
						position, tokenIndex, depth = position116, tokenIndex116, depth116
					}
					if buffer[position] != rune('}') {
						goto l101
					}
					position++
					goto l102
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
			l102:
				if !_rules[rulenewline_or_eot]() {
					goto l93
				}
				if !_rules[ruleAction5]() {
					goto l93
				}
				depth--
				add(rulerelation_info, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 11 relation_left <- <(<string> Action6)> */
		func() bool {
			position117, tokenIndex117, depth117 := position, tokenIndex, depth
			{
				position118 := position
				depth++
				{
					position119 := position
					depth++
					if !_rules[rulestring]() {
						goto l117
					}
					depth--
					add(rulePegText, position119)
				}
				if !_rules[ruleAction6]() {
					goto l117
				}
				depth--
				add(rulerelation_left, position118)
			}
			return true
		l117:
			position, tokenIndex, depth = position117, tokenIndex117, depth117
			return false
		},
		/* 12 cardinality_left <- <(<cardinality> Action7)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				{
					position122 := position
					depth++
					if !_rules[rulecardinality]() {
						goto l120
					}
					depth--
					add(rulePegText, position122)
				}
				if !_rules[ruleAction7]() {
					goto l120
				}
				depth--
				add(rulecardinality_left, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 13 relation_right <- <(<string> Action8)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				{
					position125 := position
					depth++
					if !_rules[rulestring]() {
						goto l123
					}
					depth--
					add(rulePegText, position125)
				}
				if !_rules[ruleAction8]() {
					goto l123
				}
				depth--
				add(rulerelation_right, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 14 cardinality_right <- <(<cardinality> Action9)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				{
					position128 := position
					depth++
					if !_rules[rulecardinality]() {
						goto l126
					}
					depth--
					add(rulePegText, position128)
				}
				if !_rules[ruleAction9]() {
					goto l126
				}
				depth--
				add(rulecardinality_right, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 15 title_attribute <- <(attribute_key space* ':' space* attribute_value Action10)> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l129
				}
			l131:
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l132
					}
					goto l131
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				if buffer[position] != rune(':') {
					goto l129
				}
				position++
			l133:
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l134
					}
					goto l133
				l134:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
				}
				if !_rules[ruleattribute_value]() {
					goto l129
				}
				if !_rules[ruleAction10]() {
					goto l129
				}
				depth--
				add(ruletitle_attribute, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 16 table_attribute <- <(attribute_key space* ':' space* attribute_value Action11)> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l135
				}
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l138
					}
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				if buffer[position] != rune(':') {
					goto l135
				}
				position++
			l139:
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l140
					}
					goto l139
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
				if !_rules[ruleattribute_value]() {
					goto l135
				}
				if !_rules[ruleAction11]() {
					goto l135
				}
				depth--
				add(ruletable_attribute, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 17 column_attribute <- <(attribute_key space* ':' space* attribute_value Action12)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l141
				}
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				if buffer[position] != rune(':') {
					goto l141
				}
				position++
			l145:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
				if !_rules[ruleattribute_value]() {
					goto l141
				}
				if !_rules[ruleAction12]() {
					goto l141
				}
				depth--
				add(rulecolumn_attribute, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 18 relation_attribute <- <(attribute_key space* ':' space* attribute_value Action13)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l147
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				if buffer[position] != rune(':') {
					goto l147
				}
				position++
			l151:
				{
					position152, tokenIndex152, depth152 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l152
					}
					goto l151
				l152:
					position, tokenIndex, depth = position152, tokenIndex152, depth152
				}
				if !_rules[ruleattribute_value]() {
					goto l147
				}
				if !_rules[ruleAction13]() {
					goto l147
				}
				depth--
				add(rulerelation_attribute, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 19 attribute_key <- <(<string> Action14)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				{
					position155 := position
					depth++
					if !_rules[rulestring]() {
						goto l153
					}
					depth--
					add(rulePegText, position155)
				}
				if !_rules[ruleAction14]() {
					goto l153
				}
				depth--
				add(ruleattribute_key, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 20 attribute_value <- <(bare_value / quoted_value)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if !_rules[rulebare_value]() {
						goto l159
					}
					goto l158
				l159:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					if !_rules[rulequoted_value]() {
						goto l156
					}
				}
			l158:
				depth--
				add(ruleattribute_value, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 21 bare_value <- <(<string> Action15)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					position162 := position
					depth++
					if !_rules[rulestring]() {
						goto l160
					}
					depth--
					add(rulePegText, position162)
				}
				if !_rules[ruleAction15]() {
					goto l160
				}
				depth--
				add(rulebare_value, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 22 quoted_value <- <(<('"' string_in_quote '"')> Action16)> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				{
					position165 := position
					depth++
					if buffer[position] != rune('"') {
						goto l163
					}
					position++
					if !_rules[rulestring_in_quote]() {
						goto l163
					}
					if buffer[position] != rune('"') {
						goto l163
					}
					position++
					depth--
					add(rulePegText, position165)
				}
				if !_rules[ruleAction16]() {
					goto l163
				}
				depth--
				add(rulequoted_value, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 23 attribute_sep <- <(space* ',' space*)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
			l168:
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l169
					}
					goto l168
				l169:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
				}
				if buffer[position] != rune(',') {
					goto l166
				}
				position++
			l170:
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l171
					}
					goto l170
				l171:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
				}
				depth--
				add(ruleattribute_sep, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 24 comment_string <- <(!('\r' / '\n') .)*> */
		func() bool {
			{
				position173 := position
				depth++
			l174:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						{
							position177, tokenIndex177, depth177 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l178
							}
							position++
							goto l177
						l178:
							position, tokenIndex, depth = position177, tokenIndex177, depth177
							if buffer[position] != rune('\n') {
								goto l176
							}
							position++
						}
					l177:
						goto l175
					l176:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
					}
					if !matchDot() {
						goto l175
					}
					goto l174
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
				depth--
				add(rulecomment_string, position173)
			}
			return true
		},
		/* 25 ws <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l184
					}
					position++
					goto l183
				l184:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if buffer[position] != rune('\t') {
						goto l185
					}
					position++
					goto l183
				l185:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if buffer[position] != rune('\r') {
						goto l186
					}
					position++
					goto l183
				l186:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if buffer[position] != rune('\n') {
						goto l179
					}
					position++
				}
			l183:
			l181:
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					{
						position187, tokenIndex187, depth187 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l188
						}
						position++
						goto l187
					l188:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('\t') {
							goto l189
						}
						position++
						goto l187
					l189:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('\r') {
							goto l190
						}
						position++
						goto l187
					l190:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('\n') {
							goto l182
						}
						position++
					}
				l187:
					goto l181
				l182:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
				}
				depth--
				add(rulews, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 26 newline <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l194
					}
					position++
					if buffer[position] != rune('\n') {
						goto l194
					}
					position++
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('\n') {
						goto l195
					}
					position++
					goto l193
				l195:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('\r') {
						goto l191
					}
					position++
				}
			l193:
				depth--
				add(rulenewline, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 27 newline_or_eot <- <(newline / EOT)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if !_rules[rulenewline]() {
						goto l199
					}
					goto l198
				l199:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
					if !_rules[ruleEOT]() {
						goto l196
					}
				}
			l198:
				depth--
				add(rulenewline_or_eot, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 28 space <- <(' ' / '\t')+> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l205
					}
					position++
					goto l204
				l205:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
					if buffer[position] != rune('\t') {
						goto l200
					}
					position++
				}
			l204:
			l202:
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					{
						position206, tokenIndex206, depth206 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l207
						}
						position++
						goto l206
					l207:
						position, tokenIndex, depth = position206, tokenIndex206, depth206
						if buffer[position] != rune('\t') {
							goto l203
						}
						position++
					}
				l206:
					goto l202
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
				depth--
				add(rulespace, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 29 string <- <(!('"' / '\t' / '\r' / '\n' / '/' / ':' / ',' / '[' / ']' / '{' / '}' / ' ') .)+> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					{
						position213, tokenIndex213, depth213 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l214
						}
						position++
						goto l213
					l214:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('\t') {
							goto l215
						}
						position++
						goto l213
					l215:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('\r') {
							goto l216
						}
						position++
						goto l213
					l216:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('\n') {
							goto l217
						}
						position++
						goto l213
					l217:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('/') {
							goto l218
						}
						position++
						goto l213
					l218:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune(':') {
							goto l219
						}
						position++
						goto l213
					l219:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune(',') {
							goto l220
						}
						position++
						goto l213
					l220:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('[') {
							goto l221
						}
						position++
						goto l213
					l221:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune(']') {
							goto l222
						}
						position++
						goto l213
					l222:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('{') {
							goto l223
						}
						position++
						goto l213
					l223:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('}') {
							goto l224
						}
						position++
						goto l213
					l224:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune(' ') {
							goto l212
						}
						position++
					}
				l213:
					goto l208
				l212:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
				}
				if !matchDot() {
					goto l208
				}
			l210:
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						{
							position226, tokenIndex226, depth226 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l227
							}
							position++
							goto l226
						l227:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('\t') {
								goto l228
							}
							position++
							goto l226
						l228:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('\r') {
								goto l229
							}
							position++
							goto l226
						l229:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('\n') {
								goto l230
							}
							position++
							goto l226
						l230:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('/') {
								goto l231
							}
							position++
							goto l226
						l231:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune(':') {
								goto l232
							}
							position++
							goto l226
						l232:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune(',') {
								goto l233
							}
							position++
							goto l226
						l233:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('[') {
								goto l234
							}
							position++
							goto l226
						l234:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune(']') {
								goto l235
							}
							position++
							goto l226
						l235:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('{') {
								goto l236
							}
							position++
							goto l226
						l236:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune('}') {
								goto l237
							}
							position++
							goto l226
						l237:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
							if buffer[position] != rune(' ') {
								goto l225
							}
							position++
						}
					l226:
						goto l211
					l225:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
					}
					if !matchDot() {
						goto l211
					}
					goto l210
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
				depth--
				add(rulestring, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 30 string_in_quote <- <(!('"' / '\t' / '\r' / '\n') .)+> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					{
						position243, tokenIndex243, depth243 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l244
						}
						position++
						goto l243
					l244:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if buffer[position] != rune('\t') {
							goto l245
						}
						position++
						goto l243
					l245:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if buffer[position] != rune('\r') {
							goto l246
						}
						position++
						goto l243
					l246:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if buffer[position] != rune('\n') {
							goto l242
						}
						position++
					}
				l243:
					goto l238
				l242:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
				}
				if !matchDot() {
					goto l238
				}
			l240:
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					{
						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						{
							position248, tokenIndex248, depth248 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l249
							}
							position++
							goto l248
						l249:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if buffer[position] != rune('\t') {
								goto l250
							}
							position++
							goto l248
						l250:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if buffer[position] != rune('\r') {
								goto l251
							}
							position++
							goto l248
						l251:
							position, tokenIndex, depth = position248, tokenIndex248, depth248
							if buffer[position] != rune('\n') {
								goto l247
							}
							position++
						}
					l248:
						goto l241
					l247:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
					}
					if !matchDot() {
						goto l241
					}
					goto l240
				l241:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
				}
				depth--
				add(rulestring_in_quote, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 31 cardinality <- <('0' / '1' / '*' / '+')> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				{
					position254, tokenIndex254, depth254 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l255
					}
					position++
					goto l254
				l255:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
					if buffer[position] != rune('1') {
						goto l256
					}
					position++
					goto l254
				l256:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
					if buffer[position] != rune('*') {
						goto l257
					}
					position++
					goto l254
				l257:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
					if buffer[position] != rune('+') {
						goto l252
					}
					position++
				}
			l254:
				depth--
				add(rulecardinality, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		nil,
		/* 34 Action0 <- <{p.Err(begin, buffer)}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 35 Action1 <- <{p.Err(begin, buffer)}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 36 Action2 <- <{ p.ClearTableAndColumn() }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 37 Action3 <- <{ p.AddTable(text) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 38 Action4 <- <{ p.AddColumn(text) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 39 Action5 <- <{ p.AddRelation() }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 40 Action6 <- <{ p.SetRelationLeft(text) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 41 Action7 <- <{ p.SetCardinalityLeft(text)}> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 42 Action8 <- <{ p.SetRelationRight(text) }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 43 Action9 <- <{ p.SetCardinalityRight(text)}> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 44 Action10 <- <{ p.AddTitleKeyValue() }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 45 Action11 <- <{ p.AddTableKeyValue() }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 46 Action12 <- <{ p.AddColumnKeyValue() }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 47 Action13 <- <{ p.AddRelationKeyValue() }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 48 Action14 <- <{ p.SetKey(text) }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 49 Action15 <- <{ p.SetValue(text) }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 50 Action16 <- <{ p.SetValue(text) }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
	}
	p.rules = _rules
}

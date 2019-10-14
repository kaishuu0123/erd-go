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
	rulecolor_info
	rulecolor_key_value
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
	ruleAction17

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
	"color_info",
	"color_key_value",
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
	"Action17",

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
	rules  [54]func() bool
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
			p.AddColorDefine()
		case ruleAction4:
			p.AddTable(text)
		case ruleAction5:
			p.AddColumn(text)
		case ruleAction6:
			p.AddRelation()
		case ruleAction7:
			p.SetRelationLeft(text)
		case ruleAction8:
			p.SetCardinalityLeft(text)
		case ruleAction9:
			p.SetRelationRight(text)
		case ruleAction10:
			p.SetCardinalityRight(text)
		case ruleAction11:
			p.AddTitleKeyValue()
		case ruleAction12:
			p.AddTableKeyValue()
		case ruleAction13:
			p.AddColumnKeyValue()
		case ruleAction14:
			p.AddRelationKeyValue()
		case ruleAction15:
			p.SetKey(text)
		case ruleAction16:
			p.SetValue(text)
		case ruleAction17:
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
		/* 2 expression <- <(title_info / color_info / relation_info / table_info / comment_line / empty_line)*> */
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
						if !_rules[rulecolor_info]() {
							goto l20
						}
						goto l18
					l20:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[rulerelation_info]() {
							goto l21
						}
						goto l18
					l21:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[ruletable_info]() {
							goto l22
						}
						goto l18
					l22:
						position, tokenIndex, depth = position18, tokenIndex18, depth18
						if !_rules[rulecomment_line]() {
							goto l23
						}
						goto l18
					l23:
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
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[rulews]() {
					goto l24
				}
				if !_rules[ruleAction2]() {
					goto l24
				}
				depth--
				add(ruleempty_line, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 4 comment_line <- <(space* '#' comment_string newline)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
			l28:
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
				if buffer[position] != rune('#') {
					goto l26
				}
				position++
				if !_rules[rulecomment_string]() {
					goto l26
				}
				if !_rules[rulenewline]() {
					goto l26
				}
				depth--
				add(rulecomment_line, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 5 color_info <- <('c' 'o' 'l' 'o' 'r' 's' ws* '{' ws* (color_key_value ws* attribute_sep? ws*)* ws* '}' newline)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if buffer[position] != rune('c') {
					goto l30
				}
				position++
				if buffer[position] != rune('o') {
					goto l30
				}
				position++
				if buffer[position] != rune('l') {
					goto l30
				}
				position++
				if buffer[position] != rune('o') {
					goto l30
				}
				position++
				if buffer[position] != rune('r') {
					goto l30
				}
				position++
				if buffer[position] != rune('s') {
					goto l30
				}
				position++
			l32:
				{
					position33, tokenIndex33, depth33 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
				}
				if buffer[position] != rune('{') {
					goto l30
				}
				position++
			l34:
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l35
					}
					goto l34
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
			l36:
				{
					position37, tokenIndex37, depth37 := position, tokenIndex, depth
					if !_rules[rulecolor_key_value]() {
						goto l37
					}
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l39
						}
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					{
						position40, tokenIndex40, depth40 := position, tokenIndex, depth
						if !_rules[ruleattribute_sep]() {
							goto l40
						}
						goto l41
					l40:
						position, tokenIndex, depth = position40, tokenIndex40, depth40
					}
				l41:
				l42:
					{
						position43, tokenIndex43, depth43 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l43
						}
						goto l42
					l43:
						position, tokenIndex, depth = position43, tokenIndex43, depth43
					}
					goto l36
				l37:
					position, tokenIndex, depth = position37, tokenIndex37, depth37
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l45
					}
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				if buffer[position] != rune('}') {
					goto l30
				}
				position++
				if !_rules[rulenewline]() {
					goto l30
				}
				depth--
				add(rulecolor_info, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 6 color_key_value <- <(attribute_key space* ':' space* attribute_value Action3)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l46
				}
			l48:
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l49
					}
					goto l48
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
				if buffer[position] != rune(':') {
					goto l46
				}
				position++
			l50:
				{
					position51, tokenIndex51, depth51 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l51
					}
					goto l50
				l51:
					position, tokenIndex, depth = position51, tokenIndex51, depth51
				}
				if !_rules[ruleattribute_value]() {
					goto l46
				}
				if !_rules[ruleAction3]() {
					goto l46
				}
				depth--
				add(rulecolor_key_value, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 7 title_info <- <('t' 'i' 't' 'l' 'e' ws* '{' ws* (title_attribute ws* attribute_sep? ws*)* ws* '}' newline)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if buffer[position] != rune('t') {
					goto l52
				}
				position++
				if buffer[position] != rune('i') {
					goto l52
				}
				position++
				if buffer[position] != rune('t') {
					goto l52
				}
				position++
				if buffer[position] != rune('l') {
					goto l52
				}
				position++
				if buffer[position] != rune('e') {
					goto l52
				}
				position++
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				if buffer[position] != rune('{') {
					goto l52
				}
				position++
			l56:
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
				}
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					if !_rules[ruletitle_attribute]() {
						goto l59
					}
				l60:
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l61
						}
						goto l60
					l61:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
					}
					{
						position62, tokenIndex62, depth62 := position, tokenIndex, depth
						if !_rules[ruleattribute_sep]() {
							goto l62
						}
						goto l63
					l62:
						position, tokenIndex, depth = position62, tokenIndex62, depth62
					}
				l63:
				l64:
					{
						position65, tokenIndex65, depth65 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l65
						}
						goto l64
					l65:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
					}
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if !_rules[rulews]() {
						goto l67
					}
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				if buffer[position] != rune('}') {
					goto l52
				}
				position++
				if !_rules[rulenewline]() {
					goto l52
				}
				depth--
				add(ruletitle_info, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 8 table_info <- <('[' table_title ']' (space* '{' ws* (table_attribute ws* attribute_sep?)* ws* '}' space*)? newline_or_eot (table_column / empty_line)*)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if buffer[position] != rune('[') {
					goto l68
				}
				position++
				if !_rules[ruletable_title]() {
					goto l68
				}
				if buffer[position] != rune(']') {
					goto l68
				}
				position++
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
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
					if buffer[position] != rune('{') {
						goto l70
					}
					position++
				l74:
					{
						position75, tokenIndex75, depth75 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l75
						}
						goto l74
					l75:
						position, tokenIndex, depth = position75, tokenIndex75, depth75
					}
				l76:
					{
						position77, tokenIndex77, depth77 := position, tokenIndex, depth
						if !_rules[ruletable_attribute]() {
							goto l77
						}
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
						{
							position80, tokenIndex80, depth80 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l80
							}
							goto l81
						l80:
							position, tokenIndex, depth = position80, tokenIndex80, depth80
						}
					l81:
						goto l76
					l77:
						position, tokenIndex, depth = position77, tokenIndex77, depth77
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
					if buffer[position] != rune('}') {
						goto l70
					}
					position++
				l84:
					{
						position85, tokenIndex85, depth85 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l85
						}
						goto l84
					l85:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
					}
					goto l71
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
			l71:
				if !_rules[rulenewline_or_eot]() {
					goto l68
				}
			l86:
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					{
						position88, tokenIndex88, depth88 := position, tokenIndex, depth
						if !_rules[ruletable_column]() {
							goto l89
						}
						goto l88
					l89:
						position, tokenIndex, depth = position88, tokenIndex88, depth88
						if !_rules[ruleempty_line]() {
							goto l87
						}
					}
				l88:
					goto l86
				l87:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
				}
				depth--
				add(ruletable_info, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 9 table_title <- <(<string> Action4)> */
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
				add(ruletable_title, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 10 table_column <- <(space* column_name (space* '{' ws* (column_attribute ws* attribute_sep?)* ws* '}' space*)? newline_or_eot)> */
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
				if !_rules[rulecolumn_name]() {
					goto l93
				}
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
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
					if buffer[position] != rune('{') {
						goto l97
					}
					position++
				l101:
					{
						position102, tokenIndex102, depth102 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l102
						}
						goto l101
					l102:
						position, tokenIndex, depth = position102, tokenIndex102, depth102
					}
				l103:
					{
						position104, tokenIndex104, depth104 := position, tokenIndex, depth
						if !_rules[rulecolumn_attribute]() {
							goto l104
						}
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
						{
							position107, tokenIndex107, depth107 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l107
							}
							goto l108
						l107:
							position, tokenIndex, depth = position107, tokenIndex107, depth107
						}
					l108:
						goto l103
					l104:
						position, tokenIndex, depth = position104, tokenIndex104, depth104
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
					if buffer[position] != rune('}') {
						goto l97
					}
					position++
				l111:
					{
						position112, tokenIndex112, depth112 := position, tokenIndex, depth
						if !_rules[rulespace]() {
							goto l112
						}
						goto l111
					l112:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
					}
					goto l98
				l97:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
				}
			l98:
				if !_rules[rulenewline_or_eot]() {
					goto l93
				}
				depth--
				add(ruletable_column, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 11 column_name <- <(<string> Action5)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115 := position
					depth++
					if !_rules[rulestring]() {
						goto l113
					}
					depth--
					add(rulePegText, position115)
				}
				if !_rules[ruleAction5]() {
					goto l113
				}
				depth--
				add(rulecolumn_name, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 12 relation_info <- <(space* relation_left space* cardinality_left ('-' '-') cardinality_right space* relation_right (ws* '{' ws* (relation_attribute ws* attribute_sep? ws*)* ws* '}')? newline_or_eot Action6)> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
			l118:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l119
					}
					goto l118
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
				if !_rules[rulerelation_left]() {
					goto l116
				}
			l120:
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l121
					}
					goto l120
				l121:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
				}
				if !_rules[rulecardinality_left]() {
					goto l116
				}
				if buffer[position] != rune('-') {
					goto l116
				}
				position++
				if buffer[position] != rune('-') {
					goto l116
				}
				position++
				if !_rules[rulecardinality_right]() {
					goto l116
				}
			l122:
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l123
					}
					goto l122
				l123:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
				}
				if !_rules[rulerelation_right]() {
					goto l116
				}
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
				l126:
					{
						position127, tokenIndex127, depth127 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l127
						}
						goto l126
					l127:
						position, tokenIndex, depth = position127, tokenIndex127, depth127
					}
					if buffer[position] != rune('{') {
						goto l124
					}
					position++
				l128:
					{
						position129, tokenIndex129, depth129 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l129
						}
						goto l128
					l129:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
					}
				l130:
					{
						position131, tokenIndex131, depth131 := position, tokenIndex, depth
						if !_rules[rulerelation_attribute]() {
							goto l131
						}
					l132:
						{
							position133, tokenIndex133, depth133 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l133
							}
							goto l132
						l133:
							position, tokenIndex, depth = position133, tokenIndex133, depth133
						}
						{
							position134, tokenIndex134, depth134 := position, tokenIndex, depth
							if !_rules[ruleattribute_sep]() {
								goto l134
							}
							goto l135
						l134:
							position, tokenIndex, depth = position134, tokenIndex134, depth134
						}
					l135:
					l136:
						{
							position137, tokenIndex137, depth137 := position, tokenIndex, depth
							if !_rules[rulews]() {
								goto l137
							}
							goto l136
						l137:
							position, tokenIndex, depth = position137, tokenIndex137, depth137
						}
						goto l130
					l131:
						position, tokenIndex, depth = position131, tokenIndex131, depth131
					}
				l138:
					{
						position139, tokenIndex139, depth139 := position, tokenIndex, depth
						if !_rules[rulews]() {
							goto l139
						}
						goto l138
					l139:
						position, tokenIndex, depth = position139, tokenIndex139, depth139
					}
					if buffer[position] != rune('}') {
						goto l124
					}
					position++
					goto l125
				l124:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
				}
			l125:
				if !_rules[rulenewline_or_eot]() {
					goto l116
				}
				if !_rules[ruleAction6]() {
					goto l116
				}
				depth--
				add(rulerelation_info, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 13 relation_left <- <(<string> Action7)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				{
					position142 := position
					depth++
					if !_rules[rulestring]() {
						goto l140
					}
					depth--
					add(rulePegText, position142)
				}
				if !_rules[ruleAction7]() {
					goto l140
				}
				depth--
				add(rulerelation_left, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 14 cardinality_left <- <(<cardinality> Action8)> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				{
					position145 := position
					depth++
					if !_rules[rulecardinality]() {
						goto l143
					}
					depth--
					add(rulePegText, position145)
				}
				if !_rules[ruleAction8]() {
					goto l143
				}
				depth--
				add(rulecardinality_left, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 15 relation_right <- <(<string> Action9)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				{
					position148 := position
					depth++
					if !_rules[rulestring]() {
						goto l146
					}
					depth--
					add(rulePegText, position148)
				}
				if !_rules[ruleAction9]() {
					goto l146
				}
				depth--
				add(rulerelation_right, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 16 cardinality_right <- <(<cardinality> Action10)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				{
					position151 := position
					depth++
					if !_rules[rulecardinality]() {
						goto l149
					}
					depth--
					add(rulePegText, position151)
				}
				if !_rules[ruleAction10]() {
					goto l149
				}
				depth--
				add(rulecardinality_right, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 17 title_attribute <- <(attribute_key space* ':' space* attribute_value Action11)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l152
				}
			l154:
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l155
					}
					goto l154
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
				if buffer[position] != rune(':') {
					goto l152
				}
				position++
			l156:
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l157
					}
					goto l156
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				if !_rules[ruleattribute_value]() {
					goto l152
				}
				if !_rules[ruleAction11]() {
					goto l152
				}
				depth--
				add(ruletitle_attribute, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 18 table_attribute <- <(attribute_key space* ':' space* attribute_value Action12)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l158
				}
			l160:
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l161
					}
					goto l160
				l161:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
				}
				if buffer[position] != rune(':') {
					goto l158
				}
				position++
			l162:
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l163
					}
					goto l162
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
				if !_rules[ruleattribute_value]() {
					goto l158
				}
				if !_rules[ruleAction12]() {
					goto l158
				}
				depth--
				add(ruletable_attribute, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 19 column_attribute <- <(attribute_key space* ':' space* attribute_value Action13)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l164
				}
			l166:
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l167
					}
					goto l166
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				if buffer[position] != rune(':') {
					goto l164
				}
				position++
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
				if !_rules[ruleattribute_value]() {
					goto l164
				}
				if !_rules[ruleAction13]() {
					goto l164
				}
				depth--
				add(rulecolumn_attribute, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 20 relation_attribute <- <(attribute_key space* ':' space* attribute_value Action14)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if !_rules[ruleattribute_key]() {
					goto l170
				}
			l172:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l173
					}
					goto l172
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
				if buffer[position] != rune(':') {
					goto l170
				}
				position++
			l174:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l175
					}
					goto l174
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
				if !_rules[ruleattribute_value]() {
					goto l170
				}
				if !_rules[ruleAction14]() {
					goto l170
				}
				depth--
				add(rulerelation_attribute, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 21 attribute_key <- <(<string> Action15)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178 := position
					depth++
					if !_rules[rulestring]() {
						goto l176
					}
					depth--
					add(rulePegText, position178)
				}
				if !_rules[ruleAction15]() {
					goto l176
				}
				depth--
				add(ruleattribute_key, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 22 attribute_value <- <(bare_value / quoted_value)> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if !_rules[rulebare_value]() {
						goto l182
					}
					goto l181
				l182:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
					if !_rules[rulequoted_value]() {
						goto l179
					}
				}
			l181:
				depth--
				add(ruleattribute_value, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 23 bare_value <- <(<string> Action16)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				{
					position185 := position
					depth++
					if !_rules[rulestring]() {
						goto l183
					}
					depth--
					add(rulePegText, position185)
				}
				if !_rules[ruleAction16]() {
					goto l183
				}
				depth--
				add(rulebare_value, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 24 quoted_value <- <(<('"' string_in_quote '"')> Action17)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				{
					position188 := position
					depth++
					if buffer[position] != rune('"') {
						goto l186
					}
					position++
					if !_rules[rulestring_in_quote]() {
						goto l186
					}
					if buffer[position] != rune('"') {
						goto l186
					}
					position++
					depth--
					add(rulePegText, position188)
				}
				if !_rules[ruleAction17]() {
					goto l186
				}
				depth--
				add(rulequoted_value, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 25 attribute_sep <- <(space* ',' space*)> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
			l191:
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l192
					}
					goto l191
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
				if buffer[position] != rune(',') {
					goto l189
				}
				position++
			l193:
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if !_rules[rulespace]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
				}
				depth--
				add(ruleattribute_sep, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 26 comment_string <- <(!('\r' / '\n') .)*> */
		func() bool {
			{
				position196 := position
				depth++
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					{
						position199, tokenIndex199, depth199 := position, tokenIndex, depth
						{
							position200, tokenIndex200, depth200 := position, tokenIndex, depth
							if buffer[position] != rune('\r') {
								goto l201
							}
							position++
							goto l200
						l201:
							position, tokenIndex, depth = position200, tokenIndex200, depth200
							if buffer[position] != rune('\n') {
								goto l199
							}
							position++
						}
					l200:
						goto l198
					l199:
						position, tokenIndex, depth = position199, tokenIndex199, depth199
					}
					if !matchDot() {
						goto l198
					}
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				depth--
				add(rulecomment_string, position196)
			}
			return true
		},
		/* 27 ws <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
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
						goto l208
					}
					position++
					goto l206
				l208:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if buffer[position] != rune('\r') {
						goto l209
					}
					position++
					goto l206
				l209:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if buffer[position] != rune('\n') {
						goto l202
					}
					position++
				}
			l206:
			l204:
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					{
						position210, tokenIndex210, depth210 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l211
						}
						position++
						goto l210
					l211:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('\t') {
							goto l212
						}
						position++
						goto l210
					l212:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('\r') {
							goto l213
						}
						position++
						goto l210
					l213:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('\n') {
							goto l205
						}
						position++
					}
				l210:
					goto l204
				l205:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
				}
				depth--
				add(rulews, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 28 newline <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l217
					}
					position++
					if buffer[position] != rune('\n') {
						goto l217
					}
					position++
					goto l216
				l217:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
					if buffer[position] != rune('\n') {
						goto l218
					}
					position++
					goto l216
				l218:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
					if buffer[position] != rune('\r') {
						goto l214
					}
					position++
				}
			l216:
				depth--
				add(rulenewline, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 29 newline_or_eot <- <(newline / EOT)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if !_rules[rulenewline]() {
						goto l222
					}
					goto l221
				l222:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if !_rules[ruleEOT]() {
						goto l219
					}
				}
			l221:
				depth--
				add(rulenewline_or_eot, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 30 space <- <(' ' / '\t')+> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l228
					}
					position++
					goto l227
				l228:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
					if buffer[position] != rune('\t') {
						goto l223
					}
					position++
				}
			l227:
			l225:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position229, tokenIndex229, depth229 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l230
						}
						position++
						goto l229
					l230:
						position, tokenIndex, depth = position229, tokenIndex229, depth229
						if buffer[position] != rune('\t') {
							goto l226
						}
						position++
					}
				l229:
					goto l225
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
				depth--
				add(rulespace, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 31 string <- <(!('"' / '\t' / '\r' / '\n' / '/' / ':' / ',' / '[' / ']' / '{' / '}' / ' ') .)+> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l237
						}
						position++
						goto l236
					l237:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('\t') {
							goto l238
						}
						position++
						goto l236
					l238:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('\r') {
							goto l239
						}
						position++
						goto l236
					l239:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('\n') {
							goto l240
						}
						position++
						goto l236
					l240:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('/') {
							goto l241
						}
						position++
						goto l236
					l241:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune(':') {
							goto l242
						}
						position++
						goto l236
					l242:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune(',') {
							goto l243
						}
						position++
						goto l236
					l243:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('[') {
							goto l244
						}
						position++
						goto l236
					l244:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune(']') {
							goto l245
						}
						position++
						goto l236
					l245:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('{') {
							goto l246
						}
						position++
						goto l236
					l246:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('}') {
							goto l247
						}
						position++
						goto l236
					l247:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune(' ') {
							goto l235
						}
						position++
					}
				l236:
					goto l231
				l235:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
				}
				if !matchDot() {
					goto l231
				}
			l233:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					{
						position248, tokenIndex248, depth248 := position, tokenIndex, depth
						{
							position249, tokenIndex249, depth249 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l250
							}
							position++
							goto l249
						l250:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('\t') {
								goto l251
							}
							position++
							goto l249
						l251:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('\r') {
								goto l252
							}
							position++
							goto l249
						l252:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('\n') {
								goto l253
							}
							position++
							goto l249
						l253:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('/') {
								goto l254
							}
							position++
							goto l249
						l254:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune(':') {
								goto l255
							}
							position++
							goto l249
						l255:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune(',') {
								goto l256
							}
							position++
							goto l249
						l256:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('[') {
								goto l257
							}
							position++
							goto l249
						l257:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune(']') {
								goto l258
							}
							position++
							goto l249
						l258:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('{') {
								goto l259
							}
							position++
							goto l249
						l259:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('}') {
								goto l260
							}
							position++
							goto l249
						l260:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune(' ') {
								goto l248
							}
							position++
						}
					l249:
						goto l234
					l248:
						position, tokenIndex, depth = position248, tokenIndex248, depth248
					}
					if !matchDot() {
						goto l234
					}
					goto l233
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				depth--
				add(rulestring, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 32 string_in_quote <- <(!('"' / '\t' / '\r' / '\n') .)+> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				{
					position265, tokenIndex265, depth265 := position, tokenIndex, depth
					{
						position266, tokenIndex266, depth266 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l267
						}
						position++
						goto l266
					l267:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if buffer[position] != rune('\t') {
							goto l268
						}
						position++
						goto l266
					l268:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if buffer[position] != rune('\r') {
							goto l269
						}
						position++
						goto l266
					l269:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if buffer[position] != rune('\n') {
							goto l265
						}
						position++
					}
				l266:
					goto l261
				l265:
					position, tokenIndex, depth = position265, tokenIndex265, depth265
				}
				if !matchDot() {
					goto l261
				}
			l263:
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					{
						position270, tokenIndex270, depth270 := position, tokenIndex, depth
						{
							position271, tokenIndex271, depth271 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l272
							}
							position++
							goto l271
						l272:
							position, tokenIndex, depth = position271, tokenIndex271, depth271
							if buffer[position] != rune('\t') {
								goto l273
							}
							position++
							goto l271
						l273:
							position, tokenIndex, depth = position271, tokenIndex271, depth271
							if buffer[position] != rune('\r') {
								goto l274
							}
							position++
							goto l271
						l274:
							position, tokenIndex, depth = position271, tokenIndex271, depth271
							if buffer[position] != rune('\n') {
								goto l270
							}
							position++
						}
					l271:
						goto l264
					l270:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
					}
					if !matchDot() {
						goto l264
					}
					goto l263
				l264:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
				}
				depth--
				add(rulestring_in_quote, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 33 cardinality <- <('0' / '1' / '*' / '+')> */
		func() bool {
			position275, tokenIndex275, depth275 := position, tokenIndex, depth
			{
				position276 := position
				depth++
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l278
					}
					position++
					goto l277
				l278:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('1') {
						goto l279
					}
					position++
					goto l277
				l279:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('*') {
						goto l280
					}
					position++
					goto l277
				l280:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('+') {
						goto l275
					}
					position++
				}
			l277:
				depth--
				add(rulecardinality, position276)
			}
			return true
		l275:
			position, tokenIndex, depth = position275, tokenIndex275, depth275
			return false
		},
		nil,
		/* 36 Action0 <- <{p.Err(begin, buffer)}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 37 Action1 <- <{p.Err(begin, buffer)}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 38 Action2 <- <{ p.ClearTableAndColumn() }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 39 Action3 <- <{ p.AddColorDefine() }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 40 Action4 <- <{ p.AddTable(text) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 41 Action5 <- <{ p.AddColumn(text) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 42 Action6 <- <{ p.AddRelation() }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 43 Action7 <- <{ p.SetRelationLeft(text) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 44 Action8 <- <{ p.SetCardinalityLeft(text)}> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 45 Action9 <- <{ p.SetRelationRight(text) }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 46 Action10 <- <{ p.SetCardinalityRight(text)}> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 47 Action11 <- <{ p.AddTitleKeyValue() }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 48 Action12 <- <{ p.AddTableKeyValue() }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 49 Action13 <- <{ p.AddColumnKeyValue() }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 50 Action14 <- <{ p.AddRelationKeyValue() }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 51 Action15 <- <{ p.SetKey(text) }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 52 Action16 <- <{ p.SetValue(text) }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 53 Action17 <- <{ p.SetValue(text) }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
	}
	p.rules = _rules
}

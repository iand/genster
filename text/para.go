package text

import "strings"

type Para struct {
	sentences []string
}

func (p *Para) Text() string {
	p.FinishSentence()
	return strings.Join(p.sentences, " ")
}

// Continue continues an existing sentence
func (p *Para) Continue(ss ...string) {
	if len(ss) == 0 {
		return
	}
	current := ""
	if len(p.sentences) == 0 {
		p.sentences = append(p.sentences, "")
	} else {
		current = p.sentences[len(p.sentences)-1]
	}

	for i, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		if i == 0 && current == "" {
			current = UpperFirst(s)
			continue
		}
		if !strings.HasSuffix(current, " ") || !strings.HasPrefix(s, " ") {
			current += " "
		}
		current += s
	}

	p.sentences[len(p.sentences)-1] = current
}

// NewSentence begins a new sentence by finishing any existing sentence and combining the strings into text which becomes the current sentence.
func (p *Para) NewSentence(ss ...string) {
	p.FinishSentence()
	p.Continue(ss...)
}

// FinishSentence completes the current sentence and leaves the paragraph ready for the next one.
func (p *Para) FinishSentence() {
	if len(p.sentences) == 0 {
		return
	}

	current := p.sentences[len(p.sentences)-1]
	current = strings.TrimSpace(current)

	if current == "" {
		p.sentences[len(p.sentences)-1] = current
		return
	}

	current = strings.TrimRight(current, ",:;")
	if !strings.HasSuffix(current, ".") && !strings.HasSuffix(current, "!") && !strings.HasSuffix(current, "?") {
		current += "."
	}
	p.sentences[len(p.sentences)-1] = current
	p.sentences = append(p.sentences, "")
}

// AppendClause appends a clause to the current sentence, preceding it with a comma
// if necessary.
func (p *Para) AppendClause(clause string) {
	clause = strings.TrimSpace(clause)
	if len(clause) == 0 {
		return
	}

	if len(p.sentences) == 0 {
		p.Continue(clause)
		return
	}
	current := p.sentences[len(p.sentences)-1]
	if len(current) == 0 {
		current = UpperFirst(clause)
		p.sentences[len(p.sentences)-1] = current
		return
	}

	if !strings.HasSuffix(current, ",") {
		current += ","
	}
	current += " " + clause
	p.sentences[len(p.sentences)-1] = current
}

// AppendAsAside appends an aside to the current sentence, preceding it with a comma
// if necessary and appending a comma after the clause.
func (p *Para) AppendAsAside(clause string) {
	p.AppendClause(clause + ",")
}

func (p *Para) AppendList(ss ...string) {
	if len(ss) == 0 {
		return
	}
	p.Continue(ss[0])

	for i := 1; i < len(ss)-2; i++ {
		p.AppendClause(ss[i])
	}
	p.Continue("and " + ss[len(ss)-1])
}

package text

import (
	"strings"
)

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
	s := p.join(ss...)
	if s == "" {
		return
	}

	current := ""
	if len(p.sentences) == 0 {
		p.sentences = append(p.sentences, "")
	} else {
		current = p.sentences[len(p.sentences)-1]
	}

	if current == "" {
		current = UpperFirst(s)
	} else if !strings.HasSuffix(current, " ") || !strings.HasPrefix(s, " ") {
		current += " "
		current += s
	}

	p.sentences[len(p.sentences)-1] = current
}

func (p *Para) join(ss ...string) string {
	var str string
	for i, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if i != 0 {
			str += " "
		}
		str += s
	}
	return str
}

// NewSentence begins a new sentence by finishing any existing sentence and combining the strings into text which becomes the current sentence.
func (p *Para) NewSentence(ss ...string) {
	p.FinishSentence()
	p.Continue(ss...)
}

// DropSentence drops the current sentence
func (p *Para) DropSentence() {
	if len(p.sentences) == 0 {
		return
	}
	p.sentences = p.sentences[:len(p.sentences)-1]
}

// ReplaceSentence replaces the current sentence with s
func (p *Para) ReplaceSentence(s string) {
	p.DropSentence()
	p.NewSentence(s)
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
func (p *Para) AppendClause(ss ...string) {
	clause := p.join(ss...)
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
	if clause == "" {
		return
	}
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

func (p *Para) CurrentSentenceLength() int {
	if len(p.sentences) == 0 {
		return 0
	}
	return len(p.sentences[len(p.sentences)-1])
}

func (p *Para) CurrentSentenceWords() int {
	if len(p.sentences) == 0 {
		return 0
	}
	return len(strings.Fields(p.sentences[len(p.sentences)-1]))
}

func (p *Para) Length() int {
	l := 0
	for i := range p.sentences {
		l += len(p.sentences[i])
	}
	return l
}

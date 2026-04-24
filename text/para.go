package text

import (
	"strings"
)

type Para struct {
	sentences []string
}

// Text finalizes the current sentence and returns the full paragraph as a string.
func (p *Para) Text() string {
	p.FinishSentence()
	return strings.TrimSpace(strings.Join(p.sentences, " "))
}

// Current returns the accumulated text without finalizing the current sentence.
// Unlike Text, it does not add a period or uppercase the first letter.
func (p *Para) Current() string {
	return strings.TrimSpace(strings.Join(p.sentences, " "))
}

// Continue appends parts to the current sentence. Empty and whitespace-only
// parts are skipped. A colon ":" is appended without a preceding space.
// The first letter is not uppercased until the sentence is finalized.
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
		current = s
	} else {
		if s != ":" {
			current += " "
		}
		current += s
	}

	p.sentences[len(p.sentences)-1] = current
}

func (p *Para) join(ss ...string) string {
	var str strings.Builder
	for i, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if i != 0 && s != ":" {
			str.WriteString(" ")
		}
		str.WriteString(s)
	}
	return str.String()
}

// StartSentence begins a new sentence by finishing any existing sentence and combining
// the strings into text which becomes the current sentence. No punctuation or
// formatting is performed on the new sentence.
func (p *Para) StartSentence(ss ...string) {
	p.FinishSentence()
	p.Continue(ss...)
}

// AddCompleteSentence combines the supplied strings into a formatted sentence and adds
// it to the para. Any open sentence is finished first. The para is left at the start
// of a new sentence.
func (p *Para) AddCompleteSentence(ss ...string) {
	p.FinishSentence()
	p.StartSentence(ss...)
	p.FinishSentence()
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
	p.StartSentence(s)
}

// FinishSentence completes the current sentence, It combines the supplied strings into
// text which are added the current sentence and it then terminates the sentence with a
// full stop and leaves the paragraph ready for the next one.
func (p *Para) FinishSentence(ss ...string) {
	p.Continue(ss...)
	p.FinishSentenceWithTerminator(".")
}

// FinishSentence completes the current sentence, terminating it with t
// and leaves the paragraph ready for the next one.
func (p *Para) FinishSentenceWithTerminator(t string) {
	if len(p.sentences) == 0 {
		return
	}

	current := p.sentences[len(p.sentences)-1]
	current = strings.TrimSpace(current)

	if current == "" {
		p.sentences[len(p.sentences)-1] = current
		return
	}

	current = strings.TrimRight(current, ",;:-!?.")
	current = UpperFirst(current)
	current += t

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

func (p *Para) IsEmpty() bool {
	return len(p.sentences) == 0
}

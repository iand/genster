package pandoc

// Text is a piece of pandoc markup text
type Text string

func (m Text) String() string { return string(m) }
func (m Text) IsZero() bool   { return m == "" }

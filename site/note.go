package site

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/iand/genster/logging"
)

type NoteDoc struct {
	Filename string
	Title    string
	Author   string
	Date     string
	Type     string
	Person   string
	Markdown string
	Mentions []string
}

func LoadNotes(dir string) ([]*NoteDoc, error) {
	var notes []*NoteDoc

	fsys := os.DirFS(dir)

	fnames, err := fs.Glob(fsys, "*.md")
	if err != nil {
		return nil, fmt.Errorf("glob: %w", err)
	}

	for _, fname := range fnames {
		logging.Debug("found note", "filename", fname)
		n, err := ParseNote(fname, fsys)
		if err != nil {
			logging.Warn("note not formatted as expected", "filename", fname, "error", err)
			continue
		}
		if n != nil {
			notes = append(notes, n)
		}
	}

	return notes, nil
}

func ParseNote(fname string, fsys fs.FS) (*NoteDoc, error) {
	f, err := fsys.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open note: %w", err)
	}
	defer f.Close()

	const (
		seekingHeader       = 0
		readingHeader       = 1
		readingBody         = 2
		readingMentionsList = 3
	)

	n := &NoteDoc{
		Filename: fname,
	}

	state := seekingHeader
	body := new(strings.Builder)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		switch state {
		case seekingHeader:
			switch line {
			case "---":
				state = readingHeader
			case "":
				// no-op
			default:
				return nil, fmt.Errorf("did not find start of header marker '---'")
			}
		case readingMentionsList:
			if strings.HasPrefix(line, "- ") {
				n.Mentions = append(n.Mentions, line[2:])
				break
			}
			state = readingHeader
			fallthrough

		case readingHeader:
			switch line {
			case "---":
				state = readingBody
			case "":
				// no-op
			default:
				field, value, ok := strings.Cut(line, ":")
				if !ok {
					return nil, fmt.Errorf("found unexpected text in header: %q", line)
				}

				field = strings.TrimSpace(field)
				value = strings.TrimSpace(value)
				switch strings.ToLower(field) {
				case "person":
					n.Person = value
				case "title":
					n.Title = value
				case "type":
					n.Type = value
				case "author":
					n.Author = value
				case "date":
					n.Date = value
				case "mentions":
					if value != "" {
						n.Mentions = append(n.Mentions, value)
					} else {
						state = readingMentionsList
					}
				}

			}
		case readingBody:
			body.WriteString(line)
			body.WriteString("\n")
		}
	}
	if s.Err() != nil {
		return nil, fmt.Errorf("read note: %w", s.Err())
	}

	n.Markdown = strings.TrimSpace(body.String())

	return n, nil
}

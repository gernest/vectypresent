// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package present

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gernest/vectypresent/present/models"
)

var (
	parsers = make(map[string]ParseFunc)
	funcs   = template.FuncMap{}
)

// Template returns an empty template with the action functions in its FuncMap.
func Template() *template.Template {
	return template.New("").Funcs(funcs)
}

type ParseFunc func(ctx *Context, fileName string, lineNumber int, inputLine string) (models.Elem, error)

// Register binds the named action, which does not begin with a period, to the
// specified parser to be invoked when the name, with a period, appears in the
// present input text.
func Register(name string, parser ParseFunc) {
	if len(name) == 0 || name[0] == ';' {
		panic("bad name in Register: " + name)
	}
	parsers["."+name] = parser
}

// renderElem implements the elem template function, used to render
// sub-templates.
func renderElem(t *template.Template, e models.Elem) (template.HTML, error) {
	var data interface{} = e
	if s, ok := e.(models.Section); ok {
		data = struct {
			models.Section
			Template *template.Template
		}{s, t}
	}
	return execTemplate(t, e.TemplateName(), data)
}

func init() {
	funcs["elem"] = renderElem
}

// execTemplate is a helper to execute a template and return the output as a
// template.HTML value.
func execTemplate(t *template.Template, name string, data interface{}) (template.HTML, error) {
	b := new(bytes.Buffer)
	err := t.ExecuteTemplate(b, name, data)
	if err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}

func readLines(r io.Reader) (*models.Lines, error) {
	var lines []string
	s := bufio.NewScanner(r)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return &models.Lines{Line: 0, Text: lines}, nil
}

// A Context specifies the supporting context for parsing a presentation.
type Context struct {
	// ReadFile reads the file named by filename and returns the contents.
	ReadFile func(filename string) ([]byte, error)
}

// ParseMode represents flags for the Parse function.
type ParseMode int

const (
	// If set, parse only the title and subtitle.
	TitlesOnly ParseMode = 1
)

// Parse parses a document from r.
func (ctx *Context) Parse(r io.Reader, name string, mode ParseMode) (*models.Doc, error) {
	doc := new(models.Doc)
	lines, err := readLines(r)
	if err != nil {
		return nil, err
	}

	for i := lines.Line; i < len(lines.Text); i++ {
		if strings.HasPrefix(lines.Text[i], "*") {
			break
		}

		if isSpeakerNote(lines.Text[i]) {
			doc.TitleNotes = append(doc.TitleNotes, lines.Text[i][2:])
		}
	}

	err = parseHeader(doc, lines)
	if err != nil {
		return nil, err
	}
	if mode&TitlesOnly != 0 {
		return doc, nil
	}

	// Authors
	if doc.Authors, err = parseAuthors(lines); err != nil {
		return nil, err
	}
	// Sections
	if doc.Sections, err = parseSections(ctx, name, lines, []int{}); err != nil {
		return nil, err
	}
	return doc, nil
}

// Parse parses a document from r. Parse reads assets used by the presentation
// from the file system using ioutil.ReadFile.
func Parse(r io.Reader, name string, mode ParseMode) (*models.Doc, error) {
	ctx := Context{ReadFile: ioutil.ReadFile}
	return ctx.Parse(r, name, mode)
}

// isHeading matches any section heading.
var isHeading = regexp.MustCompile(`^\*+ `)

// lesserHeading returns true if text is a heading of a lesser or equal level
// than that denoted by prefix.
func lesserHeading(text, prefix string) bool {
	return isHeading.MatchString(text) && !strings.HasPrefix(text, prefix+"*")
}

// parseSections parses Sections from lines for the section level indicated by
// number (a nil number indicates the top level).
func parseSections(ctx *Context, name string, lines *models.Lines, number []int) ([]models.Section, error) {
	var sections []models.Section
	for i := 1; ; i++ {
		// Next non-empty line is title.
		text, ok := lines.NextNonEmpty()
		for ok && text == "" {
			text, ok = lines.Next()
		}
		if !ok {
			break
		}
		prefix := strings.Repeat("*", len(number)+1)
		if !strings.HasPrefix(text, prefix+" ") {
			lines.Back()
			break
		}
		section := models.Section{
			Number: append(append([]int{}, number...), i),
			Title:  text[len(prefix)+1:],
		}
		text, ok = lines.NextNonEmpty()
		for ok && !lesserHeading(text, prefix) {
			var e models.Elem
			r, _ := utf8.DecodeRuneInString(text)
			switch {
			case unicode.IsSpace(r):
				i := strings.IndexFunc(text, func(r rune) bool {
					return !unicode.IsSpace(r)
				})
				if i < 0 {
					break
				}
				indent := text[:i]
				var s []string
				for ok && (strings.HasPrefix(text, indent) || text == "") {
					if text != "" {
						text = text[i:]
					}
					s = append(s, text)
					text, ok = lines.Next()
				}
				lines.Back()
				pre := strings.Join(s, "\n")
				pre = strings.Replace(pre, "\t", "    ", -1) // browsers treat tabs badly
				pre = strings.TrimRightFunc(pre, unicode.IsSpace)
				e = models.Text{Lines: []string{pre}, Pre: true}
			case strings.HasPrefix(text, "- "):
				var b []string
				for ok && strings.HasPrefix(text, "- ") {
					b = append(b, text[2:])
					text, ok = lines.Next()
				}
				lines.Back()
				e = models.List{Bullet: b}
			case isSpeakerNote(text):
				section.Notes = append(section.Notes, text[2:])
			case strings.HasPrefix(text, prefix+"* "):
				lines.Back()
				subsecs, err := parseSections(ctx, name, lines, section.Number)
				if err != nil {
					return nil, err
				}
				for _, ss := range subsecs {
					section.Elem = append(section.Elem, ss)
				}
			case strings.HasPrefix(text, "."):
				args := strings.Fields(text)
				if args[0] == ".background" {
					section.Classes = append(section.Classes, "background")
					section.Styles = append(section.Styles, "background-image: url('"+args[1]+"')")
					break
				}
				parser := parsers[args[0]]
				if parser == nil {
					return nil, fmt.Errorf("%s:%d: unknown command %q\n", name, lines.Line, text)
				}
				t, err := parser(ctx, name, lines.Line, text)
				if err != nil {
					return nil, err
				}
				e = t
			default:
				var l []string
				for ok && strings.TrimSpace(text) != "" {
					if text[0] == '.' { // Command breaks text block.
						lines.Back()
						break
					}
					if strings.HasPrefix(text, `\.`) { // Backslash escapes initial period.
						text = text[1:]
					}
					l = append(l, text)
					text, ok = lines.Next()
				}
				if len(l) > 0 {
					e = models.Text{Lines: l}
				}
			}
			if e != nil {
				section.Elem = append(section.Elem, e)
			}
			text, ok = lines.NextNonEmpty()
		}
		if isHeading.MatchString(text) {
			lines.Back()
		}
		sections = append(sections, section)
	}
	return sections, nil
}

func parseHeader(doc *models.Doc, lines *models.Lines) error {
	var ok bool
	// First non-empty line starts header.
	doc.Title, ok = lines.NextNonEmpty()
	if !ok {
		return errors.New("unexpected EOF; expected title")
	}
	for {
		text, ok := lines.Next()
		if !ok {
			return errors.New("unexpected EOF")
		}
		if text == "" {
			break
		}
		if isSpeakerNote(text) {
			continue
		}
		const tagPrefix = "Tags:"
		if strings.HasPrefix(text, tagPrefix) {
			tags := strings.Split(text[len(tagPrefix):], ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			doc.Tags = append(doc.Tags, tags...)
		} else if t, ok := parseTime(text); ok {
			doc.Time = t
		} else if doc.Subtitle == "" {
			doc.Subtitle = text
		} else {
			return fmt.Errorf("unexpected header line: %q", text)
		}
	}
	return nil
}

func parseAuthors(lines *models.Lines) (authors []models.Author, err error) {
	// This grammar demarcates authors with blanks.

	// Skip blank lines.
	if _, ok := lines.NextNonEmpty(); !ok {
		return nil, errors.New("unexpected EOF")
	}
	lines.Back()

	var a *models.Author
	for {
		text, ok := lines.Next()
		if !ok {
			return nil, errors.New("unexpected EOF")
		}

		// If we find a section heading, we're done.
		if strings.HasPrefix(text, "* ") {
			lines.Back()
			break
		}

		if isSpeakerNote(text) {
			continue
		}

		// If we encounter a blank we're done with this author.
		if a != nil && len(text) == 0 {
			authors = append(authors, *a)
			a = nil
			continue
		}
		if a == nil {
			a = new(models.Author)
		}

		// Parse the line. Those that
		// - begin with @ are twitter names,
		// - contain slashes are links, or
		// - contain an @ symbol are an email address.
		// The rest is just text.
		var el models.Elem
		switch {
		case strings.HasPrefix(text, "@"):
			el = parseURL("http://twitter.com/" + text[1:])
		case strings.Contains(text, ":"):
			el = parseURL(text)
		case strings.Contains(text, "@"):
			el = parseURL("mailto:" + text)
		}
		if l, ok := el.(models.Link); ok {
			l.Label = text
			el = l
		}
		if el == nil {
			el = models.Text{Lines: []string{text}}
		}
		a.Elem = append(a.Elem, el)
	}
	if a != nil {
		authors = append(authors, *a)
	}
	return authors, nil
}

func parseURL(text string) models.Elem {
	u, err := url.Parse(text)
	if err != nil {
		log.Printf("Parse(%q): %v", text, err)
		return nil
	}
	return models.Link{URL: u}
}

func parseTime(text string) (t time.Time, ok bool) {
	t, err := time.Parse("15:04 2 Jan 2006", text)
	if err == nil {
		return t, true
	}
	t, err = time.Parse("2 Jan 2006", text)
	if err == nil {
		// at 11am UTC it is the same date everywhere
		t = t.Add(time.Hour * 11)
		return t, true
	}
	return time.Time{}, false
}

func isSpeakerNote(s string) bool {
	return strings.HasPrefix(s, ": ")
}

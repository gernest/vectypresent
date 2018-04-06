package models

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func init() {
	gob.Register(Text{})
	gob.Register(Code{})
	gob.Register(List{})
}

func Encode(o io.Writer, v interface{}) error {
	return gob.NewEncoder(o).Encode(v)
}

func Decode(o io.Reader, v interface{}) error {
	return gob.NewDecoder(o).Decode(v)
}

// TODO(adg): replace the PlayEnabled flag with something less spaghetti-like.
// Instead this will probably be determined by a template execution Context
// value that contains various global metadata required when rendering
// templates.

// NotesEnabled specifies whether presenter notes should be displayed in the
// present user interface.
var NotesEnabled = false

// PlayEnabled specifies whether runnable playground snippets should be
// displayed in the present user interface.
var PlayEnabled = false

const TimeFormat = "2 January 2006"

type Caption struct {
	Text string
}

func (c Caption) TemplateName() string { return "caption" }

type Code struct {
	Text     template.HTML
	Play     bool   // runnable code
	Edit     bool   // editable code
	FileName string // file name
	Ext      string // file extension
	Raw      []byte // content of the file
}

func (c Code) TemplateName() string { return "code" }

type HTML struct {
	template.HTML
}

func (s HTML) TemplateName() string { return "html" }

type Iframe struct {
	URL    string
	Width  int
	Height int
}

func (i Iframe) TemplateName() string { return "iframe" }

type Image struct {
	URL    string
	Width  int
	Height int
}

func (i Image) TemplateName() string { return "image" }

type Link struct {
	URL   *url.URL
	Label string
}

func (l Link) TemplateName() string { return "link" }

type Video struct {
	URL        string
	SourceType string
	Width      int
	Height     int
}

func (v Video) TemplateName() string { return "video" }

// Doc represents an entire document.
type Doc struct {
	Title      string
	Subtitle   string
	Time       time.Time
	Authors    []Author
	TitleNotes []string
	Sections   []Section
	Tags       []string
}

// Render renders the doc to the given writer using the provided template.
func (d *Doc) Render(w io.Writer, t *template.Template) error {
	data := struct {
		*Doc
		Template     *template.Template
		PlayEnabled  bool
		NotesEnabled bool
	}{d, t, PlayEnabled, NotesEnabled}
	return t.ExecuteTemplate(w, "root", data)
}

// Author represents the person who wrote and/or is presenting the document.
type Author struct {
	Elem []Elem
}

// TextElem returns the first text elements of the author details.
// This is used to display the author' name, job title, and company
// without the contact details.
func (p *Author) TextElem() (elems []Elem) {
	for _, el := range p.Elem {
		if _, ok := el.(Text); !ok {
			break
		}
		elems = append(elems, el)
	}
	return
}

// Section represents a section of a document (such as a presentation slide)
// comprising a title and a list of elements.
type Section struct {
	Number  []int
	Title   string
	Elem    []Elem
	Notes   []string
	Classes []string
	Styles  []string
}

// Render renders the section to the given writer using the provided template.
func (s *Section) Render(w io.Writer, t *template.Template) error {
	data := struct {
		*Section
		Template    *template.Template
		PlayEnabled bool
	}{s, t, PlayEnabled}
	return t.ExecuteTemplate(w, "section", data)
}

// HTMLAttributes for the section
func (s Section) HTMLAttributes() template.HTMLAttr {
	if len(s.Classes) == 0 && len(s.Styles) == 0 {
		return ""
	}

	var class string
	if len(s.Classes) > 0 {
		class = fmt.Sprintf(`class=%q`, strings.Join(s.Classes, " "))
	}
	var style string
	if len(s.Styles) > 0 {
		style = fmt.Sprintf(`style=%q`, strings.Join(s.Styles, " "))
	}
	return template.HTMLAttr(strings.Join([]string{class, style}, " "))
}

// Sections contained within the section.
func (s Section) Sections() (sections []Section) {
	for _, e := range s.Elem {
		if section, ok := e.(Section); ok {
			sections = append(sections, section)
		}
	}
	return
}

// Level returns the level of the given section.
// The document title is level 1, main section 2, etc.
func (s Section) Level() int {
	return len(s.Number) + 1
}

// FormattedNumber returns a string containing the concatenation of the
// numbers identifying a Section.
func (s Section) FormattedNumber() string {
	b := &bytes.Buffer{}
	for _, n := range s.Number {
		fmt.Fprintf(b, "%v.", n)
	}
	return b.String()
}

func (s Section) TemplateName() string { return "section" }

// Elem defines the interface for a present element. That is, something that
// can provide the name of the template used to render the element.
type Elem interface {
	TemplateName() string
}

// Text represents an optionally preformatted paragraph.
type Text struct {
	Lines []string
	Pre   bool
}

func (t Text) TemplateName() string { return "text" }

// List represents a bulleted list.
type List struct {
	Bullet []string
}

func (l List) TemplateName() string { return "list" }

// Lines is a helper for parsing line-based input.
type Lines struct {
	Line int // 0 indexed, so has 1-indexed number of last line returned
	Text []string
}

func (l *Lines) Next() (text string, ok bool) {
	for {
		current := l.Line
		l.Line++
		if current >= len(l.Text) {
			return "", false
		}
		text = l.Text[current]
		// Lines starting with # are comments.
		if len(text) == 0 || text[0] != '#' {
			ok = true
			break
		}
	}
	return
}

func (l *Lines) Back() {
	l.Line--
}

func (l *Lines) NextNonEmpty() (text string, ok bool) {
	for {
		text, ok = l.Next()
		if !ok {
			return
		}
		if len(text) > 0 {
			break
		}
	}
	return
}

type File struct {
	IsDir    bool
	Children []*File
	Name     string
	Context  *Doc `json:"-"`
}

func (d *File) Path() string {
	return d.Name
}
func (d *File) BaseName() string {
	return filepath.Base(d.Name)
}

func (d *File) URL() string {
	s := "/" + filepath.ToSlash(d.Path())
	return s
}

func (d *File) Cache(c *sync.Map) {
	c.Store(d.URL(), d)
	for _, child := range d.Children {
		child.Cache(c)
	}
}

func base(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	return fmt.Sprintf("/%s/", parts[0])
}

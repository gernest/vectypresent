package main

import (
	"bytes"
	"fmt"

	"github.com/gopherjs/vecty/event"

	"github.com/gopherjs/vecty/prop"

	"github.com/gernest/CatAcademy/present/models"
	"github.com/gernest/socrates"
	"github.com/gernest/xhr"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

func main() {
	vecty.RenderBody(&Slide{})
}

type Position int

const (
	FarPast Position = iota << 1
	Past
	Current
	Next
	FarNext
	Silent
)

func (p Position) Class() string {
	switch p {
	case FarPast:
		return "far-past"
	case Past:
		return "past"
	case Current:
		return "current"
	case Next:
		return "next"
	case FarNext:
		return "far-next"
	case Silent:
		return ""
	default:
		return ""
	}
}

type Slide struct {
	vecty.Core

	Doc         *models.Doc
	socket      *socrates.Socket
	activeSlide int
}

// Mount implements vecty.Mount interface.
//
// This opens a websocket connection which allows to remotely control the slides.
func (s *Slide) Mount() {
	// location := js.Global.Get("location").Get("href").String()
	// u, err := url.Parse(location)
	// if err != nil {
	// 	panic(err)
	// }
	// u.Scheme = "ws"
	// u.Path = "/ws/" + u.Path
	// sock, err := socrates.NewSocket(u.String(), &socrates.Options{
	// 	OnMessage: s.OnMessage,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// s.socket = sock
	go func() {
		data, err := xhr.Send("GET", "/data/", nil)
		if err != nil {
			panic(err)
		}
		doc := &models.Doc{}
		err = models.Decode(bytes.NewReader(data), &doc)
		if err != nil {
			panic(err)
		}
		s.Doc = doc
		vecty.Rerender(s)
	}()

}

func (s *Slide) UnMount() {
	s.socket.Close()
}

func (s *Slide) OnMessage(data []byte) {

}

func (s *Slide) Render() vecty.ComponentOrHTML {
	if s.Doc == nil {
		return elem.Body()
	}
	var sections vecty.List
	for i, section := range s.Doc.Sections {
		pos := Silent
		switch i {
		case s.activeSlide - 2:
			pos = FarPast
		case s.activeSlide - 1:
			pos = Past
		case s.activeSlide:
			pos = Current
		case s.activeSlide + 1:
			pos = Next
		case s.activeSlide + 2:
			pos = FarNext
		}
		sections = append(sections, &Section{s: section, Pos: pos})
	}
	return elem.Body(
		vecty.Markup(
			vecty.Style("display", "none"),
			event.KeyDown(func(e *vecty.Event) {
				s.UpdatePosition(e.Get("code").String())
			}),
		),
		elem.Section(
			vecty.Markup(
				vecty.Class("slides", "layout-widescreen"),
			),
			elem.Article(
				elem.Heading1(
					vecty.Text(s.Doc.Title),
				),
				vecty.If(s.Doc.Subtitle != "", elem.Heading3(
					vecty.Text(s.Doc.Subtitle),
				)),
				vecty.If(!s.Doc.Time.IsZero(), elem.Heading3(
					vecty.Text(s.Doc.Time.Format(models.TimeFormat)),
				)),
			),
			sections,
		),
	)
}

func (s *Slide) UpdatePosition(key string) {
	up := false
	switch key {
	case "ArrowRight", "ArrowUp":
		if s.activeSlide < len(s.Doc.Sections) {
			s.activeSlide++
			up = true
		}
	case "ArrowLeft", "ArrowDown":
		if s.activeSlide != 0 {
			s.activeSlide--
			up = true
		}
	}
	if up {
		vecty.Rerender(s)
	}
}

func (s *Slide) renderSections() vecty.List {
	var sections vecty.List
	for i, section := range s.Doc.Sections {
		pos := Silent
		switch i {
		case s.activeSlide - 2:
			pos = FarPast
		case s.activeSlide - 1:
			pos = Past
		case s.activeSlide:
			pos = Current
		case s.activeSlide + 1:
			pos = Next
		case s.activeSlide + 2:
			pos = FarNext
		}
		sections = append(sections, &Section{s: section, Pos: pos})
	}
	return sections
}

func renderElems(e []models.Elem) vecty.List {
	var o vecty.List
	for _, v := range e {
		o = append(o, renderElem(v))
	}
	return o
}

func renderElem(e models.Elem) vecty.ComponentOrHTML {
	switch v := e.(type) {
	case models.Section:
		return &Section{s: v}
	case models.List:
		return &List{list: v}
	case models.Text:
		return &Text{txt: v}
	case models.Code:
		return &Code{code: v}
	case models.Image:
		return &Image{img: v}
	case models.Link:
		return &Link{link: v}
	case models.Caption:
		return &Caption{c: v}
	default:
		return nil
	}
}

func join(v []string, by string) string {
	var o string
	for k, value := range v {
		if k == 0 {
			o = value
		} else {
			o += by + value
		}
	}
	return o
}

type Section struct {
	vecty.Core

	Pos Position `vecty:"prop"`
	s   models.Section
}

func (s *Section) Render() vecty.ComponentOrHTML {
	return elem.Article(
		vecty.Markup(
			vecty.MarkupIf(s.Pos.Class() != "", vecty.Class(s.Pos.Class())),
			vecty.MarkupIf(s.s.Classes != nil,
				vecty.Class(s.s.Classes...)),
			vecty.MarkupIf(s.s.Styles != nil,
				vecty.Attribute("style", join(s.s.Styles, " "))),
		),
		vecty.If(s.s.Elem != nil,
			vecty.List{
				elem.Heading3(vecty.Text(s.s.Title)),
				renderElems(s.s.Elem),
			},
		),
		vecty.If(s.s.Elem == nil,
			vecty.List{
				elem.Heading2(vecty.Text(s.s.Title)),
			},
		),
	)
}

type List struct {
	vecty.Core

	list models.List
}

func (l *List) Render() vecty.ComponentOrHTML {
	var items vecty.List
	for _, bullet := range l.list.Bullet {
		items = append(items, elem.ListItem(
			vecty.Text(bullet),
		))
	}
	return elem.UnorderedList(items)
}

func newLine() *vecty.HTML {
	return elem.Break()
}

type Code struct {
	vecty.Core

	code models.Code
}

func (c *Code) Render() vecty.ComponentOrHTML {
	class := vecty.ClassMap{
		"code":       true,
		"playground": c.code.Play,
	}
	return elem.Div(
		vecty.Markup(class,
			vecty.MarkupIf(c.code.Edit,
				vecty.Attribute("contenteditable", "true"),
				vecty.Attribute("spellcheck", "false"),
			),
		),
		vecty.Text(string(c.code.Text)),
	)
}

type Text struct {
	vecty.Core

	txt models.Text
}

func (t *Text) Render() vecty.ComponentOrHTML {
	if t.txt.Pre {
		var s string
		for k, v := range t.txt.Lines {
			if k == 0 {
				s += v
			} else {
				s += "\n" + v
			}
		}
		return elem.Div(
			vecty.Markup(vecty.Class("code")),
			elem.Preformatted(
				vecty.Text(s),
			),
		)
	}
	var s string
	for k, v := range t.txt.Lines {
		if k == 0 {
			s += v
		} else {
			s += "<br>" + v
		}
	}
	return elem.Paragraph(
		vecty.Markup(vecty.UnsafeHTML(s)),
	)
}

type Image struct {
	vecty.Core

	img models.Image
}

func (i *Image) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Markup(vecty.Class("image")),
		elem.Image(
			vecty.Markup(
				prop.Src(i.img.URL),
				vecty.MarkupIf(i.img.Height != 0,
					vecty.Attribute("height", fmt.Sprint(i.img.Height))),
				vecty.MarkupIf(i.img.Width != 0,
					vecty.Attribute("height", fmt.Sprint(i.img.Width))),
			),
		),
	)
}

type IFrame struct {
	vecty.Core

	frame models.Iframe
}

func (f *IFrame) Render() vecty.ComponentOrHTML {
	return elem.InlineFrame(
		vecty.Markup(
			prop.Src(f.frame.URL),
			vecty.MarkupIf(f.frame.Height != 0,
				vecty.Attribute("height", fmt.Sprint(f.frame.Height))),
			vecty.MarkupIf(f.frame.Width != 0,
				vecty.Attribute("height", fmt.Sprint(f.frame.Width))),
		),
	)
}

type Video struct {
	vecty.Core

	v models.Video
}

func (v *Video) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Markup(
			vecty.Class("video"),
		),
		elem.Video(
			vecty.Markup(
				vecty.MarkupIf(v.v.Height != 0,
					vecty.Attribute("height", fmt.Sprint(v.v.Height))),
				vecty.MarkupIf(v.v.Width != 0,
					vecty.Attribute("height", fmt.Sprint(v.v.Width))),
				vecty.Attribute("controls", ""),
			),
			elem.Source(
				vecty.Markup(
					prop.Src(v.v.URL),
					vecty.Attribute("type", v.v.SourceType),
				),
			),
		),
	)

}

type Link struct {
	vecty.Core

	link models.Link
}

func (l *Link) Render() vecty.ComponentOrHTML {
	return elem.Paragraph(
		vecty.Markup(vecty.Class("link")),
		elem.Anchor(
			vecty.Markup(
				prop.Href(l.link.URL.String()),
				vecty.Attribute("target", "_blank"),
			),
			vecty.Text(l.link.Label),
		),
	)
}

type Caption struct {
	vecty.Core

	c models.Caption
}

func (c *Caption) Render() vecty.ComponentOrHTML {
	return elem.FigureCaption(
		vecty.Text(c.c.Text),
	)
}

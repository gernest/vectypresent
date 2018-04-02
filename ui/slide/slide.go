package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gernest/locstor"

	"github.com/gopherjs/vecty/prop"

	"github.com/gernest/CatAcademy/present/models"
	"github.com/gernest/socrates"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

func main() {
	vecty.RenderBody(&Slide{})
}

type Slide struct {
	vecty.Core

	Doc    *models.Doc
	socket *socrates.Socket
}

// Mount implements vecty.Mount interface.
//
// This opens a websocket connection which allows to remotely control the slides.
func (s *Slide) Mount() {
	location := js.Global.Get("location").Get("href").String()
	u, err := url.Parse(location)
	if err != nil {
		panic(err)
	}
	u.Scheme = "ws"
	u.Path = "/ws/" + u.Path
	sock, err := socrates.NewSocket(u.String(), &socrates.Options{
		OnMessage: s.OnMessage,
	})
	if err != nil {
		panic(err)
	}
	s.socket = sock
	data, err := locstor.GetItem("slideData")
	if err != nil {
		panic(err)
	}
	doc := &models.Doc{}
	err = json.Unmarshal([]byte(data), doc)
	if err != nil {
		panic(err)
	}
	s.Doc = doc
	vecty.Rerender(s)
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
	return elem.Body(
		vecty.Markup(
			vecty.Style("display", "none"),
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
			s.renderSections(),
		),
	)
}

func (s *Slide) renderSections() vecty.List {
	var sections vecty.List
	for _, section := range s.Doc.Sections {
		sections = append(sections, &Section{s: section})
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

	s models.Section
}

func (s *Section) Render() vecty.ComponentOrHTML {
	return elem.Article(
		vecty.Markup(
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
		var s strings.Builder
		for k, v := range t.txt.Lines {
			if k == 0 {
				s.WriteString(v)
			} else {
				s.WriteRune('\n')
				s.WriteString(v)
			}
		}
		return elem.Div(
			vecty.Markup(vecty.Class("code")),
			elem.Preformatted(
				vecty.Text(s.String()),
			),
		)
	}
	var s strings.Builder
	for k, v := range t.txt.Lines {
		if k == 0 {
			s.WriteString(v)
		} else {
			s.WriteString("<br>")
			s.WriteString(v)
		}
	}
	return elem.Paragraph(
		vecty.Markup(vecty.UnsafeHTML(s.String())),
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

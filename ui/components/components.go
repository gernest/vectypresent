package components

import (
	"fmt"
	"strings"

	"github.com/gernest/vectypresent/present/models"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

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

// Section is a single slide show unit/page.
type Section struct {
	vecty.Core

	Pos Position `vecty:"prop"`
	S   models.Section
}

func (s *Section) Render() vecty.ComponentOrHTML {
	return elem.Article(
		vecty.Markup(
			vecty.MarkupIf(s.Pos.Class() != "", vecty.Class(s.Pos.Class())),
			vecty.MarkupIf(s.S.Classes != nil,
				vecty.Class(s.S.Classes...)),
			vecty.MarkupIf(s.S.Styles != nil,
				vecty.Attribute("style", strings.Join(s.S.Styles, " "))),
		),
		vecty.If(s.S.Elem != nil,
			vecty.List{
				elem.Heading3(vecty.Text(s.S.Title)),
				RenderElems(s.S.Elem),
			},
		),
		vecty.If(s.S.Elem == nil,
			vecty.List{
				elem.Heading2(vecty.Text(s.S.Title)),
			},
		),
	)
}

func RenderElems(e []models.Elem) vecty.List {
	var o vecty.List
	for _, v := range e {
		o = append(o, RenderElem(v))
	}
	return o
}

func RenderElem(e models.Elem) vecty.ComponentOrHTML {
	switch v := e.(type) {
	case models.Section:
		return &Section{S: v}
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

type List struct {
	vecty.Core

	list models.List
}

func (l *List) Render() vecty.ComponentOrHTML {
	var items vecty.List
	for _, bullet := range l.list.Bullet {
		items = append(items, elem.ListItem(
			vecty.Markup(
				vecty.UnsafeHTML(string(models.Style(bullet))),
			),
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
			vecty.UnsafeHTML(string(c.code.Text)),
		),
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

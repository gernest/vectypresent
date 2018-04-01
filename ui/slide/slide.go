package main

import (
	"encoding/json"
	"net/url"

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

	data := js.Global.Get("slideData").String()
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
		sections = append(sections, elem.Article(
			vecty.Markup(
				vecty.MarkupIf(section.Classes != nil,
					vecty.Class(section.Classes...)),
				vecty.MarkupIf(section.Styles != nil,
					vecty.Attribute("style", join(section.Styles, " "))),
			),
			vecty.If(section.Elem != nil,
				vecty.List{
					elem.Heading3(vecty.Text(section.Title)),
					s.renderElems(section.Elem),
				},
			),
			vecty.If(section.Elem == nil,
				vecty.List{
					elem.Heading2(vecty.Text(section.Title)),
				},
			),
		))
	}
	return sections
}

func (s *Slide) renderElems(e []models.Elem) vecty.ComponentOrHTML {
	return nil
}

func (s *Slide) renderElem(e models.Elem) vecty.ComponentOrHTML {
	return nil
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

package article

import (
	"bytes"
	"net/url"
	"path/filepath"

	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"

	"github.com/gernest/vectypresent/present/models"
	"github.com/gernest/vectypresent/ui/components"
	"github.com/gernest/vectypresent/ui/util"
	"github.com/gernest/xhr"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
)

const (
	articleSheet = "/static/article.css"
)

type Article struct {
	vecty.Core

	doc *models.Doc
}

func (a *Article) Mount() {
	location := js.Global.Get("location")
	addStyle(location.Get("origin").String())
	href := location.Get("href").String()
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}
	u.Path = filepath.Join("/files", u.Path)
	go func() {
		data, err := xhr.Send("GET", u.String(), nil)
		if err != nil {
			panic(err)
		}
		doc := &models.Doc{}
		err = models.Decode(bytes.NewReader(data), &doc)
		if err != nil {
			panic(err)
		}
		a.doc = doc
		vecty.SetTitle(doc.Title)
		vecty.Rerender(a)
	}()
}

func addStyle(origin string) {
	slideHref := origin + articleSheet
	hasSlideSheet := false
	util.ListSheets(func(sheet *js.Object) bool {
		href := sheet.Get("href").String()
		switch href {
		case slideHref:
			sheet.Set("disabled", false)
			hasSlideSheet = true
		default:
			sheet.Set("disabled", true)
		}
		return true
	})
	if !hasSlideSheet {
		link := js.Global.Get("document").Call("createElement", "link")
		link.Set("rel", "stylesheet")
		link.Set("href", slideHref)
		js.Global.Get("document").Get("head").Call("appendChild", link)
	}
}

func (a *Article) Unmount() {
	// restoreStyle(js.Global.Get("location").Get("origin").String())
}

func restoreStyle(origin string) {
	slideHref := origin + articleSheet
	util.ListSheets(func(sheet *js.Object) bool {
		href := sheet.Get("href").String()
		switch href {
		case slideHref:
			sheet.Set("disabled", true)
		default:
			sheet.Set("disabled", false)
		}
		return true
	})
}

func (a *Article) Render() vecty.ComponentOrHTML {
	if a.doc == nil {
		return elem.Body()
	}
	var authors vecty.List
	for _, author := range a.doc.Authors {
		authors = append(authors, elem.Div(
			vecty.Markup(vecty.Class("author")),
			components.RenderElems(author.Elem),
		))
	}
	var sections vecty.List
	for _, v := range a.doc.Sections {
		sections = append(sections, &components.Section{S: v})
	}
	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Class("wide"),
				prop.ID("topbar"),
			),
			elem.Div(
				vecty.Markup(
					vecty.Class("container"),
				),
				elem.Div(
					vecty.Markup(
						prop.ID("heading"),
					),
					authors,
				),
			),
		),
		elem.Div(
			vecty.Markup(
				vecty.Class("wide"),
				prop.ID("page"),
			),
			elem.Div(
				vecty.Markup(
					vecty.Class("container"),
				),
				&components.TOC{Sections: a.doc.Sections},
				sections,
			),
		),
	)
}

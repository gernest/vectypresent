package main

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/gernest/vectypresent/present/models"
	"github.com/gernest/vectypresent/ui/dir"
	"github.com/gernest/vectypresent/ui/router"
	"github.com/gernest/vectypresent/ui/slide"
	"github.com/gernest/xhr"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

func main() {
	r := router.NewRouter()

	cache := &sync.Map{}
	r.NotFound = func(ctx ...interface{}) vecty.ComponentOrHTML {
		if len(ctx) > 0 {
			if key, ok := ctx[0].(string); ok {
				if vk, ok := cache.Load(key); ok {
					val := vk.(*models.File)
					if val.IsDir {
						return &dir.Dir{Dir: val, Router: r}
					}
					if val.IsSlide() {
						return &slide.Slide{}
					}
					return &PlainText{}
				}
			}
		}
		return elem.Body(
			vecty.Text("404"),
		)
	}
	r.Handle("/", func(ctx ...interface{}) vecty.ComponentOrHTML {
		return &Home{cache: cache, router: r}
	})
	vecty.RenderBody(r)
}

type Home struct {
	vecty.Core

	dir    *models.File
	cache  *sync.Map
	router *router.Router
}

func (h *Home) Mount() {
	go func() {
		data, err := xhr.Send("GET", "/context", nil)
		if err != nil {
			panic(err)
		}
		dir := &models.File{}
		err = json.Unmarshal(data, dir)
		if err != nil {
			panic(err)
		}
		h.dir = dir
		h.dir.Cache(h.cache)
		vecty.SetTitle(dir.Name)
		vecty.Rerender(h)
	}()
}

func (h *Home) Render() vecty.ComponentOrHTML {
	if h.dir != nil {
		return elem.Body(
			&dir.Dir{Dir: h.dir, Router: h.router},
		)
	}
	return elem.Body()
}

type PlainText struct {
	vecty.Core
	txt string
}

func (p *PlainText) Mount() {
	location := js.Global.Get("location")
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
		p.txt = string(data)
		vecty.Rerender(p)
	}()
}

func (p *PlainText) Render() vecty.ComponentOrHTML {
	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Class("code"),
			),
			elem.Code(
				elem.Preformatted(
					vecty.Markup(
						vecty.Style("text-align", "initial"),
						vecty.UnsafeHTML(p.txt),
					),
				),
			),
		),
	)
}

package main

import (
	"encoding/json"
	"sync"

	"github.com/gernest/CatAcademy/present/models"
	"github.com/gernest/CatAcademy/ui/dir"
	"github.com/gernest/CatAcademy/ui/router"
	"github.com/gernest/CatAcademy/ui/slide"
	"github.com/gernest/xhr"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

func main() {
	r := router.NewRouter()
	r.NotFound = func(...interface{}) vecty.ComponentOrHTML {
		return elem.Body(
			vecty.Text("404"),
		)
	}

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

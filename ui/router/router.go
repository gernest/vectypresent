package router

import (
	"github.com/gernest/locstor"
	"github.com/gernest/vectypresent/ui/components"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

// Router is a component which simplifies registering of components and proxying
// them depending on the path. This uses pushstate for navigation.
type Router struct {
	vecty.Core

	handlers map[string]HandlerFunc
	active   string
	NotFound HandlerFunc
	context  []interface{}
	ready    bool
}

const ActiveRoute = "ACTIVE_ROUTE"

type HandlerFunc func(...interface{}) vecty.ComponentOrHTML

// NewRouter returns new Router instance.
func NewRouter() *Router {
	active := "/"
	if a, err := locstor.GetItem(ActiveRoute); err == nil {
		active = a
	}
	r := &Router{
		active:   active,
		handlers: make(map[string]HandlerFunc),
	}
	return r
}

func (r *Router) BeforeRendering() func() {
	return func() {
		if !r.ready {
			r.ready = true
			vecty.Rerender(r)
		}
	}
}

// Mount registers event listener for onpopstate global event.
func (r *Router) Mount() {
	js.Global.Set("onpopstate", func() {
		go func() {
			path := js.Global.Get("location").Get("pathname").String()
			r.active = path
			vecty.Rerender(r)
		}()
	})
}

// PushState re renders component registered on path.
func (r *Router) PushState(path string, ctx ...interface{}) {
	js.Global.Get("history").Call("pushState", nil, "", path)
	r.active = path
	r.context = ctx
	locstor.SetItem(ActiveRoute, path)
	vecty.Rerender(r)
}

// Render renders the active component register in the router. If there was any
// context supplied which calling pushState, it is passed to the handler.
func (r *Router) Render() vecty.ComponentOrHTML {
	if !r.ready {
		return &components.Spinner{}
	}
	if r.active == "" {
		r.active = "/"
	}
	if h, ok := r.handlers[r.active]; ok {
		if r.context != nil {
			return h(r.context...)
		}
		return h()
	}
	if r.NotFound != nil {
		return r.NotFound(r.active)
	}
	return elem.Body(vecty.Text("Not Found"))
}

// Handle register the handler for the matched pattern. At the moment a map is
// used to store the pattern/handler so matching is done on a full string
// comparison, the pattern is used as key.
//
// TODO: use trie for pattern matching.
func (r *Router) Handle(pattern string, h HandlerFunc) {
	r.handlers[pattern] = h
}

// Unmount stops listening for onpopstate events when the component is
// unmounted.
func (r *Router) Unmount() {
	js.Global.Set("onpopstate", nil)
	locstor.Clear()
}

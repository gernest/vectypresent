package slide

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty/event"

	"github.com/gernest/vectypresent/present/models"
	"github.com/gernest/vectypresent/ui/components"
	"github.com/gernest/vectypresent/ui/util"
	"github.com/gernest/xhr"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

type Slide struct {
	vecty.Core

	doc         *models.Doc
	activeSlide int
	remote      *RemoteControl
	recording   bool
	auto        bool
	startTime   time.Time
}

const (
	slideSheet = "/static/styles.css"
)

func (s *Slide) Mount() {
	location := js.Global.Get("location")
	addStyle(location.Get("origin").String())
	href := location.Get("href").String()
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}
	s.remote = &RemoteControl{
		events: make(map[int]TickEvent),
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
		s.doc = doc
		vecty.SetTitle(doc.Title)
		vecty.Rerender(s)
	}()

}

func (s *Slide) Unmount() {
	restoreStyle(js.Global.Get("location").Get("origin").String())
}

// disable all other stylesheets and only leave styles.css
func addStyle(origin string) {
	slideHref := origin + slideSheet
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

func restoreStyle(origin string) {
	slideHref := origin + slideSheet
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

func getPos(active, n int) components.Position {
	switch n {
	case active - 2:
		return components.FarPast
	case active - 1:
		return components.Past
	case active:
		return components.Current
	case active + 1:
		return components.Next
	case active + 2:
		return components.FarNext
	default:
		return components.Silent
	}
}

func (s *Slide) Render() vecty.ComponentOrHTML {
	if s.doc == nil {
		return elem.Body()
	}
	var sections vecty.List
	for i, section := range s.doc.Sections {
		pos := getPos(s.activeSlide, i+1)
		sections = append(sections, &components.Section{S: section, Pos: pos, Slide: true})
	}
	var authors vecty.List
	for _, author := range s.doc.Authors {
		authors = append(authors, elem.Div(
			vecty.Markup(vecty.Class("presenter")),
			components.RenderElems(author.Elem),
		))
	}
	initPos := getPos(s.activeSlide, 0)
	return elem.Body(
		vecty.Markup(
			vecty.Style("display", "none"),
			event.KeyDown(func(e *vecty.Event) {
				s.KeyPress(e.Get("code").String())
			}),
			vecty.MarkupIf(s.recording, vecty.Style("background", "red")),
		),
		elem.Section(
			vecty.Markup(
				vecty.Class("slides", "layout-widescreen"),
			),
			elem.Article(
				vecty.Markup(
					vecty.MarkupIf(initPos.Class() != "",
						vecty.Class(initPos.Class())),
				),
				elem.Heading1(
					vecty.Text(s.doc.Title),
				),
				vecty.If(s.doc.Subtitle != "", elem.Heading3(
					vecty.Text(s.doc.Subtitle),
				)),
				vecty.If(!s.doc.Time.IsZero(), elem.Heading3(
					vecty.Text(s.doc.Time.Format(models.TimeFormat)),
				)),
				authors,
			),
			sections,
		),
	)
}
func (s *Slide) showSlide(n int) {
	s.activeSlide = n
	vecty.Rerender(s)
}

func (s *Slide) KeyPress(key string) {
	up := false
	switch key {
	case "ArrowRight", "ArrowUp":
		if s.activeSlide < len(s.doc.Sections) {
			s.activeSlide++
			up = true
		}
	case "ArrowLeft", "ArrowDown":
		if s.activeSlide != 0 {
			s.activeSlide--
			up = true
		}
	case "KeyR":
		if !s.recording {
			s.recording = true
			s.startTime = time.Now()
		} else {
			s.remote.length = time.Now().Sub(s.startTime)
			s.recording = false
		}
		up = true
	case "Space":
		if s.recording {
			s.remote.Add(s.activeSlide+1, time.Now().Sub(s.startTime))
		}
		if s.activeSlide < len(s.doc.Sections) {
			s.activeSlide++
			up = true
		}
	case "KeyP":
		if !s.auto {
			s.auto = true
			s.play()
		}
	default:
		println(key)
	}
	if up {
		vecty.Rerender(s)
	}
}

func (s *Slide) play() {
	go func() {
		start := time.Now()
		tick := time.NewTicker(time.Second)
		s.showSlide(0)
		for {
			select {
			case next := <-tick.C:
				dur := next.Sub(start)
				if dur > s.remote.length {
					tick.Stop()
					s.auto = false
					return
				}
				sec := int(dur.Seconds())
				if e, ok := s.remote.events[sec]; ok {
					s.showSlide(e.Slide)
				}
			}
		}
	}()
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

type RemoteControl struct {
	length time.Duration
	events map[int]TickEvent
}

func (r *RemoteControl) Add(n int, duration time.Duration) {
	e := TickEvent{
		Time:  duration,
		Slide: n,
	}
	r.events[int(duration.Seconds())] = e
}

type TickEvent struct {
	Time  time.Duration
	Slide int
}

func (t TickEvent) String() string {
	return fmt.Sprintf("%d|%v", t.Slide, int(t.Time.Seconds()))
}

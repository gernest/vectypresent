package slide

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty/event"

	"github.com/gopherjs/vecty/prop"

	"github.com/gernest/CatAcademy/present/models"
	"github.com/gernest/xhr"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

type position int

const (
	farPast position = iota << 1
	past
	current
	next
	farNext
	silent
)

func (p position) Class() string {
	switch p {
	case farPast:
		return "far-past"
	case past:
		return "past"
	case current:
		return "current"
	case next:
		return "next"
	case farNext:
		return "far-next"
	case silent:
		return ""
	default:
		return ""
	}
}

type Slide struct {
	vecty.Core

	doc         *models.Doc
	activeSlide int
	remote      *RemoteControl
	recording   bool
	auto        bool
	startTime   time.Time
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
	s.remote = &RemoteControl{
		events: make(map[int]TickEvent),
	}
	u.Path = filepath.Join("/slide", u.Path)
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
		vecty.Rerender(s)
	}()

}

func (s *Slide) OnMessage(data []byte) {

}

func getPos(active, n int) position {
	switch n {
	case active - 2:
		return farPast
	case active - 1:
		return past
	case active:
		return current
	case active + 1:
		return next
	case active + 2:
		return farNext
	default:
		return silent
	}
}
func (s *Slide) Render() vecty.ComponentOrHTML {
	if s.doc == nil {
		return elem.Body()
	}
	var sections vecty.List
	for i, section := range s.doc.Sections {
		pos := getPos(s.activeSlide, i+1)
		sections = append(sections, &Section{s: section, Pos: pos})
	}
	var authors vecty.List
	for _, author := range s.doc.Authors {
		authors = append(authors, elem.Div(
			vecty.Markup(vecty.Class("presenter")),
			renderElems(author.Elem),
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

// Section is a single slide show unit/page.
type Section struct {
	vecty.Core

	Pos position `vecty:"prop"`
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

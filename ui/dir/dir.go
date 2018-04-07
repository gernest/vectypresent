package dir

import (
	"github.com/gernest/vectypresent/present/models"
	"github.com/gernest/vectypresent/ui/router"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Dir struct {
	vecty.Core

	Dir *models.File `vecty:"prop"`

	Router *router.Router
}

func (d *Dir) Render() vecty.ComponentOrHTML {
	var dirList []*models.File
	var slideList []*models.File
	var filesList []*models.File
	for _, child := range d.Dir.Children {
		switch {
		case child.IsDir:
			dirList = append(dirList, child)
		case child.IsSlide():
			slideList = append(slideList, child)
		default:
			filesList = append(filesList, child)
		}
	}
	var list vecty.List
	if len(slideList) > 0 {
		list = append(list, elem.Heading4(
			vecty.Text("Slide decks:"),
		))
		for _, child := range slideList {
			url := child.URL()
			list = append(list, elem.Description(
				elem.Description(
					elem.Anchor(
						vecty.Markup(
							prop.Href(child.URL()),
							event.Click(func(e *vecty.Event) {
								d.Router.PushState(url)
							}).PreventDefault(),
						),
						vecty.Text(child.BaseName()),
					),
				),
			))
		}
	}
	if len(dirList) > 0 {
		list = append(list, elem.Heading4(
			vecty.Text("Sub-directories:"),
		))
		for _, child := range dirList {
			url := child.URL()
			list = append(list, elem.Description(
				elem.Description(
					elem.Anchor(
						vecty.Markup(
							prop.Href(child.URL()),
							event.Click(func(e *vecty.Event) {
								d.Router.PushState(url)
							}).PreventDefault(),
						),
						vecty.Text(child.BaseName()),
					),
				),
			))
		}
	}
	if len(filesList) > 0 {
		list = append(list, elem.Heading4(
			vecty.Text("Files:"),
		))
		for _, child := range filesList {
			url := child.URL()
			list = append(list, elem.Description(
				elem.Description(
					elem.Anchor(
						vecty.Markup(
							prop.Href(child.URL()),
							event.Click(func(e *vecty.Event) {
								d.Router.PushState(url)
							}).PreventDefault(),
						),
						vecty.Text(child.BaseName()),
					),
				),
			))
		}
	}
	return elem.Div(
		vecty.Markup(
			prop.ID("page"),
		),
		elem.Heading2(vecty.Text(d.Dir.BaseName())),
		elem.DescriptionList(list),
	)
}

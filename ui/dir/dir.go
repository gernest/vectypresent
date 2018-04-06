package dir

import (
	"github.com/gernest/CatAcademy/present/models"
	"github.com/gernest/CatAcademy/ui/router"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Dir struct {
	vecty.Core

	Dir    *models.File `vecty:prop"`
	Router *router.Router
}

func (d *Dir) Render() vecty.ComponentOrHTML {
	var list vecty.List
	for _, child := range d.Dir.Children {
		url := child.URL()
		if child.IsDir && child.Children == nil {
			continue
		}
		list = append(list, elem.Description(
			elem.Description(
				elem.Anchor(
					vecty.Markup(
						prop.Href(child.URL()),
						event.Click(func(e *vecty.Event) {
							println(url)
							d.Router.PushState(url)
						}).PreventDefault(),
					),
					vecty.Text(child.BaseName()),
				),
			),
		))
	}
	return elem.Div(
		vecty.Markup(
			prop.ID("page"),
		),
		elem.DescriptionList(list),
	)
}

package present

import (
	"errors"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/gernest/vectypresent/present/models"
)

func init() {
	Register("html", parseHTML)
}

func parseHTML(ctx *Context, fileName string, lineno int, text string) (models.Elem, error) {
	p := strings.Fields(text)
	if len(p) != 2 {
		return nil, errors.New("invalid .html args")
	}
	name := filepath.Join(filepath.Dir(fileName), p[1])
	b, err := ctx.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return models.HTML{HTML: template.HTML(b)}, nil
}

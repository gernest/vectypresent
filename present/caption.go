// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package present

import (
	"strings"

	"github.com/gernest/CatAcademy/present/models"
)

func init() {
	Register("caption", parseCaption)
}

func parseCaption(_ *Context, _ string, _ int, text string) (models.Elem, error) {
	text = strings.TrimSpace(strings.TrimPrefix(text, ".caption"))
	return models.Caption{Text: text}, nil
}

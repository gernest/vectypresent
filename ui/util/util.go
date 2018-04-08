package util

import (
	"github.com/gopherjs/gopherjs/js"
)

func ListSheets(f func(*js.Object) bool) {
	sheets := js.Global.Get("document").Get("styleSheets")
	length := sheets.Get("length").Int()
	for i := 0; i < length; i++ {
		sheet := sheets.Index(i)
		if !f(sheet) {
			return
		}
	}
}

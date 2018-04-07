//+build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func Build() error {
	return sh.RunV("go", "build", "-o", "vpresent")
}

const pkg = "github.com/gernest/vectypresent"

func Ui() error {
	return sh.RunV("gopherjs", "build", "-o", "static/ui.js", pkg+"/ui")
}

func Serve() error {
	if err := Build(); err != nil {
		return err
	}
	return sh.RunV("./vpresent", "serve")
}

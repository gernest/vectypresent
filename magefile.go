//+build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func Graph() error {
	files := []struct {
		src, output string
	}{
		{"workflow.mmd", "workflow.png"},
		{"services.mmd", "services.png"},
	}
	for _, file := range files {
		if err := sh.RunV("mmdc", "-t", "forest", "-i", file.src, "-o", file.output); err != nil {
			return err
		}
	}
	return nil
}

func Build() error {
	return sh.RunV("go", "build", "-o", "catac")
}

const pkg = "github.com/gernest/CatAcademy"

func Ui() error {
	return sh.RunV("gopherjs", "build", "-o", "static/ui.js", pkg+"/ui")
}

func Serve() error {
	if err := Build(); err != nil {
		return err
	}
	return sh.RunV("./catac", "serve")
}

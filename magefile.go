//+build mage

package main

import (
	"github.com/magefile/mage/sh"
)

const manifestFile = "docker-manifest.yaml"

func Build() error {
	return sh.RunV("go", "build", "-o", "vpresent")
}

const pkg = "github.com/gernest/vectypresent"

func Ui() error {
	return sh.RunV("gopherjs", "build", "-m", "-o", "static/ui.js", pkg+"/ui")
}

func Serve() error {
	if err := Build(); err != nil {
		return err
	}
	return sh.RunV("./vpresent", "serve")
}

func Release() error {
	return sh.RunV("goreleaser", "--rm-dist")
}

func Manifest() error {
	return sh.RunV("manifest-tool",
		"--username", "$DOCKERHUB_USER",
		"--password", "$DOCKERHUB_PASSWORD",
		"push", "from-spec", manifestFile)
}

func Assets() error {
	return sh.RunV("go-bindata", "-o", "data/assets.gen.go",
		"-pkg", "data", "-prefix", "static/", "static/...",
	)
}

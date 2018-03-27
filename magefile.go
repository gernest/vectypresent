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

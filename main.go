package main

import (
	"fmt"
	"os"

	"github.com/gernest/CatAcademy/generator"
	"github.com/gernest/CatAcademy/server"
	"github.com/urfave/cli"
)

func main() {
	a := cli.NewApp()
	a.Name = "catac"
	a.Usage = "simple programming lessons companion"
	a.Commands = []cli.Command{
		static.Command(),
		server.Command(),
	}
	if err := a.Run(os.Args); err != nil {
		fmt.Printf("catac: %v\n", err)
		os.Exit(1)
	}
}

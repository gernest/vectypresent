package main

import (
	"fmt"
	"os"

	"github.com/gernest/CatAcademy/static"
	"github.com/urfave/cli"
)

func main() {
	a := cli.NewApp()
	a.Name = "catac"
	a.Description = "learn programming like a boss"
	a.Commands = []cli.Command{
		static.Command(),
	}

	if err := a.Run(os.Args); err != nil {
		fmt.Printf("catac: %v\n", err)
		os.Exit(1)
	}
}

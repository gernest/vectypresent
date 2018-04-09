package main

import (
	"fmt"
	"os"

	"github.com/gernest/vectypresent/server"
	"github.com/urfave/cli"
)

func main() {
	a := cli.NewApp()
	a.Name = "vectypresent"
	a.Usage = "present with vecty frontend"
	a.Commands = []cli.Command{
		server.Command(),
	}
	if err := a.Run(os.Args); err != nil {
		fmt.Printf("catac: %v\n", err)
		os.Exit(1)
	}
}

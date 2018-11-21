package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rk_mft"
	app.Usage = "RK multi File Transferer"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		uploadCommand(),
	}
	app.Run(os.Args)
}

package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rk_mft"
	app.Usage = "RK multi File Transferer Server"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		StartServerCommand(),
	}
	app.Run(os.Args)
}

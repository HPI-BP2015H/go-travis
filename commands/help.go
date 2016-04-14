package commands

import (
	"github.com/fatih/color"
	"github.com/mislav/go-utils/cli"
)

func init() {
	cli.Register("help", helpCmd)
}

func helpCmd(cmd *cli.Cmd) {
	println("Usage: travis COMMAND ...\n ")
	println("Available commands: \n ")

	print("	branches	")
	color.Yellow("shows the branches of your current repository")
	print("	history 	")
	color.Yellow("does sdfgh")
	print("	login   	")
	color.Yellow("authenticates against the API and stores the token")
	print("	logout  	")
	color.Yellow("deletes the stored API token")
	print("	whoami  	")
	color.Yellow("outputs the current user")

}
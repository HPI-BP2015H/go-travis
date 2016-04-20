package commands

import (
	"github.com/HPI-BP2015H/go-travis/config"
	"github.com/HPI-BP2015H/go-utils/cli"
)

func init() {
	cli.AppInstance().RegisterCommand(
		cli.Command{
			Name:     "logout",
			Help:     "deletes the stored API token",
			Function: logoutCmd,
		},
	)
}

func logoutCmd(cmd *cli.Cmd) int {
	if NotLoggedIn(cmd) {
		return 1
	}
	env := cmd.Env.(config.TravisCommandConfig)
	user, _ := CurrentUser(env.Client)
	env.Config.DeleteTravisTokenForEndpoint(env.Endpoint)
	cmd.Stdout.Cprintf("%C(boldgreen)%s%C(reset)%C(green) is now logged out.%C(reset)\n", user.Name)
	return 0
}

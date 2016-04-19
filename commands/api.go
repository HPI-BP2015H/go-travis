package commands

import (
	"io"
	"strings"

	"github.com/HPI-BP2015H/go-travis/client"
	"github.com/HPI-BP2015H/go-utils/cli"
	"github.com/HPI-BP2015H/go-utils/utils"
)

func init() {
	cmd1 := cli.Command{
		Name:     "api",
		Help:     "makes an (authenticated) API call and prints out the raw result",
		Function: apiCmd,
	}
	cmd2 := cli.Command{
		Name:     "raw",
		Help:     "alias for api",
		Function: apiCmd,
	}
	flag := cli.Flag{
		Short: "-i",
		Long:  "--include-headers",
		Ftype: false,
	}

	cmd1.RegisterFlag(flag)
	cmd2.RegisterFlag(flag)

	cli.AppInstance().RegisterCommand(cmd1)
	cli.AppInstance().RegisterCommand(cmd2)

}

func checkUnusedArgs(cmd *cli.Cmd, args *cli.Args) {
	if args.Length() > 0 {
		cmd.Stderr.Printf("error: unrecognized arguments %q\n", args)
		cmd.Exit(1)
	}
}

func apiCmd(cmd *cli.Cmd) {
	path := ""
	args := cmd.Args
	if args.Length() > 0 {
		path, args = args.Shift()
	}

	if path == "manifest" {
		showResource := ""
		if args.Length() > 0 {
			showResource, args = args.Shift()
		}
		checkUnusedArgs(cmd, args)

		showManifest(cmd, showResource)
		return
	} else if path == "" {
		cmd.Stderr.Println("error: missing PATH argument for request")
		cmd.Exit(1)
	} else {
		checkUnusedArgs(cmd, args)
	}

	res, err := client.Travis().PerformRequest("GET", path, nil, nil)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if cmd.Flags.Bool("--include-headers") {
		cmd.Stdout.Printf("%s %s\r\n", res.Proto, res.Status)
		for name, values := range res.Header {
			value := strings.Join(values, ",")
			cmd.Stdout.Printf("%s: %s\r\n", name, value)
		}
		cmd.Stdout.Print("\r\n")
	}

	if res.StatusCode < 300 {
		io.Copy(cmd.Stdout, res.Body)
	} else {
		io.Copy(cmd.Stderr, res.Body)
		cmd.Exit(1)
	}
}

func showManifest(cmd *cli.Cmd, showResource string) {
	manifest, _ := client.Travis().Manifest()

	if showResource == "" {
		for _, resource := range manifest.AllResources() {
			actionNames := utils.NewSet()
			for _, action := range resource.AllActions() {
				actionNames.Add(action.Name)
			}
			cmd.Stdout.Print(resource.Name)
			if actionNames.Length() > 0 {
				cmd.Stdout.Printf(": %s", strings.Join(actionNames.Values(), ", "))
			}
			cmd.Stdout.Print("\n")
		}
	} else {
		resource := manifest.Resource(showResource)
		if resource == nil {
			cmd.Stderr.Printf("error: could not find the %q resource\n", showResource)
			cmd.Exit(1)
		} else {
			cmd.Stdout.Cprintf("%C(blue)ATTRIBUTES%C(reset) %v\n", resource.Attributes)
			cmd.Stdout.Cprintf("%C(blue)SORTABLE%C(reset) %v\n", resource.SortableBy)
			if resource.DefaultSort != "" {
				cmd.Stdout.Cprintf("%C(blue)SORTABLE%C(reset) default %1\n", resource.DefaultSort)
			}
			for _, action := range resource.AllActions() {
				cmd.Stdout.Cprintf("%C(blue)ACTION%C(reset) %s %s %s\n", action.Name, action.RequestMethod, action.UriTemplate)
			}
		}
	}
}

package commands

import (
	"os"

	"github.com/HPI-BP2015H/go-travis/client"
	"github.com/HPI-BP2015H/go-utils/cli"
	"github.com/fatih/color"
)

func init() {
	cli.Register("branches", "displays the most recent build for each branch", branchesCmd)
}

type Branches struct {
	Branches []Branch `json:"branches"`
}

type Branch struct {
	Name       string      `json:"name"`
	LastBuild  *Build      `json:"last_build"`
	Repository *Repository `json:"repo"`
}

func branchesCmd(cmd *cli.Cmd) {
	params := map[string]string{
		"repository.slug": os.Getenv("TRAVIS_REPO"),
	}

	res, err := client.Travis().PerformAction("branches", "find", params)
	if err != nil {
		panic(err)
	}
	if res.StatusCode > 299 {
		cmd.Stderr.Printf("Unexpected HTTP status: %d\n", res.StatusCode)
		cmd.Exit(1)
	}

	branches := Branches{}
	res.Unmarshal(&branches)

	for _, branch := range branches.Branches {
		printBranch(branch)
	}
}

func printBranch(branch Branch) {
	color.Yellow("%s:  ", branch.Name)
	c := color.New(color.FgRed, color.Bold).PrintfFunc()
	if branch.LastBuild.State == "passed" {
		c = color.New(color.FgGreen, color.Bold).PrintfFunc()
	}
	c("  #%s  %s\n", branch.LastBuild.Number, branch.LastBuild.State)
}

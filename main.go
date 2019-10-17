package main

import (
	"fmt"

	"github.com/allocine/postgresql-streamer-go/cmd"
	"github.com/allocine/postgresql-streamer-go/utils"
)

func main() {
	printVersion()
	cmd.Execute()
}

func printVersion() {
	fmt.Printf(
		"PostgreSQL Streamer CLI (%s-%s) (Go %s)\n",
		utils.GetVersion(),
		utils.GetCommit(),
		utils.GetRuntimeVersion(),
	)
}

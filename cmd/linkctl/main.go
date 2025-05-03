package main

import (
	"fmt"

	"git.omicron.one/omicron/linkshare/internal/util"
)

func main() {
	paths, err := util.FindDirectories("")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Paths:")
	fmt.Println("  Schema:", paths.SchemaDir)
	fmt.Println("  Database:", paths.DatabaseFile)
}

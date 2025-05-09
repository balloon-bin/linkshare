package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func configPreRun(cmd *cobra.Command, args []string) error {
	return setupDb()
}

func configPostRun(cmd *cobra.Command, args []string) error {
	return cleanupDb()
}

func configSetHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented")
}

func configGetHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented")
}

func configListHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented")
}

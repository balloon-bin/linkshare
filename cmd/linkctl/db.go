package main

import (
	"fmt"

	"git.omicron.one/omicron/linkshare/internal/database"
	"git.omicron.one/omicron/linkshare/internal/version"
	"github.com/spf13/cobra"
)

func dbPreRun(cmd *cobra.Command, args []string) error {
	return setupDb()
}

func dbPostRun(cmd *cobra.Command, args []string) error {
	return cleanupDb()
}

func dbInitHandler(cmd *cobra.Command, args []string) {
	err := db.Initialize(paths.SchemaDir)
	if err == database.ErrAlreadyInitialized {
		fmt.Printf("Database %q is already initialized\n", dbPath)
		return
	}
	if err == nil {
		fmt.Printf("Initialized database %q with schema version %d\n", dbPath, version.SchemaVersion)
		return
	}

	fmt.Printf("Failed to initialize database %q: %v\n", dbPath, err)
}

func dbBackupHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented")
}

func dbUpdateHandler(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented")
}

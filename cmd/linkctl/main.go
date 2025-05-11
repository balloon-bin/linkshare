package main

import (
	"fmt"
	"os"

	"git.omicron.one/omicron/linkshare/internal/database"
	"git.omicron.one/omicron/linkshare/internal/util"
	"git.omicron.one/omicron/linkshare/internal/version"
	"github.com/spf13/cobra"
)

var (
	dbPath    string
	verbosity int
)

var (
	paths *util.AppPaths
	db    *database.DB
)

func exitIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupPaths() error {
	if paths != nil {
		return nil
	}
	paths_, err := util.FindDirectories(dbPath)
	if err != nil {
		return err
	}
	paths = paths_
	return nil
}

func setupDb() error {
	if db != nil {
		return nil
	}
	err := setupPaths()
	if err != nil {
		return err
	}

	db_, err := database.Open(dbPath)
	if err != nil {
		return err
	}
	db = db_
	return nil
}

func cleanupDb() error {
	if db != nil {
		err := db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "linkctl",
		Short: "LinkShare CLI tool",
		Long:  `Command line tool to manage your self-hosted LinkShare service.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&dbPath, "db", "d", "", "Database file path")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity level")

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPreRunE:  configPreRun,
		PersistentPostRunE: configPostRun,
	}

	configSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set a configuration value",
		Run:   configSetHandler,
	}

	configGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get a configuration value",
		Run:   configGetHandler,
	}

	configListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Run:   configListHandler,
	}

	configCmd.AddCommand(configSetCmd, configGetCmd, configListCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPreRunE:  dbPreRun,
		PersistentPostRunE: dbPostRun,
	}

	dbInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the database",
		Run:   dbInitHandler,
	}

	dbBackupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup the database",
		Run:   dbBackupHandler,
	}

	dbUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the database schema",
		Run:   dbUpdateHandler,
	}

	dbCmd.AddCommand(dbInitCmd, dbBackupCmd, dbUpdateCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.Print()
		},
	}

	rootCmd.AddCommand(configCmd, dbCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourorg/drp/internal/config"
	"github.com/yourorg/drp/internal/db"
	"github.com/yourorg/drp/internal/output"
)

var dbTablesCmd = &cobra.Command{
	Use:   "db:tables",
	Short: "List all tables in the configured database",
	RunE: func(c *cobra.Command, _ []string) error {
		cfg, conn, err := connectDB()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		tables, err := db.TableNames(conn.DB, conn.DriverName)
		if err != nil {
			output.Fail("Could not list tables: %v", err)
			return err
		}
		if len(tables) == 0 {
			output.Info("No tables found in database %q", cfg.DBName)
			return nil
		}
		for _, t := range tables {
			fmt.Println(" ", t)
		}
		return nil
	},
}

var dbStatusCmd = &cobra.Command{
	Use:   "db:status",
	Short: "Show database connection status and basic info",
	RunE: func(c *cobra.Command, _ []string) error {
		cfg, conn, err := connectDB()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		if err := db.Ping(conn.DB); err != nil {
			output.Fail("%v", err)
			return err
		}
		output.Success("Connected to %q on %s:%s (%s)", cfg.DBName, cfg.DBHost, cfg.DBPort, cfg.DBDriver)
		return nil
	},
}

var dbDropCmd = &cobra.Command{
	Use:   "db:drop",
	Short: "Drop all tables in the configured database",
	RunE: func(c *cobra.Command, _ []string) error {
		dryRun, _ := c.Flags().GetBool("dry-run")
		cfg, conn, err := connectDB()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		tables, err := db.TableNames(conn.DB, conn.DriverName)
		if err != nil {
			output.Fail("Could not list tables: %v", err)
			return err
		}
		if len(tables) == 0 {
			output.Info("No tables to drop in %q", cfg.DBName)
			return nil
		}

		if dryRun {
			output.Warn("[dry-run] Would drop %d table(s) in %q:", len(tables), cfg.DBName)
			for _, t := range tables {
				output.Info("  - %s", t)
			}
			return nil
		}

		if err := db.DropAllTables(conn.DB, conn.DriverName); err != nil {
			output.Fail("%v", err)
			return err
		}
		output.Success("Dropped %d table(s) from %q", len(tables), cfg.DBName)
		return nil
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "db:reset",
	Short: "Drop all tables and re-run all migrations from scratch",
	RunE: func(c *cobra.Command, args []string) error {
		return migrateFreshCmd.RunE(c, args)
	},
}

// connectDB is a shared helper for db:* commands.
func connectDB() (*config.Config, *db.Connection, error) {
	cfg, err := config.Load(globalEnvFile)
	if err != nil {
		output.Fail("%v", err)
		return nil, nil, err
	}
	conn, err := db.Connect(cfg)
	if err != nil {
		output.Fail("%v", err)
		return nil, nil, err
	}
	return cfg, conn, nil
}

func init() {
	rootCmd.AddCommand(dbTablesCmd, dbStatusCmd, dbDropCmd, dbResetCmd)
	dbDropCmd.Flags().Bool("dry-run", false, "Print tables that would be dropped without executing")
	dbResetCmd.Flags().Bool("dry-run", false, "Print what would be dropped/applied without executing")
}

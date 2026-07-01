package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/DarpanAdhikari/drp-go-cli/internal/config"
	"github.com/DarpanAdhikari/drp-go-cli/internal/db"
	"github.com/DarpanAdhikari/drp-go-cli/internal/migration"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
)

// migrationsDir is the conventional migrations directory inside a drp project.
const migrationsDir = "database/migrations"

// migrateCmd is the parent command; running `drp migrate` is equivalent to
// `drp migrate:up`.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending database migrations",
	RunE:  runMigrateUp,
}

var migrateUpCmd = &cobra.Command{
	Use:   "migrate:up",
	Short: "Run all pending migrations",
	RunE:  runMigrateUp,
}

func runMigrateUp(c *cobra.Command, _ []string) error {
	engine, conn, err := newMigrationEngine()
	if err != nil {
		return err
	}
	defer conn.DB.Close()

	if err := engine.Up(); err != nil {
		if errors.Is(err, migration.ErrNothingToMigrate) {
			output.Info(err.Error())
			return nil
		}
		output.Fail("Migration failed: %v", err)
		return err
	}
	output.Success("Migrations ran successfully")
	return nil
}

var migrateCreateCmd = &cobra.Command{
	Use:   "migrate:create [name]",
	Short: "Create a new migration file pair (up/down)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		f, err := migration.NewFile(migrationsDir, args[0])
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		output.Success("Created %s", f.UpPath)
		output.Success("Created %s", f.DownPath)
		return nil
	},
}

var migrateRollbackCmd = &cobra.Command{
	Use:   "migrate:rollback",
	Short: "Roll back the last batch of migrations",
	RunE: func(c *cobra.Command, _ []string) error {
		dryRun, _ := c.Flags().GetBool("dry-run")
		engine, conn, err := newMigrationEngine()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		if dryRun {
			output.Info("[dry-run] Would roll back the latest migration batch")
			return nil
		}
		if err := engine.Rollback(); err != nil {
			if errors.Is(err, migration.ErrNothingToRollback) {
				output.Info(err.Error())
				return nil
			}
			output.Fail("Rollback failed: %v", err)
			return err
		}
		output.Success("Rolled back successfully")
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "migrate:down",
	Short: "Roll back a single migration step",
	RunE: func(c *cobra.Command, _ []string) error {
		dryRun, _ := c.Flags().GetBool("dry-run")
		engine, conn, err := newMigrationEngine()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		if dryRun {
			output.Info("[dry-run] Would roll back the single latest migration")
			return nil
		}
		if err := engine.Down(); err != nil {
			if errors.Is(err, migration.ErrNothingToRollback) {
				output.Info(err.Error())
				return nil
			}
			output.Fail("Step-down failed: %v", err)
			return err
		}
		output.Success("Stepped down one migration")
		return nil
	},
}

var migrateFreshCmd = &cobra.Command{
	Use:   "migrate:fresh",
	Short: "Drop all tables and re-run all migrations from scratch",
	RunE: func(c *cobra.Command, _ []string) error {
		dryRun, _ := c.Flags().GetBool("dry-run")
		engine, conn, err := newMigrationEngine()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		if err := engine.Fresh(dryRun); err != nil {
			output.Fail("Fresh migration failed: %v", err)
			return err
		}
		if !dryRun {
			output.Success("Database refreshed and all migrations applied")
		}
		return nil
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "migrate:status",
	Short: "Show which migrations have run and which are pending",
	RunE: func(c *cobra.Command, _ []string) error {
		engine, conn, err := newMigrationEngine()
		if err != nil {
			return err
		}
		defer conn.DB.Close()

		entries, err := engine.Status()
		if err != nil {
			output.Fail("Could not read migration status: %v", err)
			return err
		}
		print(migration.FormatStatus(entries))
		return nil
	},
}

var migrateSeedCmd = &cobra.Command{
	Use:   "migrate:seed",
	Short: "Run pending migrations, then run database seeders",
	RunE: func(c *cobra.Command, _ []string) error {
		if err := runMigrateUp(c, nil); err != nil {
			return err
		}
		fresh, _ := c.Flags().GetBool("fresh")
		return runSeeders(fresh)
	},
}

// newMigrationEngine is a shared helper that loads config and returns a ready
// Engine plus the open Connection (caller must close conn.DB).
func newMigrationEngine() (*migration.Engine, *db.Connection, error) {
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
	output.Success("Connected to %s (%s)", cfg.DBName, cfg.DBDriver)

	return migration.NewEngine(conn, migrationsDir), conn, nil
}

func init() {
	rootCmd.AddCommand(
		migrateCmd,
		migrateCreateCmd,
		migrateUpCmd,
		migrateRollbackCmd,
		migrateDownCmd,
		migrateFreshCmd,
		migrateStatusCmd,
		migrateSeedCmd,
	)

	migrateRollbackCmd.Flags().Bool("dry-run", false, "Print what would be rolled back without executing")
	migrateDownCmd.Flags().Bool("dry-run", false, "Print what would be stepped down without executing")
	migrateFreshCmd.Flags().Bool("dry-run", false, "Print what would be dropped/applied without executing")
	migrateSeedCmd.Flags().Bool("fresh", false, "Clear seed history and re-run all seeders")
}

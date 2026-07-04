package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/DarpanAdhikari/drp-go-cli/internal/config"
	"github.com/DarpanAdhikari/drp-go-cli/internal/db"
	"github.com/DarpanAdhikari/drp-go-cli/internal/migration"
	"github.com/DarpanAdhikari/drp-go-cli/internal/output"
	"github.com/DarpanAdhikari/drp-go-cli/internal/seeder"
)

const seedersDir = "database/seeders"

var seederCreateCmd = &cobra.Command{
	Use:   "seeder:create [name]",
	Short: "Create a new seeder file",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		f, err := seeder.NewFile(seedersDir, args[0])
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		output.Success("Created %s", f.Path)
		return nil
	},
}

var dbSeedCmd = &cobra.Command{
	Use:   "db:seed",
	Short: "Run database seeders",
	RunE: func(c *cobra.Command, _ []string) error {
		fresh, _ := c.Flags().GetBool("fresh")
		return runSeeders(fresh)
	},
}

// runSeeders is shared by db:seed and migrate:seed.
func runSeeders(fresh bool) error {
	cfg, err := config.Load(globalEnvFile)
	if err != nil {
		output.Fail("%v", err)
		return err
	}
	conn, err := db.Connect(cfg)
	if err != nil {
		output.Fail("%v", err)
		return err
	}
	defer conn.DB.Close()

	if fresh {
		output.Info("Dropping all tables and re-running migrations...")
		if err := db.DropAllTables(conn.DB, conn.DriverName); err != nil {
			output.Fail("Failed to drop tables: %v", err)
			return err
		}
		if err := db.EnsureSchemaHistoryTable(conn.DB, conn.DriverName); err != nil {
			output.Fail("Failed to create schema_history: %v", err)
			return err
		}
		migEngine := migration.NewEngine(conn, migrationsDir)
		if err := migEngine.Up(); err != nil && !errors.Is(err, migration.ErrNothingToMigrate) {
			output.Fail("Migration failed: %v", err)
			return err
		}
	}

	engine := seeder.NewEngine(conn, seedersDir)
	if err := engine.Run(fresh); err != nil {
		if errors.Is(err, seeder.ErrNoSeeders) || errors.Is(err, seeder.ErrNothingToSeed) {
			output.Info(err.Error())
			return nil
		}
		output.Fail("Seeding failed: %v", err)
		return err
	}
	output.Success("Database seeded successfully")
	return nil
}

var seederStatusCmd = &cobra.Command{
	Use:   "seeder:status",
	Short: "Show which seeders have run and which are pending",
	RunE: func(c *cobra.Command, _ []string) error {
		cfg, err := config.Load(globalEnvFile)
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		conn, err := db.Connect(cfg)
		if err != nil {
			output.Fail("%v", err)
			return err
		}
		defer conn.DB.Close()

		engine := seeder.NewEngine(conn, seedersDir)
		entries, err := engine.Status()
		if err != nil {
			output.Fail("Could not read seeder status: %v", err)
			return err
		}
		print(seeder.FormatStatus(entries))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(seederCreateCmd, dbSeedCmd, seederStatusCmd)
	dbSeedCmd.Flags().Bool("fresh", false, "Drop all tables, re-run migrations, then re-run all seeders from scratch")
}

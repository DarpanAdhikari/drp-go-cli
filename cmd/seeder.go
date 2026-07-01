package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/DarpanAdhikari/drp-go-cli/internal/config"
	"github.com/DarpanAdhikari/drp-go-cli/internal/db"
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

func init() {
	rootCmd.AddCommand(seederCreateCmd, dbSeedCmd)
	dbSeedCmd.Flags().Bool("fresh", false, "Clear seed history and re-run all seeders")
}

package cmd

import (
	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/entities"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Execute database migration files in the migrations directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := db.New("fasim.db")
		if err != nil {
			return err
		}

		return db.RunMigrations(entities.GetModels()...)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

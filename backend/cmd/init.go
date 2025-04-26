package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fasim/backend/internal/repositories/db"
	"github.com/fasim/backend/internal/repositories/entities"
)

var initDBCmd = &cobra.Command{
	Use:   "init-db",
	Short: "Initialize the database",
	Long: `Initialize the SQLite database with required tables.
This command will create a new database file if it doesn't exist
and run all necessary migrations to set up the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// データベース接続を作成
		database, err := db.New("fasim.db")
		if err != nil {
			return fmt.Errorf("failed to create database connection: %w", err)
		}

		// マイグレーションを実行
		if err := database.RunMigrations(entities.GetModels()...); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Println("Database initialized successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initDBCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
	"path"
	"tg_todo_bot/config"
	"tg_todo_bot/kernel/db"
	zap_logger "tg_todo_bot/kernel/logger"
	"tg_todo_bot/kernel/migrate"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all 'up' migrations",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap_logger.InitLogger()
		logger.Infof("Start execute '%s %s' command", cmd.Parent().Name(), cmd.Name())

		conf, err := config.GetConfig()
		if err != nil {
			logger.Panicw("config.GetMigrateConfig()", "error", err.Error())
		}

		pg := db.NewPG(
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Database,
			conf.Database.User,
			conf.Database.Password,
		)

		pgUrl := pg.GetConnUrl()
		migrationDir := path.Join(".", "migrations")
		m, err := migrate.NewPostgresMigrateInstance(pgUrl, migrationDir)
		if err != nil {
			logger.Panicw("migrate.NewPostgresMigrateInstance(pgUrl, migrationDir)",
				"error", err.Error(), "pgUrl", pgUrl, "migrationDir", migrationDir,
			)
		}

		err = m.Up()
		if err != nil {
			if err.Error() != "no change" {
				logger.Panicw("m.Up()", "error", err.Error())
			}
		}
	},
}

func init() {
	migrateCmd.AddCommand(upCmd)
}

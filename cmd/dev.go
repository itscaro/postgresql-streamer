package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/allocine/postgresql-streamer-go/services/dev"
	"github.com/allocine/postgresql-streamer-go/services/postgresql"
	"github.com/google/martian/log"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/cobra"
)

func createDevCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "dev",
		Short:             "",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: rootCmdPreRun,
		RunE:              devCmdE,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			printMemory()
		},
	}

	return cmd
}

func devCmdE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("need at least one argument")
	}
	if err := postgresql.InitPostgreSQLSettings(settings.DBDsn(), rootCmdOpts.logLevel >= LogLevelDebug); err != nil {
		return err
	}
	db := postgresql.Connect()
	switch args[0] {
	case "init-db":
		db.AutoMigrate(&dev.TestModel{})
	case "reinit-db":
		db.DropTableIfExists(&dev.TestModel{})
		db.AutoMigrate(&dev.TestModel{})
	case "gendata":
		if len(args) < 2 {
			return fmt.Errorf("Give number of entries to generate as second argument")
		}
		if nbEntries, err := strconv.Atoi(args[1]); err == nil {
			for i := 0; i < nbEntries; i++ {
				db.Create(&dev.TestModel{
					Source: postgres.Jsonb{
						RawMessage: json.RawMessage(`{"name": "test-source"}`),
					},
					Content: postgres.Jsonb{
						RawMessage: json.RawMessage(fmt.Sprintf(`{"name": "test-content-%d"}`, i)),
					},
				})
			}
			if errs := db.GetErrors(); len(errs) > 0 {
				for err := range errs {
					log.Errorf("Error while generating data: %s", err)
				}
			}
		}
	case "gendata-trx":
		if len(args) < 2 {
			return fmt.Errorf("Give number of entries to generate as second argument")
		}
		if nbEntries, err := strconv.Atoi(args[1]); err == nil {
			trx := db.Begin()
			for i := 0; i < nbEntries; i++ {
				trx.Create(&dev.TestModel{
					Source: postgres.Jsonb{
						RawMessage: json.RawMessage(`{"name": "test-source"}`),
					},
					Content: postgres.Jsonb{
						RawMessage: json.RawMessage(fmt.Sprintf(`{"name": "test-content-%d"}`, i)),
					},
				})
			}
			trx.Commit()
			if errs := trx.GetErrors(); len(errs) > 0 {
				for err := range errs {
					log.Errorf("Error while generating data: %s", err)
				}
			}
		}
	}

	return nil
}

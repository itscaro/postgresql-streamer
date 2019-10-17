package postgresql

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createSlotCmd(t *testing.T) {
	var tests = []struct {
		dsn               string
		slotName          string
		expectedCmdString string
		expectedPassword  string
	}{
		{
			"postgres://test-user@test-hostname/test_db?sslmode=disable",
			"testing_slot_name",
			"pg_recvlogical -h test-hostname -U test-user -d test_db --slot testing_slot_name --create-slot --if-not-exists -P wal2json",
			"",
		},
		{
			"postgres://test-user1:test-pass1@test-hostname1/test_db1?sslmode=disable",
			"testing_slot_name1",
			"pg_recvlogical -h test-hostname1 -U test-user1 -d test_db1 --slot testing_slot_name1 --create-slot --if-not-exists -P wal2json",
			"PGPASSWORD=test-pass1",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			_ = InitPostgreSQLSettings(tt.dsn, false)

			cmd := createSlotCmd(tt.slotName)
			assert.Equal(t, tt.expectedCmdString, strings.Join(cmd.Args, " "))
			if len(tt.expectedPassword) != 0 {
				assert.Contains(t, cmd.Env, tt.expectedPassword)
			}
		})
	}
}

func Test_dropSlotCmd(t *testing.T) {
	var tests = []struct {
		dsn               string
		slotName          string
		expectedCmdString string
		expectedPassword  string
	}{
		{
			"postgres://test-user@test-hostname/test_db?sslmode=disable",
			"testing_slot_name",
			"pg_recvlogical -h test-hostname -U test-user -d test_db --slot testing_slot_name --drop-slot",
			"",
		},
		{
			"postgres://test-user1:test-pass1@test-hostname1/test_db1?sslmode=disable",
			"testing_slot_name1",
			"pg_recvlogical -h test-hostname1 -U test-user1 -d test_db1 --slot testing_slot_name1 --drop-slot",
			"PGPASSWORD=test-pass1",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			_ = InitPostgreSQLSettings(tt.dsn, false)

			cmd := dropSlotCmd(tt.slotName)
			assert.Equal(t, tt.expectedCmdString, strings.Join(cmd.Args, " "))
			if len(tt.expectedPassword) != 0 {
				assert.Contains(t, cmd.Env, tt.expectedPassword)
			}
		})
	}
}

func Test_ListenSlotCmd(t *testing.T) {
	var tests = []struct {
		dsn              string
		slotName         string
		environ          map[string]string
		expectedCmd      []string
		expectedPassword string
	}{
		{
			"postgres://test-user@test-hostname/test_db?sslmode=disable",
			"testing_slot_name",
			map[string]string{},
			strings.Split("pg_recvlogical -h test-hostname -U test-user -d test_db --slot testing_slot_name --start -f - -o include-types=0 -o include-timestamp=1", " "),
			"",
		},
		{
			"postgres://test-user1:test-pass1@test-hostname1/test_db1?sslmode=disable",
			"testing_slot_name1",
			map[string]string{
				"PGS_WAL2JSON_INCLUDE_TIMESTAMP": "0",
				"PGS_WAL2JSON_ADD_TABLES":        "public.*",
			},
			strings.Split("pg_recvlogical -h test-hostname1 -U test-user1 -d test_db1 --slot testing_slot_name1 --start -f - -o add-tables=public.* -o include-types=0 -o include-timestamp=0", " "),
			"PGPASSWORD=test-pass1",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			for k, v := range tt.environ {
				os.Setenv(k, v)
			}
			_ = InitPostgreSQLSettings(tt.dsn, false)
			InitWal2JsonSettings()
			cmd := ListenSlotCmd(tt.slotName)
			assert.ElementsMatch(t, tt.expectedCmd, cmd.Args)
			if len(tt.expectedPassword) != 0 {
				assert.Contains(t, cmd.Env, tt.expectedPassword)
			}
			resetWal2JsonSettings()
			for k, _ := range tt.environ {
				os.Unsetenv(k)
			}
		})
	}
}

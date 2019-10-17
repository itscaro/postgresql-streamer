package postgresql

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetWal2JsonSettings() {
	for k, _ := range Wal2JsonSettings {
		Wal2JsonSettings[k] = nil
	}
}

func TestInitWal2JsonSettings_DefaultValues(t *testing.T) {
	os.Setenv("PGS_WAL2JSON_PRETTY_PRINT", "1")
	os.Setenv("PGS_WAL2JSON_INCLUDE_TYPE_OIDS", "1")
	InitWal2JsonSettings()
	assert.Equal(t, "1", Wal2JsonSettings["include-type-oids"])
	assert.Equal(t, "1", Wal2JsonSettings["pretty-print"])
	assert.Equal(t, "0", Wal2JsonSettings["include-types"])
	assert.Equal(t, "1", Wal2JsonSettings["include-timestamp"])
	os.Unsetenv("PGS_WAL2JSON_PRETTY_PRINT")
	os.Unsetenv("PGS_WAL2JSON_INCLUDE_TYPE_OIDS")
	resetWal2JsonSettings()
}

func TestInitWal2JsonSettings_ModifiedDefaultValues(t *testing.T) {
	os.Setenv("PGS_WAL2JSON_INCLUDE_TYPES", "1")
	os.Setenv("PGS_WAL2JSON_INCLUDE_TIMESTAMP", "0")
	InitWal2JsonSettings()
	assert.Equal(t, nil, Wal2JsonSettings["include-type-oids"])
	assert.Equal(t, "1", Wal2JsonSettings["include-types"])
	assert.Equal(t, "0", Wal2JsonSettings["include-timestamp"])
	os.Unsetenv("PGS_WAL2JSON_INCLUDE_TYPES")
	os.Unsetenv("PGS_WAL2JSON_INCLUDE_TIMESTAMP")
	resetWal2JsonSettings()
}

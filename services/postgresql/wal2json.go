package postgresql

import (
	"fmt"
	"os"
	"strings"
)

type Wal2Json struct {
	Change    []*Wal2JsonChange `json:"change"`
	Timestamp string            `json:"timestamp,omitempty"`
}

type Wal2JsonChange struct {
	Timestamp    string           `json:"timestamp,omitempty"`
	Kind         string           `json:"kind"`
	Schema       string           `json:"schema"`
	Table        string           `json:"table"`
	ColumnNames  []interface{}    `json:"columnnames,omitempty"`
	ColumnValues []interface{}    `json:"columnvalues,omitempty"`
	OldKeys      *Wal2JsonOldKeys `json:"oldkeys,omitempty"`
}

type Wal2JsonOldKeys struct {
	KeyNames  []interface{} `json:"keynames,omitempty"`
	KeyValues []interface{} `json:"keyvalues,omitempty"`
}

var Wal2JsonSettings = map[string]interface{}{
	"include-xids":        nil,
	"include-timestamp":   nil,
	"include-schemas":     nil,
	"include-types":       nil,
	"include-typmod":      nil,
	"include-type-oids":   nil,
	"include-not-null":    nil,
	"pretty-print":        nil,
	"write-in-chunks":     nil,
	"include-lsn":         nil,
	"filter-tables":       nil,
	"add-tables":          nil,
	"filter-msg-prefixes": nil,
	"add-msg-prefixes":    nil,
	"format-version":      nil,
}

func InitWal2JsonSettings() {
	for k, _ := range Wal2JsonSettings {
		val := os.Getenv(fmt.Sprintf("PGS_WAL2JSON_%s", strings.ToUpper(strings.ReplaceAll(k, "-", "_"))))
		if len(val) != 0 {
			Wal2JsonSettings[k] = val
		}
	}

	// Overrides some default settings
	if Wal2JsonSettings["include-types"] == nil {
		Wal2JsonSettings["include-types"] = "0"
	}
	if Wal2JsonSettings["include-timestamp"] == nil {
		Wal2JsonSettings["include-timestamp"] = "1"
	}
}

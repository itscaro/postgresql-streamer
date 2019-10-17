package postgresql

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

type postgreSQLSettings struct {
	debug                 bool
	dsn                   string
	hostname              string
	port                  string
	dbname                string
	username              string
	password              string
	pgRecvLogicalBaseArgs []string
}

var postgreSQL postgreSQLSettings

func InitPostgreSQLSettings(DBDsn string, debug bool) error {
	// Reset
	postgreSQL = postgreSQLSettings{}

	postgreSQL.dsn = DBDsn
	postgreSQL.debug = debug

	u, err := url.Parse(postgreSQL.dsn)
	if err != nil {
		return err
	}
	postgreSQL.hostname = u.Hostname()
	postgreSQL.port = u.Port()
	if u.User != nil {
		postgreSQL.username = u.User.Username()
		postgreSQL.password, _ = u.User.Password()
	}
	if u.Path != "" {
		postgreSQL.dbname = u.Path[1:]
	}

	if len(postgreSQL.hostname) != 0 {
		postgreSQL.pgRecvLogicalBaseArgs = []string{
			"-h",
			postgreSQL.hostname,
		}
	}
	if len(postgreSQL.port) != 0 {
		postgreSQL.pgRecvLogicalBaseArgs = append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"-p",
			postgreSQL.port,
		)
	}
	if len(postgreSQL.username) != 0 {
		postgreSQL.pgRecvLogicalBaseArgs = append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"-U",
			postgreSQL.username,
		)
	}
	if len(postgreSQL.dbname) != 0 {
		postgreSQL.pgRecvLogicalBaseArgs = append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"-d",
			postgreSQL.dbname,
		)
	}

	log.Debugf("Using DB DSN %s\n", postgreSQL.dsn)

	return nil
}

func createSlotCmd(slotName string) (cmd *exec.Cmd) {
	log.WithFields(log.Fields{
		"slot": slotName,
	}).Warn("Creating slot if not exists")

	cmd = exec.Command(
		"pg_recvlogical",
		append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"--slot",
			slotName,
			"--create-slot",
			"--if-not-exists",
			"-P",
			"wal2json",
		)...,
	)
	setCommandEnviron(cmd)

	return
}

func CreateSlot(slotName string) error {
	return execute("createSlotCmd", slotName)
}

func ListenSlotCmd(slotName string) *exec.Cmd {
	log.WithFields(log.Fields{
		"slot": slotName,
	}).Warn("Listen to slot")

	cmd := exec.Command(
		"pg_recvlogical",
		append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"--slot",
			slotName,
			"--start",
			"-f",
			"-",
		)...,
	)
	for option, value := range Wal2JsonSettings {
		if value != nil {
			// TODO escape characters
			cmd.Args = append(
				cmd.Args,
				"-o",
				fmt.Sprintf("%s=%s", option, value.(string)),
			)
		}
	}
	setCommandEnviron(cmd)
	return cmd
}

func CleanSlot(slotName string) error {
	log.WithFields(log.Fields{
		"slot": slotName,
	}).Warn("Cleaning slot")

	if err := terminateRunningQueries(slotName); err != nil {
		log.Error(err)
		return err
	} else {
		if err := execute("dropSlotCmd", slotName); err != nil {
			return err
		}
	}

	return nil
}

func dropSlotCmd(slotName string) (cmd *exec.Cmd) {
	log.WithFields(log.Fields{
		"slot": slotName,
	}).Warn("Drop slot")

	cmd = exec.Command(
		"pg_recvlogical",
		append(
			postgreSQL.pgRecvLogicalBaseArgs,
			"--slot",
			slotName,
			"--drop-slot",
		)...,
	)
	setCommandEnviron(cmd)

	return cmd
}

func execute(funcName string, slotName string) (err error) {
	handlers := map[string]func(slotName string) (cmd *exec.Cmd){
		"dropSlotCmd":   dropSlotCmd,
		"createSlotCmd": createSlotCmd,
	}

	if f, ok := handlers[funcName]; ok {
		output, err := f(slotName).CombinedOutput()
		if err != nil {
			log.Error(string(output))
		}
	} else {
		err = fmt.Errorf("function '%s' does not exist", funcName)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"slot": slotName,
		}).Error(err)
	}

	return
}

func terminateRunningQueries(slotName string) error {
	db := Connect()
	errs := db.Exec(`
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity pa
INNER JOIN pg_replication_slots pr ON pa.pid=pr.active_pid
WHERE pr.slot_name=?
`,
		slotName,
	).GetErrors()

	if len(errs) > 0 {
		return errs[len(errs)-1]
	}
	return nil
}

func Connect() *gorm.DB {
	db, err := gorm.Open("postgres", postgreSQL.dsn)
	if err != nil {
		log.Errorf("failed to connect database")
		os.Exit(1)
	}
	db.SingularTable(true)
	db.LogMode(postgreSQL.debug)

	return db
}

func setCommandEnviron(cmd *exec.Cmd) {
	cmd.Env = os.Environ()
	if len(postgreSQL.password) != 0 {
		cmd.Env = append(
			cmd.Env,
			"PGPASSWORD="+postgreSQL.password,
		)
	}
	log.Debugf("Command: %s", strings.Join(cmd.Args, " "))
}

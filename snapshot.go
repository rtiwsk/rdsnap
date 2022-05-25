package rdsnap

import (
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

var (
	Info  *log.Logger
	Error *log.Logger
)

type config struct {
	instanceId string
	engine     string
	host       string
	port       int64
	user       string
	password   string
	dbtables   []dbTable
	wlog       io.Writer
	logFlag    int
}

type dbTable struct {
	db    string
	table string
}

func SetConfig(instance, engine, user, password, tables string, wlog io.Writer, logFlag int) config {
	c := config{}
	c.instanceId = instance
	c.engine = engine
	c.wlog = wlog
	c.logFlag = logFlag

	if tables != "" {
		c.user = user
		c.password = password

		for _, dbtbl := range strings.Split(tables, ",") {
			dt := strings.Split(dbtbl, ".")
			c.dbtables = append(c.dbtables, dbTable{db: dt[0], table: dt[1]})
		}
	}

	return c
}

func Run(cfg config, svc rdsiface.RDSAPI) (err error) {
	Info = log.New(cfg.wlog, "rdsnap: ", cfg.logFlag)
	Error = log.New(cfg.wlog, "rdsnap: error: ", cfg.logFlag)

	rc := rdsClient{svc: svc}

	Info.Println("Start creating a DB snapshot.")

	snapshotId, err := rc.createDBSnapshot(cfg.instanceId)
	if err != nil {
		Error.Println(err)
		return
	}
	Info.Printf("Created DB snapshot: %s\n", snapshotId)

	if len(cfg.dbtables) == 0 {
		return
	}

	rescfg, err := rc.restoreDBInstanceFromDBSnapshot(cfg, snapshotId)
	if err != nil {
		Error.Println(err)
		if err = rc.deleteDBSnapshot(snapshotId); err != nil {
			Error.Println(err)
		}
		return
	}
	Info.Printf("Restored DB instance: %s\n", rescfg.instanceId)

	var db *db
	for _, dbtable := range rescfg.dbtables {
		db, err = connectDB(
			rescfg.engine,
			rescfg.user,
			rescfg.password,
			rescfg.host,
			dbtable.db,
			rescfg.port,
		)
		if err != nil {
			Error.Println(err)
			continue
		}

		if err = db.ping(); err != nil {
			Error.Println(err)
			continue
		}

		if err = db.truncateTable(dbtable.table); err != nil {
			Error.Println(err)
			continue
		}
		Info.Printf("Truncated the table: %s.%s\n", dbtable.db, dbtable.table)

		db.close()
	}

	if err == nil {
		resSnapshotId, err := rc.createDBSnapshot(rescfg.instanceId)
		if err != nil {
			Error.Println(err)
		}
		Info.Printf("Created DB snapshot: %s\n", resSnapshotId)
	}

	err = rc.deleteDBInstance(rescfg.instanceId)
	if err != nil {
		Error.Println(err)
	}
	Info.Printf("Delete DB instance: %s\n", rescfg.instanceId)

	err = rc.deleteDBSnapshot(snapshotId)
	if err != nil {
		Error.Println(err)
	}
	Info.Printf("Delete DB snapshot: %s\n", snapshotId)

	return
}

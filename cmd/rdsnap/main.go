package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/rtiwsk/rdsnap"
)

var help = `Usage: 
  rdsnap [options...]

Options:
  -instance    Amazon RDS DB Instance ID. (Not support Aurora)
               Required option.
  -engine      DB Engine. Select from 'mysql' and 'postgres'.
               Required option.
  -log         Logfile path. Default is Stdout.
  -user        DB Username. Specify when truncating tables.
  -password    DB Password. Specify when truncating tables.
  -tables      DB tables. Specify when truncating tables.
               table as <database>.<tablename>.
               e.g. company.sales,company.account
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, help)
	}

	instance := flag.String("instance", "", "")
	engine := flag.String("engine", "", "")
	logFile := flag.String("log", "", "")
	user := flag.String("user", "", "")
	password := flag.String("password", "", "")
	tables := flag.String("tables", "", "")

	flag.Parse()

	var (
		f       = os.Stdout
		err     error
		logFlag int
	)

	if *logFile != "" {
		f, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()

		logFlag = log.LstdFlags
	}

	if err := validateRequiredOption(*instance, *engine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := validateDBEngineType(*engine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := validateTruncateTableOption(*user, *password, *tables); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := rdsnap.SetConfig(*instance, *engine, *user, *password, *tables, f, logFlag)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := rds.New(sess)

	if err := rdsnap.Run(cfg, svc); err != nil {
		os.Exit(1)
	}
}

func validateRequiredOption(instance, engine string) error {
	if instance == "" || engine == "" {
		return fmt.Errorf("-instance and -engine is required.")
	}

	return nil
}

func validateDBEngineType(engine string) error {
	engineTypes := []string{"mysql", "postgres"}
	for _, engineType := range engineTypes {
		if engineType == engine {
			return nil
		}
	}

	return fmt.Errorf("Select the correct engine.")
}

func validateTruncateTableOption(user, password, tables string) error {
	if user != "" && password != "" && tables != "" {
		return nil
	} else if user != "" || password != "" || tables != "" {
		return fmt.Errorf("Insufficient options to truncate DB tables.")
	}

	return nil
}

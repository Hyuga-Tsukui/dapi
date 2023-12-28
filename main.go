package main

import (
	"dapi/internal/aurora"
	"dapi/internal/tui"
	"log"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"
)

var (
	flags       = flag.NewFlagSet("dapi", flag.ExitOnError)
	resourceArn = flags.String("resource-arn", "", "RDS resource ARN")
	secretArn   = flags.String("secret-arn", "", "RDS secret ARN")
	dbname      = flags.StringP("dbname", "d", "", "RDS database name")
	region      = flags.String("region", "ap-northeast-1", "AWS region")
	help        = flags.BoolP("help", "h", false, "Show help")
)

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("failed to parse flags: %v", err)
		return
	}

	if *help {
		flags.PrintDefaults()
		return
	}

	db, err := aurora.New(
		*resourceArn,
		*secretArn,
		*dbname,
		*region,
	)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close db: %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close db: %v", err)
		}
		os.Exit(0)
	}()

	startApp(db)
}

func startApp(aurora *aurora.DataSource) {
	t := tui.New(aurora)
	if err := t.Run(); err != nil {
		panic(err)
	}
}

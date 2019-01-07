package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/alash3al/go-color"
	"github.com/jmoiron/sqlx"
)

func init() {
	flag.Parse()

	if _, err := sqlx.Connect(*flagDBDriver, *flagDBDSN); err != nil {
		fmt.Println(color.RedString("[%s] - connection error - (%s)", *flagDBDriver, err.Error()))
		os.Exit(0)
	}

	manager, err := NewManager(*flagConfigFile)
	if err != nil {
		fmt.Println(color.RedString("(%s)", err.Error()))
		os.Exit(0)
	}

	macrosManager = manager
}

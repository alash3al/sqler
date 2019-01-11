// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bwmarrin/snowflake"

	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/kshvakov/clickhouse"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/alash3al/go-color"
	"github.com/jmoiron/sqlx"
)

func init() {
	usage := flag.Usage
	flag.Usage = func() {
		fmt.Println(color.MagentaString(sqlerBrand))
		usage()
	}

	flag.Parse()
	runtime.GOMAXPROCS(*flagWorkers)

	{
		tstconn, err := sqlx.Connect(*flagDBDriver, *flagDBDSN)
		if err != nil {
			fmt.Println(color.RedString("[%s] %s - connection error - (%s)", *flagDBDriver, *flagDBDSN, err.Error()))
			os.Exit(0)
		}
		tstconn.Close()
	}

	{
		manager, err := NewManager(*flagAPIFile)
		if err != nil {
			fmt.Println(color.RedString("(%s)", err.Error()))
			os.Exit(0)
		}
		macrosManager = manager
	}

	{
		var err error

		snow, err = snowflake.NewNode(1)
		if err != nil {
			fmt.Println(color.RedString("(%s)", err.Error()))
			os.Exit(0)
		}
	}

	if *flagREDISAddr != "" {
		var err error

		cacher, err = NewCacher(*flagREDISAddr)
		if err != nil {
			fmt.Println(color.RedString("redis - (%s)", err.Error()))
			os.Exit(0)
		}
	}
}

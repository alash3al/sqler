// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
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

	manager, err := NewManager(*flagAPIFile)
	if err != nil {
		fmt.Println(color.RedString("(%s)", err.Error()))
		os.Exit(0)
	}

	macrosManager = manager
}

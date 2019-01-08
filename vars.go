// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"flag"
)

var (
	flagDBDriver   = flag.String("engine", "mysql", "the sql engine/driver to be used")
	flagDBDSN      = flag.String("dsn", "root:root@tcp(127.0.0.1)/test?multiStatements=true", "the data source name for the selected engine")
	flagAPIFile    = flag.String("api", "./api.example.hcl", "the validators used before processing the sql, it accepts a glob style pattern")
	flagListenAddr = flag.String("listen", ":8025", "the rest api listen address")
)

var (
	macrosManager *Manager
)

const (
	sqlerVersion = "v1.0"
	sqlerBrand   = `
	
			 ____   ___  _              
			/ ___| / _ \| |    ___ _ __ 
			\___ \| | | | |   / _ \ '__|
			 ___) | |_| | |__|  __/ |   
			|____/ \__\_\_____\___|_|   
											
	turn your SQL queries into safe valid RESTful apis.
	
	`
)

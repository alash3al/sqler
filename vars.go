// Copyright 2018 The SQLer Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"errors"
	"flag"
	"runtime"
)

var (
	flagDBDriver       = flag.String("driver", "mysql", "the sql driver to be used")
	flagDBDSN          = flag.String("dsn", "root:root@tcp(127.0.0.1)/test?multiStatements=true", "the data source name for the selected engine")
	flagAPIFile        = flag.String("config", "./config.example.hcl", "the config file(s) that contains your endpoints configs, it accepts comma seprated list of glob style pattern")
	flagRESTListenAddr = flag.String("rest", ":8025", "the http restful api listen address")
	flagRESPListenAddr = flag.String("resp", ":3678", "the resp (redis protocol) server listen address")
	flagWorkers        = flag.Int("workers", runtime.NumCPU(), "the maximum workers count")
)

var (
	errNoMacroFound       = errors.New("Resource not found")
	errValidationError    = errors.New("Validation error")
	errAuthorizationError = errors.New("Authorization Error")
)

var (
	errStatusCodeMap = map[error]int{
		errNoMacroFound:       404,
		errValidationError:    422,
		errAuthorizationError: 401,
	}
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

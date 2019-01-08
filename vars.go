package main

import (
	"flag"
)

var (
	flagDBDriver   = flag.String("engine", "sqlite3", "the sql engine/driver to be used")
	flagDBDSN      = flag.String("dsn", "./database.sqlite", "the data source name for the selected engine")
	flagConfigFile = flag.String("config", "./config.hcl", "the validators used before processing the sql, it accepts a glob style pattern")
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
											
	turn your SQL queries into safe valid RESTful apis

	
	`
)

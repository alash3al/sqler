SQLer
=====
> `SQL-er` is a tiny http server that applies the old `CGI` concept but for `SQL` queries, it enables you to an endpoint and assign a SQL query to be executed when anyone hits it, also it enables you to define validation rules so you can validate the request body/query params. `sqler` uses `nginx` style configuration language ([`HCL`](https://github.com/hashicorp/hcl)).

Features
========
- Standalone with no dependencies.
- Works with most of SQL databases out there including (`SQL Server`, `MYSQL`, `SQLITE`, `PostgreSQL`, `Cockroachdb`)
- Built-in Validators
- Built-in `sql escaper` function
- Uses ([`HCL`](https://github.com/hashicorp/hcl)) configuration language
- You can load multiple configuration files not just one, based on `unix glob` style pattern
- Each `SQL` query could be named as `Macro`
- You can use [`Go tex/template`](https://golang.org/pkg/text/template/) within each macro
- Each macro have its own `Context` (`query params` + `body params` + `helper functions`)

Configuration Overview
======================
```hcl
// create a macro/endpoint called "_boot",
// this macro is private "used within other macros" 
// because it starts with "_".
_boot {
    // the query we want to execute
    exec = <<SQL
        CREATE TABLE IF NOT EXISTS `users` (
            `ID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            `name` VARCHAR(30) DEFAULT "@anonymous",
            `email` VARCHAR(30) DEFAULT "@anonymous" 
        );
    SQL
}

// adduser macro/endpoint, just hit `/adduser` with
// a `?user_name=&user_email=` or json `POST` request
// with the same fields.
adduser {
    // what request method will this macro be called
    // default: ["ANY"]
    methods = ["POST"]

    // the validation rules
    // you can specifiy seprated rules for each request method!
    // validation rules uses this package: https://github.com/asaskevich/govalidator
    rules {
        user_name = ["required"]
        user_email =  ["required"]
    }

    // the query to be executed
    exec = <<SQL
        /* include the "_boot" macro */
        {{ template "_boot" }}

        INSERT INTO users(name, email) VALUES('{{ .Input.user_name | .SQLEscape }}', '{{ .Input.user_email | .SQLEscape }}');
        SELECT * FROM users WHERE id = LAST_INSERT_ID();
    SQL
}

// an endpoint for `GET /databases` method to display all databases
databases {
    exec = "SHOW DATABASES"
}
```

Supported SQL Engines
=====================
- `sqlite3`
- `mysql`
- `postgresql`
- `cockroachdb`
- `sqlserver`

Supported Validation Rules
==========================
- Simple Validations methods with no args: [here](https://godoc.org/github.com/asaskevich/govalidator#TagMap)
- Advanced Validations methods with args: [here](https://godoc.org/github.com/asaskevich/govalidator#ParamTagMap) 

License
========
> Copyright 2018 The SQLer Authors. All rights reserved.
> Use of this source code is governed by a Apache 2.0
> license that can be found in the [LICENSE](/License) file.


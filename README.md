SQLer
=====
> `SQL-er` is a tiny portable server enables you to write APIs using SQL query to be executed when anyone hits it, also it enables you to define validation rules so you can validate the request body/query params, as well as data transformation using simple `javascript` syntax. `sqler` uses `nginx` style configuration language ([`HCL`](https://github.com/hashicorp/hcl)) amd `Go` [`text/template`](https://golang.org/pkg/text/template).

Table Of Contents
=================
- [SQLer](#sqler)
- [Table Of Contents](#table-of-contents)
- [Features](#features)
- [Quick Tour](#quick-tour)
- [Configuration Overview](#configuration-overview)
- [Supported Validation Rules](#supported-validation-rules)
- [Supported Utils](#supported-utils)
- [REST vs RESP](#rest-vs-resp)
- [Data Transformation](#data-transformation)
- [Issue/Suggestion/Contribution ?](#issuesuggestioncontribution)
- [Author](#author)
- [License](#license)

Features
========
- Standalone with no dependencies.
- Works with most of SQL databases out there including (`SQL Server`, `MYSQL`, `SQLITE`, `PostgreSQL`, `Cockroachdb`)
- Built-in RESTful server
- Built-in RESP `Redis Protocol`, you connect to `SQLer` using any `redis` client
- Built-in `Javascript` interpreter to easily transform the result
- Built-in Validators
- Automatically uses prepared statements
- Uses ([`HCL`](https://github.com/hashicorp/hcl)) configuration language
- You can load multiple configuration files not just one, based on `unix glob` style pattern
- Each `SQL` query could be named as `Macro`
- You can use `Go` [`text/template`](https://golang.org/pkg/text/template/) within each macro
- Each macro has its own `Context` (`query params` + `body params`) as `.Input` which is `map[string]interface{}`, and `.Utils` which is a list of helper functions, currently it contains only `SQLEscape`.
- You can define `authorizers`, an `authorizer` is just a simple webhook that enables `sqler` to verify whether the request should be done or not.

Quick Tour
==========
- You install `sqler` using the right binary for your `os` from the [releases](https://github.com/alash3al/sqler/releases) page.
- Let's say that you downloaded `sqler_darwin_amd64`
- Let's rename it to `sqler`, and copy it to `/usr/local/bin`
- Now just run `sqler -h`, you will the next  
```bash
                         ____   ___  _
                        / ___| / _ \| |    ___ _ __
                        \___ \| | | | |   / _ \ '__|
                         ___) | |_| | |__|  __/ |
                        |____/ \__\_\_____\___|_|

        turn your SQL queries into safe valid RESTful apis.


  -config string
        the config file(s) that contains your endpoints configs, it accepts comma seprated list of glob style pattern (default "./config.example.hcl")
  -driver string
        the sql driver to be used (default "mysql")
  -dsn string
        the data source name for the selected engine (default "root:root@tcp(127.0.0.1)/test?multiStatements=true")
  -resp string
        the resp (redis protocol) server listen address (default ":3678")
  -rest string
        the http restful api listen address (default ":8025")
  -workers int
        the maximum workers count (default 4)
```
- you can specifiy multiple files for `-config` as [configuration](#configuration-overview), i.e `-config="/my/config/dir/*.hcl,/my/config/dir2/*.hcl"`
- you need specify which driver you need and its `dsn` from the following:

| Driver                 | DSN |
---------| ----|
| `mysql`| `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
| `postgres`| `postgresql://username:password@server:port/dbname?option1=value1`|
| `sqlite3`| `/path/to/db.sqlite?option1=value1`|
| `sqlserver` | `sqlserver://username:password@host/instance?param1=value&param2=value` |
|             | `sqlserver://username:password@host:port?param1=value&param2=value`|
| `tidb`| `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
|             | `sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30`|
| `mssql` | `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`|
|         | `server=localhost;user id=sa;database=master;app name=MyAppName`|
|         | `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName` |
|         | `odbc:server=localhost;user id=sa;database=master;app name=MyAppName` |
| `hdb` (SAP HANA) |   `hdb://user:password@host:port` |
| `clickhouse` (Yandex ClickHouse) |   `tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000` |


Configuration Overview
======================
```hcl
// create a macro/endpoint called "_boot",
// this macro is private "used within other macros" 
// because it starts with "_".
// this rule only used within `RESTful` context.
_boot {
    // the query we want to execute
    exec = <<SQL
        CREATE TABLE IF NOT EXISTS `users` (
            `ID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            `name` VARCHAR(30) DEFAULT "@anonymous",
            `email` VARCHAR(30) DEFAULT "@anonymous",
            `password` VARCHAR(200) DEFAULT "",
            `time` INT UNSIGNED
        );
    SQL
}

// adduser macro/endpoint, just hit `/adduser` with
// a `?user_name=&user_email=` or json `POST` request
// with the same fields.
adduser {
    // what request method will this macro be called
    // default: ["ANY"]
    // this only used within `RESTful` context.
    methods = ["POST"]

    // authorizers,
    // sqler will attempt to send the incoming authorization header
    // to the provided endpoint(s) as `Authorization`,
    // each endpoint MUST return `200 OK` so sqler can continue, other wise,
    // sqler will break the request and return back the client with the error occurred.
    // each authorizer has a method and a url.
    // this only used within `RESTful` context.
    authorizers = ["GET http://web.hook/api/authorize", "GET http://web.hook/api/allowed?roles=admin,root,super_admin"]

    // the validation rules
    // you can specify separated rules for each request method!
    rules {
        user_name = ["required"]
        user_email =  ["required", "email"]
        user_password = ["required", "stringlength: 5,50"]
    }

    // the query to be executed
    exec = <<SQL
       {{ template "_boot" }}

        /* let's bind a vars to be used within our internal prepared statement */
        {{ .BindVar "name" .Input.user_name }}
        {{ .BindVar "email" .Input.user_email }}
        {{ .BindVar "emailx" .Input.user_email }}

        INSERT INTO users(name, email, password, time) VALUES(
            /* we added it above */
            :name,

            /* we added it above */
            :email,

            /* it will be secured anyway because it is encoded */
            '{{ .Input.user_password | .Hash "bcrypt" }}',

            /* generate a unix timestamp "seconds" */
            {{ .UnixTime }}
        );

        SELECT * FROM users WHERE id = LAST_INSERT_ID();
    SQL
}

// list all databases, and run a transformer function
databases {
    exec = "SHOW DATABASES"

    transformer = <<JS
        // there is a global variable called `$result`,
        // `$result` holds the result of the sql execution.
        (function(){
            newResult = []

            for ( i in $result ) {
                newResult.push($result[i].Database)
            }

            return newResult
        })()
    JS
}

```

Supported Validation Rules
==========================
- Simple Validations methods with no args: [here](https://godoc.org/github.com/asaskevich/govalidator#TagMap)
- Advanced Validations methods with args: [here](https://godoc.org/github.com/asaskevich/govalidator#ParamTagMap) 

Supported Utils
===============
- `.Hash <method>` - hash the specified input using the specified method [md5, sha1, sha256, sha512, bcrypt], `{{ "data" | .Hash "md5" }}`
- `.UnixTime` - returns the unix time in seconds, `{{ .UnixTime }}`
- `.UnixNanoTime` - returns the unix time in nanoseconds, `{{ .UnixNanoTime }}`
- `.Uniqid` - returns a unique id, `{{ .Uniqid }}`


REST vs RESP
=============
> RESTful server could be used to interact directly with i.e `mobile, browser, ... etc`, in this mode `SQLer` is protected by `authorizers`, which gives you the abbility to check authorization against another 3rd-party api.  
> Each macro you add to the configuration file(s) you can access to it by issuing a http request to `/<macro-name>`, every query param and json body will be passed to the macro `.Input`.

> RESP server is just a basic `REDIS` compatible server, you connect to it using any `REDIS` client out there, even `redis-cli`, just open `redis-cli -p 3678 list` to list all available macros (`commands`), you can execute any macro as a redis command and pass the arguments as a json encoded data, i.e `redis-cli -p 3678 adduser "{\"user_name\": \"u6\", \"user_email\": \"email@tld.com\", \"user_password\":\"pass@123\"}"`.

Data Transformation
====================
> In some cases we need to transform the resulted data into something more friendly to our API consumers, so I added `javascript` interpreter to `SQLer` so we can transform our data, each js code has a global variable called `$result`, it holds the result of the `exec` section, you should write your code like the following:

```hcl
// list all databases, and run a transformer function
databases {
    exec = "SHOW DATABASES"

    transformer = <<JS
        // there is a global variable called `$result`,
        // `$result` holds the result of the sql execution.
        (function(){
            newResult = []

            for ( i in $result ) {
                newResult.push($result[i].Database)
            }

            return newResult
        })()
    JS
}
```

Issue/Suggestion/Contribution ?
===============================
`SQLer` is your software, feel free to open an issue with your feature(s), suggestions, ... etc, also you can easily contribute even you aren't a `Go` developer, you can write wikis it is open for all, let's make `SQLer` more powerful.

Author
=======
> I'm Mohamed Al Ashaal, just a problem solver :), you can view more projects from me [here](https://github.com/alash3al), and here is my email [m7medalash3al@gmail.com](mailto:m7medalash3al@gmail.com)

License
========
> Copyright 2019 The SQLer Authors. All rights reserved.
> Use of this source code is governed by a Apache 2.0
> license that can be found in the [LICENSE](/License) file.


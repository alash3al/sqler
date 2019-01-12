SQLer
=====
> `SQL-er` is a tiny portable server enables you to write APIs using SQL query to be executed when anyone hits it, also it enables you to define validation rules so you can validate the request body/query params, as well as data transformation using simple `javascript` syntax. `sqler` uses `nginx` style configuration language ([`HCL`](https://github.com/hashicorp/hcl)) amd `javascript` engine for custom expressions.

Table Of Contents
=================
- [SQLer](#sqler)
- [Table Of Contents](#table-of-contents)
- [Features](#features)
- [Quick Tour](#quick-tour)
- [Supprted DBMSs](#supprted-dbmss)
- [Configuration Overview](#configuration-overview)
- [REST vs RESP](#rest-vs-resp)
- [Sanitization](#sanitization)
- [Validation](#validation)
- [Authorization](#authorization)
- [Data Transformation](#data-transformation)
- [Aggregators](#aggregators)
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
- Uses `Javascript` custom expressions.
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

| Driver | DSN |
---------| ------ |
| `mysql`| `usrname:password@tcp(server:port)/dbname?option1=value1&...`|
| `postgres`| `postgresql://username:password@server:port/dbname?option1=value1`|
| `sqlite3`| `/path/to/db.sqlite?option1=value1`|
| `sqlserver` | `sqlserver://username:password@host/instance?param1=value&param2=value` |
|             | `sqlserver://username:password@host:port?param1=value&param2=value`|
|             | `sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30`|
| `mssql` | `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`|
|         | `server=localhost;user id=sa;database=master;app name=MyAppName`|
|         | `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName` |
|         | `odbc:server=localhost;user id=sa;database=master;app name=MyAppName` |
| `hdb` (SAP HANA) |   `hdb://user:password@host:port` |
| `clickhouse` (Yandex ClickHouse) |   `tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000` |


Supprted DBMSs
===============
- `MYSQL`, `TiDB`, `MariaDB`, `Percona` and any MYSQL compatible server uses `mysql` driver.
- `PostgreSQL`, `CockroachDB` and any PostgreSQL compatible server uses `postgres` driver.
- `SQL Server`, `MSSQL`, `ADO`, `ODBC` uses `sqlserver` or `mssql` driver.
- `SQLITE`, uses `sqlite3` driver.
- `HANA` (SAP), uses `hdb` driver.
- `Clickhouse`, uses `clickhouse` driver.

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
    validators {
        user_name_is_empty = "$input.user_name && $input.user_name.trim().length > 0"
        user_email_is_empty = "$input.user_email && $input.user_email.trim(' ').length > 0"
        user_password_is_not_ok = "$input.user_password && $input.user_password.trim(' ').length > 5"
    }

    bind {
        name = "$input.user_name"
        email = "$input.user_email"
        password = "$input.user_password"
    }

    methods = ["POST"]

    authorizer = <<JS
        (function(){
            log("use this for debugging")
            token = $input.http_authorization
            response = fetch("http://requestbin.fullcontact.com/zxpjigzx", {
                headers: {
                    "Authorization": token
                }
            })
            if ( response.statusCode != 200 ) {
                return false
            }
            return true
        })()
    JS

    // include some macros we declared before
    include = ["_boot"]

    exec = <<SQL
        INSERT INTO users(name, email, password, time) VALUES(:name, :email, :password, UNIX_TIMESTAMP());
        SELECT * FROM users WHERE id = LAST_INSERT_ID();
    SQL
}

// list all databases, and run a transformer function
databases {
    exec = "SHOW DATABASES"
}

// list all tables from all databases
tables {
    exec = "SELECT `table_schema` as `database`, `table_name` as `table` FROM INFORMATION_SCHEMA.tables"
}

// a macro that aggregates `databases` macro and `tables` macro into one macro
databases_tables {
    aggregate = ["databases", "tables"]
}

```

REST vs RESP
=============
> RESTful server could be used to interact directly with i.e `mobile, browser, ... etc`, in this mode `SQLer` is protected by `authorizers`, which gives you the abbility to check authorization against another 3rd-party api.  
> Each macro you add to the configuration file(s) you can access to it by issuing a http request to `/<macro-name>`, every query param and json body will be passed to the macro `.Input`.

> RESP server is just a basic `REDIS` compatible server, you connect to it using any `REDIS` client out there, even `redis-cli`, just open `redis-cli -p 3678 list` to list all available macros (`commands`), you can execute any macro as a redis command and pass the arguments as a json encoded data, i.e `redis-cli -p 3678 adduser "{\"user_name\": \"u6\", \"user_email\": \"email@tld.com\", \"user_password\":\"pass@123\"}"`.

Sanitization
=============
> `SQLer` uses prepared statements, you can bind use inputs like the following:


```hcl
addpost {
    // $input is a global variable holds all request inputs,
    // including the http headers too (prefixed with `http_`)
    // all http header keys are normalized to be in this form 
    // `http_x_header_example`, `http_authorization` ... etc in lower case.
    bind {
        title = "$input.post_title"
        content = "$input.post_content"
        user_id = "$input.post_user"
    }

    exec = <<SQL
        INSERT INTO posts(user_id, title, content) VALUES(:user_id, :title, :content);
        SELECT * FROM posts WHERE id = LAST_INSERT_ID();
    SQL
}
```


Validation
===========
> Data validation is very easy in `SQLer`, it is all about simple `javascript` expression like this:

```hcl
addpost {
    // if any rule returns false, 
    // SQLer will return 422 code, with invalid rules.
    // 
    // $input is a global variable holds all request inputs,
    // including the http headers too (prefixed with `http_`)
    // all http header keys are normalized to be in this form 
    // `http_x_header_example`, `http_authorization` ... etc in lower case.
    validators {
        post_title_length = "$input.post_title && $input.post_title.trim().length > 0"
        post_content_length = "$input.post_content && $input.post_content.length > 0"
        post_user = "$input.post_user"
    }

    bind {
        title = "$input.post_title"
        content = "$input.post_content"
        user_id = "$input.post_user"
    }

    exec = <<SQL
        INSERT INTO posts(user_id, title, content) VALUES(:user_id, :title, :content);
        SELECT * FROM posts WHERE id = LAST_INSERT_ID();
    SQL
}
```

Authorization
=============
> If you want to expose `SQLer` as a direct api to API consumers, you will need to add an authorization layer on top of it, let's see how to do that

```hcl
addpost {
    authorizer = <<JS
        (function(){
            // $input is a global variable holds all request inputs,
            // including the http headers too (prefixed with `http_`)
            // all http header keys are normalized to be in this form 
            // `http_x_header_example`, `http_authorization` ... etc in lower case.
            token = $input.http_authorization
            response = fetch("http://requestbin.fullcontact.com/zxpjigzx", {
                headers: {
                    "Authorization": token
                }
            })
            if ( response.statusCode != 200 ) {
                return false
            }
            return true
        })()
    JS
}
```

> using that trick, you can use any third-party Authentication service that will remove that hassle from your code.

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

Aggregators
============
> `SQLer` helps you to merge multiple macros into one to minimize the API calls number, see the example bellow

```hcl
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

tables {
    exec = "SELECT `table_schema` as `database`, `table_name` as `table` FROM INFORMATION_SCHEMA.tables"
    transformer = <<JS
        (function(){
            $ret = {}
            for (  i in $result ) {
                if ( ! $ret[$result[i].database] ) {
                    $ret[$result[i].database] = [];
                }
                $ret[$result[i].database].push($result[i].table)
            }
            return $ret
        })()
    JS
}

databasesAndTables {
    aggregate {
        databases = "current_databases"
        tables = "current_tables"
    }
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


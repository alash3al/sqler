SQLer
=====
> `SQL-er` is a tiny http server that applies the old `CGI` concept but for `SQL` queries, it enables you to an endpoint and assign a SQL query to be executed when anyone hits it, also it enables you to define validation rules so you can validate the request body/query params. `sqler` uses `nginx` style configuration language ([`HCL`](https://github.com/hashicorp/hcl)).

Configuration Overview
======================
```hcl
// create a macro/endpoint called "_boot",
// this macro is private "used within other macros" 
// because it starts with "_".
_boot {
    // the query we want to execute
    exec = <<EOF
        CREATE TABLE IF NOT EXISTS `users` (
            `ID` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
            `name` VARCHAR(30) DEFAULT "@anonymous",
            `email` VARCHAR(30) DEFAULT "@anonymous" 
        );
    EOF
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
    exec = <<EOF
        /* include the "_boot" macro */
        {{ template "_boot" }}

        INSERT INTO users(name, email) VALUES('{{ .Input.user_name | .SQL }}', '{{ .Input.user_email | .SQL }}');
        SELECT * FROM users WHERE id = LAST_INSERT_ID();
    EOF
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
- `mssql`
// create a macro/endpoint called "_boot",
// this macro is private "used within other macros" 
// because it starts with "_".
_boot {
    // the query we want to execute
    exec = <<SQL
        CREATE TABLE IF NOT EXISTS datax (
            ID SERIAL PRIMARY KEY,
            data JSONB DEFAULT NULL
        );

        CREATE INDEX IF NOT EXISTS datax_gin_index ON datax USING GIN(data);
    SQL

    extend {
        data {}
        call {}
    }

    trigger {
        mail {

        }

        http {

        }

        call {
            macro = ""
            input = ""
        }
    }
}

addpost {
    include = ["_boot"]
    methods = ["POST"]

    validators {
        title_is_empty = "$input.title && $input.title.trim().length > 0"
        content_is_empty = "$input.content"
    }

    bind {
        data = <<JS
            JSON.stringify({
                "title": $input.title,
                "content": $input.content
            })
        JS
    }

    exec = <<SQL
        INSERT INTO datax(ID, data) VALUES(default, :data) RETURNING id, data;
    SQL
}

// adduser macro/endpoint, just hit `/adduser` with
// a `?user_name=&user_email=` or json `POST` request
// with the same fields.
adduser {
    validators {
        user_name_is_empty = "$input.user_name && $input.user_name.trim().length > 0"
        user_email_is_empty = "$input.user_email && $input.user_email.trim().length > 0"
        user_password_is_not_ok = "$input.user_password && $input.user_password.trim().length > 5"
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
    // include = ["_boot"]
    exec = "SHOW DATABASES"
    transformer = <<JS
        (function(){
            // $result
            $new = [];
            for ( i in $result ) {
                $new.push($result[i].Database)
            }
            return $new
        })()
    JS
}

// list all tables from all databases
tables {
    exec = "SELECT `table_name` as `table`, `table_schema` as `database` FROM INFORMATION_SCHEMA.tables"
    transformer = <<SQL
        (function(){
            $ret = []
            for ( i in $result ){
                $ret.push({
                    table: $result[i].table,
                    database: $result[i].database,
                })
            }
            return $ret
        })()
    SQL
}

data {
    bind {
        limit = 2
        field = "'id'"
    }
    
    exec = "SELECT id FROM data limit 5"
}

// a macro that aggregates `databases` macro and `tables` macro into one macro
databases_tables {
    aggregate = ["databases", "tables"]
}

vars {
    exec = "show variables"
}
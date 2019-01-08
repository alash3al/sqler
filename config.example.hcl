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
    rules {
        user_name = ["required"]
        user_email =  ["required"]
    }

    // the query to be executed
    exec = <<EOF
        {{ template "_boot" }}

        INSERT INTO users(name, email) VALUES('{{ .Input.user_name | .SQL }}', '{{ .Input.user_email | .SQL }}');
        SELECT * FROM users WHERE id = LAST_INSERT_ID();
    EOF
}

proclist {
    exec = "SHOW PROCESSLIST"
}

tables {
    exec = "SELECT * FROM information_schema.tables"
}

databases {
    exec = "SHOW DATABASES"
}
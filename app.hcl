connection mysql {
    dsn = "mysql://localhost/dbname"
}

query "users/list_all" {
    connection = "mysql"

    sql = "SELECT * FROM users"

    pagination {
        algo = "limit-offset"
        page_name = "page"
    }
}

http "GET" "/users/" {
    middlewares = []

    query = "users/list_all"

    validator {
        
    }
}
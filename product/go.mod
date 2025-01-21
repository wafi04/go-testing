module github.com/wafi04/go-testing/product

go 1.22

require (
    github.com/jmoiron/sqlx v1.4.0
    github.com/wafi04/go-testing/common v0.0.0
)

replace (
    github.com/wafi04/go-testing/auth => ../auth
    github.com/wafi04/go-testing/category => ../category
    github.com/wafi04/go-testing/product => ../product
    github.com/wafi04/go-testing/common => ../common
)
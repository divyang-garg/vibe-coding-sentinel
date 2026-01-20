module github.com/divyang-garg/sentinel-hub-api

go 1.24.1

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728
	github.com/lib/pq v1.10.9
	github.com/nguyenthenguyen/docx v0.0.0-20230621112118-9c8e795a11db
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.43.0
	sentinel-hub-api v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace sentinel-hub-api => ./hub/api

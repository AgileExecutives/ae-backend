module github.com/unburdy/invoice-module

go 1.24.5

require (
	github.com/ae-base-server v0.0.0
	github.com/gin-gonic/gin v1.10.0
	gorm.io/datatypes v1.2.5
	gorm.io/gorm v1.25.12
)

replace github.com/ae-base-server => ../../base-server

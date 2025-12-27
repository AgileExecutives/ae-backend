module github.com/unburdy/templates-module

go 1.24.5

require (
	github.com/ae-base-server/pkg/core v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/redis/go-redis/v9 v9.7.0
	gorm.io/datatypes v1.2.5
	gorm.io/gorm v1.25.12
)

replace github.com/ae-base-server/pkg/core => ../../base-server/pkg/core

replace github.com/ae-base-server/api => ../../base-server/api

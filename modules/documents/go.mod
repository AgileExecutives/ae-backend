module github.com/unburdy/documents-module

go 1.24.5

require (
	github.com/ae-base-server v0.0.0
	github.com/chromedp/cdproto v0.0.0-20241110205750-a72e6703cd9b
	github.com/chromedp/chromedp v0.11.2
	github.com/gin-gonic/gin v1.10.0
	github.com/minio/minio-go/v7 v7.0.82
	github.com/redis/go-redis/v9 v9.7.0
	github.com/unburdy/templates-module v0.0.0
	gorm.io/datatypes v1.2.5
	gorm.io/gorm v1.25.12
)

replace github.com/ae-base-server => ../../base-server

replace github.com/unburdy/templates-module => ../templates

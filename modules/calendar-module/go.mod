module github.com/ae-backend/calendar-module

go 1.24

require (
	github.com/ae-saas-basic/ae-saas-basic v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.10.1
	gorm.io/gorm v1.30.0
)

replace github.com/ae-saas-basic/ae-saas-basic => ../../base-server
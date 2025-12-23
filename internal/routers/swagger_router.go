package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupSwaggerRouter configures the Swagger documentation endpoint
func SetupSwaggerRouter(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
}

package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
    r.GET("/", HomeHandler)

    r.GET("/test" , func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"message": "TEST OK!"})
    })
}

func HomeHandler(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Welcome to Pokemon Showdown Draft League!"})
}

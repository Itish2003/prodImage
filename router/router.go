package router

import (
	"io"
	"log"
	"os"
	"prodImage/service"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	file, err := os.Create("logFile.log")
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(file, os.Stdout)
	router.Use(gin.Logger())
	router.Static("/css", "./css")
	router.Static("/static", "./static")
	router.LoadHTMLGlob("template/*")

	route := router.Group("/v1")
	{
		route.PATCH("/update/:name", service.UpdateImage)
		route.POST("/upload", service.UploadImage)
		route.GET("/get/:name", service.GetImage)
		route.GET("/getall", service.GetAllImage)
		route.GET("/home", service.GetHome)
		route.GET("/download/:name", service.DownloadImage)
		route.DELETE("/delete/:name", service.DeleteImage)
	}
	return router
}

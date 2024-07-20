package main

import "prodImage/router"

const (
	port string = "8080"
)

func main() {
	r := router.NewRouter()
	r.Run(":" + port)
}

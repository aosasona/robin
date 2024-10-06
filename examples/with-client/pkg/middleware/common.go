package middleware

import (
	"fmt"

	"go.trulyao.dev/robin"
)

func Cors(c *robin.Context) error {
	fmt.Println("Cors")

	c.Response.Header().Set("Access-Control-Allow-Origin", "*")
	// c.Response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	// c.Response.Header().
	// 	Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

	return nil
}

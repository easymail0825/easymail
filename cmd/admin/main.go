package main

import (
	"easymail/cmd/admin/router"
)

func main() {
	r := router.New()
	_ = r.Run("0.0.0.0:8080")
}

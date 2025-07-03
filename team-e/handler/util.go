package handler

import (
	"feather/types"

	"github.com/gin-gonic/gin"
)

func response(c *gin.Context, s int, res interface{}, data ...string) {
	c.JSON(s, types.NewRes(s, res, data...))
}

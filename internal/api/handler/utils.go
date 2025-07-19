package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

func getUserId(c *gin.Context) (int, error) {
	//userIDStr := c.Query("userID")

	userIDStr := c.GetHeader("userID")
	fmt.Println(userIDStr)

	return strconv.Atoi(userIDStr)
}

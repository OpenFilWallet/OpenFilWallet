package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewError(code int, msg string) *client.Response {
	return &client.Response{Code: code, Message: msg}
}

var (
	ParamErr = NewError(1001, "parameter mismatch")
	AuthErr  = NewError(1002, "permission verification failed")
)

func ReturnOk(c *gin.Context, data interface{}) {
	if data == nil {
		data = client.Response{
			Code:    200,
			Message: "success",
		}
	}

	c.JSON(http.StatusOK, data)
}

func ReturnError(c *gin.Context, res *client.Response) {
	c.JSON(http.StatusOK, res)
}

func addr2Str(addrs []address.Address) []string {
	var addrsStr []string
	for _, addr := range addrs {
		addrsStr = append(addrsStr, addr.String())
	}
	return addrsStr
}

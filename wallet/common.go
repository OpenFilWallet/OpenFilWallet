package wallet

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Error struct {
	code int    `json:"code"`
	msg  string `json:"msg"`
}

func NewError(code int, msg string) *Error {
	return &Error{code: code, msg: msg}
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Msg() string {
	return e.msg
}

func (e *Error) Error() string {
	return fmt.Sprintf("code：%d, msg:：%s", e.Code(), e.Msg())
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

func ReturnError(c *gin.Context, err *Error) {
	data := client.Response{
		Code:    err.code,
		Message: err.msg,
	}

	c.JSON(http.StatusOK, data)
}

func addr2Str(addrs []address.Address) []string {
	var addrsStr []string
	for _, addr := range addrs {
		addrsStr = append(addrsStr, addr.String())
	}
	return addrsStr
}

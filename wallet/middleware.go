package wallet

import (
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/gin-gonic/gin"
	"strings"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic recover err: %v", err)

				ReturnError(c, NewError(500, fmt.Sprintf("panic recover err: %v", err)))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func (w *Wallet) MustUnlock() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.String(), "status") ||
			strings.Contains(c.Request.URL.String(), "login") {
			c.Next()
			return
		}

		if w.lock {
			ReturnError(c, NewError(500, "wallet is locked, please login"))
			c.Abort()
			return
		}

		// Reset lock Ticker
		w.unlock()

		c.Next()
	}
}

func (w *Wallet) MustHaveNode() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.String(), "wallet") ||
			strings.Contains(c.Request.URL.String(), "miner") ||
			strings.Contains(c.Request.URL.String(), "msig") ||
			strings.Contains(c.Request.URL.String(), "transfer") {
			if w.node == nil {
				ReturnError(c, NewError(504, "no node available"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func (w *Wallet) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		allow, err := app.AuthVerify(token)
		if err != nil {
			ReturnError(c, NewError(505, err.Error()))
			c.Abort()
			return
		}

		if !VerifyPermission(c.Request.URL.String(), allow) {
			ReturnError(c, NewError(505, "Insufficient Permission"))
			c.Abort()
			return
		}

		c.Next()
	}
}

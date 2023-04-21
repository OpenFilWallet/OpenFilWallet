package wallet

import (
	"bytes"
	"fmt"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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
			strings.Contains(c.Request.URL.String(), "login") ||
			strings.Contains(c.Request.URL.String(), "logout") {
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

func (w *Wallet) IfOfflineWallet() gin.HandlerFunc {
	return func(c *gin.Context) {
		if w.offline && strings.Contains(c.Request.URL.String(), "send") {
			ReturnError(c, NewError(504, "Offline wallet, does not support sending transactions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func (w *Wallet) JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.URL.String()
		if method != "/login" && method != "/getRouters" && method != "/logout" {
			token := c.GetHeader("Authorization")
			tokens := strings.Split(token, " ")
			if len(tokens) != 2 {
				ReturnError(c, NewError(505, "invalid token"))
				c.Abort()
				return
			}
			allow, err := app.AuthVerify(tokens[1])
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
		}

		c.Next()
	}
}

type LoggerWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (l LoggerWriter) Write(p []byte) (int, error) {
	if n, err := l.body.Write(p); err != nil {
		return n, err
	}
	return l.ResponseWriter.Write(p)
}

func (w *Wallet) TraceLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyWriter := &LoggerWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter

		method := c.Request.URL.String()
		start := time.Now()
		c.Next()
		log.Infow("TraceLogger", "method", method, "cost", time.Since(start).String())

		// The login response contains token,
		// which is sensitive information, skip it
		if method != "/login" {
			request := ""
			response := bodyWriter.body.String()
			if c.Request.Method == http.MethodPost {
				body, err := ioutil.ReadAll(c.Request.Body)
				if err != nil {
					log.Warnw("TraceLogger: ReadAll Request.Body Failed", "err", err)
				}

				request = string(body)
			} else {
				request = c.Request.URL.RawQuery
			}

			log.Debugw("TraceLogger", "method", method, "request", request, "response", response)

		}
	}
}

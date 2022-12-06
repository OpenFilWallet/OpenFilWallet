package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/build"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/gin-gonic/gin"
)

// Status Get
func (w *Wallet) Status(c *gin.Context) {
	ReturnOk(c, client.StatusInfo{
		Lock:    w.lock,
		Offline: w.offline,
		Version: build.Version(),
	})
}

package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/crypto"
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/gin-gonic/gin"
	"time"
)

const lockDuration = 10 * time.Minute

type login struct {
	lock       bool
	lockTicker *time.Ticker
	close      <-chan struct{}
}

func newLogin(close <-chan struct{}) *login {
	l := &login{
		lock:       true,
		lockTicker: time.NewTicker(lockDuration),
		close:      close,
	}

	go l.loop()

	return l
}

// Login : Post
func (w *Wallet) Login(c *gin.Context) {
	param := client.LoginRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		ReturnError(c, ParamErr)
		return
	}

	loginScryptKey, err := w.db.GetLoginPassword()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	isOk, err := crypto.VerifyScrypt(param.LoginPassword, loginScryptKey)
	if err != nil || !isOk {
		ReturnError(c, AuthErr)
		return
	}

	token, err := app.AuthNew(app.SignPermissions)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	w.unlock()
	ReturnOk(c, client.LoginInfo{
		Code:    200,
		Message: "unlock wallet success",
		Token:   string(token),
	})
}

// SignOut Post
func (w *Wallet) SignOut(c *gin.Context) {
	w.lock = true
}

func (l *login) unlock() {
	l.lock = false
	l.lockTicker.Reset(lockDuration)
}

func (l *login) loop() {
	for {
		select {
		case <-l.lockTicker.C:
			l.lock = true
		case <-l.close:
			return
		}
	}
}

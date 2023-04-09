package wallet

import (
	"context"
	"encoding/hex"
	"github.com/OpenFilWallet/OpenFilWallet/chain"
	"github.com/OpenFilWallet/OpenFilWallet/client"
	"github.com/OpenFilWallet/OpenFilWallet/datastore"
	"github.com/gin-gonic/gin"
	"time"
)

// SignMsg Post
func (w *Wallet) SignMsg(c *gin.Context) {
	param := client.SingRequest{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("SignMsg: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	msg, err := hex.DecodeString(param.HexMessage)
	if err != nil {
		log.Warnw("SignMsg: DecodeString", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	sign, err := w.signer.Sign(param.From, msg)
	if err != nil {
		log.Warnw("SignMsg: Sign", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	sigBytes := append([]byte{byte(sign.Type)}, sign.Data...)

	ReturnOk(c, client.Response{
		Code:    200,
		Message: hex.EncodeToString(sigBytes),
	})
}

// Sign Post
func (w *Wallet) Sign(c *gin.Context) {
	param := chain.Message{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Sign: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	msg, err := chain.DecodeMessage(&param)
	if err != nil {
		log.Warnw("Sign: DecodeMessage", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	signedMsg, err := w.signer.SignMsg(msg)
	if err != nil {
		log.Warnw("Sign: SignMsg", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	mySignedMsg, err := chain.BuildSignedMessage(&param, signedMsg.Signature)
	if err != nil {
		log.Warnw("Sign: BuildSignedMessage", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ReturnOk(c, mySignedMsg)
}

// SignAndSend Post
func (w *Wallet) SignAndSend(c *gin.Context) {
	param := chain.Message{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("SignAndSend: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	msg, err := chain.DecodeMessage(&param)
	if err != nil {
		log.Warnw("SignAndSend: DecodeMessage", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	signedMsg, err := w.signer.SignMsg(msg)
	if err != nil {
		log.Warnw("SignAndSend: SignMsg", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cid, err := w.Api.MpoolPush(ctx, signedMsg)
	cancel()
	if err != nil {
		log.Warnw("SignAndSend: MpoolPush", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	if ok := w.signer.HasSigner(signedMsg.Message.From.String()); ok {
		w.txTracker.trackTx(&datastore.History{
			Version:    signedMsg.Message.Version,
			To:         signedMsg.Message.To.String(),
			From:       signedMsg.Message.From.String(),
			Nonce:      signedMsg.Message.Nonce,
			Value:      signedMsg.Message.Value.Int64(),
			GasLimit:   signedMsg.Message.GasLimit,
			GasFeeCap:  signedMsg.Message.GasFeeCap.Int64(),
			GasPremium: signedMsg.Message.GasPremium.Int64(),
			Method:     uint64(signedMsg.Message.Method),
			Params:     param.Params.Params,
			ParamName:  param.Params.Name,
			TxCid:      cid.String(),
			TxState:    datastore.Pending,
		})
	}

	ReturnOk(c, client.Response{
		Code:    200,
		Message: cid.String(),
	})
}

// Send Post
func (w *Wallet) Send(c *gin.Context) {
	param := chain.SignedMessage{}
	err := c.BindJSON(&param)
	if err != nil {
		log.Warnw("Send: BindJSON", "err", err.Error())
		ReturnError(c, ParamErr)
		return
	}

	signedMsg, err := chain.DecodeSignedMessage(&param)
	if err != nil {
		log.Warnw("Send: DecodeSignedMessage", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cid, err := w.Api.MpoolPush(ctx, signedMsg)
	cancel()
	if err != nil {
		log.Warnw("Send: MpoolPush", "err", err.Error())
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	w.txTracker.trackTx(&datastore.History{
		Version:    signedMsg.Message.Version,
		To:         signedMsg.Message.To.String(),
		From:       signedMsg.Message.From.String(),
		Nonce:      signedMsg.Message.Nonce,
		Value:      signedMsg.Message.Value.Int64(),
		GasLimit:   signedMsg.Message.GasLimit,
		GasFeeCap:  signedMsg.Message.GasFeeCap.Int64(),
		GasPremium: signedMsg.Message.GasPremium.Int64(),
		Method:     uint64(signedMsg.Message.Method),
		Params:     param.Message.Params.Params,
		ParamName:  param.Message.Params.Name,
		TxCid:      cid.String(),
		TxState:    datastore.Pending,
	})

	ReturnOk(c, client.Response{
		Code:    200,
		Message: cid.String(),
	})
}

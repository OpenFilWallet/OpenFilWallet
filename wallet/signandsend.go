package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
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
		ReturnError(c, ParamErr)
		return
	}

	msg, err := hex.DecodeString(param.HexMessage)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	sign, err := w.signer.Sign(param.From, msg)
	if err != nil {
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
		ReturnError(c, ParamErr)
		return
	}

	msg, err := chain.DecodeMessage(&param)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	signedMsg, err := w.signer.SignMsg(msg)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	mySignedMsg, err := chain.BuildSignedMessage(&param, signedMsg.Signature)
	if err != nil {
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
		ReturnError(c, ParamErr)
		return
	}

	msg, err := chain.DecodeMessage(&param)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	signedMsg, err := w.signer.SignMsg(msg)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cid, err := w.Api.MpoolPush(ctx, signedMsg)
	cancel()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	w.RecordTx(&datastore.History{
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
	})

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
		ReturnError(c, ParamErr)
		return
	}

	signedMsg, err := chain.DecodeSignedMessage(&param)
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cid, err := w.Api.MpoolPush(ctx, signedMsg)
	cancel()
	if err != nil {
		ReturnError(c, NewError(500, err.Error()))
		return
	}

	w.RecordTx(&datastore.History{
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
	})

	ReturnOk(c, client.Response{
		Code:    200,
		Message: cid.String(),
	})
}

func (w *Wallet) RecordTx(msg *datastore.History) {
	err := w.db.SetHistory(msg)
	if err != nil {
		log.Warnw("RecordTx fail", "msg", fmt.Sprintf("From: %s To: %s Method: %d", msg.From, msg.To, msg.Method), "err", err)
	}
}

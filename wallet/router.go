package wallet

import (
	"github.com/OpenFilWallet/OpenFilWallet/modules/app"
	"github.com/gin-gonic/gin"
	"strings"
)

func (w *Wallet) NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(Recovery())
	r.Use(w.MustUnlock())
	r.Use(w.MustHaveNode())
	r.Use(w.JWT())

	r.GET("/status", w.Status)

	r.POST("/login", w.Login)
	r.POST("/signout", w.SignOut)

	r.POST("/chain/decode", w.Decode)
	r.POST("/chain/encode", w.Encode)

	r.POST("/node/add", w.NodeAdd)
	r.POST("/node/update", w.NodeUpdate)
	r.POST("/node/delete", w.NodeDelete)
	r.POST("/node/use_node", w.UseNode)
	r.GET("/node/list", w.NodeList)
	r.GET("/node/best", w.NodeBest)

	r.POST("/wallet/create", w.WalletCreate)
	r.GET("/wallet/list", w.WalletList)

	r.GET("/balance", w.Balance)

	r.POST("/transfer", w.Transfer)

	r.POST("/send", w.Send)

	r.GET("/tx_history", w.TxHistory)

	r.POST("/sign_msg", w.SignMsg)
	r.POST("/sign", w.Sign)
	r.POST("/sign_send", w.SignAndSend)

	r.POST("/miner/withdraw", w.Withdraw)
	r.POST("/miner/change_owner", w.ChangeOwner)
	r.POST("/miner/change_worker", w.ChangeWorker)
	r.POST("/miner/confirm_change_worker", w.ConfirmChangeWorker)
	r.POST("/miner/change_control", w.ChangeControl)
	r.GET("/miner/control_list", w.ControlList)
	r.POST("/miner/change_beneficiary", w.ChangeBeneficiary)
	r.POST("/miner/confirm_change_beneficiary", w.ConfirmChangeBeneficiary)

	r.GET("/msig/list", w.MsigWalletList)
	r.GET("/msig/inspect", w.MsigInspect)
	r.POST("/msig/create", w.MsigCreate)
	r.POST("/msig/approve", w.MsigApprove)
	r.POST("/msig/cancel", w.MsigCancel)
	r.POST("/msig/transfer_propose", w.MsigTransferPropose)
	r.POST("/msig/transfer_approve", w.MsigTransferApprove)
	r.POST("/msig/transfer_cancel", w.MsigTransferCancel)
	r.POST("/msig/add_signer_propose", w.MsigAddPropose)
	r.POST("/msig/add_signer_approve", w.MsigAddApprove)
	r.POST("/msig/add_signer_cancel", w.MsigAddCancel)
	r.POST("/msig/swap_propose", w.MsigSwapPropose)
	r.POST("/msig/swap_approve", w.MsigSwapApprove)
	r.POST("/msig/swap_cancel", w.MsigSwapCancel)
	r.POST("/msig/lock_propose", w.MsigLockPropose)
	r.POST("/msig/lock_approve", w.MsigLockApprove)
	r.POST("/msig/lock_cancel", w.MsigLockCancel)
	r.POST("/msig/threshold_propose", w.MsigThresholdPropose)
	r.POST("/msig/threshold_approve", w.MsigThresholdApprove)
	r.POST("/msig/threshold_cancel", w.MsigThresholdCancel)
	r.POST("/msig/change_owner_propose", w.MsigChangeOwnerPropose)
	r.POST("/msig/change_owner_approve", w.MsigChangeOwnerApprove)
	r.POST("/msig/withdraw_propose", w.MsigWithdrawPropose)
	r.POST("/msig/withdraw_approve", w.MsigWithdrawApprove)
	r.POST("/msig/change_worker_propose", w.MsigChangeWorkerPropose)
	r.POST("/msig/change_worker_approve", w.MsigChangeWorkerApprove)
	r.POST("/msig/confirm_change_worker_propose", w.MsigConfirmChangeWorkerPropose)
	r.POST("/msig/confirm_change_worker_approve", w.MsigConfirmChangeWorkerApprove)
	r.POST("/msig/set_control_propose", w.MsigSetControlPropose)
	r.POST("/msig/set_control_approve", w.MsigSetControlApprove)
	r.POST("/msig/change_beneficiary_propose", w.MsigChangeBeneficiaryPropose)
	r.POST("/msig/change_beneficiary_approve", w.MsigChangeBeneficiaryApprove)
	r.POST("/msig/confirm_change_beneficiary_propose", w.MsigConfirmChangeBeneficiaryPropose)
	r.POST("/msig/confirm_change_beneficiary_approve", w.MsigConfirmChangeBeneficiaryApprove)

	return r
}

var HandlePermMap = map[string]app.Permission{
	"/status":                                  app.PermRead,
	"/login":                                   app.PermWrite,
	"/signout":                                 app.PermWrite,
	"/chain/decode":                            app.PermRead,
	"/chain/encode":                            app.PermRead,
	"/node/add":                                app.PermWrite,
	"/node/update":                             app.PermWrite,
	"/node/delete":                             app.PermWrite,
	"/node/use_node":                           app.PermWrite,
	"/node/list":                               app.PermRead,
	"/node/best":                               app.PermRead,
	"/wallet/create":                           app.PermWrite,
	"/wallet/list":                             app.PermRead,
	"/balance":                                 app.PermRead,
	"/transfer":                                app.PermWrite,
	"/tx_history":                              app.PermRead,
	"/send":                                    app.PermWrite,
	"/sign_msg":                                app.PermSign,
	"/sign":                                    app.PermSign,
	"/sign_send":                               app.PermSign,
	"/miner/withdraw":                          app.PermWrite,
	"/miner/change_owner":                      app.PermWrite,
	"/miner/change_worker":                     app.PermWrite,
	"/miner/confirm_change_worker":             app.PermWrite,
	"/miner/change_control":                    app.PermWrite,
	"/miner/control_list":                      app.PermRead,
	"/miner/change_beneficiary":                app.PermWrite,
	"/miner/confirm_change_beneficiary":        app.PermWrite,
	"/msig/list":                               app.PermRead,
	"/msig/inspect":                            app.PermRead,
	"/msig/create":                             app.PermWrite,
	"/msig/approve":                            app.PermWrite,
	"/msig/cancel":                             app.PermWrite,
	"/msig/transfer_propose":                   app.PermWrite,
	"/msig/transfer_approve":                   app.PermWrite,
	"/msig/transfer_cancel":                    app.PermWrite,
	"/msig/add_signer_propose":                 app.PermWrite,
	"/msig/add_signer_approve":                 app.PermWrite,
	"/msig/add_signer_cancel":                  app.PermWrite,
	"/msig/swap_propose":                       app.PermWrite,
	"/msig/swap_approve":                       app.PermWrite,
	"/msig/swap_cancel":                        app.PermWrite,
	"/msig/lock_propose":                       app.PermWrite,
	"/msig/lock_approve":                       app.PermWrite,
	"/msig/lock_cancel":                        app.PermWrite,
	"/msig/threshold_propose":                  app.PermWrite,
	"/msig/threshold_approve":                  app.PermWrite,
	"/msig/threshold_cancel":                   app.PermWrite,
	"/msig/change_owner_propose":               app.PermWrite,
	"/msig/change_owner_approve":               app.PermWrite,
	"/msig/withdraw_propose":                   app.PermWrite,
	"/msig/withdraw_approve":                   app.PermWrite,
	"/msig/change_worker_propose":              app.PermWrite,
	"/msig/change_worker_approve":              app.PermWrite,
	"/msig/confirm_change_worker_propose":      app.PermWrite,
	"/msig/confirm_change_worker_approve":      app.PermWrite,
	"/msig/set_control_propose":                app.PermWrite,
	"/msig/set_control_approve":                app.PermWrite,
	"/msig/change_beneficiary_propose":         app.PermWrite,
	"/msig/change_beneficiary_approve":         app.PermWrite,
	"/msig/confirm_change_beneficiary_propose": app.PermWrite,
	"/msig/confirm_change_beneficiary_approve": app.PermWrite,
}

func VerifyPermission(requestUrl string, allows []app.Permission) bool {
	for key, perm := range HandlePermMap {
		if strings.Contains(requestUrl, key) {
			for _, allow := range allows {
				if allow == perm {
					return true
				}
			}
		}
	}
	return false
}

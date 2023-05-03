package wallet

import "github.com/gin-gonic/gin"

type RouterResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []Routers `json:"data"`
}

type Routers struct {
	Name       string     `json:"name,omitempty"`
	Path       string     `json:"path"`
	Hidden     bool       `json:"hidden"`
	Redirect   string     `json:"redirect,omitempty"`
	Component  string     `json:"component"`
	AlwaysShow bool       `json:"alwaysShow,omitempty"`
	Meta       Meta       `json:"meta"`
	Children   []Children `json:"children"`
}

type Meta struct {
	Title   string `json:"title"`
	Icon    string `json:"icon"`
	NoCache bool   `json:"noCache"`
}

type Children struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Hidden    bool   `json:"hidden"`
	Component string `json:"component"`
	Meta      Meta   `json:"meta"`
}

// transferRouter
var transferRouter = Routers{
	Name:       "Transfer",
	Path:       "/transfer",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,
	Children: []Children{
		{
			Name:      "Transfer",
			Path:      "/transfer",
			Hidden:    false,
			Component: "openfil/transfer/transfer",
			Meta: Meta{
				Title:   "Transfer",
				Icon:    "guide",
				NoCache: false,
			},
		},
	},
}

// minerRouter
var minerRouter = Routers{
	Name:       "Miner",
	Path:       "/miner",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,
	Meta: Meta{
		Title:   "Miner",
		Icon:    "system",
		NoCache: false,
	},
	Children: []Children{
		{
			Name:      "Withdraw",
			Path:      "/miner/withdraw",
			Hidden:    false,
			Component: "openfil/miner/withdraw",
			Meta: Meta{
				Title:   "Withdraw",
				Icon:    "system",
				NoCache: false,
			},
		},
		{
			Name:      "Owner",
			Path:      "/miner/owner",
			Hidden:    false,
			Component: "openfil/miner/owner",
			Meta: Meta{
				Title:   "Owner",
				Icon:    "system",
				NoCache: false,
			},
		},
		{
			Name:      "Worker",
			Path:      "/miner/worker",
			Hidden:    false,
			Component: "openfil/miner/worker",
			Meta: Meta{
				Title:   "Worker",
				Icon:    "system",
				NoCache: false,
			},
		},
		{
			Name:      "Control",
			Path:      "/miner/control",
			Hidden:    false,
			Component: "openfil/miner/control",
			Meta: Meta{
				Title:   "Control",
				Icon:    "system",
				NoCache: false,
			},
		},
		{
			Name:      "Beneficiary",
			Path:      "/miner/beneficiary",
			Hidden:    false,
			Component: "openfil/miner/beneficiary",
			Meta: Meta{
				Title:   "Beneficiary",
				Icon:    "system",
				NoCache: false,
			},
		},
	},
}

// msigRouter
var msigRouter = Routers{
	Name:       "Msig",
	Path:       "/msig",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,
	Meta: Meta{
		Title:   "Msig",
		Icon:    "peoples",
		NoCache: false,
	},
	Children: []Children{
		{
			Name:      "MsigTx",
			Path:      "/msig/msig",
			Hidden:    false,
			Component: "openfil/msig/msig",
			Meta: Meta{
				Title:   "Msig Tx",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigTransfer",
			Path:      "/msig/transfer",
			Hidden:    false,
			Component: "openfil/msig/transfer",
			Meta: Meta{
				Title:   "Transfer",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigWithdraw",
			Path:      "/msig/withdraw",
			Hidden:    false,
			Component: "openfil/msig/withdraw",
			Meta: Meta{
				Title:   "Withdraw",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigOwner",
			Path:      "/msig/owner",
			Hidden:    false,
			Component: "openfil/msig/owner",
			Meta: Meta{
				Title:   "Owner",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigWorker",
			Path:      "/msig/worker",
			Hidden:    false,
			Component: "openfil/msig/worker",
			Meta: Meta{
				Title:   "Worker",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigControl",
			Path:      "/msig/control",
			Hidden:    false,
			Component: "openfil/msig/control",
			Meta: Meta{
				Title:   "Control",
				Icon:    "peoples",
				NoCache: false,
			},
		},
		{
			Name:      "MsigBeneficiary",
			Path:      "/msig/beneficiary",
			Hidden:    false,
			Component: "openfil/msig/beneficiary",
			Meta: Meta{
				Title:   "Beneficiary",
				Icon:    "peoples",
				NoCache: false,
			},
		},
	},
}

// signRouter
var signRouter = Routers{
	Name:       "Sign",
	Path:       "/sign",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: true,
	Redirect:   "noRedirect",
	Meta: Meta{
		Title:   "Sign",
		Icon:    "documentation",
		NoCache: false,
	},
	Children: []Children{
		{
			Name:      "SignTx",
			Path:      "/sign_tx",
			Hidden:    false,
			Component: "openfil/sign/sign-tx",
			Meta: Meta{
				Title:   "Sign Tx",
				Icon:    "documentation",
				NoCache: false,
			},
		},
		{
			Name:      "SignMsg",
			Path:      "/sign_msg",
			Hidden:    false,
			Component: "openfil/sign/sign-msg",
			Meta: Meta{
				Title:   "Sign Msg",
				Icon:    "documentation",
				NoCache: false,
			},
		},
	},
}

// signAndSendChildrenRouter
var signAndSendChildrenRouter = Children{
	Name:      "SignAndSend",
	Path:      "/sign_send",
	Hidden:    false,
	Component: "openfil/sign/sign-send",
	Meta: Meta{
		Title:   "Sign And Send",
		Icon:    "documentation",
		NoCache: false,
	},
}

// sendRouter
var sendRouter = Routers{
	Name:       "Send",
	Path:       "/send",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,

	Children: []Children{
		{
			Name:      "Send",
			Path:      "/send",
			Hidden:    false,
			Component: "openfil/send/send",
			Meta: Meta{
				Title:   "Send",
				Icon:    "druid",
				NoCache: false,
			},
		},
	},
}

// nodeRouter
var nodeRouter = Routers{
	Name:       "Node",
	Path:       "/node",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,
	Redirect:   "noRedirect",

	Children: []Children{
		{
			Name:      "Node",
			Path:      "/node",
			Hidden:    false,
			Component: "openfil/node/index",
			Meta: Meta{
				Title:   "Node",
				Icon:    "monitor",
				NoCache: false,
			},
		},
	},
}

// toolRouter
var toolRouter = Routers{
	Name:       "Tool",
	Path:       "/tool",
	Hidden:     false,
	Component:  "Layout",
	AlwaysShow: false,
	Redirect:   "noRedirect",

	Children: []Children{
		{
			Name:      "Tool",
			Path:      "/tool",
			Hidden:    false,
			Component: "openfil/tool/index",
			Meta: Meta{
				Title:   "Tool",
				Icon:    "tool",
				NoCache: false,
			},
		},
	},
}

var isAddSignAndSendChildrenRouter bool = false

func (w *Wallet) GetRouters(c *gin.Context) {
	if !w.offline {
		if !isAddSignAndSendChildrenRouter {
			signRouter.Children = append(signRouter.Children, signAndSendChildrenRouter)
			isAddSignAndSendChildrenRouter = true
		}

		ReturnOk(c, RouterResponse{
			Code:    200,
			Message: "success",
			Data:    []Routers{transferRouter, minerRouter, msigRouter, signRouter, sendRouter, nodeRouter, toolRouter},
		})
		return
	}

	ReturnOk(c, RouterResponse{
		Code:    200,
		Message: "success",
		Data:    []Routers{transferRouter, minerRouter, msigRouter, signRouter, nodeRouter, toolRouter},
	})
}

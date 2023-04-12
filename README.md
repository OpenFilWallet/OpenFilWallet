# OpenFilWallet

OpenFilWallet focuses on creating an easy-to-use miner HD wallet, which is convenient for users to initiate ordinary transactions and multi-signature transactions simply and safely. OpenFilWallet allows users to easily initiate transactions without importing private keys to guardian nodes. OpenFilWallet provides offline functionality, allowing it to act as an offline signature machine, providing maximum security.

OpenFilWallet provides a safe and simple way to send Filecoin transactions, and introduces a mnemonic function to facilitate account management, and ensures wallet security through encryption. The focus is to provide a simple multi-signature transaction experience, lower the threshold for using the multi-signature function, and promote network stability.

See the docs at [https://docs.openfilwallet.com](https://docs.openfilwallet.com/) to get started.

Note: Because the git submodule is used in the compilation process, the code of `filecoin-ffi` needs to be automatically pulled during compilation, so you need to configure the ssh key in your github

## Warn

A release will be made when it has been thoroughly tested. It is not recommended to use it on the mainnet, unless you know what you are doing!

## TODO

- Web UI
- Display transactions through QR codes
- Support docker
- ...

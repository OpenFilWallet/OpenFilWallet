<template>
    <div class="container">
        <div class="transaction-group" v-for="msigWalletGroup in msigWalletGroups" :key="msigWalletGroup.msig_addr">
            <div class="header">{{ msigWalletGroup.msig_addr }}</div>
            <template v-if="msigWalletGroup.transactions && msigWalletGroup.transactions.length > 0">
                <el-card class="card" v-for="transaction in msigWalletGroup.transactions" :key="transaction.txid">
                    <div class="card-content">
                        <div class="card-item">
                            <span class="label">TxId:</span>
                            <span>{{ transaction.txid }}</span>
                        </div>
                        <div class="card-item">
                            <span class="label">To:</span>
                            <span>{{ transaction.to }}</span>
                        </div>
                        <div class="card-item">
                            <span class="label">Value:</span>
                            <span>{{ transaction.value }}</span>
                        </div>
                        <div class="card-item">
                            <span class="label">Method:</span>
                            <span>{{ transaction.method }}</span>
                        </div>
                        <div class="card-item">
                            <span class="label">Params:</span>
                            <span>{{ transaction.params }}</span>
                        </div>
                        <div class="card-item">
                            <span class="label">Approved By:</span>
                            <div class="approved-list">
                                <span v-for="approver in transaction.approved" :key="approver" class="approved-item">{{
                                    approver
                                }}</span>
                            </div>
                        </div>
                        <div class="btn-container">
                            <div class="btn-group">
                                <el-button :loading="cancelLoading" size="medium"
                                    @click="selectAddress(false, msigWalletGroup.msig_addr, transaction.txid)">
                                    <span v-if="!cancelLoading">{{ $t("Cancel") }}</span>
                                    <span v-else>cancel ...</span>
                                </el-button>
                                <el-button :loading="loading" size="medium" type="primary"
                                    @click="selectAddress(true, msigWalletGroup.msig_addr, transaction.txid)">
                                    <span v-if="!loading">{{ $t("Approve") }}</span>
                                    <span v-else>approve ...</span>
                                </el-button>
                            </div>
                        </div>
                    </div>
                </el-card>
            </template>
            <template v-else>
                <el-card class="card">
                    <div class="card-content">
                        <div class="no-transactions">{{ $t("No transactions") }}</div>
                    </div>
                </el-card>
            </template>
        </div>
        <el-dialog title="Select from address" :visible.sync="dialogVisible" width="30%" center>
            <el-form label="From:" required>
                <el-select v-model="selectedFromAddress" placeholder="please select from address">
                    <el-option v-for="account in fromAddresses" :key="account.address" :value="account.address">
                    </el-option>
                </el-select>
            </el-form>

            <span slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">Cancel</el-button>
                <el-button type="primary" @click="makeTransaction" :disabled="!selectedFromAddress">OK</el-button>
            </span>
        </el-dialog>

        <el-dialog title="Transaction Result" :visible.sync="dialogTxVisible">
            <pre>{{ transactionResult }}</pre>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogTxVisible = false">Close</el-button>
                <el-button type="primary" @click="copyToClipboard">Copy</el-button>
            </div>
        </el-dialog>
    </div>
</template>
  
<script>
import { walletList, msigWalletList } from "@/api/openfil/wallet.js";
import { msigInspect } from "@/api/openfil/tool.js";
import { msigApprove, msigCancel } from "@/api/openfil/msig.js";
export default {
    data() {
        return {
            formData: {},
            addresses: [],
            msigWallets: [],
            msigWalletGroups: [],
            transactionResult: '',
            fromAddresses: [],
            loading: false,
            cancelLoading: false,
            dialogVisible: false,
            dialogTxVisible: false,
            isApprove: false,
            selectedFromAddress: '',
            msigAddr: '',
            txid: '',
        };
    },
    created() {
        this.fetchWalletList();
        this.fetchMsigWalletList();
    },
    methods: {
        fetchWalletList() {
            walletList(this.queryParams).then(response => {
                this.addresses = response;
            }).catch(error => {
                console.log(error);
            });
        },
        fetchMsigWalletList() {
            msigWalletList(this.queryParams).then(response => {
                this.msigWallets = response;
                this.fetchMsigWalletGroups();
            }).catch(error => {
                console.log(error);
            });
        },
        fetchMsigWalletGroups() {
            this.msigWalletGroups = [];
            this.msigWallets.forEach(msigWallet => {
                msigInspect(msigWallet.address).then(response => {
                    this.msigWalletGroups.push(response);
                }).catch(error => {
                    console.log(error);
                });
            });
        },
        selectAddress(isApprove, msigAddr, txid) {
            this.isApprove = isApprove;
            const msigWalletGroup = this.msigWalletGroups.find(
                (msigGroup) => msigGroup.msig_addr === msigAddr
            );
            this.fromAddresses = this.addresses.filter((address) => {
                return msigWalletGroup.signers.includes(address.id);
            });
            this.selectedFromAddress = '';
            this.dialogVisible = true;
            this.msigAddr = msigAddr;
            this.txid = txid;

        },
        makeTransaction() {
            if (!this.isApprove) {
                this.$confirm(`Are you sure to cancel txId: ${this.txid} of ${this.msigAddr}?`, 'Warning', {
                    confirmButtonText: 'OK',
                    cancelButtonText: 'Cancel',
                    type: 'warning'
                }).then(() => {
                    this.dialogVisible = false;
                    this.cancelLoading = true;
                    msigCancel(this.selectedFromAddress, this.msigAddr, this.txid.toString())
                        .then((response) => {
                            this.transactionResult = JSON.stringify(response, null, 4);
                            this.cancelLoading = false;
                            this.dialogTxVisible = true;
                        })
                        .catch((error) => {
                            console.log(error);
                            this.cancelLoading = false;
                        });
                });
            } else {
                this.$confirm(`Are you sure to approve txId: ${this.txid} of ${this.msigAddr}?`, 'Warning', {
                    confirmButtonText: 'OK',
                    cancelButtonText: 'Cancel',
                    type: 'warning'
                }).then(() => {
                    this.dialogVisible = false;
                    this.loading = true;
                    msigApprove(this.selectedFromAddress, this.msigAddr, this.txid.toString())
                        .then((response) => {
                            this.transactionResult = JSON.stringify(response, null, 4);
                            this.loading = false;
                            this.dialogTxVisible = true;
                        })
                        .catch((error) => {
                            console.log(error);
                            this.loading = false;
                        });
                });
            }
        },
        copyToClipboard() {
            const textarea = document.createElement('textarea');
            textarea.setAttribute('readonly', true);
            textarea.style.position = 'absolute';
            textarea.style.left = '-9999px';
            textarea.value = this.transactionResult;
            document.body.appendChild(textarea);
            textarea.select();
            document.execCommand('copy');
            document.body.removeChild(textarea);
            this.$message({
                message: 'Transaction copied!',
                type: 'success'
            });
        },
    },
};
</script>
  
<style scoped>
.container {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    margin: 50px;
}

.transaction-group {
    width: 100%;
    max-width: 800px;
    margin-bottom: 30px;
}

.header {
    font-size: 24px;
    font-weight: bold;
    margin-bottom: 10px;
}

.card {
    margin-bottom: 20px;
}

.card-content {
    display: flex;
    flex-direction: column;
    padding: 20px;
}

.card-item {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
}

.label {
    font-weight: bold;
    margin-right: 10px;
}

.approved-list {
    display: flex;
    flex-wrap: wrap;
}

.approved-item {
    background-color: #9af0a8;
    border-radius: 10px;
    padding: 2px 10px;
    margin-right: 10px;
    margin-bottom: 5px;
}

.btn-group {
    display: flex;
    justify-content: center;
    align-items: center;
}
</style>
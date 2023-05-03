<template>
    <div class="withdraw-page">
        <h1 class="title">Withdraw Available Balance Propose</h1>
        <el-form class="withdraw-form" :model="form" ref="form" label-position="left" label-width="120px">
            <el-form-item label="Msig:" required>
                <el-select v-model="form.msig" placeholder="please select msig wallet">
                    <el-option v-for="account in msigAddressLabels" :key="account.address" :label="account.labelWithId"
                        :value="account.address">
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="From:" required>
                <el-select v-model="form.from" placeholder="please select from address">
                    <el-option v-for="account in filteredAddressLabels" :key="account.address" :label="account.labelWithId"
                        :value="account.address">
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="MinerID:" required>
                <el-input v-model="form.minerId" placeholder="please enter to minerID"></el-input>
            </el-form-item>
            <el-form-item label="Amount:" required>
                <el-input-number v-model="form.amount" :min="0" :step="1"></el-input-number>
            </el-form-item>
            <el-form-item>
                <el-button :loading="loading" size="medium" type="primary" @click="submit">
                    <span v-if="!loading">{{ $t("Withdraw") }}</span>
                    <span v-else>withdraw...</span>
                </el-button>
            </el-form-item>
        </el-form>
        <el-dialog title="Transaction Result" :visible.sync="dialogVisible">
            <pre>{{ transactionResult }}</pre>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">Close</el-button>
                <el-button type="primary" @click="copyToClipboard">Copy</el-button>
            </div>
        </el-dialog>
    </div>
</template>
  
<script>
import { walletList, msigWalletList } from "@/api/openfil/wallet.js";
import { msigWithdraw } from "@/api/openfil/msig.js";
export default {
    data() {
        return {
            msigWallets: [],
            addresses: [],
            form: {
                msig: '',
                from: '',
                minerId: '',
                amount: '',
            },
            queryParams: { balance: true },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
        };
    },
    created() {
        this.fetchWalletList();
        this.fetchMsigWalletList();
    },
    computed: {
        addressLabels() {
            return this.addresses.map(account => ({
                address: account.address,
                balance: account.balance,
                id: account.id,
                labelWithId: `${account.address} (${account.id}) ${account.balance}`,
            }));
        },
        msigAddressLabels() {
            return this.msigWallets.map(account => ({
                address: account.address,
                labelWithId: `${account.address} (${account.id}) ${account.balance}`,
            }));
        },
        filteredAddressLabels() {
            if (!this.form.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.form.msig)
                return this.addressLabels.filter(address => selectedMsigWallet.signers.includes(address.id))
            }
        }
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
            }).catch(error => {
                console.log(error);
            });
        },
        submit() {
            this.$refs.form.validate(valid => {
                if (valid) {
                    this.loading = true;
                    msigWithdraw(this.form.from, this.form.msig, this.form.minerId, this.form.amount.toString()).then(response => {
                        console.log(response);
                        this.transactionResult = JSON.stringify(response, null, 4);
                        this.dialogVisible = true;
                        this.loading = false;
                    }).catch(error => {
                        this.loading = false;
                        console.error(error);
                    });
                }
            });
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
    }
}
</script>
  
<style scoped>
.withdraw-page {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-top: 50px;
}

.title {
    font-size: 32px;
    margin-bottom: 30px;
}

.withdraw-form {
    max-width: 1000px;
    width: 100%;
}

.el-select-dropdown__item span.balance {
    float: right;
    font-size: 12px;
}

.el-input__inner {
    height: 80px !important;
    resize: none;
}
</style>
  
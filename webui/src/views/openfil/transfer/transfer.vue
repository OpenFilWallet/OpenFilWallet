<template>
    <div class="transfer-page">
        <h1 class="title">Transfer FIL</h1>
        <el-form class="transfer-form" :model="form" ref="form" label-position="left" label-width="120px">
            <el-form-item label="From:" required>
                <el-select v-model="form.from" placeholder="please select from address" @change="onFromChange">
                    <el-option v-for="account in addressLabels" :key="account.address" :label="account.labelWithBalance"
                        :value="account.address">
                        <div slot="label">
                            <span>{{ account.label }}</span>
                            <span class="balance">{{ account.balance }}</span>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="To:" required>
                <el-input v-model="form.to" placeholder="please enter to address"></el-input>
            </el-form-item>
            <el-form-item label="Amount:" required>
                <el-input-number v-model="form.amount" :min="0" :step="0.5"></el-input-number>
            </el-form-item>
            <el-form-item>
                <el-button :loading="loading" size="medium" type="primary" @click="submit">
                    <span v-if="!loading">{{ $t("Transfer") }}</span>
                    <span v-else>transferring ...</span>
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
import { walletList } from "@/api/openfil/wallet.js";
import { transfer } from "@/api/openfil/transfer.js";
export default {
    data() {
        return {
            addresses: [],
            form: {
                from: '',
                to: '',
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
    },
    computed: {
        addressLabels() {
            return this.addresses.map(account => ({
                address: account.address,
                balance: account.balance,
                labelWithBalance: `${account.address} (${account.balance})`,
            }));
        },
    },
    methods: {
        onFromChange() {
            const selectedAddress = this.addresses.find(account => account.address === this.form.from);
            if (selectedAddress) {
                this.$set(selectedAddress, 'labelWithBalance', `${selectedAddress.address} (${selectedAddress.balance} FIL)`);
            }
        },
        fetchWalletList() {
            walletList(this.queryParams).then(response => {
                this.addresses = response;
            }).catch(error => {
                console.log(error);
            });
        },
        submit() {
            this.$refs.form.validate(valid => {
                if (valid) {
                    this.loading = true;
                    transfer(this.form.from, this.form.to, this.form.amount.toString()).then(response => {
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
.transfer-page {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-top: 50px;
}

.title {
    font-size: 32px;
    margin-bottom: 30px;
}

.transfer-form {
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
  
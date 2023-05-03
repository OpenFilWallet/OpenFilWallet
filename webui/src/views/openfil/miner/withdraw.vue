<template>
    <div class="withdraw-page">
        <h1 class="title">Withdraw Available Balance</h1>
        <el-form class="withdraw-form" :model="form" ref="form" label-position="left" label-width="120px">
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
import { withdraw } from "@/api/openfil/miner.js";
export default {
    data() {
        return {
            form: {
                minerId: '',
                amount: '',
            },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
        };
    },

    methods: {
        submit() {
            this.$refs.form.validate(valid => {
                if (valid) {
                    this.loading = true;
                    withdraw(this.form.minerId, this.form.amount.toString()).then(response => {
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
  
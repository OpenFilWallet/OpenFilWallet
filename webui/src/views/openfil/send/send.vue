<template>
    <div class="sign-transaction">
        <div class="input-container">
            <el-input class="input-box" type="textarea" :rows="10" placeholder="Enter signed transaction data here"
                v-model="txData"></el-input>
            <el-button class="sign-button" :loading="loading" size="medium" type="primary" @click="signTransaction">
                <span v-if="!loading">{{ $t("Send") }}</span>
                <span v-else>sending ...</span>
            </el-button>

        </div>

        <el-dialog title="Transaction Cid" :visible.sync="dialogVisible">
            <pre>{{ transactionResult }}</pre>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">Close</el-button>
                <el-button type="primary" @click="copyToClipboard">Copy</el-button>
            </div>
        </el-dialog>
    </div>
</template>
  
  
<script>
import { send } from "@/api/openfil/send.js";
export default {
    data() {
        return {
            txData: "",
            dialogVisible: false,
            transactionResult: '',
            loading: false,
        };
    },
    methods: {
        signTransaction() {
            this.loading = true;
            send(this.txData).then(response => {
                console.log(response);
                this.transactionResult = response.message;
                this.dialogVisible = true;
                this.loading = false;
            }).catch(error => {
                this.loading = false;
                console.error(error);
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
                message: 'Transaction cid copied!',
                type: 'success'
            });
        }
    },

}
</script>
  
<style scoped>
.sign-transaction {
    display: flex;
    flex-direction: column;
    align-items: center;
}

.input-container {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    width: 80%;
    margin-top: 50px;
    margin-bottom: 50px;
}

.input-box {
    max-width: none;
    width: 100%;
    margin-bottom: 20px;
    resize: none;
}

.sign-button {
    width: 40%;
    margin-top: 20px;
}
</style>
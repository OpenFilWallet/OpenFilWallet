<template>
    <div class="sign-transaction">
        <div class="input-container">
            <el-select v-model="selectedAddress" placeholder="Select sign address">
                <el-option v-for="account in addresses" :key="account.address" :label="account.address"
                    :value="account.address" />
            </el-select>
            <el-input class="input-box" type="textarea" :rows="10" placeholder="Enter sign data here"
                v-model="hexMsg"></el-input>
            <el-button class="sign-button" type="primary" @click="signHexMessage">Sign Msg</el-button>
        </div>

        <el-dialog title="Signed Result" :visible.sync="dialogVisible">
            <pre>{{ transactionResult }}</pre>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogVisible = false">Close</el-button>
                <el-button type="primary" @click="copyToClipboard">Copy</el-button>
            </div>
        </el-dialog>
    </div>
</template>
  
  
<script>
import { signMsg } from "@/api/openfil/sign.js";
import { walletList } from "@/api/openfil/wallet.js";
export default {
    data() {
        return {
            addresses: [],
            selectedAddress: "",
            hexMsg: "",
            dialogVisible: false,
            transactionResult: '',
            loading: false,
        };
    },
    created() {
        this.fetchWalletList();
    },
    methods: {
        fetchWalletList() {
            walletList(this.queryParams).then(response => {
                this.addresses = response;
            }).catch(error => {
                console.log(error);
            });
        },
        signHexMessage() {
            this.loading = true;
            signMsg(this.selectedAddress, this.hexMsg).then(response => {
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
                message: 'Signature copied!',
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
</style>∂
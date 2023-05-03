<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Miner Owner</h3>
                    <el-form ref="changeOwnerForm" :model="changeOwnerFormData" :rules="ownerFormRules" label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId">
                            <el-input placeholder="please input minerId" v-model="changeOwnerFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="NewOwner" prop="newOwnerAddress">
                            <el-input placeholder="please input new owner address"
                                v-model="changeOwnerFormData.newOwnerAddress"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="loading" size="medium" type="primary" @click="submitChangeOwnerForm">
                                <span v-if="!loading">{{ $t("Change") }}</span>
                                <span v-else>change ...</span>
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-card>
            </div>
            <div class="card-container">
                <el-card class="card">
                    <h3>Receive Miner Owner</h3>
                    <el-form ref="receiveOwnerForm" :model="receiveOwnerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId">
                            <el-input placeholder="please input minerId" v-model="receiveOwnerFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="NewOwner" prop="newOwnerAddress">
                            <el-input placeholder="please input new owner address"
                                v-model="receiveOwnerFormData.newOwnerAddress"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="receieLoading" size="medium" type="primary"
                                @click="submitReceiveOwnerForm">
                                <span v-if="!receieLoading">{{ $t("Receive") }}</span>
                                <span v-else>receive ...</span>
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-card>
            </div>
        </el-row>
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
import { changeOwner } from "@/api/openfil/miner.js";

export default {
    data() {
        return {
            changeOwnerFormData: {
                minerId: '',
                newOwnerAddress: '',
            },
            receiveOwnerFormData: {
                minerId: '',
                newOwnerAddress: '',
            },
            ownerFormRules: {
                minerId: [
                    { required: true, message: 'Please input Miner ID', trigger: 'blur' },
                ],
                newOwnerAddress: [
                    { required: true, message: 'Please input new Owner Address', trigger: 'blur' },
                ],
            },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
            receieLoading: false,
        };
    },
    methods: {
        submitChangeOwnerForm() {
            this.loading = true;
            changeOwner(this.changeOwnerFormData.minerId, this.changeOwnerFormData.newOwnerAddress, '').then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
                this.dialogVisible = true;
                this.loading = false;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            });
        },
        submitReceiveOwnerForm() {
            this.receieLoading = true;
            changeOwner(this.receiveOwnerFormData.minerId, this.receiveOwnerFormData.newOwnerAddress, this.receiveOwnerFormData.newOwnerAddress).then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
                this.dialogVisible = true;
                this.receieLoading = false;
            }).catch(error => {
                this.receieLoading = false;
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
            // this.dialogVisible = false; // Hide the dialog after copying
            this.$message({
                message: 'Transaction copied!',
                type: 'success'
            });
        },
    },
}
</script>

  
<style>
.container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
}

.card {
    width: 800px;
    margin: 50px;
}
</style>
  

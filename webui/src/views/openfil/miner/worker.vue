<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Worker</h3>
                    <el-form ref="changeWorkerForm" :model="changeWorkerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId">
                            <el-input placeholder="please input minerId" v-model="changeWorkerFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="NewWorker" prop="newWorkerAddress">
                            <el-input placeholder="please input new worker address"
                                v-model="changeWorkerFormData.newWorkerAddress"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="loading" size="medium" type="primary" @click="submitChangeWorkerForm">
                                <span v-if="!loading">{{ $t("Change") }}</span>
                                <span v-else>change ...</span>
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-card>
            </div>
            <div class="card-container">
                <el-card class="card">
                    <h3>Confirm Worker</h3>
                    <el-form ref="confirmWorkerForm" :model="confirmWorkerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId">
                            <el-input placeholder="please input minerId" v-model="confirmWorkerFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="NewWorker" prop="newWorkerAddress">
                            <el-input placeholder="please input new worker address"
                                v-model="confirmWorkerFormData.newWorkerAddress"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="confirmLoading" size="medium" type="primary"
                                @click="submitConfirmWorkerForm">
                                <span v-if="!confirmLoading">{{ $t("Confirm") }}</span>
                                <span v-else>confirm ...</span>
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
import { changeWorker, confirmWorker } from "@/api/openfil/miner.js";

export default {
    data() {
        return {
            changeWorkerFormData: {
                minerId: '',
                newWorkerAddress: '',
            },
            confirmWorkerFormData: {
                minerId: '',
                newWorkerAddress: '',
            },
            ownerFormRules: {
                minerId: [
                    { required: true, message: 'Please input Miner ID', trigger: 'blur' },
                ],
                newWorkerAddress: [
                    { required: true, message: 'Please input new worker Address', trigger: 'blur' },
                ],
            },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
            confirmLoading: false,
        };
    },
    methods: {
        submitChangeWorkerForm() {
            this.loading = true;
            changeWorker(this.changeWorkerFormData.minerId, this.changeWorkerFormData.newWorkerAddress).then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
                this.dialogVisible = true;
                this.loading = false;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            });
        },
        submitConfirmWorkerForm() {
            this.confirmLoading = true;
            confirmWorker(this.confirmWorkerFormData.minerId, this.confirmWorkerFormData.newWorkerAddress).then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
                this.dialogVisible = true;
                this.confirmLoading = false;
            }).catch(error => {
                this.confirmLoading = false;
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
  

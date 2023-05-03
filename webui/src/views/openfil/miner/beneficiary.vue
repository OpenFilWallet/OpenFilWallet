<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Beneficiary</h3>
                    <el-form ref="changeBeneficiaryForm" :model="changeBeneficiaryFormData" :rules="formRules"
                        label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId" label-width="120px">
                            <el-input placeholder="please input minerId"
                                v-model="changeBeneficiaryFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="NewBeneficiary" prop="newBeneficiaryAddress" label-width="120px">
                            <el-input placeholder="please input new beneficiary address"
                                v-model="changeBeneficiaryFormData.newBeneficiaryAddress"></el-input>
                        </el-form-item>
                        <el-form-item label="Quota" prop="quota" label-width="120px">
                            <el-input placeholder="please input quota (FIL)"
                                v-model="changeBeneficiaryFormData.quota"></el-input>
                        </el-form-item>
                        <el-form-item label="Expiration" prop="expiration" label-width="120px">
                            <el-input placeholder="please input expiration (epoch)"
                                v-model="changeBeneficiaryFormData.expiration"></el-input>
                        </el-form-item>
                        <el-form-item label="Overwrite" prop="overwrite" label-width="120px">
                            <el-checkbox v-model="changeBeneficiaryFormData.overwrite"></el-checkbox>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="loading" size="medium" type="primary" @click="submitChangeBeneficiaryForm">
                                <span v-if="!loading">{{ $t("Change") }}</span>
                                <span v-else>change ...</span>
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-card>
            </div>
            <div class="card-container">
                <el-card class="card">
                    <h3>Confirm Beneficiary</h3>
                    <el-form ref="confirmBeneficiaryForm" :model="confirmBeneficiaryFormData" :rules="formRules"
                        label-width="100px">
                        <el-form-item label="Miner ID" prop="minerId" label-width="120px">
                            <el-input placeholder="please input minerId"
                                v-model="confirmBeneficiaryFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="confirmLoading" size="medium" type="primary"
                                @click="submitConfirmBeneficiaryForm">
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
import { changeBeneficiary, confirmBeneficiary } from "@/api/openfil/miner.js";

export default {
    data() {
        return {
            changeBeneficiaryFormData: {
                minerId: '',
                newBeneficiaryAddress: '',
                quota: '',
                expiration: '',
                overwrite: false
            },
            confirmBeneficiaryFormData: {
                minerId: '',
            },
            formRules: {
                minerId: [
                    { required: true, message: 'Please input Miner ID', trigger: 'blur' },
                ],
                newBeneficiaryAddress: [
                    { required: true, message: 'Please input new worker Address', trigger: 'blur' },
                ],
                quota: [
                    { required: true, message: 'Quota is required', trigger: 'blur' }
                ],
                expiration: [
                    { required: true, message: 'Expiration is required', trigger: 'blur' }
                ]
            },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
            confirmLoading: false,
        };
    },
    methods: {
        submitChangeBeneficiaryForm() {
            this.loading = true;
            changeBeneficiary(this.changeBeneficiaryFormData.minerId, this.changeBeneficiaryFormData.newBeneficiaryAddress, this.changeBeneficiaryFormData.quota, this.changeBeneficiaryFormData.expiration, this.changeBeneficiaryFormData.overwrite).then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
                this.dialogVisible = true;
                this.loading = false;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            });
        },
        submitConfirmBeneficiaryForm() {
            this.confirmLoading = true;
            confirmBeneficiary(this.confirmBeneficiaryFormData.minerId).then(response => {
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
    margin: 70px;
}
</style>
  

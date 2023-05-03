<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Worker Propose</h3>
                    <el-form ref="changeWorkerForm" :model="changeWorkerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Msig:" required>
                            <el-select v-model="changeWorkerFormData.msig" placeholder="please select msig wallet">
                                <el-option v-for="account in msigAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="From:" required>
                            <el-select v-model="changeWorkerFormData.from" placeholder="please select from address">
                                <el-option v-for="account in filteredAddressLabels2" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
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
                    <h3>Confirm Worker Propose</h3>
                    <el-form ref="confirmWorkerForm" :model="confirmWorkerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Msig:" required>
                            <el-select v-model="confirmWorkerFormData.msig" placeholder="please select msig wallet">
                                <el-option v-for="account in msigAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="From:" required>
                            <el-select v-model="confirmWorkerFormData.from" placeholder="please select from address">
                                <el-option v-for="account in filteredAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
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
import { walletList, msigWalletList } from "@/api/openfil/wallet.js";
import { msigChangeWorker, msigConfirmChangeWorker } from "@/api/openfil/msig.js";
export default {
    data() {
        return {
            msigWallets: [],
            addresses: [],
            changeWorkerFormData: {
                msig: '',
                from: '',
                minerId: '',
                newWorkerAddress: '',
            },
            confirmWorkerFormData: {
                msig: '',
                from: '',
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
            queryParams: { balance: true },
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
            if (!this.confirmWorkerFormData.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.confirmWorkerFormData.msig)
                return this.addressLabels.filter(address => selectedMsigWallet.signers.includes(address.id))
            }
        },
        filteredAddressLabels2() {
            if (!this.changeWorkerFormData.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.changeWorkerFormData.msig)
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
        submitChangeWorkerForm() {
            this.loading = true;
            msigChangeWorker(this.changeWorkerFormData.from, this.changeWorkerFormData.msig, this.changeWorkerFormData.minerId, this.changeWorkerFormData.newWorkerAddress).then(response => {
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
            msigConfirmChangeWorker(this.confirmWorkerFormData.from, this.confirmWorkerFormData.msig, this.confirmWorkerFormData.minerId, this.confirmWorkerFormData.newWorkerAddress).then(response => {
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
  

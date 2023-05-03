<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Miner Owner Propose</h3>
                    <el-form ref="changeOwnerForm" :model="changeOwnerFormData" :rules="ownerFormRules" label-width="100px">
                        <el-form-item label="Msig:" required>
                            <el-select v-model="changeOwnerFormData.msig" placeholder="please select msig wallet">
                                <el-option v-for="account in msigAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="From:" required>
                            <el-select v-model="changeOwnerFormData.from" placeholder="please select from address">
                                <el-option v-for="account in filteredAddressLabels2" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
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
                    <h3>Receive Miner Owner Propose</h3>
                    <el-form ref="receiveOwnerForm" :model="receiveOwnerFormData" :rules="ownerFormRules"
                        label-width="100px">
                        <el-form-item label="Msig:" required>
                            <el-select v-model="receiveOwnerFormData.msig" placeholder="please select msig wallet"
                                @change="onFromMsigChange">
                                <el-option v-for="account in msigAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                    <div slot="label">
                                        <span>{{ account.label }}</span>
                                        <span class="id">{{ account.id }}</span>
                                    </div>
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="From:" required>
                            <el-select v-model="receiveOwnerFormData.from" placeholder="please select from address"
                                @change="onFromChange">
                                <el-option v-for="account in filteredAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                    <div slot="label">
                                        <span>{{ account.label }}</span>
                                        <span class="balance">{{ account.balance }}</span>
                                    </div>
                                </el-option>
                            </el-select>
                        </el-form-item>
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
import { walletList, msigWalletList } from "@/api/openfil/wallet.js";
import { msigChangeOwner } from "@/api/openfil/msig.js";
export default {
    data() {
        return {
            msigWallets: [],
            addresses: [],
            changeOwnerFormData: {
                msig: '',
                from: '',
                minerId: '',
                newOwnerAddress: '',
            },
            receiveOwnerFormData: {
                msig: '',
                from: '',
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
            if (!this.receiveOwnerFormData.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.receiveOwnerFormData.msig)
                return this.addressLabels.filter(address => selectedMsigWallet.signers.includes(address.id))
            }
        },
        filteredAddressLabels2() {
            if (!this.changeOwnerFormData.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.changeOwnerFormData.msig)
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
        submitChangeOwnerForm() {
            this.loading = true;
            msigChangeOwner(this.changeOwnerFormData.minerId, this.changeOwnerFormData.newOwnerAddress, '').then(response => {
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
            msigChangeOwner(this.receiveOwnerFormData.minerId, this.receiveOwnerFormData.newOwnerAddress, this.receiveOwnerFormData.newOwnerAddress).then(response => {
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
  

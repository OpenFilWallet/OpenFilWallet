<template>
    <div class="container">
        <el-row>
            <div class="card-container">
                <el-card class="card">
                    <h3>Change Miner Control Address Propose</h3>
                    <el-form ref="changeControlForm" :model="changeControlFormData" :rules="formRules" label-width="100px">
                        <el-form-item label="Msig:" required>
                            <el-select v-model="changeControlFormData.msig" placeholder="please select msig wallet">
                                <el-option v-for="account in msigAddressLabels" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="From:" required>
                            <el-select v-model="changeControlFormData.from" placeholder="please select from address">
                                <el-option v-for="account in filteredAddressLabels2" :key="account.address"
                                    :label="account.labelWithId" :value="account.address">
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item label="Miner ID" prop="minerId" label-width="120px">
                            <el-input placeholder="please input minerId" v-model="changeControlFormData.minerId"></el-input>
                        </el-form-item>
                        <el-form-item label="Controllers" prop="controlAddrs" label-width="120px">
                            <div v-for="(control, index) in changeControlFormData.controlAddrs" :key="index">
                                <el-input v-model="control.address" />
                                <el-button v-if="index === changeControlFormData.controlAddrs.length - 1"
                                    @click="addControlAddress(index)">
                                    +
                                </el-button>
                                <el-button v-else @click="removeControlAddress(index)">-</el-button>
                            </div>
                        </el-form-item>
                        <el-form-item>
                            <el-button :loading="loading" size="medium" type="primary" @click="submitchangeControlForm">
                                <span v-if="!loading">{{ $t("Change") }}</span>
                                <span v-else>change ...</span>
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
import { msigChangeControl } from "@/api/openfil/msig.js";

export default {
    data() {
        return {
            msigWallets: [],
            addresses: [],
            changeControlFormData: {
                msig: '',
                from: '',
                minerId: '',
                controlAddrs: [{ address: '' }]
            },
            formRules: {
                minerId: [
                    { required: true, message: 'Please input Miner ID', trigger: 'blur' },
                ],
                controlAddrs: [
                    { required: true, message: 'Please input new control addrs', trigger: 'blur' },
                ],
            },
            dialogVisible: false,
            transactionResult: '',
            loading: false,
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
            if (!this.changeControlFormData.msig) {
                return this.addressLabels
            } else {
                const selectedMsigWallet = this.msigWallets.find(wallet => wallet.address === this.changeControlFormData.msig)
                return this.addressLabels.filter(address => selectedMsigWallet.signers.includes(address.id))
            }
        },
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
        addControlAddress(index) {
            this.changeControlFormData.controlAddrs.splice(index + 1, 0, { address: '' })
        },
        removeControlAddress(index) {
            this.changeControlFormData.controlAddrs.splice(index, 1)
        },
        submitchangeControlForm() {
            this.loading = true;
            const controlAddresses = this.changeControlFormData.controlAddrs.map(control => control.address)
            msigChangeControl(this.changeControlFormData.from, this.changeControlFormData.msig, this.changeControlFormData.minerId, controlAddresses).then(response => {
                console.log(response);
                this.transactionResult = JSON.stringify(response, null, 4);
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
  

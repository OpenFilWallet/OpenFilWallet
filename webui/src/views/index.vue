<template>
  <div class="app-container home">
    <div v-if="false">
      <wallet-guide></wallet-guide>
    </div>
    <div v-else id="console">
      <el-row :gutter="20">
        <el-col :span="23">
          <el-card>
            <div slot="header"
              style="font-size: 20px; display: flex; justify-content: space-between; align-items: center;">
              <span>Normal Wallet</span>
              <div style="display: flex;">
                <el-button type="primary" style="flex: 1; width: 50%;" @click="showCreateDialog">Create</el-button>
              </div>
            </div>
            <div>
              <el-table :data="wallets" stripe border>
                <el-table-column label="Address" prop="address" min-width="190"></el-table-column>
                <el-table-column label="ID" prop="id" min-width="30"></el-table-column>
                <el-table-column label="Balance" prop="balance" min-width="60"></el-table-column>
                <el-table-column label="Path" prop="path" min-width="40"></el-table-column>
                <el-table-column label="Tx history" min-width="30">
                  <template slot-scope="scope">
                    <span style="color: blue; text-decoration: underline; cursor: pointer;"
                      @click="showTransactionList(scope.row.address)">detail</span>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>
        </el-col>
      </el-row>
      <div style="height: 20px;"></div>
      <el-row :gutter="20">
        <el-col :span="23">
          <el-card>
            <div slot="header"
              style="font-size: 20px; display: flex; justify-content: space-between; align-items: center;">
              <span>Multisig Wallet</span>
              <div style="display: flex;">
                <el-button type="primary" style="flex: 1; width: 50%;" @click="showMsigCreateDialog">Create</el-button>
                <el-button type="primary" style="flex: 1; width: 50%;" @click="showAddMsigDialog">Add</el-button>
              </div>
            </div>
            <div>
              <el-table :data="msigWallets" stripe border>
                <el-table-column label="Address" prop="address" min-width="100"></el-table-column>
                <el-table-column label="ID" prop="id" min-width="50"></el-table-column>
                <el-table-column label="Balance" prop="balance" min-width="60"></el-table-column>
                <el-table-column label="Threshold" prop="num_approvals_threshold" min-width="60"></el-table-column>
                <el-table-column label="Signers" min-width="70">
                  <template slot-scope="scope">
                    {{ formatSigners(scope.row.signers) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-dialog title="Create Wallet" :visible.sync="createDialogVisible">
        <el-form :model="createForm" ref="createForm">
          <el-form-item label="Index" prop="index">
            <el-input v-model.number="createForm.index"
              placeholder="Please enter the wallet index, defaults to increment if not entered"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="createDialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="createWallet">Create</el-button>
        </div>
      </el-dialog>
      <el-dialog title="Create Msig Wallet" :visible.sync="msigCreateDialogVisible">
        <el-form :model="createMsigForm" ref="createMsigForm">
          <el-form-item label="From" prop="from">
            <el-select v-model="createMsigForm.from" placeholder="Select wallet">
              <el-option v-for="(account, index) in wallets" :key="index" :label="account.address"
                :value="account.address" />
            </el-select>
          </el-form-item>
          <el-form-item label="Required" prop="required">
            <el-input v-model="createMsigForm.required" />
          </el-form-item>
          <el-form-item label="Duration" prop="duration">
            <el-input v-model="createMsigForm.duration" />
          </el-form-item>
          <el-form-item label="Value" prop="value">
            <el-input v-model="createMsigForm.value" />
          </el-form-item>
          <el-form-item label="Signers" prop="signers">
            <div v-for="(signer, index) in createMsigForm.signers" :key="index">
              <el-input v-model="signer.address" />
              <el-button v-if="index === createMsigForm.signers.length - 1" @click="addSigner(index)">
                +
              </el-button>
              <el-button v-else @click="removeSigner(index)">-</el-button>
            </div>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="msigCreateDialogVisible = false">Cancel</el-button>
          <el-button :loading="loading" size="medium" type="primary" @click="createMsigWallet">
            <span v-if="!loading">{{ $t("Create") }}</span>
            <span v-else>create...</span>
          </el-button>
        </div>
      </el-dialog>
      <el-dialog title="Add Msig Wallet" :visible.sync="addMsigDialogVisible">
        <el-form :model="addMsigForm" ref="createForm">
          <el-form-item label="msigAddress" prop="index">
            <el-input v-model.number="addMsigForm.msigAddress" placeholder="Please enter msig wallet address"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="addMsigDialogVisible = false">Cancel</el-button>
          <el-button :loading="loading" size="medium" type="primary" @click="addMsigWallet">
            <span v-if="!loading">{{ $t("Add") }}</span>
            <span v-else>add...</span>
          </el-button>
        </div>
      </el-dialog>
      <el-dialog title="Transaction List" :visible.sync="transactionListVisible" :width="'70%'" :height="'70vh'">
        <div v-if="currentTransactionList.length === 0">No local transaction history found.</div>
        <el-table :data="currentTransactionList" stripe border v-else>
          <el-table-column label="Version" prop="version"></el-table-column>
          <el-table-column label="From" prop="from"></el-table-column>
          <el-table-column label="To" prop="to"></el-table-column>
          <el-table-column label="Nonce" prop="nonce"></el-table-column>
          <el-table-column label="Value" prop="value"></el-table-column>
          <el-table-column label="Gas Limit" prop="gas_limit"></el-table-column>
          <el-table-column label="Gas FeeCap" prop="gas_feecap"></el-table-column>
          <el-table-column label="Gas Premium" prop="gas_premium"></el-table-column>
          <el-table-column label="Method" prop="method"></el-table-column>
          <el-table-column label="Params" prop="params"></el-table-column>
          <el-table-column label="Tx CID" prop="tx_cid"></el-table-column>
          <el-table-column label="State" prop="tx_state"></el-table-column>
        </el-table>
      </el-dialog>
      <el-dialog title="Msig Create Transaction Result" :visible.sync="dialogMsigCreateVisible">
        <pre>{{ transactionResult }}</pre>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogMsigCreateVisible = false">Close</el-button>
          <el-button type="primary" @click="copyToClipboard">Copy</el-button>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script>
import { walletList, msigWalletList, create, msigCreate, msigAdd, txHistory } from "@/api/openfil/wallet.js";
export default {
  data() {
    return {
      wallets: [],
      msigWallets: [],
      queryParams: { balance: true },
      createDialogVisible: false,
      msigCreateDialogVisible: false,
      addMsigDialogVisible: false,
      createForm: {
        index: ''
      },
      addMsigForm: {
        msigAddress: '',
      },
      createMsigForm: {
        from: '',
        required: '',
        duration: '',
        value: '',
        signers: [{ address: '' }]
      },

      transactionListVisible: false,
      currentTransactionList: [],
      transactionResult: '',
      dialogMsigCreateVisible: false,
      loading: false,
    };
  },
  created() {
    this.fetchWalletList();
    this.fetchMsigWalletList();
  },
  methods: {
    fetchWalletList() {
      walletList(this.queryParams).then(response => {
        this.wallets = response;
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
    formatSigners(signers) {
      return signers.join(', ');
    },
    showCreateDialog() {
      this.createDialogVisible = true
    },
    showMsigCreateDialog() {
      this.msigCreateDialogVisible = true
    },
    showAddMsigDialog() {
      this.addMsigDialogVisible = true
    },
    createWallet() {
      let index = -1;
      const parsedIndex = parseInt(this.createForm.index);
      if (!isNaN(parsedIndex)) {
        index = parsedIndex;
      } else {
        console.error('Unable to parse index as integer');
      }

      create(index).then(response => {
        walletList(this.queryParams).then(response => {
          this.wallets = response;
        }).catch(error => {
          console.log(error);
        });
      }).catch(error => {
        console.log(error);
      });
      this.createDialogVisible = false;
      this.createForm.index = '';
    },
    showTransactionList(row) {
      this.transactionListVisible = true;
      txHistory(row).then(response => {
        this.currentTransactionList = response;
      }).catch(error => {
        console.log(error);
      });
      this.currentTransactionList = [];
    },
    addSigner(index) {
      this.createMsigForm.signers.splice(index + 1, 0, { address: '' })
    },
    removeSigner(index) {
      this.createMsigForm.signers.splice(index, 1)
    },
    createMsigWallet() {
      this.loading = true;
      const required = parseInt(this.createMsigForm.required);
      const duration = parseInt(this.createMsigForm.duration);
      const signerAddresses = this.createMsigForm.signers.map(signer => signer.address)
      msigCreate(this.createMsigForm.from, required, duration, this.createMsigForm.value, signerAddresses).then(response => {
        this.transactionResult = JSON.stringify(response, null, 4);
        this.dialogMsigCreateVisible = true;
        this.msigCreateDialogVisible = false;
      }).catch(error => {
        console.log(error);
      });

      this.loading = false;
    },
    addMsigWallet() {
      this.loading = true;
      msigAdd(this.addMsigForm.msigAddress).then(response => {
        msigWalletList(this.queryParams).then(response => {
          this.msigWallets = response;
        }).catch(error => {
          console.log(error);
        });
      }).catch(error => {
        console.log(error);
      });
      this.loading = false;
      this.addMsigDialogVisible = false;
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
};
</script>

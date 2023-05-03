<template>
    <div class="page-container">
        <div class="tool-container">
            <el-tabs v-model="activeTab" @tab-click="reset" active-tab-class="active-tab">
                <el-tab-pane label="Encode">
                    <div class="tool-content">
                        <el-form ref="encodeForm" label-width="100px">
                            <el-form-item label="To">
                                <el-input v-model="encodeForm.to" placeholder="To Address, such as：t01001"></el-input>
                            </el-form-item>
                            <el-form-item label="Method">
                                <el-input v-model="encodeForm.methodId" placeholder="Method ID, such as：23"></el-input>
                            </el-form-item>
                            <el-form-item label="Params">
                                <el-input v-model="encodeForm.params" placeholder='Params, such as："t01001"'></el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button :loading="loading" size="medium" type="primary" @click="handleEncode">
                                    <span v-if="!loading">{{ $t("Encode") }}</span>
                                    <span v-else>encode...</span>
                                </el-button>
                            </el-form-item>
                        </el-form>
                        <el-card v-if="encodedResult" class="result-card">
                            <div class="result-title">Encoded Result:</div>
                            <div class="result-value">{{ encodedResult }}</div>
                        </el-card>
                    </div>
                </el-tab-pane>
                <el-tab-pane label="Decode">
                    <div class="tool-content">
                        <el-form ref="decodeForm" label-width="100px">
                            <el-form-item label="To">
                                <el-input v-model="decodeForm.to" placeholder="To Address, such as：t01001"></el-input>
                            </el-form-item>
                            <el-form-item label="Method">
                                <el-input v-model="decodeForm.methodId" placeholder="Method ID, such as：23"></el-input>
                            </el-form-item>
                            <el-form-item label="Params">
                                <el-input v-model="decodeForm.params" placeholder='Params, such as：4300e807'></el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button :loading="loading" size="medium" type="primary" @click="handleDecode">
                                    <span v-if="!loading">{{ $t("Decode") }}</span>
                                    <span v-else>decode...</span>
                                </el-button>
                            </el-form-item>
                        </el-form>
                        <el-card v-if="decodedResult" class="result-card">
                            <div class="result-title">Decoded Result:</div>
                            <div class="result-value">{{ decodedResult }}</div>
                        </el-card>
                    </div>
                </el-tab-pane>

                <el-tab-pane label="Balance">
                    <div class="tool-content">
                        <el-form ref="balanceForm" label-width="100px">
                            <el-form-item label="Address">
                                <el-input v-model="balanceForm.address" placeholder="Inquiry address"></el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button :loading="loading" size="medium" type="primary" @click="handleBalance">
                                    <span v-if="!loading">{{ $t("Balance") }}</span>
                                    <span v-else>balance...</span>
                                </el-button>
                            </el-form-item>
                        </el-form>
                        <el-card v-if="balanceResult" class="result-card">
                            <div class="result-title">Balance:</div>
                            <div class="result-value">{{ balanceResult }}</div>
                        </el-card>
                    </div>
                </el-tab-pane>

                <el-tab-pane label="Msig Inspect">
                    <div class="tool-content">
                        <el-form ref="msigInspectForm" label-width="100px">
                            <el-form-item label="Address">
                                <el-input v-model="msigInspectForm.address" placeholder="Inquiry msig address"></el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button :loading="loading" size="medium" type="primary" @click="handleMsigInspect">
                                    <span v-if="!loading">{{ $t("Msig Inspect") }}</span>
                                    <span v-else>inspect...</span>
                                </el-button>
                            </el-form-item>
                        </el-form>
                        <el-card v-if="msigInspectResult" class="result-card">
                            <div class="result-title">Msig Inspect:</div>
                            <pre class="result-value">{{ msigInspectResult }}</pre>
                        </el-card>
                    </div>
                </el-tab-pane>
                <el-tab-pane label="Miner Control">
                    <div class="tool-content">
                        <el-form ref="minerForm" label-width="100px">
                            <el-form-item label="minerId">
                                <el-input v-model="minerForm.minerId" placeholder="Inquiry miner id"></el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button :loading="loading" size="medium" type="primary" @click="handleMinerControl">
                                    <span v-if="!loading">{{ $t("Miner Control") }}</span>
                                    <span v-else>control...</span>
                                </el-button>
                            </el-form-item>
                        </el-form>
                        <el-card v-if="minerControlResult" class="result-card">
                            <div class="result-title">Miner Control:</div>
                            <pre class="result-value">{{ minerControlResult }}</pre>
                        </el-card>
                    </div>
                </el-tab-pane>
            </el-tabs>
        </div>
    </div>
</template>
  
<script>
import { decode, encode, balance, msigInspect, minerControl } from "@/api/openfil/tool.js";
export default {
    data() {
        return {
            activeTab: '',
            encodeForm: {
                to: '',
                methodId: '',
                params: "",
            },
            decodeForm: {
                to: '',
                methodId: '',
                params: "",
            },
            balanceForm: {
                address: '',
            },
            msigInspectForm: {
                address: '',
            },
            minerForm: {
                miner: '',
            },
            encodedResult: '',
            decodedResult: '',
            balanceResult: '',
            msigInspectResult: '',
            minerControlResult: '',
            loading: false,
        };
    },
    methods: {
        handleEncode() {
            this.loading = true;
            encode(this.encodeForm.to, parseInt(this.encodeForm.methodId), this.encodeForm.params).then(response => {
                this.loading = false;
                this.encodedResult = response;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            })
        },

        handleDecode() {
            this.loading = true;
            decode(this.decodeForm.to, parseInt(this.decodeForm.methodId), this.decodeForm.params).then(response => {
                this.loading = false;
                this.decodedResult = response;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            })
        },
        handleBalance() {
            this.loading = true;
            balance(this.balanceForm.address).then(response => {
                this.loading = false;
                this.balanceResult = response.amount;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            })
        },
        handleMsigInspect() {
            this.loading = true;
            msigInspect(this.msigInspectForm.address).then(response => {
                this.loading = false;
                this.msigInspectResult = response;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            })
        },
        handleMinerControl() {
            this.loading = true;
            minerControl(this.minerForm.minerId).then(response => {
                this.loading = false;
                this.minerControlResult = response;
            }).catch(error => {
                this.loading = false;
                console.error(error);
            })
        },
        reset() {
            this.encodedResult = '';
            this.decodedResult = '';
            this.balanceResult = '';
            this.msigInspectResult = '';
            this.minerControlResult = '';
        }

    }

};
</script>

<style>
.page-container {
    margin: 0 auto;
    max-width: 900px;
    padding: 40px;
}

.tool-container {
    background-color: #f0f2f5;
    border-radius: 5px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    margin-bottom: 20px;
    overflow: hidden;
}

.el-tabs {
    background-color: #fff;
    border-bottom: 1px solid #ddd;
}

.tool-content {
    padding: 30px;
}

.el-form-item__label {
    font-weight: bold;
}

.result-card {
    margin-top: 20px;
}

.result-title {
    font-weight: bold;
    margin-bottom: 10px;
}

.result-value {
    word-break: break-all;
}

@media (max-width: 767px) {
    .page-container {
        padding: 10px;
    }
}
</style>

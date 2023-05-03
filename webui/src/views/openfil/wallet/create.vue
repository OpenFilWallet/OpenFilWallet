<template>
    <div>
        <el-form ref="form" :model="form" label-width="120px">
            <el-form-item label="钱包类型">
                <el-radio-group v-model="form.walletType">
                    <el-radio label="normal">Normal</el-radio>
                    <el-radio label="multisig">Multisig</el-radio>
                </el-radio-group>
            </el-form-item>
        </el-form>
        <div class="modal-footer">
            <el-button @click="cancel()">取消</el-button>
            <el-button type="primary" @click="create()">创建</el-button>
        </div>
    </div>
</template>
  
<script>
export default {
    name: "WalletCreateModal",
    data() {
        return {
            form: {
                walletType: "normal",
            },
        };
    },
    methods: {
        create() {
            // 处理创建钱包的逻辑，可以在这里调用API
            console.log("创建钱包：", this.form);
            this.$emit("create-wallet", this.form);
            this.$emit("close");
        },
        cancel() {
            this.$emit("close");
        },
    },
};
</script>
  
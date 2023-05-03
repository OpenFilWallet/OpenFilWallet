<template>
    <div>
        <div class="btn-container" style="justify-content: flex-start;">
            <el-button type="primary" class="add-node-btn" @click="addDialogVisible = true">Add Node</el-button>
            <el-button type="success" class="check-node-btn" @click="fetchNodeList">Refresh</el-button>
        </div>
        <div class="wrapper">
            <el-table :data="nodeLists" style="width: 100%" :index="true">
                <el-table-column prop="name" label="Name" width="140"></el-table-column>
                <el-table-column prop="endpoint" label="Endpoint">
                    <template slot-scope="{row}">
                        <el-tooltip class="cell-ellipsis" :content="row.endpoint" placement="top-start" :open-delay="500">
                            <div class="cell-ellipsis">{{ row.endpoint }}</div>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column prop="token" label="Token">
                    <template slot-scope="{row}">
                        <el-tooltip class="cell-ellipsis" :content="row.token" placement="top-start" :open-delay="500">
                            <div class="cell-ellipsis">{{ row.token }}</div>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column prop="blockHeight" label="Block Height" width="160"></el-table-column>
                <el-table-column label="Using" width="100">
                    <template slot-scope="scope">
                        <i class="el-icon-check" v-if="scope.row.isUsing"></i>
                    </template>
                </el-table-column>
                <el-table-column label="Action">
                    <template slot-scope="scope">
                        <div style="display: flex; align-items: center">
                            <el-button type="success" size="small" @click="handleUseNode(scope.row)">
                                Use
                            </el-button>
                            <el-button type="primary" size="small" @click="handleEditNode(scope.row)">
                                Edit
                            </el-button>
                            <el-button type="danger" size="small" @click="handleDeleteNode(scope.row)">
                                Delete
                            </el-button>
                        </div>
                    </template>
                </el-table-column>

            </el-table>

            <el-dialog :visible.sync="addDialogVisible" title="Add Node">
                <el-form :model="newNode" label-width="80px">
                    <el-form-item label="Name">
                        <el-input v-model="newNode.name"></el-input>
                    </el-form-item>
                    <el-form-item label="Endpoint">
                        <el-input v-model="newNode.endpoint"></el-input>
                    </el-form-item>
                    <el-form-item label="Token">
                        <el-input v-model="newNode.token"></el-input>
                    </el-form-item>
                </el-form>
                <div slot="footer" class="dialog-footer">
                    <el-button @click="addDialogVisible = false">Cancel</el-button>
                    <el-button type="primary" @click="addNode">Add</el-button>
                </div>
            </el-dialog>

            <el-dialog :visible.sync="dialogVisible" title="Edit Node">
                <el-form :model="currentNode" label-width="80px">
                    <el-form-item label="Name">
                        <el-input v-model="currentNode.name"></el-input>
                    </el-form-item>
                    <el-form-item label="Endpoint">
                        <el-input v-model="currentNode.endpoint"></el-input>
                    </el-form-item>
                    <el-form-item label="Token">
                        <el-input v-model="currentNode.token"></el-input>
                    </el-form-item>
                </el-form>
                <div slot="footer" class="dialog-footer">
                    <el-button @click="dialogVisible = false">Cancel</el-button>
                    <el-button type="primary" @click="saveNode">Save</el-button>
                </div>
            </el-dialog>
        </div>
    </div>
</template>
<script>
import { nodeList, nodeAdd, nodeUpdate, useNode, nodeDelete } from "@/api/openfil/node.js";
export default {
    data() {
        return {
            nodeLists: [],
            dialogVisible: false,
            currentNode: {
                name: '',
                endpoint: '',
                token: ''
            },
            addDialogVisible: false,
            newNode: {
                name: '',
                endpoint: '',
                token: ''
            },
        }
    },
    created() {
        this.fetchNodeList();
    },
    methods: {
        fetchNodeList() {
            nodeList().then(response => {
                this.nodeLists = response;
            }).catch(error => {
                console.log(error);
            });
        },
        handleEditNode(node) {
            this.dialogVisible = true;
            this.currentNode = Object.assign({}, node);
        },
        handleUseNode(node) {
            useNode(node.name).then(response => {
                nodeList().then(response => {
                    this.nodeLists = response;
                }).catch(error => {
                    console.log(error);
                });
            }).catch(error => {
                console.log(error);
            });
        },
        handleDeleteNode(node) {
            this.$confirm('Are you sure to delete this node?', 'Warning', {
                confirmButtonText: 'OK',
                cancelButtonText: 'Cancel',
                type: 'warning'
            }).then(() => {
                nodeDelete(node.name).then(response => {
                    nodeList().then(response => {
                        this.nodeLists = response;
                    }).catch(error => {
                        console.log(error);
                    });
                }).catch(error => {
                    console.log(error);
                });
            });
        },
        saveNode() {
            nodeUpdate(this.currentNode.name, this.currentNode.endpoint, this.currentNode.token).then(response => {
                nodeList().then(response => {
                    this.nodeLists = response;
                }).catch(error => {
                    console.log(error);
                });
            }).catch(error => {
                console.log(error);
            });
            this.dialogVisible = false;
        },
        addNode() {
            nodeAdd(this.newNode.name, this.newNode.endpoint, this.newNode.token).then(response => {
                nodeList().then(response => {
                    this.nodeLists = response;
                }).catch(error => {
                    console.log(error);
                });
            }).catch(error => {
                console.log(error);
            });
            this.newNode.name = '';
            this.newNode.endpoint = '';
            this.newNode.token = '';
            this.addDialogVisible = false;
        },
    }
};
</script>
<style>
.wrapper {
    border: 1px solid #ccc;
    padding: 10px;
    margin: 10px;
}


.btn-container {
    display: flex;
    justify-content: center;
    margin: 20px 0;
}

.add-node-btn {
    margin-left: 10px;
    margin-right: 20px;
}

.cell-ellipsis {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
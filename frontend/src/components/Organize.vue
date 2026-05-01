<script setup>
import { ref, onMounted } from "vue";

import { 
  OrganizeFolder, 
  SelectDirectory, 
  GetHistory, 
  Rollback, 
  GetCustomRules, 
  SaveCustomRule, 
  DeleteCustomRule 
} from "/wailsjs/go/main/App";

const folderPaths = ref([]);
const historyRecords = ref([]);
const customRules = ref([]);
const showRuleModal = ref(false);
const activeTab = ref("organize");

const newRule = ref({
  id: "",
  name: "",
  conditionType: "extension",
  condition: "",
  targetFolder: ""
});

const conditionTypes = [
  { value: "extension", label: "扩展名" },
  { value: "filename_contains", label: "文件名包含" },
  { value: "filename_starts_with", label: "文件名前缀" },
  { value: "size_larger", label: "大小大于 (KB)" },
  { value: "size_smaller", label: "大小小于 (KB)" }
];

const getAllOrganizeOptions = () => {
  const defaultOptions = ["Year", "Month", "File Type"];
  const ruleOptions = customRules.value.map(rule => ({
    id: rule.id,
    name: rule.name
  }));
  return { defaultOptions, ruleOptions };
};

const loadHistory = async () => {
  try {
    historyRecords.value = await GetHistory();
  } catch (err) {
    console.log("Failed to load history:", err);
    historyRecords.value = [];
  }
};

const loadCustomRules = async () => {
  try {
    customRules.value = await GetCustomRules();
  } catch (err) {
    console.log("Failed to load custom rules:", err);
    customRules.value = [];
  }
};

const formatDate = (timestamp) => {
  const date = new Date(timestamp * 1000);
  return date.toLocaleString("zh-CN");
};

const organizeF = async () => {
  try {
    for (let i = 0; i < folderPaths.value.length; i++) {
      const folder = folderPaths.value[i];
      await OrganizeFolder(folder.path, folder.organizeBy);
      console.log(folder.path, folder.organizeBy);
    }
    while (folderPaths.value.length > 0) {
      deleteFolder(0);
    }
    await loadHistory();
  } catch (err) {
    console.log(err);
  }
};

const rollbackRecord = async (recordId) => {
  try {
    await Rollback(recordId);
    await loadHistory();
  } catch (err) {
    console.log("Rollback failed:", err);
    alert("回滚失败: " + err);
  }
};

const selectFolder = async () => {
  try {
    const path = await SelectDirectory();
    if (path) {
      const name = path.split(/[\\/]/).pop();
      folderPaths.value.push({ name: name, path, organizeBy: "File Type" });
    }
  } catch (error) {
    console.error("Error selecting folder:", error);
  }
};

const deleteFolder = (index) => {
  folderPaths.value.splice(index, 1);
};

const updateOrganizeBy = (index, value) => {
  folderPaths.value[index].organizeBy = value;
};

const openRuleModal = (rule = null) => {
  if (rule) {
    newRule.value = { ...rule };
  } else {
    newRule.value = {
      id: "",
      name: "",
      conditionType: "extension",
      condition: "",
      targetFolder: ""
    };
  }
  showRuleModal.value = true;
};

const closeRuleModal = () => {
  showRuleModal.value = false;
};

const saveRule = async () => {
  if (!newRule.value.name || !newRule.value.condition || !newRule.value.targetFolder) {
    alert("请填写所有字段");
    return;
  }

  try {
    await SaveCustomRule(newRule.value);
    closeRuleModal();
    await loadCustomRules();
  } catch (err) {
    console.log("Save rule failed:", err);
    alert("保存规则失败: " + err);
  }
};

const deleteRule = async (ruleId) => {
  if (!confirm("确定要删除这个规则吗？")) {
    return;
  }

  try {
    await DeleteCustomRule(ruleId);
    await loadCustomRules();
  } catch (err) {
    console.log("Delete rule failed:", err);
    alert("删除规则失败: " + err);
  }
};

const getOrganizeByDisplay = (organizeBy) => {
  const { defaultOptions, ruleOptions } = getAllOrganizeOptions();
  if (defaultOptions.includes(organizeBy)) {
    switch (organizeBy) {
      case "Year": return "年份";
      case "Month": return "月份";
      case "File Type": return "文件类型";
      default: return organizeBy;
    }
  }
  const rule = ruleOptions.find(r => r.id === organizeBy);
  return rule ? rule.name : organizeBy;
};

onMounted(async () => {
  await loadHistory();
  await loadCustomRules();
});
</script>

<template>
  <div class="organizer-container">
    <div class="tab-header">
      <button 
        class="tab-button" 
        :class="{ active: activeTab === 'organize' }"
        @click="activeTab = 'organize'"
      >
        整理文件
      </button>
      <button 
        class="tab-button" 
        :class="{ active: activeTab === 'history' }"
        @click="activeTab = 'history'"
      >
        历史记录
      </button>
      <button 
        class="tab-button" 
        :class="{ active: activeTab === 'rules' }"
        @click="activeTab = 'rules'"
      >
        自定义规则
      </button>
    </div>

    <div v-show="activeTab === 'organize'" class="tab-content">
      <div class="organizer">
        <h1>Organize Your files</h1>
      </div>
      <div class="upload-section">
        <h3>Upload Your Folder</h3>
        <div class="button-container">
          <button class="upload-button" @click="selectFolder">Choose Folder</button>
          <button class="upload-button organize-button" @click="organizeF">
            Organize
          </button>
        </div>
      </div>
      <div class="folder-table" style="overflow-x: auto">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Path</th>
              <th>Organize by</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in folderPaths" :key="index">
              <td class="ellipsis" data-label="Name">{{ item.name }}</td>
              <td class="ellipsis" data-label="Path">{{ item.path }}</td>
              <td class="organize-by" data-label="OrganizeBy">
                <select
                  v-model="item.organizeBy"
                  @change="updateOrganizeBy(index, item.organizeBy)"
                >
                  <optgroup label="默认选项">
                    <option value="Year">年份</option>
                    <option value="Month">月份</option>
                    <option value="File Type">文件类型</option>
                  </optgroup>
                  <optgroup label="自定义规则" v-if="customRules.length > 0">
                    <option 
                      v-for="rule in customRules" 
                      :key="rule.id"
                      :value="rule.id"
                    >
                      {{ rule.name }}
                    </option>
                  </optgroup>
                </select>
              </td>
              <td data-label="Action">
                <button class="remove-folder" @click="deleteFolder(index)">
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-show="activeTab === 'history'" class="tab-content">
      <div class="organizer">
        <h1>整理历史记录</h1>
      </div>
      <div class="history-section">
        <div v-if="historyRecords.length === 0" class="empty-state">
          <p>暂无整理历史记录</p>
        </div>
        <div v-else class="history-table" style="overflow-x: auto">
          <table>
            <thead>
              <tr>
                <th>时间</th>
                <th>目录路径</th>
                <th>整理方式</th>
                <th>文件数量</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(record, index) in historyRecords" :key="index">
                <td>{{ formatDate(record.timestamp) }}</td>
                <td class="ellipsis" title="record.folderPath">{{ record.folderPath }}</td>
                <td>{{ getOrganizeByDisplay(record.organizeBy) }}</td>
                <td>{{ record.operations.length }}</td>
                <td>
                  <span 
                    class="status-badge"
                    :class="{ 
                      'rolled-back': record.isRolledBack, 
                      'active': !record.isRolledBack 
                    }"
                  >
                    {{ record.isRolledBack ? '已回滚' : '已整理' }}
                  </span>
                </td>
                <td>
                  <button 
                    class="rollback-button" 
                    :disabled="record.isRolledBack"
                    @click="rollbackRecord(record.id)"
                  >
                    回滚
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'rules'" class="tab-content">
      <div class="organizer">
        <h1>自定义整理规则</h1>
      </div>
      <div class="rules-section">
        <div class="button-container">
          <button class="upload-button" @click="openRuleModal()">添加规则</button>
        </div>
        <div v-if="customRules.length === 0" class="empty-state">
          <p>暂无自定义规则</p>
        </div>
        <div v-else class="rules-table" style="overflow-x: auto">
          <table>
            <thead>
              <tr>
                <th>规则名称</th>
                <th>条件类型</th>
                <th>条件值</th>
                <th>目标文件夹</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(rule, index) in customRules" :key="index">
                <td>{{ rule.name }}</td>
                <td>
                  {{ conditionTypes.find(ct => ct.value === rule.conditionType)?.label || rule.conditionType }}
                </td>
                <td>{{ rule.condition }}</td>
                <td>{{ rule.targetFolder }}</td>
                <td class="action-buttons">
                  <button class="edit-button" @click="openRuleModal(rule)">编辑</button>
                  <button class="remove-folder" @click="deleteRule(rule.id)">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-if="showRuleModal" class="modal-overlay" @click.self="closeRuleModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ newRule.id ? '编辑规则' : '添加规则' }}</h2>
          <button class="close-button" @click="closeRuleModal">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>规则名称</label>
            <input 
              type="text" 
              v-model="newRule.name" 
              placeholder="例如：大文件整理"
            />
          </div>
          <div class="form-group">
            <label>条件类型</label>
            <select v-model="newRule.conditionType">
              <option 
                v-for="type in conditionTypes" 
                :key="type.value"
                :value="type.value"
              >
                {{ type.label }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>条件值</label>
            <input 
              type="text" 
              v-model="newRule.condition" 
              placeholder="根据条件类型填写，如：txt、报告、1024"
            />
            <p class="hint">
              <span v-if="newRule.conditionType === 'extension'">示例：txt, pdf, jpg（不带点）</span>
              <span v-else-if="newRule.conditionType === 'filename_contains'">示例：报告、工作、文档</span>
              <span v-else-if="newRule.conditionType === 'filename_starts_with'">示例：img_、doc_</span>
              <span v-else-if="newRule.conditionType === 'size_larger'">示例：1024（表示大于1MB）</span>
              <span v-else-if="newRule.conditionType === 'size_smaller'">示例：100（表示小于100KB）</span>
            </p>
          </div>
          <div class="form-group">
            <label>目标文件夹名称</label>
            <input 
              type="text" 
              v-model="newRule.targetFolder" 
              placeholder="例如：大文件、文档、图片"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="cancel-button" @click="closeRuleModal">取消</button>
          <button class="save-button" @click="saveRule">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.organizer-container {
  min-height: 100vh;
}

.tab-header {
  display: flex;
  gap: 10px;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.tab-button {
  padding: 10px 20px;
  border: none;
  background-color: #f0f0f0;
  cursor: pointer;
  border-radius: 6px;
  font-family: "Outfit";
  font-size: 14px;
  transition: all 0.2s ease;
}

.tab-button:hover {
  background-color: #e0e0e0;
}

.tab-button.active {
  background-color: #4c956c;
  color: white;
}

.tab-content {
  padding: 20px;
}

.organize-by select {
  outline: none;
  border: none;
  font-family: "Outfit";
}
table {
  width: 100%;
  border: none;
}

th,
td {
  text-align: left;
  padding: 8px;
}
.folder-table, .history-table, .rules-table {
  box-sizing: border-box;
  margin-left: 160px;
  margin-right: 10px;
}

th {
  color: #1a2821;
  opacity: 62%;
  font-family: "Outfit";
}
td {
  color: #1b2821;
  font-family: "Outfit";
}

.ellipsis {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 200px;
}
.organize-button {
  justify-content: flex-end;
  align-items: flex-end;
}
.button-container {
  display: flex;
  justify-content: flex-start;
  gap: 10px;
}
.remove-folder {
  border: none;
  background-color: rgb(241, 91, 91);
  border-radius: 6px;
  width: 80px;
  height: 30px;
  font-family: "Outfit";
  color: #fafefc;
}
.remove-folder:hover {
  background-color: rgb(218, 55, 55);
}

.edit-button {
  border: none;
  background-color: #4c956c;
  border-radius: 6px;
  width: 80px;
  height: 30px;
  font-family: "Outfit";
  color: #fafefc;
  margin-right: 10px;
}

.edit-button:hover {
  background-color: #3b7955;
}

.rollback-button {
  border: none;
  background-color: #f0ad4e;
  border-radius: 6px;
  width: 80px;
  height: 30px;
  font-family: "Outfit";
  color: #fff;
}

.rollback-button:hover:not(:disabled) {
  background-color: #ec971f;
}

.rollback-button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.upload-section {
  height: 140px;
  display: flex;
  flex-direction: column;
  justify-content: left;
  margin-left: 190px;
  margin-top: 20px;
}

.history-section, .rules-section {
  margin-left: 190px;
  margin-top: 20px;
}

.upload-section h3, .history-section h3, .rules-section h3 {
  color: #1b2821;
  font-family: "Outfit";
  font-size: 25px;
}

.upload-button {
  color: #fafefc;
  font-family: "Outfit";
  font-weight: 600;
  font-size: 15px;
  background-color: #4c956c;
  margin-top: 20px;
  width: 155px;
  height: 35px;
  border-radius: 8px;
  border: none;
}

.upload-button:hover{
  background-color: #3b7955;
}

.organizer {
  height: 100px;
  display: flex;
  justify-content: center;
  margin-top: 50px;
  text-align: center;
}

.organizer h1 {
  color: #1b2821;
  font-family: "Outfit";
  font-size: 45px;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.active {
  background-color: #d4edda;
  color: #155724;
}

.status-badge.rolled-back {
  background-color: #f8d7da;
  color: #721c24;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #666;
  font-family: "Outfit";
}

.action-buttons {
  display: flex;
  gap: 5px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.modal-header h2 {
  margin: 0;
  font-family: "Outfit";
  color: #1b2821;
}

.close-button {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
}

.close-button:hover {
  color: #000;
}

.modal-body {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-family: "Outfit";
  color: #1b2821;
  font-weight: 600;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: "Outfit";
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #4c956c;
}

.hint {
  margin-top: 5px;
  font-size: 12px;
  color: #666;
  font-family: "Outfit";
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid #e0e0e0;
}

.cancel-button,
.save-button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-family: "Outfit";
  font-size: 14px;
  cursor: pointer;
}

.cancel-button {
  background-color: #f0f0f0;
  color: #333;
}

.cancel-button:hover {
  background-color: #e0e0e0;
}

.save-button {
  background-color: #4c956c;
  color: white;
}

.save-button:hover {
  background-color: #3b7955;
}
</style>

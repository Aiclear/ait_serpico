<script setup>
import { ref } from "vue";

import { OrganizeFolder, OrganizeFolderWithScript } from "/wailsjs/go/main/App";
import { SelectDirectory } from "/wailsjs/go/main/App";

const folderPaths = ref([]);
const showScriptEditor = ref(false);
const currentScriptIndex = ref(-1);
const customScript = ref(`// 自定义文件整理脚本示例
// 每个文件都会执行此脚本
// file 对象包含以下属性：
// - file.name: 文件名 (例如: "document.txt")
// - file.path: 完整路径
// - file.extension: 文件扩展名 (例如: ".txt")
// - file.size: 文件大小 (字节)
// - file.created: 创建时间
// - file.modified: 修改时间
// - file.isDirectory: 是否是目录
// - file.baseName: 文件名 (不含扩展名)

// 必须设置 result 对象，包含：
// - shouldOrganize: 是否整理此文件 (true/false)
// - targetDirectory: 目标目录名称 (相对于源目录)
// - newFileName: 可选，重命名后的文件名 (留空则不重命名)

var result = {
  shouldOrganize: false,
  targetDirectory: "",
  newFileName: ""
};

// 示例：根据文件扩展名分类
if (file.extension === ".txt" || file.extension === ".doc" || file.extension === ".docx") {
  result.shouldOrganize = true;
  result.targetDirectory = "Documents";
} else if (file.extension === ".jpg" || file.extension === ".png" || file.extension === ".jpeg") {
  result.shouldOrganize = true;
  result.targetDirectory = "Images";
} else if (file.extension === ".mp3" || file.extension === ".wav" || file.extension === ".flac") {
  result.shouldOrganize = true;
  result.targetDirectory = "Audio";
} else if (file.extension === ".mp4" || file.extension === ".avi" || file.extension === ".mkv") {
  result.shouldOrganize = true;
  result.targetDirectory = "Videos";
} else if (file.extension === ".zip" || file.extension === ".rar" || file.extension === ".7z") {
  result.shouldOrganize = true;
  result.targetDirectory = "Archives";
} else if (file.size > 100 * 1024 * 1024) { // 大于100MB
  result.shouldOrganize = true;
  result.targetDirectory = "Large Files";
}`);

const organizeF = async () => {
  try {
    for (let i = 0; i < folderPaths.value.length; i++) {
      const folder = folderPaths.value[i];
      console.log(`正在整理: ${folder.path}, 方式: ${folder.organizeBy}`);
      
      if (folder.organizeBy === "Custom Script") {
        await OrganizeFolderWithScript(folder.path, folder.organizeBy, folder.customScript || "");
      } else {
        await OrganizeFolder(folder.path, folder.organizeBy);
      }
      
      console.log(`整理完成: ${folder.path}`);
    }
    while (folderPaths.value.length > 0) {
      deleteFolder(0);
    }
  } catch (err) {
    console.error("整理过程中发生错误:", err);
    alert(`错误: ${err.message || err}`);
  }
};

const selectFolder = async () => {
  try {
    const path = await SelectDirectory();
    if (path) {
      const name = path.split(/[\\/]/).pop();
      folderPaths.value.push({ 
        name: name, 
        path, 
        organizeBy: "File Type",
        customScript: customScript.value
      });
    }
  } catch (error) {
    console.error("Error selecting folder:", error);
  }
};

const organizeOptions = ["Year", "Month", "File Type", "Custom Script"];

const deleteFolder = (index) => {
  folderPaths.value.splice(index, 1);
};

const openScriptEditor = (index) => {
  currentScriptIndex.value = index;
  customScript.value = folderPaths.value[index].customScript || customScript.value;
  showScriptEditor.value = true;
};

const saveScript = () => {
  if (currentScriptIndex.value >= 0) {
    folderPaths.value[currentScriptIndex.value].customScript = customScript.value;
  }
  showScriptEditor.value = false;
  currentScriptIndex.value = -1;
};

const updateOrganizeBy = (index, value) => {
  folderPaths.value[index].organizeBy = value;
};
</script>

<template>
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
          <th>Script</th>
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
              <option
                v-for="option in organizeOptions"
                :key="option"
                :value="option"
              >
                {{ option }}
              </option>
            </select>
          </td>
          <td data-label="Script">
            <button 
              v-if="item.organizeBy === 'Custom Script'"
              class="edit-script-button" 
              @click="openScriptEditor(index)"
            >
              Edit Script
            </button>
            <span v-else class="not-applicable">N/A</span>
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

  <div v-if="showScriptEditor" class="script-editor-overlay">
    <div class="script-editor">
      <div class="script-editor-header">
        <h3>Custom Script Editor</h3>
        <button class="close-button" @click="showScriptEditor = false">×</button>
      </div>
      <div class="script-editor-body">
        <p class="script-help">
          编写JavaScript脚本来定义文件整理规则。每个文件都会执行此脚本。<br>
          <strong>file</strong> 对象包含文件信息，必须设置 <strong>result</strong> 对象来决定如何处理文件。
        </p>
        <textarea 
          v-model="customScript" 
          class="script-textarea"
          spellcheck="false"
        ></textarea>
      </div>
      <div class="script-editor-footer">
        <button class="cancel-button" @click="showScriptEditor = false">Cancel</button>
        <button class="save-button" @click="saveScript">Save</button>
      </div>
    </div>
  </div>
</template>

<style>
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
  /* max-width: 100px; */
}
.folder-table {
  box-sizing: border-box;
  margin-left: 160px;
  margin-right: 10px;
  /* width: calc(100% - 150px); */
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
}
.organize-button {
  /* display: flex; */
  justify-content: flex-end;
  align-items: flex-end;
  /* margin-left: 450px; */
}
.button-container {
  display: flex;
  justify-content: flex-start; /* Align buttons to the left */
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
  /* transition:  0.1s ease;  */
}
.remove-folder:hover {
  background-color: rgb(218, 55, 55);
}

.upload-section {
  height: 140px;
  display: flex;
  flex-direction: column;
  justify-content: left;
  /* align-items: left; */
  margin-left: 190px;
  margin-top: 20px;
  /* text-align: center; */
}

.upload-section h3 {
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
  /* margin-left: 150px; Adjust this value to match the sidebar width */
  height: 100px;
  display: flex;
  justify-content: center;
  /* align-items: center; */
  margin-top: 50px;
  text-align: center;
}

.organizer h1 {
  color: #1b2821;
  font-family: "Outfit";
  font-size: 45px;
}

.edit-script-button {
  border: none;
  background-color: #4c956c;
  border-radius: 6px;
  width: 100px;
  height: 30px;
  font-family: "Outfit";
  color: #fafefc;
  cursor: pointer;
}

.edit-script-button:hover {
  background-color: #3b7955;
}

.not-applicable {
  color: #888;
  font-style: italic;
}

.script-editor-overlay {
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

.script-editor {
  background-color: white;
  border-radius: 12px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.script-editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.script-editor-header h3 {
  margin: 0;
  color: #1b2821;
  font-family: "Outfit";
  font-size: 20px;
}

.close-button {
  background: none;
  border: none;
  font-size: 28px;
  color: #888;
  cursor: pointer;
  line-height: 1;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
}

.close-button:hover {
  background-color: #f0f0f0;
  color: #333;
}

.script-editor-body {
  padding: 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: auto;
}

.script-help {
  margin: 0 0 15px 0;
  color: #555;
  font-family: "Outfit";
  font-size: 14px;
  line-height: 1.5;
}

.script-textarea {
  flex: 1;
  min-height: 400px;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 15px;
  font-family: "Consolas", "Monaco", monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: none;
  background-color: #fafafa;
  color: #333;
}

.script-textarea:focus {
  outline: none;
  border-color: #4c956c;
  background-color: white;
}

.script-editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid #e0e0e0;
}

.cancel-button {
  border: 1px solid #ccc;
  background-color: white;
  color: #666;
  border-radius: 8px;
  padding: 10px 20px;
  font-family: "Outfit";
  font-size: 14px;
  cursor: pointer;
}

.cancel-button:hover {
  background-color: #f5f5f5;
}

.save-button {
  border: none;
  background-color: #4c956c;
  color: white;
  border-radius: 8px;
  padding: 10px 20px;
  font-family: "Outfit";
  font-size: 14px;
  cursor: pointer;
}

.save-button:hover {
  background-color: #3b7955;
}
</style>

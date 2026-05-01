<script setup>
import { ref, computed } from "vue";

import { OrganizeFolder, PreviewOrganizeFolder, ExecuteOrganizeFolder } from "/wailsjs/go/main/App";
import { SelectDirectory } from "/wailsjs/go/main/App";

const folderPaths = ref([]);
const showPreview = ref(false);
const currentPreview = ref(null);
const isExecuting = ref(false);

const organizeOptions = ["Year", "Month", "File Type"];

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

const previewOrganize = async () => {
  try {
    if (folderPaths.value.length === 0) {
      alert("Please select at least one folder to organize");
      return;
    }

    // For simplicity, we'll preview the first folder first
    // In a real implementation, you might want to handle multiple folders
    const firstFolder = folderPaths.value[0];
    const preview = await PreviewOrganizeFolder(firstFolder.path, firstFolder.organizeBy);
    
    currentPreview.value = preview;
    showPreview.value = true;
  } catch (error) {
    console.error("Error previewing organization:", error);
    alert("Error previewing organization: " + error.message);
  }
};

const executeOrganize = async () => {
  try {
    if (!currentPreview.value) return;

    isExecuting.value = true;

    // Execute the organization with user selections
    await ExecuteOrganizeFolder(currentPreview.value);

    // Remove the first folder from the list since we just organized it
    if (folderPaths.value.length > 0) {
      folderPaths.value.splice(0, 1);
    }

    // Close the preview
    closePreview();
    alert("Organization completed successfully!");
  } catch (error) {
    console.error("Error executing organization:", error);
    alert("Error executing organization: " + error.message);
  } finally {
    isExecuting.value = false;
  }
};

const closePreview = () => {
  showPreview.value = false;
  currentPreview.value = null;
};

const toggleFileSelection = (index) => {
  if (currentPreview.value && currentPreview.value.filesToMove[index]) {
    currentPreview.value.filesToMove[index].selected = !currentPreview.value.filesToMove[index].selected;
  }
};

const toggleAllConflicts = (selected) => {
  if (currentPreview.value) {
    currentPreview.value.filesToMove.forEach(file => {
      if (file.isConflict) {
        file.selected = selected;
      }
    });
  }
};

const selectedFilesCount = computed(() => {
  if (!currentPreview.value) return 0;
  return currentPreview.value.filesToMove.filter(f => f.selected).length;
});

const conflictFilesCount = computed(() => {
  if (!currentPreview.value) return 0;
  return currentPreview.value.filesToMove.filter(f => f.isConflict).length;
});
</script>

<template>
  <div class="organizer">
    <h1>Organize Your files</h1>
  </div>
  <div class="upload-section">
    <h3>Upload Your Folder</h3>
    <div class="button-container">
      <button class="upload-button" @click="selectFolder">Choose Folder</button>
      <button 
        class="upload-button organize-button" 
        @click="previewOrganize"
        :disabled="folderPaths.length === 0"
      >
        Preview & Organize
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
          <td data-label="Action">
            <button class="remove-folder" @click="deleteFolder(index)">
              Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Preview Modal -->
  <div v-if="showPreview && currentPreview" class="preview-modal">
    <div class="preview-content">
      <div class="preview-header">
        <h2>Organization Preview</h2>
        <button class="close-button" @click="closePreview">&times;</button>
      </div>

      <div class="preview-summary">
        <div class="summary-item">
          <span class="summary-label">Source Folder:</span>
          <span class="summary-value">{{ currentPreview.sourcePath }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Organize by:</span>
          <span class="summary-value">{{ currentPreview.organizeBy }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Directories to create:</span>
          <span class="summary-value">{{ currentPreview.directoriesToCreate.length }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">Files to move:</span>
          <span class="summary-value">{{ currentPreview.totalFiles }}</span>
        </div>
        <div class="summary-item conflict-summary" v-if="conflictFilesCount > 0">
          <span class="summary-label">Conflicting files:</span>
          <span class="summary-value conflict-count">{{ conflictFilesCount }}</span>
          <div class="conflict-actions">
            <button class="small-button" @click="toggleAllConflicts(true)">Select All</button>
            <button class="small-button" @click="toggleAllConflicts(false)">Deselect All</button>
          </div>
        </div>
      </div>

      <!-- Directories to Create -->
      <div class="preview-section" v-if="currentPreview.directoriesToCreate.length > 0">
        <h3>Directories to Create:</h3>
        <div class="directory-list">
          <div v-for="(dir, index) in currentPreview.directoriesToCreate" :key="index" class="directory-item">
            <span class="directory-icon">📁</span>
            <span class="directory-name">{{ dir }}</span>
          </div>
        </div>
      </div>

      <!-- Files to Move -->
      <div class="preview-section">
        <h3>Files to Move ({{ selectedFilesCount }} selected):</h3>
        <div class="files-table-container">
          <table class="files-table">
            <thead>
              <tr>
                <th>Select</th>
                <th>File Name</th>
                <th>Source</th>
                <th>Destination</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="(file, index) in currentPreview.filesToMove" 
                :key="index"
                :class="{ 'conflict-row': file.isConflict }"
              >
                <td>
                  <input 
                    type="checkbox" 
                    :checked="file.selected"
                    @change="toggleFileSelection(index)"
                  />
                </td>
                <td :class="{ 'conflict-text': file.isConflict }">
                  {{ file.fileName }}
                </td>
                <td class="ellipsis-path">{{ file.sourcePath }}</td>
                <td class="ellipsis-path">{{ file.destPath }}</td>
                <td>
                  <span v-if="file.isConflict" class="conflict-badge">
                    ⚠️ Conflict
                  </span>
                  <span v-else class="ok-badge">
                    ✓ OK
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="preview-actions">
        <button class="cancel-button" @click="closePreview" :disabled="isExecuting">
          Cancel
        </button>
        <button 
          class="confirm-button" 
          @click="executeOrganize"
          :disabled="isExecuting || selectedFilesCount === 0"
        >
          {{ isExecuting ? 'Organizing...' : `Confirm Organize (${selectedFilesCount} files)` }}
        </button>
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
  width: 170px;
  height: 35px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
}

.upload-button:hover{
  background-color: #3b7955;
}

.upload-button:disabled {
  background-color: #a0a0a0;
  cursor: not-allowed;
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

/* Preview Modal Styles */
.preview-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.preview-content {
  background-color: white;
  border-radius: 12px;
  width: 90%;
  max-width: 1000px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  background-color: #f8f9fa;
}

.preview-header h2 {
  margin: 0;
  font-family: "Outfit";
  color: #1b2821;
}

.close-button {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #666;
  line-height: 1;
}

.close-button:hover {
  color: #333;
}

.preview-summary {
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  background-color: #fafafa;
}

.summary-item {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  font-family: "Outfit";
}

.summary-item:last-child {
  margin-bottom: 0;
}

.summary-label {
  font-weight: 600;
  color: #555;
  min-width: 140px;
}

.summary-value {
  color: #1b2821;
  word-break: break-all;
}

.conflict-summary {
  background-color: #fff3f3;
  padding: 10px;
  border-radius: 6px;
  margin-top: 10px;
}

.conflict-count {
  color: #d32f2f;
  font-weight: 600;
}

.conflict-actions {
  margin-left: 15px;
  display: flex;
  gap: 8px;
}

.small-button {
  padding: 4px 10px;
  font-size: 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: white;
  cursor: pointer;
  font-family: "Outfit";
}

.small-button:hover {
  background-color: #f0f0f0;
}

.preview-section {
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  flex: 1;
  overflow-y: auto;
}

.preview-section h3 {
  margin: 0 0 15px 0;
  font-family: "Outfit";
  color: #1b2821;
  font-size: 18px;
}

.directory-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.directory-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #e8f5e9;
  border-radius: 6px;
  font-family: "Outfit";
  font-size: 14px;
}

.directory-icon {
  font-size: 16px;
}

.directory-name {
  color: #2e7d32;
  word-break: break-all;
}

.files-table-container {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
}

.files-table {
  width: 100%;
  border-collapse: collapse;
}

.files-table th,
.files-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #e0e0e0;
  font-family: "Outfit";
}

.files-table th {
  background-color: #f5f5f5;
  font-weight: 600;
  color: #555;
  position: sticky;
  top: 0;
}

.files-table tr:last-child td {
  border-bottom: none;
}

.ellipsis-path {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  color: #666;
}

.conflict-row {
  background-color: #fff3f3;
}

.conflict-text {
  color: #d32f2f;
  font-weight: 600;
}

.conflict-badge {
  display: inline-block;
  padding: 4px 8px;
  background-color: #ffebee;
  color: #c62828;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.ok-badge {
  display: inline-block;
  padding: 4px 8px;
  background-color: #e8f5e9;
  color: #2e7d32;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.preview-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid #e0e0e0;
  background-color: #f8f9fa;
}

.cancel-button,
.confirm-button {
  padding: 12px 24px;
  border-radius: 8px;
  font-family: "Outfit";
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: background-color 0.2s;
}

.cancel-button {
  background-color: #e0e0e0;
  color: #333;
}

.cancel-button:hover {
  background-color: #bdbdbd;
}

.cancel-button:disabled {
  background-color: #f0f0f0;
  color: #999;
  cursor: not-allowed;
}

.confirm-button {
  background-color: #4c956c;
  color: white;
}

.confirm-button:hover:not(:disabled) {
  background-color: #3b7955;
}

.confirm-button:disabled {
  background-color: #a0a0a0;
  cursor: not-allowed;
}
</style>

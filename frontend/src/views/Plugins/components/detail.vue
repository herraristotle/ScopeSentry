<script setup lang="ts">
import { ref, onBeforeMount, watch } from 'vue'
import {
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElButton,
  ElCol,
  ElRow,
  ElMessage,
  ElTabs,
  ElTabPane,
  ElSpace
} from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { Codemirror } from 'vue-codemirror'
import { javascript } from '@codemirror/lang-javascript'
import { oneDark } from '@codemirror/theme-one-dark'
import { useI18n } from '@/hooks/web/useI18n'
import { getPluginDetailApi, savePluginDataApi } from '@/api/plugins'
import { getManagetListApi, getPortDictDataApi } from '@/api/DictionaryManagement'
import type { fileData, portDictData } from '@/api/DictionaryManagement/types'

const { t } = useI18n()

// Receive the id passed from the parent component
const props = defineProps<{
  closeDialog: () => void
  getList: () => void
  id: string
  tp?: string
  hash?: string
}>()

// Parameter type definition
interface ParameterItem {
  name: string
  type: 'string' | 'bool' | 'dict' | 'port'
  defaultValue?: string
  dictCategory?: string
  dictName?: string
  portName?: string
}

// Form data
const form = ref({
  name: '',
  version: '',
  module: '',
  parameter: '',
  help: '',
  introduction: '',
  source: '',
  parameterList: [] as ParameterItem[]
})

const resetForm = () => {
  console.log('Clear')
  form.value = {
    name: '',
    version: '',
    module: '',
    parameter: '',
    help: '',
    introduction: '',
    source: '',
    parameterList: []
  }
  content.value = '' // Clear the Codemirror editor content
}

// Validation rules
const rules = ref({
  name: [{ required: true, message: '', trigger: 'blur' }],
  module: [{ required: props.tp == 'scan' ? true : false, message: '', trigger: 'change' }],
  source: [{ required: true, message: '', trigger: 'blur' }]
})

// Module options
const moduleOptions = ref([
  { label: 'TargetHandler', value: 'TargetHandler' },
  { label: 'SubdomainScan', value: 'SubdomainScan' },
  { label: 'SubdomainSecurity', value: 'SubdomainSecurity' },
  { label: 'PortScanPreparation', value: 'PortScanPreparation' },
  { label: 'PortScan', value: 'PortScan' },
  { label: 'AssetMapping', value: 'AssetMapping' },
  { label: 'URLScan', value: 'URLScan' },
  { label: 'WebCrawler', value: 'WebCrawler' },
  { label: 'DirScan', value: 'DirScan' },
  { label: 'VulnerabilityScan', value: 'VulnerabilityScan' },
  { label: 'AssetHandle', value: 'AssetHandle' },
  { label: 'PortFingerprint', value: 'PortFingerprint' },
  { label: 'URLSecurity', value: 'URLSecurity' },
  { label: 'PassiveScan', value: 'PassiveScan' }
])

// Codemirror configuration
const content = ref('')
const extensions = [javascript(), oneDark]

// Dictionary and port data
const dictList = ref<fileData[]>([])
const portList = ref<portDictData[]>([])

// Load the dictionary list
const loadDictList = async () => {
  try {
    const res = await getManagetListApi()
    if (res.code === 200) {
      dictList.value = res.data.list
    }
  } catch (error) {
    console.error('Failed to load dictionary list:', error)
  }
}

// Load the port list
const loadPortList = async () => {
  try {
    const res = await getPortDictDataApi('', 1, 1000)
    if (res.code === 200) {
      portList.value = res.data.list
    }
  } catch (error) {
    console.error('Failed to load port list:', error)
  }
}

// Format the parameter list into a parameter string
const formatParameters = (params: ParameterItem[]): string => {
  if (!params || params.length === 0) {
    return ''
  }
  return params
    .filter((param) => param && param.name && param.type)
    .map((param) => {
      if (param.type === 'dict') {
        // dict type: -name {dict.category.name}
        if (param.dictCategory && param.dictName) {
          return `-${param.name} {dict.${param.dictCategory}.${param.dictName}}`
        }
        return null
      }
      if (param.type === 'port') {
        // port type: -name {port.name}
        if (param.portName) {
          return `-${param.name} {port.${param.portName}}`
        }
        return null
      }
      // string and bool types: only format when the default value is set
      if (
        param.defaultValue !== undefined &&
        param.defaultValue !== null &&
        param.defaultValue !== ''
      ) {
        return `-${param.name} ${param.defaultValue}`
      }
      return null
    })
    .filter((item) => item !== null)
    .join(' ')
}

// Handle parameter list changes and update the parameter field
const handleParameterListChange = () => {
  setTimeout(() => {
    const formattedParams = formatParameters(form.value.parameterList)
    form.value.parameter = formattedParams
  }, 0)
}

// Add a parameter
const addParameter = () => {
  form.value.parameterList.push({
    name: '',
    type: 'string'
  })
  handleParameterListChange()
}

// Remove a parameter
const removeParameter = (index: number) => {
  form.value.parameterList.splice(index, 1)
  handleParameterListChange()
}

// Clear related fields when the parameter type changes
const handleParameterTypeChange = (index: number) => {
  const param = form.value.parameterList[index]
  param.defaultValue = undefined
  param.dictCategory = undefined
  param.dictName = undefined
  param.portName = undefined
  handleParameterListChange()
}

// Filter dictionary names by category
const getFilteredDictNames = (category: string | undefined) => {
  if (!category) {
    return []
  }
  return dictList.value.filter((dict) => (dict.category || dict.name) === category)
}

// Handle category change and clear the name
const handleDictCategoryChange = (index: number) => {
  const param = form.value.parameterList[index]
  param.dictName = undefined
  handleParameterListChange()
}

// Check whether this is a create or an edit
onBeforeMount(async () => {
  resetForm()
  // Load the dictionary and port lists
  await loadDictList()
  await loadPortList()
  if (props.id) {
    // If there is an id, query the content for that id
    await fetchData(props.id)
  }
})
const pluginKey = ref('')

const LoadPluginKey = () => {
  const key = localStorage.getItem(`plugin_key`) as string
  pluginKey.value = key
}
LoadPluginKey()
const isSystem = ref(false)
const activeTab = ref('basic')
// Query configuration data by id
const fetchData = async (id: string) => {
  try {
    const res = await getPluginDetailApi(id)
    if (res.code === 200) {
      const data = res.data
      form.value.name = data.name
      form.value.version = data.version
      form.value.module = data.module
      form.value.parameter = data.parameter
      form.value.help = data.help
      form.value.introduction = data.introduction
      content.value = data.source
      isSystem.value = data.isSystem
      // Parse the parameterList JSON string
      if (data.parameterList) {
        try {
          form.value.parameterList = JSON.parse(data.parameterList) as ParameterItem[]
        } catch (error) {
          console.error('Failed to parse parameterList:', error)
          form.value.parameterList = []
        }
      } else {
        form.value.parameterList = []
      }
    } else {
      ElMessage.error(`Failed to load data: ${res.message}`)
    }
  } catch (error) {
    console.error('Error querying data:', error)
  }
}

const saveLoading = ref(false)

// Watch for id changes; re-fetch data if the id changes
watch(
  () => props.id,
  async (newId) => {
    if (newId === '') {
      resetForm()
    } else {
      await fetchData(newId)
    }
  }
)
const save = async () => {
  saveLoading.value = true // start loading state
  if (form.value.name == '') {
    ElMessage.error('name cannot be empty')
    saveLoading.value = false // end loading state
    return
  }
  if (form.value.module == '' && props.tp == 'scan') {
    ElMessage.error('module cannot be empty')
    saveLoading.value = false // end loading state
    return
  }
  if (!isSystem.value) {
    if (content.value == '') {
      ElMessage.error('source code cannot be empty')
      saveLoading.value = false // end loading state
      return
    }
  }
  try {
    // Serialize parameterList to a JSON string
    const parameterListStr = JSON.stringify(form.value.parameterList)
    const res = await savePluginDataApi(
      props.id,
      form.value.name,
      form.value.version,
      form.value.module,
      form.value.parameter,
      form.value.help,
      form.value.introduction,
      content.value,
      pluginKey.value,
      parameterListStr,
      props.tp,
      props.hash
    )
    if (res.code == 505) {
      localStorage.removeItem('plugin_key')
    }
    props.closeDialog()
    props.getList()
  } catch (error) {
    console.error('Error saving data:', error)
    ElMessage.error('Failed to save, please try again later.')
  } finally {
    saveLoading.value = false // end loading state
  }
}
</script>

<template>
  <ElForm :model="form" :rules="rules" label-width="100px">
    <ElTabs v-model="activeTab">
      <!-- Basic info tab -->
      <ElTabPane :label="t('plugin.basicInfo')" name="basic">
        <ElRow :gutter="20">
          <!-- Name -->
          <ElCol :span="12">
            <ElFormItem :label="t('plugin.name')" prop="name">
              <ElInput v-model="form.name" :disabled="isSystem" />
            </ElFormItem>
          </ElCol>

          <!-- Module -->
          <ElCol :span="12" v-if="props.tp == 'scan'">
            <ElFormItem :label="t('plugin.module')" prop="module">
              <ElSelect v-model="form.module" :disabled="isSystem">
                <ElOption
                  v-for="option in moduleOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </ElSelect>
            </ElFormItem>
          </ElCol>

          <!-- Version -->
          <ElCol :span="12">
            <ElFormItem :label="t('plugin.version')" prop="version">
              <ElInput v-model="form.version" :disabled="isSystem" />
            </ElFormItem>
          </ElCol>

          <!-- Help -->
          <ElCol :span="12" v-if="props.tp == 'scan'">
            <ElFormItem :label="t('plugin.help')" prop="help">
              <ElInput v-model="form.help" />
            </ElFormItem>
          </ElCol>

          <!-- Introduction -->
          <ElCol :span="24">
            <ElFormItem :label="t('plugin.introduction')" prop="introduction">
              <ElInput v-model="form.introduction" />
            </ElFormItem>
          </ElCol>

          <!-- Parameter configuration -->
          <ElCol :span="24" v-if="props.tp == 'scan'">
            <ElFormItem :label="t('plugin.parameterConfig')">
              <ElRow :gutter="20">
                <template v-for="(param, index) in form.parameterList" :key="index">
                  <ElCol :span="24" style="margin-bottom: 16px">
                    <div style="padding: 12px; border: 1px solid #dcdfe6; border-radius: 4px">
                      <ElSpace :size="10" style="width: 100%">
                        <ElFormItem :prop="`parameterList.${index}.name`" style="margin-bottom: 0">
                          <ElInput
                            v-model="param.name"
                            :placeholder="t('plugin.parameterName')"
                            style="width: 120px"
                            @input="handleParameterListChange"
                          />
                        </ElFormItem>
                        <ElFormItem :prop="`parameterList.${index}.type`" style="margin-bottom: 0">
                          <ElSelect
                            v-model="param.type"
                            :placeholder="t('plugin.parameterType')"
                            style="width: 90px"
                            @change="handleParameterTypeChange(index)"
                          >
                            <ElOption label="string" value="string" />
                            <ElOption label="bool" value="bool" />
                            <ElOption label="dict" value="dict" />
                            <ElOption label="port" value="port" />
                          </ElSelect>
                        </ElFormItem>

                        <!-- Default value for string and bool types -->
                        <template v-if="param.type === 'string' || param.type === 'bool'">
                          <ElFormItem
                            :prop="`parameterList.${index}.defaultValue`"
                            style="margin-bottom: 0"
                          >
                            <ElSelect
                              v-if="param.type === 'bool'"
                              v-model="param.defaultValue"
                              :placeholder="t('plugin.defaultValue')"
                              style="width: 90px"
                              @change="handleParameterListChange"
                            >
                              <ElOption label="true" value="true" />
                              <ElOption label="false" value="false" />
                            </ElSelect>
                            <ElInput
                              v-else
                              v-model="param.defaultValue"
                              :placeholder="t('plugin.defaultValue')"
                              style="width: 120px"
                              @input="handleParameterListChange"
                            />
                          </ElFormItem>
                        </template>

                        <!-- Category and name for dict type -->
                        <template v-if="param.type === 'dict'">
                          <ElFormItem
                            :prop="`parameterList.${index}.dictCategory`"
                            style="margin-bottom: 0"
                          >
                            <ElSelect
                              v-model="param.dictCategory"
                              :placeholder="t('plugin.category')"
                              style="width: 120px"
                              filterable
                              @change="handleDictCategoryChange(index)"
                            >
                              <ElOption
                                v-for="dict in dictList"
                                :key="dict.id"
                                :label="dict.category || dict.name"
                                :value="dict.category || dict.name"
                              />
                            </ElSelect>
                          </ElFormItem>
                          <ElFormItem
                            :prop="`parameterList.${index}.dictName`"
                            style="margin-bottom: 0"
                          >
                            <ElSelect
                              v-model="param.dictName"
                              :placeholder="t('plugin.dictName')"
                              style="width: 120px"
                              filterable
                              :disabled="!param.dictCategory"
                              @change="handleParameterListChange"
                            >
                              <ElOption
                                v-for="dict in getFilteredDictNames(param.dictCategory)"
                                :key="dict.id"
                                :label="dict.name"
                                :value="dict.name"
                              />
                            </ElSelect>
                          </ElFormItem>
                        </template>

                        <!-- name for port type -->
                        <template v-if="param.type === 'port'">
                          <ElFormItem
                            :prop="`parameterList.${index}.portName`"
                            style="margin-bottom: 0"
                          >
                            <ElSelect
                              v-model="param.portName"
                              :placeholder="t('plugin.portName')"
                              style="width: 150px"
                              filterable
                              @change="handleParameterListChange"
                            >
                              <ElOption
                                v-for="port in portList"
                                :key="port.id"
                                :label="port.name"
                                :value="port.name"
                              />
                            </ElSelect>
                          </ElFormItem>
                        </template>

                        <ElButton
                          :icon="Delete"
                          type="danger"
                          circle
                          size="small"
                          @click="removeParameter(index)"
                        />
                      </ElSpace>
                    </div>
                  </ElCol>
                </template>
                <ElCol :span="24" style="margin-top: 10px">
                  <ElButton :icon="Plus" style="width: 100%" @click="addParameter">
                    {{ t('plugin.addParameter') }}
                  </ElButton>
                </ElCol>
              </ElRow>
            </ElFormItem>
          </ElCol>

          <!-- Parameter (read-only, auto-generated from parameter configuration) -->
          <ElCol :span="24" v-if="props.tp == 'scan'">
            <ElFormItem :label="t('plugin.parameter')" prop="parameter">
              <ElInput
                v-model="form.parameter"
                type="textarea"
                :rows="3"
                readonly
                style="background-color: #f5f5f5"
              />
              <div style="font-size: 12px; color: #909399; margin-top: 4px">
                {{ t('plugin.parameterTip') }}
              </div>
            </ElFormItem>
          </ElCol>
        </ElRow>
      </ElTabPane>

      <!-- Source code tab -->
      <ElTabPane :label="t('plugin.sourceCode')" name="source" v-if="!isSystem">
        <div class="code-editor-container">
          <codemirror
            v-model="content"
            class="code-editor"
            :autofocus="true"
            :indent-with-tab="true"
            :tab-size="2"
            :extensions="extensions"
            :disabled="isSystem"
          />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- Action bar fixed at the bottom -->
    <div class="action-bar">
      <ElButton @click="props.closeDialog">{{ t('common.cancel') }}</ElButton>
      <ElButton type="primary" @click="save" :loading="saveLoading">
        {{ t('plugin.save') }}
      </ElButton>
    </div>
  </ElForm>
</template>

<style scoped>
.header-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

/* Source code editor container */
.code-editor-container {
  height: calc(100vh - 300px);
  min-height: 500px;
  background: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.code-editor {
  height: 100%;
  width: 100%;
}

/* Fixed bottom action bar */
.action-bar {
  position: sticky;
  bottom: 0;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 0;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-light);
  margin-top: 20px;
  z-index: 10;
}

/* Ensure tab content has enough padding */
:deep(.el-tabs__content) {
  padding: 20px 0;
}

/* Improve code editor styles */
:deep(.cm-editor) {
  height: 100%;
  font-size: 14px;
  line-height: 1.6;
}

:deep(.cm-scroller) {
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Courier New', monospace;
}

:deep(.cm-focused) {
  outline: none;
}

/* Responsive adjustments */
@media (max-height: 800px) {
  .code-editor-container {
    height: calc(100vh - 250px);
    min-height: 400px;
  }
}
</style>

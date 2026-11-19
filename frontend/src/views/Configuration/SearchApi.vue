<script setup lang="ts">
import { ElRow, ElCol, ElButton, ElCard, ElInput, ElForm, ElFormItem, ElAlert } from 'element-plus'
import { useI18n } from '@/hooks/web/useI18n'
import { ref, onBeforeMount } from 'vue'
import { getSearchApiConfigurationApi, saveSearchApiConfigurationApi } from '@/api/Configuration'
const { t } = useI18n()
const fofaKey = ref('')
const shodanKey = ref('')
const saveLoading = ref(false)
onBeforeMount(async () => {
  try {
    const res = await getSearchApiConfigurationApi()
    if (res.code === 200) {
      fofaKey.value = res.data?.fofa?.key ?? ''
      shodanKey.value = res.data?.shodan?.key ?? ''
    } else {
      console.error(`API request failed with status code ${res.code}`)
    }
  } catch (error) {
    console.error('An error occurred while fetching the asset search config:', error)
  }
})
const confirmAdd = async () => {
  const confirmed = window.confirm('Do you want to save the data?')
  if (confirmed) {
    await save()
  }
}
const save = async () => {
  saveLoading.value = true
  const res = await saveSearchApiConfigurationApi(fofaKey.value, shodanKey.value)
  saveLoading.value = false
}
</script>

<template>
  <ElCard shadow="never" class="mb-20px">
    <template #header>
      <ElRow>
        <ElCol :span="6" style="height: 100%">
          <span>{{ t('configuration.searchapi') }}</span>
        </ElCol>
        <ElCol :span="2" :offset="16">
          <ElButton type="primary" @click="confirmAdd" :loading="saveLoading">{{
            t('common.save')
          }}</ElButton>
        </ElCol>
      </ElRow>
    </template>
    <ElAlert :title="t('configuration.searchapiDesc')" type="info" :closable="false" class="mb-20px" />
    <ElForm label-width="160px">
      <ElFormItem :label="t('configuration.fofaKey')">
        <ElInput v-model="fofaKey" placeholder="fofa.info API key" show-password clearable />
      </ElFormItem>
      <ElFormItem :label="t('configuration.shodanKey')">
        <ElInput v-model="shodanKey" placeholder="api.shodan.io API key" show-password clearable />
      </ElFormItem>
    </ElForm>
  </ElCard>
</template>

<script setup lang="ts">
import { ElDrawer, ElDivider, ElMessage } from 'element-plus'
import { ref, unref, computed, watch } from 'vue'
import { useI18n } from '@/hooks/web/useI18n'
import { ThemeSwitch } from '@/components/ThemeSwitch'
import { colorIsDark, lighten, hexToRGB } from '@/utils/color'
import { useCssVar } from '@vueuse/core'
import { useAppStore } from '@/store/modules/app'
import { trim, setCssVar } from '@/utils'
import ColorRadioPicker from './components/ColorRadioPicker.vue'
import InterfaceDisplay from './components/InterfaceDisplay.vue'
import LayoutRadioPicker from './components/LayoutRadioPicker.vue'
import { useStorage } from '@/hooks/web/useStorage'
import { useClipboard } from '@vueuse/core'
import { useDesign } from '@/hooks/web/useDesign'

const { clear: storageClear } = useStorage('localStorage')

const { getPrefixCls } = useDesign()

const prefixCls = getPrefixCls('setting')

const appStore = useAppStore()

const { t } = useI18n()

const layout = computed(() => appStore.getLayout)

const drawer = ref(false)

// Theme color related
const systemTheme = ref(appStore.getTheme.elColorPrimary)

const setSystemTheme = (color: string) => {
  setCssVar('--el-color-primary', color)
  appStore.setTheme({ elColorPrimary: color })
  const leftMenuBgColor = useCssVar('--left-menu-bg-color', document.documentElement)
  setMenuTheme(trim(unref(leftMenuBgColor)))
}

// Header theme related
const headerTheme = ref(appStore.getTheme.topHeaderBgColor || '')

const setHeaderTheme = (color: string) => {
  const isDarkColor = colorIsDark(color)
  const textColor = isDarkColor ? '#fff' : 'inherit'
  const textHoverColor = isDarkColor ? lighten(color!, 6) : '#f6f6f6'
  const topToolBorderColor = isDarkColor ? color : '#eee'
  setCssVar('--top-header-bg-color', color)
  setCssVar('--top-header-text-color', textColor)
  setCssVar('--top-header-hover-color', textHoverColor)
  appStore.setTheme({
    topHeaderBgColor: color,
    topHeaderTextColor: textColor,
    topHeaderHoverColor: textHoverColor,
    topToolBorderColor
  })
  if (unref(layout) === 'top') {
    setMenuTheme(color)
  }
}

// Menu theme related
const menuTheme = ref(appStore.getTheme.leftMenuBgColor || '')

const setMenuTheme = (color: string) => {
  const primaryColor = useCssVar('--el-color-primary', document.documentElement)
  const isDarkColor = colorIsDark(color)
  const theme: Recordable = {
    // Left menu border color
    leftMenuBorderColor: isDarkColor ? 'inherit' : '#eee',
    // Left menu background color
    leftMenuBgColor: color,
    // Left menu light background color
    leftMenuBgLightColor: isDarkColor ? lighten(color!, 6) : color,
    // Left menu active background color
    leftMenuBgActiveColor: isDarkColor
      ? 'var(--el-color-primary)'
      : hexToRGB(unref(primaryColor), 0.1),
    // Left menu collapsed active background color
    leftMenuCollapseBgActiveColor: isDarkColor
      ? 'var(--el-color-primary)'
      : hexToRGB(unref(primaryColor), 0.1),
    // Left menu text color
    leftMenuTextColor: isDarkColor ? '#bfcbd9' : '#333',
    // Left menu active text color
    leftMenuTextActiveColor: isDarkColor ? '#fff' : 'var(--el-color-primary)',
    // Logo text color
    logoTitleTextColor: isDarkColor ? '#fff' : 'inherit',
    // Logo border color
    logoBorderColor: isDarkColor ? color : '#eee'
  }
  appStore.setTheme(theme)
  appStore.setCssVarTheme()
}
setMenuTheme(unref(menuTheme))
// Watch layout changes and reset some theme colors
watch(
  () => layout.value,
  (n) => {
    if (n === 'top' && !appStore.getIsDark) {
      headerTheme.value = '#fff'
      setHeaderTheme('#fff')
    } else {
      setMenuTheme(unref(menuTheme))
    }
  }
)

// Copy
const copyConfig = async () => {
  const { copy, copied, isSupported } = useClipboard({
    source: `
      // Breadcrumb
      breadcrumb: ${appStore.getBreadcrumb},
      // Breadcrumb icon
      breadcrumbIcon: ${appStore.getBreadcrumbIcon},
      // Collapse icon
      hamburger: ${appStore.getHamburger},
      // Fullscreen icon
      screenfull: ${appStore.getScreenfull},
      // Size icon
      size: ${appStore.getSize},
      // Language icon
      locale: ${appStore.getLocale},
      // Tags view
      tagsView: ${appStore.getTagsView},
      // Tags view icon
      getTagsViewIcon: ${appStore.getTagsViewIcon},
      // logo
      logo: ${appStore.getLogo},
      // Menu accordion
      uniqueOpened: ${appStore.getUniqueOpened},
      // Fixed header
      fixedHeader: ${appStore.getFixedHeader},
      // Footer
      footer: ${appStore.getFooter},
      // Grey mode
      greyMode: ${appStore.getGreyMode},
      // Layout
      layout: '${appStore.getLayout}',
      // Dark mode
      isDark: ${appStore.getIsDark},
      // Component size
      currentSize: '${appStore.getCurrentSize}',
      // Theme related
      theme: {
        // Theme color
        elColorPrimary: '${appStore.getTheme.elColorPrimary}',
        // Left menu border color
        leftMenuBorderColor: '${appStore.getTheme.leftMenuBorderColor}',
        // Left menu background color
        leftMenuBgColor: '${appStore.getTheme.leftMenuBgColor}',
        // Left menu light background color
        leftMenuBgLightColor: '${appStore.getTheme.leftMenuBgLightColor}',
        // Left menu active background color
        leftMenuBgActiveColor: '${appStore.getTheme.leftMenuBgActiveColor}',
        // Left menu collapsed active background color
        leftMenuCollapseBgActiveColor: '${appStore.getTheme.leftMenuCollapseBgActiveColor}',
        // Left menu text color
        leftMenuTextColor: '${appStore.getTheme.leftMenuTextColor}',
        // Left menu active text color
        leftMenuTextActiveColor: '${appStore.getTheme.leftMenuTextActiveColor}',
        // Logo text color
        logoTitleTextColor: '${appStore.getTheme.logoTitleTextColor}',
        // Logo border color
        logoBorderColor: '${appStore.getTheme.logoBorderColor}',
        // Header background color
        topHeaderBgColor: '${appStore.getTheme.topHeaderBgColor}',
        // Header text color
        topHeaderTextColor: '${appStore.getTheme.topHeaderTextColor}',
        // Header hover color
        topHeaderHoverColor: '${appStore.getTheme.topHeaderHoverColor}',
        // Header border color
        topToolBorderColor: '${appStore.getTheme.topToolBorderColor}'
      }
    `,
    legacy: true
  })
  if (!isSupported) {
    ElMessage.error(t('setting.copyFailed'))
  } else {
    await copy()
    if (unref(copied)) {
      ElMessage.success(t('setting.copySuccess'))
    }
  }
}

// Clear cache
const clear = () => {
  storageClear()
  window.location.reload()
}
</script>

<template>
  <div
    :class="prefixCls"
    class="fixed top-[90%] right-0 w-40px h-40px flex items-center justify-center bg-[var(--el-color-primary)] cursor-pointer z-10"
    @click="drawer = true"
  >
    <Icon icon="ant-design:setting-outlined" color="#fff" />
  </div>

  <ElDrawer v-model="drawer" direction="rtl" size="350px" :z-index="4000">
    <template #header>
      <span class="text-16px font-700">{{ t('setting.projectSetting') }}</span>
    </template>

    <div class="text-center">
      <!-- Theme -->
      <ElDivider>{{ t('setting.theme') }}</ElDivider>
      <ThemeSwitch />

      <!-- Layout -->
      <ElDivider>{{ t('setting.layout') }}</ElDivider>
      <LayoutRadioPicker />

      <!-- System theme -->
      <ElDivider>{{ t('setting.systemTheme') }}</ElDivider>
      <ColorRadioPicker
        v-model="systemTheme"
        :schema="[
          '#409eff',
          '#009688',
          '#536dfe',
          '#ff5c93',
          '#ee4f12',
          '#0096c7',
          '#9c27b0',
          '#ff9800'
        ]"
        @change="setSystemTheme"
      />

      <!-- Header theme -->
      <ElDivider>{{ t('setting.headerTheme') }}</ElDivider>
      <ColorRadioPicker
        v-model="headerTheme"
        :schema="[
          '#fff',
          '#151515',
          '#5172dc',
          '#e74c3c',
          '#24292e',
          '#394664',
          '#009688',
          '#383f45'
        ]"
        @change="setHeaderTheme"
      />

      <!-- Menu theme -->
      <template v-if="layout !== 'top'">
        <ElDivider>{{ t('setting.menuTheme') }}</ElDivider>
        <ColorRadioPicker
          v-model="menuTheme"
          :schema="[
            '#fff',
            '#001529',
            '#212121',
            '#273352',
            '#191b24',
            '#383f45',
            '#001628',
            '#344058'
          ]"
          @change="setMenuTheme"
        />
      </template>
    </div>

    <!-- Interface display -->
    <ElDivider>{{ t('setting.interfaceDisplay') }}</ElDivider>
    <InterfaceDisplay />

    <ElDivider />
    <div>
      <BaseButton type="primary" class="w-full" @click="copyConfig">{{
        t('setting.copy')
      }}</BaseButton>
    </div>
    <div class="mt-5px">
      <BaseButton type="danger" class="w-full" @click="clear">
        {{ t('setting.clearAndReset') }}
      </BaseButton>
    </div>
  </ElDrawer>
</template>

<style lang="less" scoped>
@prefix-cls: ~'@{namespace}-setting';

.@{prefix-cls} {
  border-radius: 6px 0 0 6px;
}
</style>

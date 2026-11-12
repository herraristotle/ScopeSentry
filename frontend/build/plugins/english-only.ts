import type { Plugin } from 'vite'

/**
 * Strip Chinese from third-party bundles at build time.
 *
 * The application source is English-only, but several dependencies embed
 * non-English locale payloads inside their published bundles (they are not
 * separate modules we can simply avoid importing):
 *
 *   - ECharts imports and registers a Chinese locale unconditionally.
 *   - xgplayer's preset modules import the Chinese dictionary to prime i18n,
 *     and its barrel file re-exports every supported language.
 *   - wangEditor inlines seven `zh-CN` resource bundles, sets `zh-CN` as the
 *     i18next fallback, and carries a few bilingual console messages plus a
 *     list of Chinese font-family names in its prebuilt dist file.
 *   - Element Plus puts the Chinese name for Microsoft YaHei in its default
 *     font stack.
 *
 * None of that text is ever displayed - every component is pinned to English -
 * but it still ends up in the shipped JS/CSS. This plugin removes it so the
 * build output matches the source: English only.
 *
 * Every Chinese string this plugin needs to match is written with `\uXXXX`
 * escapes so this file itself stays pure ASCII.
 */

/** Any CJK Unified Ideograph. */
const CJK = /[\u4e00-\u9fff]/

/**
 * Exact source literals in wangEditor's dist bundle that contain Chinese, and
 * what to replace them with. Matching the surrounding quotes keeps this safe: a
 * literal is either found verbatim in the vendor file or left untouched.
 *
 * Chinese font names are replaced with the Latin names browsers resolve to the
 * same fonts, because the font-family dropdown needs real identifiers.
 */
const WANGEDITOR_LITERALS: Array<[from: string, to: string]> = [
  // Bilingual console messages. The English half is already present in each of
  // these, so keep it and drop the Chinese half.
  [
    // Chinese half of the bilingual "unsupported IE browser" message.
    '"\u62b1\u6b49\uff0cwangEditor V5+ \u7248\u672c\u5f00\u59cb\uff0c\u4e0d\u5728\u652f\u6301 IE \u6d4f\u89c8\u5668\\n Sorry, wangEditor V5+ versions do not support IE browser."',
    '"wangEditor V5+ versions do not support IE browser."'
  ],
  [
    // Chinese half of the bilingual full-screen error.
    '"Can not set full screen, cause toolbar DOM parent is not equal to textarea DOM parent\\n\u4e0d\u80fd\u8bbe\u7f6e\u5168\u5c4f\uff0c\u56e0\u4e3a toolbar DOM \u7236\u8282\u70b9\u548c textarea DOM \u7236\u8282\u70b9\u4e0d\u4e00\u81f4"',
    '"Can not set full screen, cause toolbar DOM parent is not equal to textarea DOM parent"'
  ],
  [
    // Chinese half of the bilingual textarea-height warning.
    '"\u7f16\u8f91\u533a\u57df\u9ad8\u5ea6 < 300px \u8fd9\u53ef\u80fd\u4f1a\u5bfc\u81f4 modal hoverbar \u5b9a\u4f4d\u5f02\u5e38\\nTextarea height < 300px . This may be cause modal and hoverbar position error"',
    '"Textarea height < 300px . This may be cause modal and hoverbar position error"'
  ],
  [
    // Chinese half of the missing-upload-address error.
    '"Cannot get upload server address\\n\u6ca1\u6709\u914d\u7f6e\u4e0a\u4f20\u5730\u5740"',
    '"Cannot get upload server address"'
  ],
  [
    // Chinese half of the missing-fieldName error.
    '"Cannot get fieldName\\n\u6ca1\u6709\u914d\u7f6e fieldName"',
    '"Cannot get fieldName"'
  ],
  [
    // Chinese half of the "cannot instantiate" error.
    '"\u4e0d\u80fd\u5b9e\u4f8b\u5316\\nCan not construct a instance"',
    '"Can not construct a instance"'
  ],
  // Warning for a disabled-scroll editor. The code appends an English sentence
  // to this literal immediately afterwards, so drop the Chinese half.
  [
    '"\u7f16\u8f91\u5668\u7981\u7528\u4e86 scroll \uff0c\u7f16\u8f91\u5668\u5185\u5bb9\u65e0\u6cd5\u6eda\u52a8\uff0c\u8bf7\u81ea\u884c\u5b9e\u73b0\u8be5\u529f\u80fd"',
    '""'
  ],
  // Duplicated-menu warning is built by concatenation; drop this trailing half.
  ['"\'\\n\u91cd\u590d\u6ce8\u518c\u4e86\u83dc\u5355\u680f menu \'"', '""'],
  // "element not found" error is built by concatenation around the element id.
  ['"\\n \u901a\u8fc7 id \'"', '"\\n element with id \'"'],
  [
    '"\' \u627e\u5230\u7684 element \u4e0d\u5728 editor DOM \u4e4b\u5185"',
    '"\' was not found inside the editor DOM"'
  ],
  // Bilingual <video> fallback markup for unsupported browsers.
  [
    // Chinese half of the bilingual <video> fallback markup.
    '"Sorry, your browser doesn\'t support embedded videos.\\n \u62b1\u6b49\uff0c\u6d4f\u89c8\u5668\u4e0d\u652f\u6301 video \u89c6\u9891"',
    '"Sorry, your browser doesn\'t support embedded videos."'
  ],
  // Font family identifiers.
  ['"\u9ed1\u4f53"', '"SimHei"'],
  [
    '{name:"\u4eff\u5b8b",value:"\u4eff\u5b8b"}',
    '{name:"FangSong",value:"FangSong"}'
  ],
  ['"\u6977\u4f53"', '"KaiTi"'],
  ['"\u6807\u6977\u4f53"', '"DFKai-SB"'],
  ['"\u534e\u6587\u4eff\u5b8b"', '"STFangsong"'],
  ['"\u534e\u6587\u6977\u4f53"', '"STKaiti"'],
  ['{name:"\u5b8b\u4f53",value:"\u5b8b\u4f53"}', '{name:"SimSun",value:"SimSun"}'],
  ['"\u5fae\u8f6f\u96c5\u9ed1"', '"Microsoft YaHei"']
]

/** The Chinese name for Microsoft YaHei, as it appears in vendor font stacks. */
const MS_YAHEI_ZH = '"\u5fae\u8f6f\u96c5\u9ed1"'

/**
 * Find the closing brace of the object literal starting at `start`, skipping
 * over string contents so braces inside strings are not counted.
 */
function matchBrace(code: string, start: number): number {
  let depth = 0
  let quote: string | null = null
  let escaped = false

  for (let i = start; i < code.length; i++) {
    const ch = code[i]
    if (quote) {
      if (escaped) {
        escaped = false
        continue
      }
      if (ch === '\\') {
        escaped = true
        continue
      }
      if (ch === quote) quote = null
      continue
    }
    if (ch === '"' || ch === "'" || ch === '`') {
      quote = ch
      continue
    }
    if (ch === '{') depth++
    else if (ch === '}') {
      depth--
      if (depth === 0) return i
    }
  }
  return -1
}

/** Remove every `MP("zh-CN", {...})` registration call from wangEditor. */
function removeWangEditorZhBundles(code: string): { code: string; removed: number } {
  let out = code
  let removed = 0

  for (;;) {
    const match = /MP\("zh-CN",/.exec(out)
    if (!match) break

    const callStart = match.index
    const braceStart = out.indexOf('{', callStart)
    if (braceStart < 0) throw new Error('[english-only] malformed wangEditor locale call')

    const braceEnd = matchBrace(out, braceStart)
    if (braceEnd < 0) throw new Error('[english-only] unbalanced wangEditor locale call')

    // The calls sit in a comma sequence. Drop the preceding comma when there is
    // one, otherwise the trailing comma, so the expression stays valid.
    let end = braceEnd + 1
    if (out[end] === ')') end++

    let start = callStart
    if (out[start - 1] === ',') {
      start--
    } else if (out[end] === ',') {
      end++
    }

    out = out.slice(0, start) + out.slice(end)
    removed++
  }

  return { code: out, removed }
}

export function englishOnly(): Plugin {
  return {
    name: 'scope-sentry:english-only',
    enforce: 'pre',

    transform(code, id) {
      // ECharts registers the Chinese locale unconditionally. Every chart is
      // initialised with `locale: 'EN'`, so the Chinese slot is never read.
      if (/[\\/]echarts[\\/](?:lib|esm)[\\/]core[\\/]locale\.js$/.test(id)) {
        const transformed = code.replace(
          /import\s+langZH\s+from\s+['"][^'"]*langZH(?:\.js)?['"];?/,
          'const langZH = {};'
        )
        if (transformed === code) {
          this.warn('ECharts locale module did not match the expected import shape')
          return null
        }
        return { code: transformed, map: null }
      }

      if (/[\\/]xgplayer[\\/]es[\\/].*\.js$/.test(id)) {
        // Preset modules import the Chinese dictionary only to prime i18n; the
        // English dictionary is equivalent for an English-only app. The barrel
        // file re-exports every language, so drop the Chinese re-exports.
        let transformed = code.replace(
          /import\s+ZH\s+from\s+['"][^'"]*zh-cn\.js['"];?/g,
          'import ZH from "../lang/en.js";'
        )
        transformed = transformed.replace(
          /export\s*\{\s*default\s+as\s+langZh(?:Cn|Hk)\s*\}\s*from\s*['"][^'"]*zh-(?:cn|hk)\.js['"];?/g,
          ''
        )
        return transformed === code ? null : { code: transformed, map: null }
      }

      if (/[\\/]@wangeditor[\\/]editor[\\/]dist[\\/]index\.esm\.js$/.test(id)) {
        const stripped = removeWangEditorZhBundles(code)
        if (stripped.removed === 0) {
          this.warn('wangEditor bundle no longer contains zh-CN registrations')
          return null
        }

        let transformed = stripped.code.replace(/lng:"zh-CN"/g, 'lng:"en"')
        let replaced = 0
        for (const [from, to] of WANGEDITOR_LITERALS) {
          if (transformed.includes(from)) {
            transformed = transformed.split(from).join(to)
            replaced++
          }
        }

        const leftovers = transformed.match(new RegExp(CJK, 'g'))
        if (leftovers) {
          this.warn(
            `wangEditor bundle still contains ${leftovers.length} CJK characters; ` +
              `${replaced}/${WANGEDITOR_LITERALS.length} known literals were replaced`
          )
        }
        return { code: transformed, map: null }
      }

      // The Vue wrapper for wangEditor builds a bilingual developer warning from
      // `\uXXXX` escapes, which grep cannot see. Drop the Chinese half; the
      // English sentence is appended to it on the next line.
      if (/[\\/]@wangeditor[\\/]editor-for-vue[\\/]dist[\\/]index\.esm\.js$/.test(id)) {
        const transformed = code.replace(
          /let info = `[^`]*`;\s*\n\s*info \+= `\nPlease use/,
          'let info = `Please use'
        )
        if (transformed === code) {
          this.warn('wangEditor Vue wrapper warning did not match the expected shape')
          return null
        }
        return { code: transformed, map: null }
      }

      // ECharts and zrender measure the advance width of a full-width glyph to
      // estimate line height and the rendered width of non-ASCII text. The probe
      // character is never displayed, so use a full-width Latin letter instead of
      // an ideograph: it has the same East Asian full-width metrics but is not
      // Chinese.
      if (
        /[\\/]zrender[\\/]lib[\\/]contain[\\/]text\.js$/.test(id) ||
        /[\\/]zrender[\\/]lib[\\/]graphic[\\/]helper[\\/]parseText\.js$/.test(id) ||
        /[\\/]echarts[\\/]lib[\\/]component[\\/]visualMap[\\/]ContinuousView\.js$/.test(id)
      ) {
        const transformed = code.replace(/'[\u4e00-\u9fff]'/g, "'\uFF21'")
        return transformed === code ? null : { code: transformed, map: null }
      }

      // echarts-wordcloud ships a font stack containing the Traditional Chinese
      // name for Microsoft JhengHei. Use the Latin name instead.
      if (/[\\/]echarts-wordcloud[\\/].*\.js$/.test(id)) {
        const transformed = code.replace(/\u5fae\u8edf\u6b63\u9ed1\u9ad4/g, 'Microsoft JhengHei')
        return transformed === code ? null : { code: transformed, map: null }
      }

      return null
    },

    generateBundle(_options, bundle) {
      // Element Plus ships `"Microsoft YaHei", "<Chinese name>"`; the Chinese
      // entry is a duplicate of the Latin one, so drop it from the stylesheet.
      const pattern = new RegExp(MS_YAHEI_ZH + ',\\s*', 'g')
      for (const file of Object.values(bundle)) {
        if (file.type !== 'asset' || !/\.css$/.test(file.fileName)) continue
        const source = typeof file.source === 'string' ? file.source : ''
        if (!CJK.test(source)) continue
        file.source = source.replace(pattern, '')
      }
    }
  }
}

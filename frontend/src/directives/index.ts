import type { App } from 'vue'
import { setupPermissionDirective } from './permission/hasPermi'

/**
 * Export directives: v-xxx
 * @methods hasPermi button permission, usage: v-hasPermi
 */
export const setupPermission = (app: App<Element>) => {
  setupPermissionDirective(app)
}

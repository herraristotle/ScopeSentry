import type { RouteRecordRaw } from 'vue-router'
import { defineComponent } from 'vue'

/**
* redirect: noredirect        When noredirect is set, the route cannot be clicked in breadcrumb navigation
* name:'router-name'          Set the route name; it must be filled in, otherwise various problems occur when using <keep-alive>
* meta : {
    hidden: true              When set to true, the route will not appear in the sidebar, such as 404, login pages, etc. (default false)

    alwaysShow: true          When a route has more than 1 child route declared, it automatically becomes nested mode;
                              when there is only one, that child route is displayed in the sidebar as the root route;
                              if you want your root route to be displayed regardless of the number of child routes declared,
                              you can set alwaysShow: true, which ignores the previous rules
                              and always displays the root route (default false)

    title: 'title'            Set the name displayed for this route in the sidebar and breadcrumb

    icon: 'svg-name'          Set the icon for this route

    noCache: true             If set to true, it will not be cached by <keep-alive> (default false)

    breadcrumb: false         If set to false, it will not be displayed in the breadcrumb (default true)

    affix: true               If set to true, it will always be fixed in the tag items (default false)

    noTagsView: true          If set to true, it will not appear in tags (default false)

    activeMenu: '/dashboard'  The route path to highlight

    canTo: true               Set to true so that route navigation is still possible even if hidden is true (default false)

    permission: ['edit','add', 'delete']    Set the permissions for this route
  }
**/

interface RouteMetaCustom extends Record<string | number | symbol, unknown> {
  hidden?: boolean
  alwaysShow?: boolean
  title?: string
  icon?: string
  noCache?: boolean
  breadcrumb?: boolean
  affix?: boolean
  activeMenu?: string
  noTagsView?: boolean
  canTo?: boolean
  permission?: string[]
}

declare module 'vue-router' {
  interface RouteMeta extends RouteMetaCustom {}
}

type Component<T = any> =
  | ReturnType<typeof defineComponent>
  | (() => Promise<typeof import('*.vue')>)
  | (() => Promise<T>)

declare global {
  declare interface AppRouteRecordRaw extends Omit<RouteRecordRaw, 'meta' | 'children'> {
    name: string
    meta: RouteMetaCustom
    component?: Component | string
    children?: AppRouteRecordRaw[]
    props?: Recordable
    fullPath?: string
  }

  declare interface AppCustomRouteRecordRaw
    extends Omit<RouteRecordRaw, 'meta' | 'component' | 'children'> {
    name: string
    meta: RouteMetaCustom
    component: string
    path: string
    redirect: string
    children?: AppCustomRouteRecordRaw[]
  }
}

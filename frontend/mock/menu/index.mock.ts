import { faker } from '@faker-js/faker'
import { SUCCESS_CODE } from '@/constants'

const timeout = 1000

export default [
  // List endpoint
  {
    url: '/mock/menu/list',
    method: 'get',
    timeout,
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: {
          list: [
            {
              path: '/dashboard',
              component: '#',
              redirect: '/dashboard/analysis',
              name: 'Dashboard',
              status: faker.number.int({ min: 0, max: 1 }),
              id: 1,
              type: 0,
              parentId: undefined,
              title: 'Home',
              meta: {
                title: 'Home',
                icon: 'ant-design:dashboard-filled',
                alwaysShow: true
              },
              children: [
                {
                  path: 'analysis',
                  component: 'views/Dashboard/Analysis',
                  name: 'Analysis',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 2,
                  type: 1,
                  parentId: 1,
                  title: 'Analysis Page',
                  permissionList: [
                    {
                      label: 'Add',
                      value: 'add'
                    },
                    {
                      label: 'Edit',
                      value: 'edit'
                    }
                  ],
                  meta: {
                    title: 'Analysis Page',
                    noCache: true,
                    permission: ['add', 'edit']
                  }
                },
                {
                  path: 'workplace',
                  component: 'views/Dashboard/Workplace',
                  name: 'Workplace',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 3,
                  type: 1,
                  parentId: 1,
                  title: 'Workplace',
                  permissionList: [
                    {
                      label: 'Add',
                      value: 'add'
                    },
                    {
                      label: 'Edit',
                      value: 'edit'
                    },
                    {
                      label: 'Delete',
                      value: 'delete'
                    }
                  ],
                  meta: {
                    title: 'Workplace',
                    noCache: true
                  }
                }
              ]
            },
            {
              path: '/external-link',
              component: '#',
              meta: {
                title: 'Docs',
                icon: 'clarity:document-solid'
              },
              name: 'ExternalLink',
              status: faker.number.int({ min: 0, max: 1 }),
              id: 4,
              type: 0,
              parentId: undefined,
              title: 'Docs',
              children: [
                {
                  path: 'https://element-plus-admin-doc.cn/',
                  name: 'DocumentLink',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 5,
                  type: 1,
                  parentId: 4,
                  title: 'Docs',
                  meta: {
                    title: 'Docs'
                  }
                }
              ]
            },
            {
              path: '/level',
              component: '#',
              redirect: '/level/menu1/menu1-1/menu1-1-1',
              name: 'Level',
              status: faker.number.int({ min: 0, max: 1 }),
              id: 6,
              type: 0,
              parentId: undefined,
              title: 'Menu',
              meta: {
                title: 'Menu',
                icon: 'carbon:skill-level-advanced'
              },
              children: [
                {
                  path: 'menu1',
                  name: 'Menu1',
                  component: '##',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 7,
                  type: 0,
                  parentId: 6,
                  title: 'Menu 1',
                  redirect: '/level/menu1/menu1-1/menu1-1-1',
                  meta: {
                    title: 'Menu 1'
                  },
                  children: [
                    {
                      path: 'menu1-1',
                      name: 'Menu11',
                      component: '##',
                      status: faker.number.int({ min: 0, max: 1 }),
                      id: 8,
                      type: 0,
                      parentId: 7,
                      title: 'Menu 1-1',
                      redirect: '/level/menu1/menu1-1/menu1-1-1',
                      meta: {
                        title: 'Menu 1-1',
                        alwaysShow: true
                      },
                      children: [
                        {
                          path: 'menu1-1-1',
                          name: 'Menu111',
                          component: 'views/Level/Menu111',
                          status: faker.number.int({ min: 0, max: 1 }),
                          id: 9,
                          type: 1,
                          parentId: 8,
                          title: 'Menu 1-1-1',
                          meta: {
                            title: 'Menu 1-1-1'
                          }
                        }
                      ]
                    },
                    {
                      path: 'menu1-2',
                      name: 'Menu12',
                      component: 'views/Level/Menu12',
                      status: faker.number.int({ min: 0, max: 1 }),
                      id: 10,
                      type: 1,
                      parentId: 7,
                      title: 'Menu 1-2',
                      meta: {
                        title: 'Menu 1-2'
                      }
                    }
                  ]
                },
                {
                  path: 'menu2',
                  name: 'Menu2Demo',
                  component: 'views/Level/Menu2',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 11,
                  type: 1,
                  parentId: 6,
                  title: 'Menu 2',
                  meta: {
                    title: 'Menu 2'
                  }
                }
              ]
            },
            {
              path: '/example',
              component: '#',
              redirect: '/example/example-dialog',
              name: 'Example',
              status: faker.number.int({ min: 0, max: 1 }),
              id: 12,
              type: 0,
              parentId: undefined,
              title: 'Comprehensive Example',
              meta: {
                title: 'Comprehensive Example',
                icon: 'ep:management',
                alwaysShow: true
              },
              children: [
                {
                  path: 'example-dialog',
                  component: 'views/Example/Dialog/ExampleDialog',
                  name: 'ExampleDialog',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 13,
                  type: 1,
                  parentId: 12,
                  title: 'Comprehensive Example - Dialog',
                  permissionList: [
                    {
                      label: 'Add',
                      value: 'add'
                    },
                    {
                      label: 'Edit',
                      value: 'edit'
                    },
                    {
                      label: 'Delete',
                      value: 'delete'
                    },
                    {
                      label: 'View',
                      value: 'view'
                    }
                  ],
                  meta: {
                    title: 'Comprehensive Example - Dialog'
                  }
                },
                {
                  path: 'example-page',
                  component: 'views/Example/Page/ExamplePage',
                  name: 'ExamplePage',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 14,
                  type: 1,
                  parentId: 12,
                  title: 'Comprehensive Example - Page',
                  permissionList: [
                    {
                      label: 'Add',
                      value: 'edit'
                    },
                    {
                      label: 'Edit',
                      value: 'edit'
                    },
                    {
                      label: 'Delete',
                      value: 'delete'
                    },
                    {
                      label: 'View',
                      value: 'view'
                    }
                  ],
                  meta: {
                    title: 'Comprehensive Example - Page'
                  }
                },
                {
                  path: 'example-add',
                  component: 'views/Example/Page/ExampleAdd',
                  name: 'ExampleAdd',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 15,
                  type: 1,
                  parentId: 12,
                  title: 'Comprehensive Example - Add',
                  meta: {
                    title: 'Comprehensive Example - Add',
                    noTagsView: true,
                    noCache: true,
                    hidden: true,
                    showMainRoute: true,
                    activeMenu: '/example/example-page'
                  }
                },
                {
                  path: 'example-edit',
                  component: 'views/Example/Page/ExampleEdit',
                  name: 'ExampleEdit',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 16,
                  type: 1,
                  parentId: 12,
                  title: 'Comprehensive Example - Edit',
                  meta: {
                    title: 'Comprehensive Example - Edit',
                    noTagsView: true,
                    noCache: true,
                    hidden: true,
                    showMainRoute: true,
                    activeMenu: '/example/example-page'
                  }
                },
                {
                  path: 'example-detail',
                  component: 'views/Example/Page/ExampleDetail',
                  name: 'ExampleDetail',
                  status: faker.number.int({ min: 0, max: 1 }),
                  id: 17,
                  type: 1,
                  parentId: 12,
                  title: 'Comprehensive Example - Detail',
                  meta: {
                    title: 'Comprehensive Example - Detail',
                    noTagsView: true,
                    noCache: true,
                    hidden: true,
                    showMainRoute: true,
                    activeMenu: '/example/example-page'
                  }
                }
              ]
            }
          ]
        }
      }
    }
  }
]

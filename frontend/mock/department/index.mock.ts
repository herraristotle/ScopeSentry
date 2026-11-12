import { toAnyString } from '@/utils'
import { faker } from '@faker-js/faker'
import { SUCCESS_CODE } from '@/constants'

const departmentList: any = []

const citys = ['Xiamen HQ', 'Beijing Branch', 'Shanghai Branch', 'Fuzhou Branch', 'Shenzhen Branch', 'Hangzhou Branch']

for (let i = 0; i < 5; i++) {
  departmentList.push({
    // Department name
    departmentName: citys[i],
    id: toAnyString(),
    createTime: faker.date.anytime(),
    status: faker.number.int({ min: 0, max: 1 }),
    // Remark
    remark: faker.lorem.sentence(),
    children: [
      {
        // Department name
        departmentName: 'R&D Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        // Department name
        departmentName: 'Product Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        // Department name
        departmentName: 'Operations Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        // Department name
        departmentName: 'Marketing Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        // Department name
        departmentName: 'Sales Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      },
      {
        // Department name
        departmentName: 'Customer Service Department',
        createTime: faker.date.anytime(),
        // Status
        status: faker.number.int({ min: 0, max: 1 }),
        id: toAnyString(),
        remark: faker.lorem.sentence()
      }
    ]
  })
}

export default [
  // List endpoint
  {
    url: '/mock/department/list',
    method: 'get',
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: {
          list: departmentList
        }
      }
    }
  },
  {
    url: '/mock/department/table/list',
    method: 'get',
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: {
          list: departmentList,
          total: 5
        }
      }
    }
  },
  {
    url: '/mock/department/users',
    method: 'get',
    timeout: 1000,
    response: ({ query }) => {
      const { pageSize } = query
      // Create data based on pageSize
      const mockList: any = []
      for (let i = 0; i < pageSize; i++) {
        mockList.push({
          // Username
          username: faker.person.firstName(),
          // Account
          account: faker.person.lastName(),
          // Email
          email: faker.internet.email(),
          // Creation time
          createTime: faker.date.anytime(),
          // User ID
          id: toAnyString()
        })
      }
      return {
        code: SUCCESS_CODE,
        data: {
          total: 100,
          list: mockList
        }
      }
    }
  },
  // Save endpoint
  {
    url: '/mock/department/user/save',
    method: 'post',
    timeout: 1000,
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: 'success'
      }
    }
  },
  // Delete endpoint
  {
    url: '/mock/department/user/delete',
    method: 'post',
    response: ({ body }) => {
      const ids = body.ids
      if (!ids) {
        return {
          code: 500,
          message: 'Please select the data to delete'
        }
      } else {
        return {
          code: SUCCESS_CODE,
          data: 'success'
        }
      }
    }
  },
  // Save endpoint
  {
    url: '/mock/department/save',
    method: 'post',
    timeout: 1000,
    response: () => {
      return {
        code: SUCCESS_CODE,
        data: 'success'
      }
    }
  },
  // Delete endpoint
  {
    url: '/mock/department/delete',
    method: 'post',
    response: ({ body }) => {
      const ids = body.ids
      if (!ids) {
        return {
          code: 500,
          message: 'Please select the data to delete'
        }
      } else {
        return {
          code: SUCCESS_CODE,
          data: 'success'
        }
      }
    }
  }
]

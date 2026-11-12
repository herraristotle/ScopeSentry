import request from '@/axios'

// Get all dictionaries
export const getDictApi = () => {
  return request.get({ url: '/mock/dict/list' })
}

// Mock fetching a single dictionary
export const getDictOneApi = async () => {
  return request.get({ url: '/mock/dict/one' })
}

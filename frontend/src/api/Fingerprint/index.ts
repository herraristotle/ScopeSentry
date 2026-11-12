import request from '@/axios'
import axios from 'axios'
import type { fingerprintData, fingerprintRespData } from './types'

interface fingerprintDataResponse {
  list: fingerprintData[]
  total: number
}
export const getFingerprintDataApi = (
  search: string,
  pageIndex: number,
  pageSize: number
): Promise<IResponse<fingerprintDataResponse>> => {
  return request.post({ url: '/api/fingerprint/data', data: { search, pageIndex, pageSize } })
}

export const addFingerprintDataApi = (content: string): Promise<IResponse<fingerprintRespData>> => {
  return request.post({
    url: '/api/fingerprint/add',
    data: { content }
  })
}

export const updateFingerprintDataApi = (
  id: string,
  content: string
): Promise<IResponse<fingerprintRespData>> => {
  return request.post({
    url: '/api/fingerprint/update',
    data: { id, content }
  })
}

export const deleteFingerprintDataApi = (
  ids: string[]
): Promise<IResponse<fingerprintRespData>> => {
  return request.post({ url: '/api/fingerprint/delete', data: { ids } })
}

interface VersionResponse {
  version: string
}

export const getFingerprintVersionApi = (): Promise<IResponse<VersionResponse>> => {
  return request.get({ url: '/api/fingerprint/version' })
}

interface CountByUpdateTimeResponse {
  count: number
}

export const getCountByUpdateTimeApi = (
  updateTime: string
): Promise<IResponse<CountByUpdateTimeResponse>> => {
  // Use native axios to call the remote API directly, avoiding adding an Authorization header
  return axios
    .post(
      'https://api.scope-sentry.top/api/common/finger/count-by-update-time',
      { updateTime },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    )
    .then((res) => res.data)
}

interface FingerprintUpdateItem {
  id: string
  name: string
  content: string
  updateTime: string
}

interface GetFingerprintsByUpdateTimeResponse {
  data: {
    data: FingerprintUpdateItem[]
  }
}

// Get the list of fingerprints that need updating (no authentication required)
export const getFingerprintsByUpdateTimeApi = (
  updateTime: string
): Promise<GetFingerprintsByUpdateTimeResponse> => {
  // Use native axios to call the remote API directly, avoiding adding an Authorization header
  return axios
    .post(
      'https://api.scope-sentry.top/api/common/finger/query-by-update-time',
      { updateTime },
      {
        headers: {
          'Content-Type': 'application/json'
        }
      }
    )
    .then((res) => res.data)
}

// Batch add fingerprints
export const batchAddFingerprintApi = (
  items: FingerprintUpdateItem[]
): Promise<IResponse<fingerprintRespData>> => {
  return request.post({
    url: '/api/fingerprint/batch-add',
    data: { items }
  })
}

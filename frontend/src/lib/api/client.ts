import { getStoredToken } from '../storage'
import {
  ApiError,
  type ApiErrorEnvelope,
  type ApiSuccessEnvelope,
  type ApiSuccessWithMetaEnvelope,
} from './types'

const baseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') ?? ''

type UnauthorizedHandler = () => void

let unauthorizedHandler: UnauthorizedHandler | null = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null): void {
  unauthorizedHandler = handler
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  auth?: boolean
  signal?: AbortSignal
}

async function parseJson(response: Response): Promise<unknown> {
  const text = await response.text()
  if (!text) {
    return null
  }
  try {
    return JSON.parse(text) as unknown
  } catch {
    throw new ApiError(response.status, 'INVALID_RESPONSE', 'Respons server tidak valid')
  }
}

function extractError(payload: unknown, status: number): ApiError {
  if (payload && typeof payload === 'object' && 'error' in payload) {
    const envelope = payload as ApiErrorEnvelope
    const code = envelope.error?.code ?? 'UNKNOWN_ERROR'
    const message = envelope.error?.message ?? 'Terjadi kesalahan'
    return new ApiError(status, code, message)
  }
  return new ApiError(status, 'UNKNOWN_ERROR', 'Terjadi kesalahan')
}

async function rawRequest(path: string, options: RequestOptions = {}): Promise<{ response: Response; payload: unknown }> {
  const headers = new Headers({
    Accept: 'application/json',
  })

  if (options.body !== undefined) {
    headers.set('Content-Type', 'application/json')
  }

  if (options.auth !== false) {
    const token = getStoredToken()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
  }

  let response: Response
  try {
    response = await fetch(`${baseUrl}${path}`, {
      method: options.method ?? 'GET',
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
      signal: options.signal,
    })
  } catch {
    throw new ApiError(0, 'NETWORK_ERROR', 'Tidak dapat terhubung ke server')
  }

  const payload = await parseJson(response)

  if (!response.ok) {
    const error = extractError(payload, response.status)
    if (response.status === 401 && options.auth !== false) {
      unauthorizedHandler?.()
    }
    throw error
  }

  return { response, payload }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { response, payload } = await rawRequest(path, options)

  if (response.status === 204 || payload === null) {
    return undefined as T
  }

  if (payload && typeof payload === 'object' && 'data' in payload) {
    return (payload as ApiSuccessEnvelope<T>).data
  }

  throw new ApiError(response.status, 'INVALID_RESPONSE', 'Respons server tidak valid')
}

export async function apiRequestWithMeta<T, M>(
  path: string,
  options: RequestOptions = {},
): Promise<{ data: T; meta: M }> {
  const { response, payload } = await rawRequest(path, options)

  if (payload && typeof payload === 'object' && 'data' in payload && 'meta' in payload) {
    const envelope = payload as ApiSuccessWithMetaEnvelope<T, M>
    return { data: envelope.data, meta: envelope.meta }
  }

  throw new ApiError(response.status, 'INVALID_RESPONSE', 'Respons server tidak valid')
}

export function toQueryString(params: Record<string, string | number | undefined | null>): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') {
      continue
    }
    search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `?${qs}` : ''
}

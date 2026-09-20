export type ApiErrorBody = {
  code: string
  message: string
}

export type ApiSuccessEnvelope<T> = {
  data: T
}

export type ApiSuccessWithMetaEnvelope<T, M> = {
  data: T
  meta: M
}

export type ApiErrorEnvelope = {
  error: ApiErrorBody
}

export type PaginationMeta = {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

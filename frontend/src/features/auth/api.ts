import { apiRequest } from '../../lib/api/client'
import type { LoginRequest, LoginResponse, User } from './types'

export function loginRequest(payload: LoginRequest, signal?: AbortSignal): Promise<LoginResponse> {
  return apiRequest<LoginResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: payload,
    auth: false,
    signal,
  })
}

export function fetchCurrentUser(signal?: AbortSignal): Promise<User> {
  return apiRequest<User>('/api/v1/auth/me', { signal })
}

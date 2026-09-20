import { createContext, useContext } from 'react'
import type { LoginRequest, User } from './types'

export type AuthContextValue = {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isInitializing: boolean
  login: (payload: LoginRequest) => Promise<void>
  logout: () => void
}

const missingAuth = (): never => {
  throw new Error('useAuth must be used within AuthProvider')
}

const defaultAuthValue: AuthContextValue = {
  token: null,
  user: null,
  isAuthenticated: false,
  isInitializing: false,
  login: async () => missingAuth(),
  logout: missingAuth,
}

export const AuthContext = createContext<AuthContextValue>(defaultAuthValue)

export function useAuth(): AuthContextValue {
  return useContext(AuthContext)
}

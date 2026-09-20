import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { setUnauthorizedHandler } from '../../lib/api/client'
import { clearStoredToken, getStoredToken, setStoredToken } from '../../lib/storage'
import { fetchCurrentUser, loginRequest } from './api'
import { AuthContext, type AuthContextValue } from './auth-context'
import type { LoginRequest, User } from './types'

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const [token, setToken] = useState<string | null>(() => getStoredToken())
  const [user, setUser] = useState<User | null>(null)
  const [isInitializing, setIsInitializing] = useState(() => Boolean(getStoredToken()))

  const clearSession = useCallback(() => {
    clearStoredToken()
    setToken(null)
    setUser(null)
    queryClient.cancelQueries()
    queryClient.clear()
  }, [queryClient])

  const logout = useCallback(() => {
    clearSession()
  }, [clearSession])

  useEffect(() => {
    setUnauthorizedHandler(() => {
      clearSession()
    })
    return () => setUnauthorizedHandler(null)
  }, [clearSession])

  useEffect(() => {
    const stored = getStoredToken()
    if (!stored) {
      return
    }

    const controller = new AbortController()

    void (async () => {
      try {
        const me = await fetchCurrentUser(controller.signal)
        if (!controller.signal.aborted) {
          setToken(stored)
          setUser(me)
        }
      } catch {
        if (controller.signal.aborted) {
          return
        }
        clearStoredToken()
        setToken(null)
        setUser(null)
        queryClient.clear()
      } finally {
        if (!controller.signal.aborted) {
          setIsInitializing(false)
        }
      }
    })()

    return () => controller.abort()
  }, [queryClient])

  const login = useCallback(async (payload: LoginRequest) => {
    const result = await loginRequest(payload)
    setStoredToken(result.access_token)
    setToken(result.access_token)
    setUser(result.user)
    setIsInitializing(false)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      token,
      user,
      isAuthenticated: Boolean(token && user),
      isInitializing,
      login,
      logout,
    }),
    [token, user, isInitializing, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

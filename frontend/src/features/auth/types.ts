export type UserRole = 'SUPER_ADMIN' | 'ADMIN_RW'

export type User = {
  id: string
  email: string
  name: string
  role: UserRole
}

export type LoginRequest = {
  email: string
  password: string
}

export type LoginResponse = {
  access_token: string
  token_type: string
  expires_in: number
  user: User
}

export type UserRole = 'user' | 'admin'

export const UserRoleMap: Record<UserRole, string> = {
  user: '普通用户',
  admin: '管理员',
}

export interface UserInfo {
  id: number
  username: string
  email: string
  avatar: string
  bio: string
  role: UserRole
  created_at: string
}

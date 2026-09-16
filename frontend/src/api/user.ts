import request from '@/utils/request'
import type { UserInfo } from '@/constants/user'
import type { TastingNote } from '@/constants/note'

export interface LoginResult { token: string; user: UserInfo }

export interface ProfileData {
  user: UserInfo
  note_count: number
  avg_score: number
  top_origins: string[]
  followers: number
  following: number
  likes_received: number
  notes: TastingNote[]
}

export function register(payload: { username: string; email: string; password: string; bio?: string }) {
  return request.post<never, LoginResult>('/users/register', payload)
}
export function login(payload: { username: string; password: string }) {
  return request.post<never, LoginResult>('/users/login', payload)
}
export function getProfile() { return request.get<never, UserInfo>('/users/me') }
export function updateProfile(payload: { bio?: string; avatar?: string }) {
  return request.put<never, UserInfo>('/users/me', payload)
}
export function getUserProfile(id: number | string) { return request.get<never, ProfileData>(`/users/${id}/profile`) }
export function followUser(id: number) { return request.post<never, { follower_id: number }>(`/users/${id}/follow`) }
export function unfollowUser(id: number) { return request.delete<never, { unfollowed: boolean }>(`/users/${id}/follow`) }

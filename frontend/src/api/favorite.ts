import request from '@/utils/request'
import type { FavoriteItem, FavoriteResult } from '@/constants/bean'

// Favorite a bean. Repeat favorites keep a single relation (backend upsert).
export function favoriteBean(id: number) {
  return request.post<never, FavoriteResult>(`/beans/${id}/favorite`)
}

// Cancel a favorite. Idempotent: counts/list/profile read back the new state.
export function unfavoriteBean(id: number) {
  return request.delete<never, FavoriteResult>(`/beans/${id}/favorite`)
}

export interface MyFavorites {
  list: FavoriteItem[]
  total: number
}

export function listMyFavorites(limit = 20) {
  return request.get<never, MyFavorites>('/users/me/favorites', { params: { limit } })
}

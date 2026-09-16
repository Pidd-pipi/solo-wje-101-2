import request from '@/utils/request'
import type { CoffeeBean } from '@/constants/bean'
import type { PageData } from '@/types/api'

export function listBeans(params: { page?: number; page_size?: number; origin?: string; process?: string; keyword?: string }) {
  return request.get<never, PageData<CoffeeBean>>('/beans', { params })
}
export function createBean(payload: Partial<CoffeeBean>) { return request.post<never, CoffeeBean>('/beans', payload) }
export function updateBean(id: number, payload: Partial<CoffeeBean>) { return request.put<never, CoffeeBean>(`/beans/${id}`, payload) }
export function deleteBean(id: number) { return request.delete<never, { deleted: boolean }>(`/beans/${id}`) }

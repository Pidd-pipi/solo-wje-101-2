import request from '@/utils/request'
import type { BrewRecipe } from '@/types/api'
import type { PageData } from '@/types/api'

export function listRecipes(params: { page?: number; page_size?: number; device?: string; keyword?: string }) {
  return request.get<never, PageData<BrewRecipe>>('/recipes', { params })
}
export function getRecipe(id: number | string) { return request.get<never, BrewRecipe>(`/recipes/${id}`) }
export function createRecipe(payload: Partial<BrewRecipe>) { return request.post<never, BrewRecipe>('/recipes', payload) }

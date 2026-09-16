export interface ApiResponse<T = unknown> { code: number; message: string; data: T }
export interface PageData<T> { list: T[]; total: number; page: number; page_size: number }

export interface BrewRecipe {
  id: number
  user_id: number
  name: string
  device: string
  water_temp: number
  grind_size: string
  ratio: string
  steps: string
  created_at: string
}

export interface Comment {
  id: number
  note_id: number
  user_id: number
  content: string
  created_at: string
}

export interface RecipeStep {
  step_number: number
  description: string
  duration_seconds: number
}

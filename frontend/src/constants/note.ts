export type RoastLevel = 'light' | 'medium' | 'dark'

export const RoastLevelMap: Record<RoastLevel, string> = {
  light: '浅烘',
  medium: '中烘',
  dark: '深烘',
}

export const ROAST_LEVELS = Object.keys(RoastLevelMap) as RoastLevel[]

export interface TastingNote {
  id: number
  user_id: number
  coffee_name: string
  origin: string
  roast_level: RoastLevel
  flavor_tags: string
  aroma_score: number
  acidity_score: number
  body_score: number
  overall_score: number
  brew_method: string
  brew_recipe_id: number
  notes_text: string
  image_url: string
  created_at: string
  updated_at: string
}

export interface NoteItem {
  note: TastingNote
  like_count: number
}

export function parseTags(raw: string): string[] {
  try {
    const arr = JSON.parse(raw || '[]')
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

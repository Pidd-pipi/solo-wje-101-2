export type ProcessMethod = 'washed' | 'natural' | 'honey' | 'anaerobic'

export const ProcessMethodMap: Record<ProcessMethod, string> = {
  washed: '水洗',
  natural: '日晒',
  honey: '蜜处理',
  anaerobic: '厌氧发酵',
}

export const PROCESS_METHODS = Object.keys(ProcessMethodMap) as ProcessMethod[]

export interface CoffeeBean {
  id: number
  name: string
  origin: string
  process_method: ProcessMethod
  flavor_tags: string
  description: string
  created_at: string
  // Latest favorite state for the logged-in viewer (anonymous viewers get false).
  is_favored?: boolean
}

export interface FavoriteResult {
  bean_id: number
  favored: boolean
  favorite_id?: number
}

export interface FavoriteItem {
  id: number
  bean_id: number
  name: string
  origin: string
  process_method: ProcessMethod
  flavor_tags: string
  favored_at: string
}

export interface PreferenceItem {
  key: string
  label: string
  count: number
}

export interface TastePreference {
  roast: PreferenceItem[]
  process: PreferenceItem[]
  flavor: PreferenceItem[]
}

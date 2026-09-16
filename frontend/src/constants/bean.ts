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
}

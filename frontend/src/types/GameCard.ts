import type { GamePlatform } from './GamePlatform'

export type GameCardStatus = 'draft' | 'published' | 'archived'

export interface GameCard {
  id: string
  name: string
  platform: GamePlatform
  description: string
  coverImage: string
  status: GameCardStatus
  createdAt: string
}

export interface GameCardQuery {
  page: number
  pageSize: number
  keyword?: string
  platformId?: string
  status?: GameCardStatus
}

export interface GameCardForm {
  id?: string
  name: string
  platformId: string
  description?: string
  coverImage?: string
  status: 'active' | 'inactive' | 'maintenance'
} 
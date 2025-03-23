import type { GameCard } from '@/types/GameCard'
import { mockGamePlatforms } from './GamePlatform'

export const mockGameCards: GameCard[] = [
  {
    id: 'game_001',
    name: 'Cyberpunk 2077',
    platform: mockGamePlatforms[0],
    description: '一款开放世界动作角色扮演游戏，故事发生在夜之城，一个充满权力、魅力和义体改造的巨型都市。',
    coverImage: 'https://example.com/cyberpunk2077.jpg',
    status: 'published',
    createdAt: '2024-01-01T00:00:00Z'
  },
  {
    id: 'game_002',
    name: 'Elden Ring',
    platform: mockGamePlatforms[1],
    description: '一款动作角色扮演游戏，由宫崎英高和乔治·R·R·马丁共同创作。',
    coverImage: 'https://example.com/eldenring.jpg',
    status: 'published',
    createdAt: '2024-01-02T00:00:00Z'
  }
] 
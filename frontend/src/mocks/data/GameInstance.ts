import type { GameInstance } from '@/types/GameInstance'
import { mockGameNodes } from './GameNode'

export const mockGameInstances: GameInstance[] = [
  {
    id: 'instance_001',
    gameId: 'game_001',
    nodeId: 'node_001',
    node: mockGameNodes[0],
    status: 'running',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    lastStartedAt: '2024-01-01T00:00:00Z'
  },
  {
    id: 'instance_002',
    gameId: 'game_002',
    nodeId: 'node_002',
    node: mockGameNodes[1],
    status: 'stopped',
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
    lastStoppedAt: '2024-01-02T00:00:00Z'
  }
] 
export * from './GameCard'
export * from './GamePlatform'
export * from './GameNode'
export * from './GameInstance'

export interface GameInstanceForm {
  id?: string;
  name: string;
  gameCardId: string;
  nodeId: string;
  config: {
    maxPlayers: number;
    port: number;
    settings: {
      map: string;
      difficulty: string;
    };
  };
} 
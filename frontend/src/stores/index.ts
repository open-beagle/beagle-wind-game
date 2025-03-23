import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    user: null
  }),
  actions: {
    setToken(token: string) {
      this.token = token
    },
    setUser(user: any) {
      this.user = user
    },
    logout() {
      this.token = ''
      this.user = null
    }
  }
})

export const useNodeStore = defineStore('node', {
  state: () => ({
    nodes: [],
    currentNode: null
  }),
  actions: {
    setNodes(nodes: any[]) {
      this.nodes = nodes
    },
    setCurrentNode(node: any) {
      this.currentNode = node
    }
  }
})

export const usePlatformStore = defineStore('platform', {
  state: () => ({
    platforms: [],
    currentPlatform: null
  }),
  actions: {
    setPlatforms(platforms: any[]) {
      this.platforms = platforms
    },
    setCurrentPlatform(platform: any) {
      this.currentPlatform = platform
    }
  }
})

export const useCardStore = defineStore('card', {
  state: () => ({
    cards: [],
    currentCard: null
  }),
  actions: {
    setCards(cards: any[]) {
      this.cards = cards
    },
    setCurrentCard(card: any) {
      this.currentCard = card
    }
  }
})

export const useInstanceStore = defineStore('instance', {
  state: () => ({
    instances: [],
    currentInstance: null
  }),
  actions: {
    setInstances(instances: any[]) {
      this.instances = instances
    },
    setCurrentInstance(instance: any) {
      this.currentInstance = instance
    }
  }
}) 
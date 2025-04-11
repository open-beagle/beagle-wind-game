import fs from 'fs'
import path from 'path'
import yaml from 'js-yaml'
import { GamePlatform } from '../src/types/GamePlatform'
import { GameCard } from '../src/types/GameCard'
import { GameNode } from '../src/types/GameNode'
import { GameInstance } from '../src/types/GameInstance'

// 文件路径配置
const CONFIG_DIR = path.resolve(__dirname, '../../config')
const DATA_DIR = path.resolve(__dirname, '../../data')
const MOCKS_DIR = path.resolve(__dirname, '../src/mocks/data')

// 确保目录存在
const ensureDir = (dir: string) => {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true })
  }
}

// 读取 YAML 文件
const readYamlFile = <T>(filePath: string): T => {
  try {
    const content = fs.readFileSync(filePath, 'utf8')
    return yaml.load(content) as T
  } catch (error) {
    console.error(`读取文件失败: ${filePath}`, error)
    throw error
  }
}

// 写入 YAML 文件
const writeYamlFile = (filePath: string, data: any) => {
  try {
    const content = yaml.dump(data, { indent: 2 })
    fs.writeFileSync(filePath, content, 'utf8')
  } catch (error) {
    console.error(`写入文件失败: ${filePath}`, error)
    throw error
  }
}

// 生成 TypeScript 文件
const generateTsFile = (filePath: string, exportName: string, data: any) => {
  const content = `// 此文件由 sync-mocks.ts 自动生成，请勿手动修改
import type { ${exportName} } from '@/types/${exportName}'

export const mock${exportName}s: ${exportName}[] = ${JSON.stringify(data, null, 2)}
`
  fs.writeFileSync(filePath, content, 'utf8')
}

// 验证游戏卡片数据
const validateGameCard = (card: GameCard): boolean => {
  // ID 格式：平台ID-发行日期
  const idPattern = /^[a-z0-9]+-\d{8}$/
  if (!idPattern.test(card.id)) {
    console.error(`游戏卡片 ID 格式错误: ${card.id}`)
    return false
  }

  // 必填字段检查
  const requiredFields = ['name', 'platformId', 'slugName', 'description', 'status']
  for (const field of requiredFields) {
    if (!card[field as keyof GameCard]) {
      console.error(`游戏卡片缺少必填字段: ${field}`)
      return false
    }
  }

  return true
}

// 验证平台数据
const validatePlatform = (platform: GamePlatform): boolean => {
  const requiredFields = ['id', 'name', 'type', 'status', 'version', 'image', 'bin', 'data']
  for (const field of requiredFields) {
    if (!platform[field as keyof GamePlatform]) {
      console.error(`平台缺少必填字段: ${field}`)
      return false
    }
  }

  return true
}

// 同步配置到 mock 数据
const syncConfigToMocks = async () => {
  console.log('开始从配置同步到 mock 数据...')
  
  try {
    // 同步平台数据
    const platforms = readYamlFile<GamePlatform[]>(path.join(CONFIG_DIR, 'gameplatforms.yaml'))
    if (platforms.every(validatePlatform)) {
      generateTsFile(
        path.join(MOCKS_DIR, 'GamePlatform.ts'),
        'GamePlatform',
        platforms
      )
    }

    // 同步游戏卡片数据
    const cards = readYamlFile<GameCard[]>(path.join(CONFIG_DIR, 'cards.yaml'))
    if (cards.every(validateGameCard)) {
      generateTsFile(
        path.join(MOCKS_DIR, 'GameCard.ts'),
        'GameCard',
        cards
      )
    }

    // TODO: 同步其他数据...

    console.log('同步完成！')
  } catch (error) {
    console.error('同步失败:', error)
    process.exit(1)
  }
}

// 同步 mock 数据到配置
const syncMocksToConfig = async () => {
  console.log('开始从 mock 数据同步到配置...')
  
  try {
    ensureDir(CONFIG_DIR)
    
    // 同步平台数据
    const platforms = require('../src/mocks/data/GamePlatform').mockGamePlatforms
    if (platforms.every(validatePlatform)) {
      writeYamlFile(
        path.join(CONFIG_DIR, 'gameplatforms.yaml'),
        platforms
      )
    }

    // 同步游戏卡片数据
    const cards = require('../src/mocks/data/GameCard').mockGameCards
    if (cards.every(validateGameCard)) {
      writeYamlFile(
        path.join(CONFIG_DIR, 'cards.yaml'),
        cards
      )
    }

    // TODO: 同步其他数据...

    console.log('同步完成！')
  } catch (error) {
    console.error('同步失败:', error)
    process.exit(1)
  }
}

// 命令行参数处理
const direction = process.argv[2]
if (direction === 'to-mocks') {
  syncConfigToMocks()
} else if (direction === 'to-config') {
  syncMocksToConfig()
} else {
  console.error('请指定同步方向: to-mocks 或 to-config')
  process.exit(1)
} 
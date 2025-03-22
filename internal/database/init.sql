-- 创建游戏机表
CREATE TABLE IF NOT EXISTS game_nodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    model VARCHAR(50) NOT NULL,
    hardware TEXT,
    network TEXT,
    location VARCHAR(200),
    status VARCHAR(20),
    resources TEXT,
    online BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_online TIMESTAMP WITH TIME ZONE
);

-- 创建游戏平台表
CREATE TABLE IF NOT EXISTS game_platforms (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    version VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    features TEXT,
    requirements TEXT,
    config TEXT,
    network TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建游戏卡片表
CREATE TABLE IF NOT EXISTS game_cards (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    sort_name VARCHAR(100) NOT NULL,
    slug_name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    cover VARCHAR(255),
    category VARCHAR(50),
    release_date DATE,
    tags TEXT,
    files TEXT,
    updates TEXT,
    patches TEXT,
    params TEXT,
    settings TEXT,
    permissions TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建游戏实例表
CREATE TABLE IF NOT EXISTS game_instances (
    id VARCHAR(36) PRIMARY KEY,
    node_id VARCHAR(36) NOT NULL REFERENCES game_nodes(id),
    platform_id VARCHAR(36) NOT NULL REFERENCES game_platforms(id),
    card_id VARCHAR(36) NOT NULL REFERENCES game_cards(id),
    status VARCHAR(20),
    resources TEXT,
    performance TEXT,
    save_data TEXT,
    config TEXT,
    backup TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    stopped_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_game_nodes_status ON game_nodes(status);
CREATE INDEX IF NOT EXISTS idx_game_nodes_online ON game_nodes(online);
CREATE INDEX IF NOT EXISTS idx_game_platforms_type ON game_platforms(type);
CREATE INDEX IF NOT EXISTS idx_game_cards_type ON game_cards(type);
CREATE INDEX IF NOT EXISTS idx_game_cards_category ON game_cards(category);
CREATE INDEX IF NOT EXISTS idx_game_instances_status ON game_instances(status);
CREATE INDEX IF NOT EXISTS idx_game_instances_node_id ON game_instances(node_id);
CREATE INDEX IF NOT EXISTS idx_game_instances_platform_id ON game_instances(platform_id);
CREATE INDEX IF NOT EXISTS idx_game_instances_card_id ON game_instances(card_id);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表添加更新时间触发器
CREATE TRIGGER update_game_nodes_updated_at
    BEFORE UPDATE ON game_nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_platforms_updated_at
    BEFORE UPDATE ON game_platforms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_cards_updated_at
    BEFORE UPDATE ON game_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_instances_updated_at
    BEFORE UPDATE ON game_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 
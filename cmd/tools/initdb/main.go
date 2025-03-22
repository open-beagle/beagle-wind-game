package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

var (
	dbPath = flag.String("db", "data/game.db", "SQLite database path")
)

func main() {
	flag.Parse()

	// 创建数据库连接
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 创建表
	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// 导入数据
	if err := importData(db); err != nil {
		log.Fatalf("Failed to import data: %v", err)
	}

	fmt.Println("Database initialized successfully!")
}

func createTables(db *sql.DB) error {
	// 读取SQL文件
	sqlFile := filepath.Join("internal", "database", "init.sql")
	sqlContent, err := os.ReadFile(sqlFile)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	// 执行SQL语句
	if _, err := db.Exec(string(sqlContent)); err != nil {
		return fmt.Errorf("failed to execute SQL: %v", err)
	}

	return nil
}

func importData(db *sql.DB) error {
	// 导入游戏机数据
	if err := importNodes(db); err != nil {
		return fmt.Errorf("failed to import nodes: %v", err)
	}

	// 导入平台数据
	if err := importPlatforms(db); err != nil {
		return fmt.Errorf("failed to import platforms: %v", err)
	}

	// 导入游戏卡片数据
	if err := importCards(db); err != nil {
		return fmt.Errorf("failed to import cards: %v", err)
	}

	return nil
}

func importNodes(db *sql.DB) error {
	data, err := os.ReadFile("data/nodes/example.yaml")
	if err != nil {
		return err
	}

	var nodes struct {
		Nodes []struct {
			ID         string `yaml:"id"`
			Name       string `yaml:"name"`
			Model      string `yaml:"model"`
			Hardware   string `yaml:"hardware"`
			Network    string `yaml:"network"`
			Location   string `yaml:"location"`
			Status     string `yaml:"status"`
			Resources  string `yaml:"resources"`
			Online     bool   `yaml:"online"`
			LastOnline string `yaml:"last_online"`
		} `yaml:"nodes"`
	}

	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return err
	}

	for _, node := range nodes.Nodes {
		_, err := db.Exec(`
			INSERT INTO game_nodes (id, name, model, hardware, network, location, status, resources, online, last_online)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, node.ID, node.Name, node.Model, node.Hardware, node.Network, node.Location, node.Status, node.Resources, node.Online, node.LastOnline)
		if err != nil {
			return err
		}
	}

	return nil
}

func importPlatforms(db *sql.DB) error {
	data, err := os.ReadFile("data/platforms/example.yaml")
	if err != nil {
		return err
	}

	var platforms struct {
		Platforms []struct {
			ID           string `yaml:"id"`
			Name         string `yaml:"name"`
			Version      string `yaml:"version"`
			Type         string `yaml:"type"`
			Features     string `yaml:"features"`
			Requirements string `yaml:"requirements"`
			Config       string `yaml:"config"`
			Network      string `yaml:"network"`
		} `yaml:"platforms"`
	}

	if err := yaml.Unmarshal(data, &platforms); err != nil {
		return err
	}

	for _, platform := range platforms.Platforms {
		_, err := db.Exec(`
			INSERT INTO game_platforms (id, name, version, type, features, requirements, config, network)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, platform.ID, platform.Name, platform.Version, platform.Type, platform.Features, platform.Requirements, platform.Config, platform.Network)
		if err != nil {
			return err
		}
	}

	return nil
}

func importCards(db *sql.DB) error {
	data, err := os.ReadFile("data/cards/example.yaml")
	if err != nil {
		return err
	}

	var cards struct {
		Cards []struct {
			ID          string `yaml:"id"`
			Name        string `yaml:"name"`
			Type        string `yaml:"type"`
			Description string `yaml:"description"`
			Cover       string `yaml:"cover"`
			Category    string `yaml:"category"`
			Tags        string `yaml:"tags"`
			Files       string `yaml:"files"`
			Updates     string `yaml:"updates"`
			Patches     string `yaml:"patches"`
			Params      string `yaml:"params"`
			Settings    string `yaml:"settings"`
			Permissions string `yaml:"permissions"`
		} `yaml:"cards"`
	}

	if err := yaml.Unmarshal(data, &cards); err != nil {
		return err
	}

	for _, card := range cards.Cards {
		_, err := db.Exec(`
			INSERT INTO game_cards (id, name, type, description, cover, category, tags, files, updates, patches, params, settings, permissions)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, card.ID, card.Name, card.Type, card.Description, card.Cover, card.Category, card.Tags, card.Files, card.Updates, card.Patches, card.Params, card.Settings, card.Permissions)
		if err != nil {
			return err
		}
	}

	return nil
}

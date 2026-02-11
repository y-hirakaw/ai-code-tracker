package main

import (
	"fmt"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// loadStorageAndConfig はストレージ初期化と設定読み込みを共通化するヘルパーです。
// handlers_checkpoint.go と handlers_commit.go で同一パターンが使用されています。
func loadStorageAndConfig() (*storage.AIctStorage, *tracker.Config, error) {
	store, err := storage.NewAIctStorage()
	if err != nil {
		return nil, nil, fmt.Errorf("initializing storage: %w", err)
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	return store, cfg, nil
}

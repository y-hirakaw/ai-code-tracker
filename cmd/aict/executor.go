package main

import "github.com/y-hirakaw/ai-code-tracker/internal/gitexec"

// newExecutor はgit Executorを生成するファクトリ関数です。
// テスト時にモック関数に差し替えることでDIを実現します。
var newExecutor = gitexec.NewExecutor

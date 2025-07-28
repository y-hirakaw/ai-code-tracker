# AI Code Tracker - セキュリティガイド

## 📋 目次

- [セキュリティ概要](#セキュリティ概要)
- [データ暗号化](#データ暗号化)
- [監査ログ](#監査ログ)
- [プライバシー保護](#プライバシー保護)
- [機密ファイル除外](#機密ファイル除外)
- [セキュリティスキャン](#セキュリティスキャン)
- [設定ガイド](#設定ガイド)

## セキュリティ概要

AI Code Tracker は企業環境での安全な利用を前提として、包括的なセキュリティ機能を提供します。

### 主要セキュリティ機能
- 🔐 **AES-256-GCM暗号化** - 機密データの完全保護
- 📋 **包括的監査ログ** - 全操作の追跡証跡
- 🛡️ **入力検証・サニタイゼーション** - 攻撃の防止
- 🔒 **プライバシー保護** - 個人情報の匿名化
- 🚫 **機密ファイル自動除外** - 秘密情報の保護
- 🔍 **セキュリティスキャン** - 定期的な脆弱性チェック

## データ暗号化

### 基本設定
```bash
# 暗号化を有効化
export AICT_ENCRYPT_DATA=true

# パスフレーズを設定
export AICT_ENCRYPTION_PASSPHRASE="your-secure-passphrase"

# 設定確認
aict security status
```

### 暗号化仕様
- **アルゴリズム**: AES-256-GCM
- **キー導出**: PBKDF2 (10,000回)
- **ソルト**: ランダム生成（32バイト）

## 監査ログ

### 設定方法
```bash
# 監査ログを有効化
export AICT_AUDIT_LOG=true

# ログの確認
aict security audit --show
```

### 記録内容
- ファイルアクセス操作
- データ変更操作
- セキュリティイベント
- エラーイベント

## プライバシー保護

### 匿名化設定
```bash
# 作成者名の匿名化
export AICT_ANONYMIZE_AUTHORS=true

# ファイルパスのハッシュ化
export AICT_HASH_FILE_PATHS=true

# データ保持期間（365日）
export AICT_DATA_RETENTION_DAYS=365
```

## 機密ファイル除外

### 基本設定
```bash
# 機密ファイル除外を有効化
export AICT_ENABLE_EXCLUSIONS=true
export AICT_EXCLUDE_SENSITIVE=true
```

### 自動除外対象
- `*.key`, `*.pem`, `*.p12` (秘密鍵)
- `.env`, `.env.*` (環境変数)
- `secrets.yml`, `credentials.json` (認証情報)
- `id_rsa`, `id_dsa` (SSH鍵)

## セキュリティスキャン

### 実行方法
```bash
# 包括的スキャン
aict security scan

# 特定項目のチェック
aict security scan --check permissions
aict security scan --check encryption

# レポート出力
aict security scan --output report.json
```

## 設定ガイド

### セキュリティモード

#### Standard（推奨）
```bash
export AICT_SECURITY_MODE=standard
export AICT_ENCRYPT_DATA=true
export AICT_AUDIT_LOG=true
export AICT_ANONYMIZE_AUTHORS=true
```

#### Strict（高セキュリティ）
```bash
export AICT_SECURITY_MODE=strict
export AICT_ENCRYPT_DATA=true
export AICT_AUDIT_LOG=true
export AICT_ANONYMIZE_AUTHORS=true
export AICT_HASH_FILE_PATHS=true
export AICT_ENABLE_EXCLUSIONS=true
```

詳細な設定については [RDD.md](../RDD.md) のセキュリティセクションを参照してください。
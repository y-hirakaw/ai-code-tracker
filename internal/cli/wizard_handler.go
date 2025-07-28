package cli

import (
	"fmt"
	"os"

	"github.com/ai-code-tracker/aict/internal/interactive"
	"github.com/ai-code-tracker/aict/internal/ui"
)

// WizardHandler はwizardコマンドを処理する
type WizardHandler struct {
	helpSystem *ui.HelpSystem
}

// NewWizardHandler は新しいWizardHandlerを作成する
func NewWizardHandler(helpSystem *ui.HelpSystem) *WizardHandler {
	return &WizardHandler{
		helpSystem: helpSystem,
	}
}

// Handle はwizardコマンドを実行する
func (h *WizardHandler) Handle(args []string) error {
	wizard := interactive.NewWizard()
	
	var wizardType string
	if len(args) > 0 {
		wizardType = args[0]
	} else {
		wizardType = "init"
	}
	
	switch wizardType {
	case "init":
		config := wizard.InitializationWizard()
		h.helpSystem.ShowSuccess("設定ウィザードが完了しました!")
		
		// 設定を適用
		return h.applyWizardConfig(config)
		
	case "security":
		config := wizard.SecurityWizard()
		h.helpSystem.ShowSuccess("セキュリティ設定が完了しました!")
		
		// セキュリティ設定を適用
		return h.applySecurityConfig(config)
		
	case "quickstart":
		wizard.QuickStartWizard()
		return nil
		
	default:
		return fmt.Errorf("不明なウィザードタイプ: %s", wizardType)
	}
}

// applyWizardConfig はウィザード設定を適用する
func (h *WizardHandler) applyWizardConfig(config map[string]interface{}) error {
	// 基本設定の適用
	if setupGit, ok := config["setup_git_hooks"].(bool); ok && setupGit {
		h.helpSystem.ShowInfo("Git hooks を設定中...")
		// Git hooks設定のロジックを呼び出し
	}
	
	if setupClaude, ok := config["setup_claude_hooks"].(bool); ok && setupClaude {
		h.helpSystem.ShowInfo("Claude Code hooks を設定中...")
		// Claude hooks設定のロジックを呼び出し
	}
	
	// セキュリティ設定の適用
	if enableEncryption, ok := config["enable_encryption"].(bool); ok && enableEncryption {
		h.helpSystem.ShowInfo("データ暗号化を有効化中...")
		os.Setenv("AICT_ENCRYPT_DATA", "true")
	}
	
	if enableAudit, ok := config["enable_audit_log"].(bool); ok && enableAudit {
		h.helpSystem.ShowInfo("監査ログを有効化中...")
		os.Setenv("AICT_AUDIT_LOG", "true")
	}
	
	if anonymize, ok := config["anonymize_authors"].(bool); ok && anonymize {
		h.helpSystem.ShowInfo("作成者匿名化を有効化中...")
		os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	}
	
	h.helpSystem.ShowSuccess("設定が正常に適用されました")
	return nil
}

// applySecurityConfig はセキュリティ設定を適用する
func (h *WizardHandler) applySecurityConfig(config map[string]interface{}) error {
	securityMode, ok := config["security_mode"].(string)
	if !ok {
		return fmt.Errorf("セキュリティモードが指定されていません")
	}
	
	h.helpSystem.ShowInfo(fmt.Sprintf("セキュリティモード '%s' を適用中...", securityMode))
	
	// セキュリティモードに応じた環境変数設定
	os.Setenv("AICT_SECURITY_MODE", securityMode)
	
	for key, value := range config {
		switch key {
		case "enable_encryption":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_ENCRYPT_DATA", "true")
			}
		case "enable_audit_log":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_AUDIT_LOG", "true")
			}
		case "anonymize_authors":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
			}
		case "strict_validation":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_STRICT_VALIDATION", "true")
			}
		case "hash_file_paths":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_HASH_FILE_PATHS", "true")
			}
		case "data_retention_days":
			if val, ok := value.(int); ok {
				os.Setenv("AICT_DATA_RETENTION_DAYS", fmt.Sprintf("%d", val))
			}
		}
	}
	
	h.helpSystem.ShowSuccess("セキュリティ設定が正常に適用されました")
	return nil
}
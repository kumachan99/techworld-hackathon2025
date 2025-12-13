// scripts/seed.go
// マスターデータを Firestore に投入するスクリプト
//
// 使用方法:
//   # 本番環境
//   GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json go run scripts/seed.go
//
//   # ローカルエミュレータ
//   FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 GOOGLE_CLOUD_PROJECT=demo-project go run scripts/seed.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"gopkg.in/yaml.v3"
)

// ============================================================================
// データ構造
// ============================================================================

// Effects は街パラメータへの影響値
type Effects struct {
	Economy     int `yaml:"economy"`
	Welfare     int `yaml:"welfare"`
	Education   int `yaml:"education"`
	Environment int `yaml:"environment"`
	Security    int `yaml:"security"`
	HumanRights int `yaml:"humanRights"`
}

// Policy は政策データ
type Policy struct {
	PolicyID    string  `yaml:"policyId"`
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	NewsFlash   string  `yaml:"newsFlash"`
	Effects     Effects `yaml:"effects"`
}

// PoliciesFile は policies.yaml のルート構造
type PoliciesFile struct {
	Policies []Policy `yaml:"policies"`
}

// Coefficients はスコア計算用係数
type Coefficients struct {
	Economy     int `yaml:"economy"`
	Welfare     int `yaml:"welfare"`
	Education   int `yaml:"education"`
	Environment int `yaml:"environment"`
	Security    int `yaml:"security"`
	HumanRights int `yaml:"humanRights"`
}

// Ideology は思想データ
type Ideology struct {
	IdeologyID   string       `yaml:"ideologyId"`
	Name         string       `yaml:"name"`
	Description  string       `yaml:"description"`
	Coefficients Coefficients `yaml:"coefficients"`
}

// IdeologiesFile は ideologies.yaml のルート構造
type IdeologiesFile struct {
	Ideologies []Ideology `yaml:"ideologies"`
}

// ============================================================================
// メイン処理
// ============================================================================

func main() {
	ctx := context.Background()

	// Firebase 初期化
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Firestore クライアント初期化
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer client.Close()

	// データファイルのパスを取得
	dataDir := getDataDir()

	// 政策マスターデータの投入
	if err := seedPolicies(ctx, client, dataDir); err != nil {
		log.Fatalf("Failed to seed policies: %v", err)
	}

	// 思想マスターデータの投入
	if err := seedIdeologies(ctx, client, dataDir); err != nil {
		log.Fatalf("Failed to seed ideologies: %v", err)
	}

	fmt.Println("✅ マスターデータの投入が完了しました")
}

// getDataDir はデータディレクトリのパスを返す
func getDataDir() string {
	// このファイルのディレクトリを基準にする
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Failed to get current file path")
	}
	return filepath.Join(filepath.Dir(filename), "data")
}

// ============================================================================
// 政策データ投入
// ============================================================================

func seedPolicies(ctx context.Context, client *firestore.Client, dataDir string) error {
	fmt.Println("📝 政策マスターデータを投入中...")

	// YAMLファイルを読み込み
	data, err := os.ReadFile(filepath.Join(dataDir, "policies.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read policies.yaml: %w", err)
	}

	var file PoliciesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse policies.yaml: %w", err)
	}

	// Firestore にバッチ書き込み
	batch := client.Batch()
	for _, policy := range file.Policies {
		docRef := client.Collection("master_policies").Doc(policy.PolicyID)
		batch.Set(docRef, map[string]interface{}{
			"policyId":    policy.PolicyID,
			"title":       policy.Title,
			"description": policy.Description,
			"newsFlash":   policy.NewsFlash,
			"effects": map[string]int{
				"economy":     policy.Effects.Economy,
				"welfare":     policy.Effects.Welfare,
				"education":   policy.Effects.Education,
				"environment": policy.Effects.Environment,
				"security":    policy.Effects.Security,
				"humanRights": policy.Effects.HumanRights,
			},
		})
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch commit failed: %w", err)
	}

	fmt.Printf("  ✓ %d 件の政策を投入しました\n", len(file.Policies))
	return nil
}

// ============================================================================
// 思想データ投入
// ============================================================================

func seedIdeologies(ctx context.Context, client *firestore.Client, dataDir string) error {
	fmt.Println("📝 思想マスターデータを投入中...")

	// YAMLファイルを読み込み
	data, err := os.ReadFile(filepath.Join(dataDir, "ideologies.yaml"))
	if err != nil {
		return fmt.Errorf("failed to read ideologies.yaml: %w", err)
	}

	var file IdeologiesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("failed to parse ideologies.yaml: %w", err)
	}

	// Firestore にバッチ書き込み
	batch := client.Batch()
	for _, ideology := range file.Ideologies {
		docRef := client.Collection("master_ideologies").Doc(ideology.IdeologyID)
		batch.Set(docRef, map[string]interface{}{
			"ideologyId":  ideology.IdeologyID,
			"name":        ideology.Name,
			"description": ideology.Description,
			"coefficients": map[string]int{
				"economy":     ideology.Coefficients.Economy,
				"welfare":     ideology.Coefficients.Welfare,
				"education":   ideology.Coefficients.Education,
				"environment": ideology.Coefficients.Environment,
				"security":    ideology.Coefficients.Security,
				"humanRights": ideology.Coefficients.HumanRights,
			},
		})
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch commit failed: %w", err)
	}

	fmt.Printf("  ✓ %d 件の思想を投入しました\n", len(file.Ideologies))
	return nil
}

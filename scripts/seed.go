// scripts/seed.go
// マスターデータを Firestore に投入するスクリプト
//
// 使用方法:
//   GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json go run scripts/seed.go

package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

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

	// 政策マスターデータの投入
	if err := seedPolicies(ctx, client); err != nil {
		log.Fatalf("Failed to seed policies: %v", err)
	}

	// 思想マスターデータの投入
	if err := seedIdeologies(ctx, client); err != nil {
		log.Fatalf("Failed to seed ideologies: %v", err)
	}

	fmt.Println("✅ マスターデータの投入が完了しました")
}

// seedPolicies は政策マスターデータを投入する
func seedPolicies(ctx context.Context, client *firestore.Client) error {
	fmt.Println("📝 政策マスターデータを投入中...")

	policies := []map[string]interface{}{
		{
			"id":          "policy_001",
			"category":    "Economy",
			"title":       "消費税廃止",
			"description": "消費者の負担を軽減し、消費を促進する大胆な経済政策",
			"newsFlash":   "【速報】消費税廃止法案が可決！商店街は歓喜に沸く一方、財政への懸念も",
			"effects":     map[string]int{"economy": 20, "welfare": -15, "education": -10, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_002",
			"category":    "Environment",
			"title":       "再生可能エネルギー推進法",
			"description": "太陽光・風力発電への補助金を大幅に増額し、脱炭素社会を目指す",
			"newsFlash":   "【特報】再エネ推進で CO2 排出量が大幅減！しかし電気代上昇に市民から不満の声も",
			"effects":     map[string]int{"economy": -10, "welfare": 0, "education": 5, "environment": 25, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_003",
			"category":    "Security",
			"title":       "防犯カメラ設置義務化",
			"description": "全ての公共スペースに監視カメラを設置し、犯罪抑止を図る",
			"newsFlash":   "【速報】犯罪発生率が激減！一方でプライバシー侵害を訴える市民団体がデモ",
			"effects":     map[string]int{"economy": -5, "welfare": 0, "education": 0, "environment": 0, "security": 20, "humanRights": -15},
		},
		{
			"id":          "policy_004",
			"category":    "Welfare",
			"title":       "ベーシックインカム導入",
			"description": "全市民に毎月一定額を支給し、最低限の生活を保障する",
			"newsFlash":   "【歴史的決定】BI 開始で貧困率が急低下！財源確保のため増税議論も",
			"effects":     map[string]int{"economy": -20, "welfare": 25, "education": 0, "environment": 0, "security": 0, "humanRights": 10},
		},
		{
			"id":          "policy_005",
			"category":    "Education",
			"title":       "教育無償化",
			"description": "幼稚園から大学まで、全ての教育費用を無償化する",
			"newsFlash":   "【朗報】教育無償化で進学率過去最高に！予算超過で他施策に影響も",
			"effects":     map[string]int{"economy": -15, "welfare": 10, "education": 30, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_006",
			"category":    "Economy",
			"title":       "大型ショッピングモール誘致",
			"description": "郊外に大型商業施設を誘致し、雇用と消費を創出する",
			"newsFlash":   "【経済】巨大モール開業で雇用 5000 人創出！周辺の自然破壊に環境団体が抗議",
			"effects":     map[string]int{"economy": 25, "welfare": 0, "education": 0, "environment": -20, "security": -5, "humanRights": 0},
		},
		{
			"id":          "policy_007",
			"category":    "Environment",
			"title":       "公園緑地化プロジェクト",
			"description": "市内の空き地を公園に整備し、緑豊かな街づくりを推進",
			"newsFlash":   "【環境】緑地面積 30% 増！市民の満足度向上も、維持費が財政を圧迫",
			"effects":     map[string]int{"economy": -10, "welfare": 10, "education": 0, "environment": 20, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_008",
			"category":    "Security",
			"title":       "警察官増員計画",
			"description": "警察官を大幅に増員し、パトロールを強化する",
			"newsFlash":   "【治安】パトロール強化で体感治安が向上！過剰取り締まりへの批判も",
			"effects":     map[string]int{"economy": -15, "welfare": 0, "education": 0, "environment": 0, "security": 25, "humanRights": -5},
		},
		{
			"id":          "policy_009",
			"category":    "Economy",
			"title":       "IT企業優遇税制",
			"description": "IT企業への減税措置により、ハイテク産業の集積を目指す",
			"newsFlash":   "【経済】IT 特区誕生でスタートアップ続々！データセンター増設で電力消費に懸念",
			"effects":     map[string]int{"economy": 20, "welfare": 0, "education": 10, "environment": -10, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_010",
			"category":    "Welfare",
			"title":       "高齢者医療費補助拡大",
			"description": "高齢者の医療費自己負担を軽減し、安心できる老後を実現",
			"newsFlash":   "【福祉】高齢者の受診率向上で健康寿命延伸！現役世代の負担増に反発も",
			"effects":     map[string]int{"economy": -15, "welfare": 20, "education": 0, "environment": 0, "security": 0, "humanRights": 5},
		},
		{
			"id":          "policy_011",
			"category":    "Environment",
			"title":       "自然保護区域拡大",
			"description": "開発制限区域を拡大し、生態系の保全を強化する",
			"newsFlash":   "【環境】希少種の生息確認相次ぐ！開発業者からは反発の声",
			"effects":     map[string]int{"economy": -20, "welfare": 0, "education": 0, "environment": 25, "security": 0, "humanRights": 5},
		},
		{
			"id":          "policy_012",
			"category":    "Security",
			"title":       "夜間外出規制条例",
			"description": "深夜帯の外出を届出制にし、犯罪発生を抑制する",
			"newsFlash":   "【治安】夜間犯罪が激減！「自由の侵害」として違憲訴訟の動きも",
			"effects":     map[string]int{"economy": 0, "welfare": -5, "education": 0, "environment": 0, "security": 15, "humanRights": -25},
		},
		{
			"id":          "policy_013",
			"category":    "Economy",
			"title":       "起業支援ファンド設立",
			"description": "スタートアップへの投資を促進し、イノベーションを加速",
			"newsFlash":   "【経済】ユニコーン企業が誕生！一方で支援を受けられない中小企業から不満",
			"effects":     map[string]int{"economy": 20, "welfare": -10, "education": 5, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_014",
			"category":    "Environment",
			"title":       "市民農園整備事業",
			"description": "市民が気軽に農業体験できる農園を各地に整備する",
			"newsFlash":   "【暮らし】市民農園が大人気！食育効果も期待、予約は半年待ちに",
			"effects":     map[string]int{"economy": 0, "welfare": 10, "education": 5, "environment": 15, "security": 0, "humanRights": 0},
		},
		{
			"id":          "policy_015",
			"category":    "HumanRights",
			"title":       "情報公開条例強化",
			"description": "行政の透明性を高め、市民の知る権利を保障する",
			"newsFlash":   "【政治】情報公開で行政の不正が次々発覚！捜査情報漏洩の懸念も",
			"effects":     map[string]int{"economy": 0, "welfare": 5, "education": 0, "environment": 0, "security": -10, "humanRights": 20},
		},
	}

	batch := client.Batch()
	for _, policy := range policies {
		docRef := client.Collection("master_policies").Doc(policy["id"].(string))
		batch.Set(docRef, policy)
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch commit failed: %w", err)
	}

	fmt.Printf("  ✓ %d 件の政策を投入しました\n", len(policies))
	return nil
}

// seedIdeologies は思想マスターデータを投入する
func seedIdeologies(ctx context.Context, client *firestore.Client) error {
	fmt.Println("📝 思想マスターデータを投入中...")

	ideologies := []map[string]interface{}{
		{
			"id":          "ideology_capitalist",
			"name":        "新自由主義者",
			"description": "経済成長こそが市民の幸福につながると信じる。規制緩和と市場原理を重視。",
			"coefficients": map[string]int{
				"economy": 3, "welfare": 0, "education": 1, "environment": -1, "security": 1, "humanRights": 0,
			},
		},
		{
			"id":          "ideology_socialist",
			"name":        "社会民主主義者",
			"description": "全ての市民に平等な福祉を提供することが最優先。格差是正を目指す。",
			"coefficients": map[string]int{
				"economy": -1, "welfare": 3, "education": 2, "environment": 0, "security": 0, "humanRights": 1,
			},
		},
		{
			"id":          "ideology_environmentalist",
			"name":        "環境保護主義者",
			"description": "持続可能な環境なくして未来はない。自然との共生を最重視。",
			"coefficients": map[string]int{
				"economy": -2, "welfare": 0, "education": 1, "environment": 3, "security": 0, "humanRights": 1,
			},
		},
		{
			"id":          "ideology_authoritarian",
			"name":        "秩序重視派",
			"description": "安全な街こそが全ての基盤。強い統治による社会の安定を求める。",
			"coefficients": map[string]int{
				"economy": 0, "welfare": -1, "education": 0, "environment": 0, "security": 3, "humanRights": -1,
			},
		},
		{
			"id":          "ideology_libertarian",
			"name":        "自由至上主義者",
			"description": "個人の自由と権利を何よりも尊重。政府の介入を最小限に。",
			"coefficients": map[string]int{
				"economy": 1, "welfare": -1, "education": 0, "environment": 0, "security": -1, "humanRights": 3,
			},
		},
		{
			"id":          "ideology_technocrat",
			"name":        "テクノクラート",
			"description": "教育と科学技術の発展が社会を前進させる。知識こそ力。",
			"coefficients": map[string]int{
				"economy": 1, "welfare": 0, "education": 3, "environment": 1, "security": 0, "humanRights": 0,
			},
		},
	}

	batch := client.Batch()
	for _, ideology := range ideologies {
		docRef := client.Collection("master_ideologies").Doc(ideology["id"].(string))
		batch.Set(docRef, ideology)
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("batch commit failed: %w", err)
	}

	fmt.Printf("  ✓ %d 件の思想を投入しました\n", len(ideologies))
	return nil
}

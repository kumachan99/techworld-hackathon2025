package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/interface/gateway/ai"
)

func main() {
	// 環境変数チェック
	token := os.Getenv("SAKURA_AI_TOKEN")
	if token == "" {
		log.Fatal("SAKURA_AI_TOKEN environment variable is not set")
	}
	fmt.Printf("SAKURA_AI_TOKEN is set: %s\n", token)

	// クライアント作成
	client := ai.NewSakuraAIClient()

	// テストケース定義
	testCases := []struct {
		name           string
		petition       string
		cityParams     entity.CityParams
		passedPolicies []*entity.MasterPolicy
		expectApproved bool // 期待する結果（参考）
	}{
		{
			name:     "人権重視国に徴兵令（却下されるべき）",
			petition: "徴兵令を発令して、軍備を強化しましょう。",
			cityParams: entity.CityParams{
				Economy: 50, Welfare: 50, Education: 50,
				Environment: 50, Security: 40, HumanRights: 75, // 人権が高い
			},
			passedPolicies: []*entity.MasterPolicy{
				{Title: "人権保護法", Description: "国民の基本的人権を保護する法律",
					Effects: map[string]int{"humanRights": 15}},
			},
			expectApproved: false,
		},
		{
			name:     "治安重視国に徴兵令（通りやすい）",
			petition: "徴兵令を発令して、軍備を強化しましょう。",
			cityParams: entity.CityParams{
				Economy: 50, Welfare: 50, Education: 50,
				Environment: 50, Security: 70, HumanRights: 35, // 治安が高く人権が低い
			},
			passedPolicies: []*entity.MasterPolicy{
				{Title: "警察増員法", Description: "警察官を大幅に増員",
					Effects: map[string]int{"security": 15, "humanRights": -10}},
			},
			expectApproved: true,
		},
		{
			name:     "経済低迷国に経済支援（通りやすい）",
			petition: "中小企業への補助金を増やして経済を活性化させましょう",
			cityParams: entity.CityParams{
				Economy: 30, Welfare: 50, Education: 50, // 経済が低い
				Environment: 50, Security: 50, HumanRights: 50,
			},
			passedPolicies: nil,
			expectApproved: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	for i, tc := range testCases {
		fmt.Printf("\n========== テスト %d: %s ==========\n", i+1, tc.name)
		fmt.Printf("陳情内容: %s\n", tc.petition)
		fmt.Printf("期待結果: %s\n", map[bool]string{true: "承認", false: "却下"}[tc.expectApproved])
		fmt.Println("審査中...")

		petitionCtx := &ai.PetitionContext{
			PetitionText:   tc.petition,
			PassedPolicies: tc.passedPolicies,
			CityParams:     tc.cityParams,
		}

		result, err := client.ReviewPetition(ctx, petitionCtx)
		if err != nil {
			fmt.Printf("❌ エラー: %v\n", err)
			continue
		}

		if result.Approved {
			fmt.Println("✅ 承認されました")
			fmt.Printf("   タイトル: %s\n", result.Policy.Title)
			fmt.Printf("   説明: %s\n", result.Policy.Description)
			fmt.Printf("   ニュース: %s\n", result.Policy.NewsFlash)
			fmt.Printf("   効果: %+v\n", result.Policy.Effects)
		} else {
			fmt.Println("❌ 却下されました")
			fmt.Printf("   理由: %s\n", result.Reason)
		}

		// 期待結果との比較
		if result.Approved == tc.expectApproved {
			fmt.Println("→ 期待通りの結果です ✓")
		} else {
			fmt.Println("→ 期待と異なる結果です ✗")
		}
	}

	fmt.Println("\n✓ テスト完了")
}

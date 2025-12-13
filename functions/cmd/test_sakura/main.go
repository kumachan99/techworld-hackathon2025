package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/techworld-hackathon/functions/internal/interface/gateway/ai"
)

func main() {
	// 環境変数チェック
	token := os.Getenv("SAKURA_AI_TOKEN")
	if token == "" {
		log.Fatal("SAKURA_AI_TOKEN environment variable is not set")
	}
	fmt.Println("✓ SAKURA_AI_TOKEN is set")

	// クライアント作成
	client := ai.NewSakuraAIClient()

	// テスト用の陳情文
	testPetitions := []string{
		"徴兵令を発令して、軍備を強化しましょう。",
		"学校の援助を大幅に増やしましょう",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for i, petition := range testPetitions {
		fmt.Printf("\n--- テスト %d ---\n", i+1)
		fmt.Printf("陳情内容: %s\n", petition)
		fmt.Println("審査中...")

		result, err := client.ReviewPetition(ctx, petition)
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
	}

	fmt.Println("\n✓ テスト完了")
}

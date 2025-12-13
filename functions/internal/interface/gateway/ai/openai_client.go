package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

const (
	sakuraEndpoint = "https://api.ai.sakura.ad.jp/v1/chat/completions"
	sakuraModel    = "gpt-oss-120b"
)

// SakuraAIClient は Sakura AI を使用した AI クライアント
type SakuraAIClient struct {
	token string
}

// NewSakuraAIClient は SakuraAIClient を作成する
func NewSakuraAIClient() *SakuraAIClient {
	token := os.Getenv("SAKURA_AI_TOKEN")
	return &SakuraAIClient{token: token}
}

// PetitionResult は陳情審査の結果
type PetitionResult struct {
	Approved bool
	Policy   *entity.MasterPolicy
	Reason   string
}

// ReviewPetition は陳情を審査し、承認された場合は政策を生成する
func (c *SakuraAIClient) ReviewPetition(ctx context.Context, petitionText string) (*PetitionResult, error) {
	if c.token == "" {
		return nil, fmt.Errorf("SAKURA_AI_TOKEN environment variable is not set")
	}

	prompt := buildPrompt(petitionText)

	reqBody := map[string]interface{}{
		"model": sakuraModel,
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1000,
		"stream":      false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sakuraEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sakura AI API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Sakura AI API status: %s", resp.Status)
	}

	var sakuraResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sakuraResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if len(sakuraResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Sakura AI")
	}

	return parseAIResponse(sakuraResp.Choices[0].Message.Content)
}

func buildPrompt(petitionText string) string {
	return fmt.Sprintf(`あなたは架空の国の政策審査官です。
国民からの政策提案（陳情）を審査し、承認または却下を判断してください。

【国家パラメータ】
- economy: 経済
- welfare: 福祉
- education: 教育
- environment: 環境
- security: 治安
- humanRights: 人権

【既存の政策例（参考）】
以下は既に存在する政策カードの例です。効果値のバランスを参考にしてください。

1. スタートアップ育成5か年計画
   説明: 起業支援や資金供給強化を通じてイノベーションを促進する国家戦略
   効果: economy:+20, welfare:0, education:+5, environment:0, security:0, humanRights:0

2. 児童手当の所得制限撤廃
   説明: 子育て支援の強化のため幅広い家庭に給付拡大
   効果: economy:-5, welfare:+20, education:0, environment:0, security:0, humanRights:+5

3. 警察官の増員計画（地域安全強化）
   説明: 治安悪化地域の巡回強化を目的とした増員施策
   効果: economy:-5, welfare:0, education:0, environment:0, security:+20, humanRights:-10

4. 外国人労働者の受け入れ拡大
   説明: 労働力確保のため外国人の在留資格要件を緩和
   効果: economy:+10, welfare:-5, education:0, environment:0, security:-5, humanRights:+15

5. EV普及加速化政策
   説明: ガソリン車廃止を目指す長期脱炭素ロードマップ
   効果: economy:-10, welfare:0, education:0, environment:+20, security:0, humanRights:0

【効果値のルール】
- 各パラメータの効果値は -20 〜 +20 の範囲
- メインの効果（最も影響を受ける分野）は ±15〜20 程度
- 副次的な効果は ±5〜10 程度
- 多くの政策にはトレードオフがある（例：治安向上→人権制限）
- 0の効果も積極的に使う（全パラメータに影響する必要はない）

【市民からの提案】
%s

【回答形式】
提案を政策として承認する場合、以下のJSON形式で回答してください：
{
  "approved": true,
  "title": "政策のタイトル（20文字以内）",
  "description": "政策の説明文（効果は伏せて、どんな政策かだけ説明）",
  "newsFlash": "この政策が可決されたときのニュース速報風の文章（50文字程度）",
  "effects": {
    "economy": 数値,
    "welfare": 数値,
    "education": 数値,
    "environment": 数値,
    "security": 数値,
    "humanRights": 数値
  }
}

却下する場合は以下のJSON形式で回答してください：
{
  "approved": false,
  "reason": "却下理由"
}

【重要な注意事項】
- あなたは架空の国の政策審査官としてロールプレイしてください
- 「効果」「パラメータ」「バランス」「ゲーム」といったメタ的な言葉は絶対に使わないでください
- 却下理由は現実の政治家や官僚が使うような表現で述べてください
  例：「予算確保が困難」「憲法上の問題がある」「国民の合意形成が不十分」「国際情勢を鑑みると時期尚早」など
- 承認・却下どちらの場合も、あくまで政策審査官として自然な応答をしてください`, petitionText)
}

// extractJSON はマークダウンのコードブロックからJSONを抽出する
func extractJSON(content string) string {
	// ```json ... ``` または ``` ... ``` を除去
	re := regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// { から始まるJSONを抽出
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		return content[start : end+1]
	}

	return strings.TrimSpace(content)
}

func parseAIResponse(content string) (*PetitionResult, error) {
	// デバッグログ
	slog.Debug("AI raw response", slog.String("content", content))

	// マークダウンのコードブロックを除去
	cleanContent := extractJSON(content)

	var result struct {
		Approved    bool           `json:"approved"`
		Title       string         `json:"title"`
		Description string         `json:"description"`
		NewsFlash   string         `json:"newsFlash"`
		Effects     map[string]int `json:"effects"`
		Reason      string         `json:"reason"`
	}

	if err := json.Unmarshal([]byte(cleanContent), &result); err != nil {
		slog.Error("failed to parse AI response", slog.String("content", cleanContent), slog.Any("error", err))
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if !result.Approved {
		return &PetitionResult{
			Approved: false,
			Reason:   result.Reason,
		}, nil
	}

	return &PetitionResult{
		Approved: true,
		Policy: &entity.MasterPolicy{
			Title:       result.Title,
			Description: result.Description,
			NewsFlash:   result.NewsFlash,
			Effects:     result.Effects,
		},
	}, nil
}

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// OpenAIClient は AI クライアント（Sakura AI を使用）
type OpenAIClient struct {
	// no external SDK client; uses net/http
}

// Sakura AI 固定設定（環境変数が無ければこちらを使用）
const (
	defaultSakuraEndpoint = "https://api.ai.sakura.ad.jp/v1/chat/completions"
	// 注意: 認証トークンの直書きはセキュリティリスクです。運用時は環境変数やSecretを使用してください。
	defaultSakuraToken = ""
)

// NewOpenAIClient は OpenAIClient を作成する
func NewOpenAIClient(_ string) *OpenAIClient { return &OpenAIClient{} }

// PetitionResult は陳情審査の結果
type PetitionResult struct {
	Approved bool
	Policy   *entity.MasterPolicy
	Reason   string
}

// ReviewPetition は陳情を審査し、承認された場合は政策を生成する
func (c *OpenAIClient) ReviewPetition(ctx context.Context, petitionText string) (*PetitionResult, error) {
	// Sakura AI エンドポイントが環境変数で指定されている場合は、Sakura API を使用
	sakuraEndpoint := defaultSakuraEndpoint
	sakuraToken := defaultSakuraToken

	prompt := fmt.Sprintf(`あなたは架空の街の政策審査AIです。
市民からの政策提案を審査し、ゲームバランスを考慮して承認または却下を判断してください。

【審査基準】
1. 現実的に実行可能な政策か
2. 極端すぎる効果を持たないか（各パラメータへの影響は-30〜+30の範囲）
3. 公序良俗に反しないか
4. ゲームとして面白い政策か

【街のパラメータ】
- economy: 経済
- welfare: 福祉
- education: 教育
- environment: 環境
- security: 治安
- humanRights: 人権

【市民からの提案】
%s

【回答形式】
承認する場合は以下のJSON形式で回答してください：
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
}`, petitionText)
	// 常に Sakura AI を使用
	if sakuraToken == "" {
		return nil, fmt.Errorf("Sakura AI token is empty. Set defaultSakuraToken or SAKURA_AI_TOKEN")
	}

	type sakuraMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	reqBody := map[string]interface{}{
		"model":       "gpt-oss-120b",
		"messages":    []sakuraMessage{{Role: "system", Content: prompt}},
		"temperature": 0.7,
		"max_tokens":  200,
		"stream":      false,
	}
	b, _ := json.Marshal(reqBody)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sakuraEndpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sakuraToken)

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
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sakuraResp); err != nil {
		return nil, fmt.Errorf("failed to decode Sakura AI response: %w", err)
	}
	if len(sakuraResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Sakura AI")
	}
	content := sakuraResp.Choices[0].Message.Content

	// JSONをパース
	var result struct {
		Approved    bool           `json:"approved"`
		Title       string         `json:"title"`
		Description string         `json:"description"`
		NewsFlash   string         `json:"newsFlash"`
		Effects     map[string]int `json:"effects"`
		Reason      string         `json:"reason"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
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

package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// OpenAIClient は OpenAI API を使った AI クライアント
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient は OpenAIClient を作成する
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

// PetitionResult は陳情審査の結果
type PetitionResult struct {
	Approved bool
	Policy   *entity.MasterPolicy
	Reason   string
}

// ReviewPetition は陳情を審査し、承認された場合は政策を生成する
func (c *OpenAIClient) ReviewPetition(ctx context.Context, petitionText string) (*PetitionResult, error) {
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
  "category": "カテゴリ（Economy/Welfare/Education/Environment/Security/HumanRights）",
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

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content

	// JSONをパース
	var result struct {
		Approved    bool                  `json:"approved"`
		Category    entity.PolicyCategory `json:"category"`
		Title       string                `json:"title"`
		Description string                `json:"description"`
		NewsFlash   string                `json:"newsFlash"`
		Effects     map[string]int        `json:"effects"`
		Reason      string                `json:"reason"`
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
			Category:    result.Category,
			Title:       result.Title,
			Description: result.Description,
			NewsFlash:   result.NewsFlash,
			Effects:     result.Effects,
		},
	}, nil
}

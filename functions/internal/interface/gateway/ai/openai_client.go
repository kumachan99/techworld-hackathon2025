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

// PetitionContext は陳情審査のコンテキスト情報
type PetitionContext struct {
	PetitionText   string                 // 陳情テキスト
	PassedPolicies []*entity.MasterPolicy // これまで採用された政策
	CityParams     entity.CityParams      // 現在の国のパラメータ
}

// ReviewPetition は陳情を審査し、承認された場合は政策を生成する
func (c *SakuraAIClient) ReviewPetition(ctx context.Context, petitionCtx *PetitionContext) (*PetitionResult, error) {
	if c.token == "" {
		return nil, fmt.Errorf("SAKURA_AI_TOKEN environment variable is not set")
	}

	prompt := buildPrompt(petitionCtx)

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

func buildPrompt(petitionCtx *PetitionContext) string {
	// 現在の国の状況を構築
	currentStatus := buildCurrentStatus(petitionCtx.CityParams)

	// 過去に採用された政策の履歴を構築
	policyHistory := buildPolicyHistory(petitionCtx.PassedPolicies)

	return fmt.Sprintf(`あなたは架空の国の政策審査官です。
国民からの政策提案（陳情）を審査し、承認または却下を判断してください。

【国家パラメータの説明】
- economy: 経済（産業・雇用・財政）
- welfare: 福祉（社会保障・医療・年金）
- education: 教育（学校・研究・人材育成）
- environment: 環境（自然保護・エネルギー・持続可能性）
- security: 治安（警察・防犯・公共安全）
- humanRights: 人権（自由・平等・プライバシー）

%s

%s

【市民からの提案】
%s

===========================================
【審査プロセス】

まず、上記の情報から「この国はどんな国か」を推論してください。
例えば：
- 北欧型福祉国家（福祉・人権重視、高税率容認）
- 自由経済国家（経済・規制緩和重視、小さな政府志向）
- 環境先進国（環境・持続可能性重視、経済コスト容認）
- 安全保障重視国家（治安・秩序重視、自由制限容認）
- バランス型国家（特定の偏りなし）
など

次に、推論した国の性格を踏まえて、この陳情が通るかどうかを判断してください。
- その国の国民が支持するか？
- その国のこれまでの政策方針と整合するか？
- その国の価値観に反していないか？

===========================================
【効果値のルール（承認時のみ使用）】
- 各パラメータの効果値は -20 〜 +20 の範囲
- メインの効果は ±15〜20 程度、副次的な効果は ±5〜10 程度
- トレードオフを意識する（例：治安向上→人権制限）
- 0の効果も積極的に使う

【回答形式】
承認する場合：
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

却下する場合：
{
  "approved": false,
  "reason": "却下理由（その国の性格を踏まえた理由）"
}

【重要な注意事項】
- 架空の国の政策審査官としてロールプレイしてください
- 「効果」「パラメータ」「バランス」「ゲーム」といったメタ的な言葉は絶対に使わないでください
- 却下理由は現実の政治家や官僚が使うような表現で述べてください
- JSONのみを出力してください（思考過程は出力しないでください）`, currentStatus, policyHistory, petitionCtx.PetitionText)
}

// buildCurrentStatus は現在の国の状況を文字列で構築する
func buildCurrentStatus(cityParams entity.CityParams) string {
	var sb strings.Builder
	sb.WriteString("【現在の国の状況】\n")
	sb.WriteString("各分野の現在値（50が初期値、0以下または100以上で国家崩壊）:\n\n")

	params := []struct {
		name  string
		value int
	}{
		{"経済", cityParams.Economy},
		{"福祉", cityParams.Welfare},
		{"教育", cityParams.Education},
		{"環境", cityParams.Environment},
		{"治安", cityParams.Security},
		{"人権", cityParams.HumanRights},
	}

	// 高い分野と低い分野を分類
	var highValues, lowValues, normalValues []string
	for _, p := range params {
		status := getStatusDescription(p.value)
		line := fmt.Sprintf("%s: %d %s", p.name, p.value, status)
		if p.value >= 60 {
			highValues = append(highValues, line)
		} else if p.value <= 40 {
			lowValues = append(lowValues, line)
		} else {
			normalValues = append(normalValues, line)
		}
	}

	// 国の価値観（高い分野）を強調
	if len(highValues) > 0 {
		sb.WriteString("★ この国が重視している価値観:\n")
		for _, v := range highValues {
			sb.WriteString(fmt.Sprintf("  - %s\n", v))
		}
		sb.WriteString("\n")
	}

	// 課題（低い分野）
	if len(lowValues) > 0 {
		sb.WriteString("▼ この国の課題:\n")
		for _, v := range lowValues {
			sb.WriteString(fmt.Sprintf("  - %s\n", v))
		}
		sb.WriteString("\n")
	}

	// 標準的な分野
	if len(normalValues) > 0 {
		sb.WriteString("- その他:\n")
		for _, v := range normalValues {
			sb.WriteString(fmt.Sprintf("  - %s\n", v))
		}
	}

	return sb.String()
}

// getStatusDescription は数値から状況の説明を返す
func getStatusDescription(value int) string {
	switch {
	case value <= 20:
		return "（危機的）"
	case value <= 35:
		return "（低迷）"
	case value <= 45:
		return "（やや低い）"
	case value <= 55:
		return "（標準）"
	case value <= 65:
		return "（やや高い）"
	case value <= 80:
		return "（好調）"
	default:
		return "（過熱気味）"
	}
}

// buildPolicyHistory は過去に採用された政策の履歴を文字列で構築する
func buildPolicyHistory(policies []*entity.MasterPolicy) string {
	if len(policies) == 0 {
		return "【これまでに採用された政策】\n（まだ政策は採用されていません）"
	}

	var sb strings.Builder
	sb.WriteString("【これまでに採用された政策】\n")
	sb.WriteString("この国では以下の政策が国民投票により可決されました:\n\n")

	for i, p := range policies {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, p.Title))
		sb.WriteString(fmt.Sprintf("   説明: %s\n", p.Description))
		// effectsも表示（審査官は把握している設定）
		sb.WriteString(fmt.Sprintf("   影響: economy:%+d, welfare:%+d, education:%+d, environment:%+d, security:%+d, humanRights:%+d\n\n",
			p.Effects["economy"], p.Effects["welfare"], p.Effects["education"],
			p.Effects["environment"], p.Effects["security"], p.Effects["humanRights"]))
	}

	return sb.String()
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

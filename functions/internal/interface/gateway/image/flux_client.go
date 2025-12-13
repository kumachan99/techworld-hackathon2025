package image

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
	"github.com/techworld-hackathon/functions/internal/domain/service"
)

const (
	defaultFluxEndpoint = "http://133.242.48.33/generate"
)

// FluxClient は FLUX.1 schnell API クライアント
type FluxClient struct {
	endpoint string
	apiKey   string
}

// NewFluxClient は FluxClient を作成する
func NewFluxClient() *FluxClient {
	endpoint := os.Getenv("FLUX_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultFluxEndpoint
	}
	apiKey := os.Getenv("FLUX_API_KEY")
	return &FluxClient{
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

// GenerateRequest は画像生成リクエスト
type GenerateRequest struct {
	Prompt            string `json:"prompt"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	NumInferenceSteps int    `json:"num_inference_steps"`
	Seed              int    `json:"seed,omitempty"`
	MaxSequenceLength int    `json:"max_sequence_length"`
}

// GenerateResponse は画像生成レスポンス
type GenerateResponse struct {
	Success bool   `json:"success"`
	Image   string `json:"image"` // Base64エンコードされたPNG
	Seed    int    `json:"seed"`
}

// インターフェースの実装を保証
var _ service.ImageGenerator = (*FluxClient)(nil)

// GenerateCityImage は街のパラメータから街の風景画像を生成する
func (c *FluxClient) GenerateCityImage(ctx context.Context, cityParams *entity.CityParams, passedPolicies []*entity.MasterPolicy) (*service.ImageGenerateResult, error) {
	prompt := buildCityPrompt(cityParams, passedPolicies)

	reqBody := GenerateRequest{
		Prompt:            prompt,
		Width:             1024,
		Height:            768,
		NumInferenceSteps: 4,
		MaxSequenceLength: 512,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FLUX API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("FLUX API status: %s", resp.Status)
	}

	var fluxResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&fluxResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &service.ImageGenerateResult{
		Image: fluxResp.Image,
		Seed:  fluxResp.Seed,
	}, nil
}

// buildCityPrompt は街のパラメータからプロンプトを生成する
func buildCityPrompt(cityParams *entity.CityParams, passedPolicies []*entity.MasterPolicy) string {
	var elements []string

	// ベースプロンプト（フォトリアルスタイル）
	baseStyle := "Photorealistic aerial view of a modern city, professional photography, golden hour lighting, ultra detailed, 8k resolution"

	// 経済に基づく要素
	elements = append(elements, describeEconomy(cityParams.Economy))

	// 環境に基づく要素
	elements = append(elements, describeEnvironment(cityParams.Environment))

	// 福祉に基づく要素
	elements = append(elements, describeWelfare(cityParams.Welfare))

	// 治安に基づく要素
	elements = append(elements, describeSecurity(cityParams.Security))

	// 教育に基づく要素
	elements = append(elements, describeEducation(cityParams.Education))

	// 人権に基づく要素
	elements = append(elements, describeHumanRights(cityParams.HumanRights))

	// 過去の政策に基づく特別な要素
	policyElements := describePolicies(passedPolicies)
	if policyElements != "" {
		elements = append(elements, policyElements)
	}

	// 全体的な雰囲気を追加
	atmosphere := describeOverallAtmosphere(cityParams)
	elements = append(elements, atmosphere)

	return baseStyle + ", " + strings.Join(elements, ", ")
}

func describeEconomy(value int) string {
	switch {
	case value >= 80:
		return "towering glass skyscrapers, luxury shopping districts, high-end cars on streets, prosperous business centers, construction cranes everywhere"
	case value >= 60:
		return "modern office buildings, busy commercial areas, well-maintained infrastructure, thriving downtown"
	case value >= 40:
		return "mixed urban landscape, some commercial activity, moderate development"
	case value >= 20:
		return "older buildings with some disrepair, vacant storefronts, reduced commercial activity"
	default:
		return "abandoned factories, boarded-up shops, crumbling infrastructure, economic depression visible"
	}
}

func describeEnvironment(value int) string {
	switch {
	case value >= 80:
		return "lush green parks everywhere, rooftop gardens, solar panels on buildings, crystal clear blue sky, clean rivers, abundant trees lining streets"
	case value >= 60:
		return "well-maintained parks, some green spaces, relatively clean air, visible environmental efforts"
	case value >= 40:
		return "limited green spaces, some urban vegetation, average air quality"
	case value >= 20:
		return "few parks, smoggy atmosphere, industrial pollution visible, brown haze in sky"
	default:
		return "heavy smog obscuring buildings, polluted waterways, dead trees, industrial smokestacks belching smoke, environmental disaster"
	}
}

func describeWelfare(value int) string {
	switch {
	case value >= 80:
		return "modern hospitals visible, community centers, accessible public facilities, well-dressed diverse citizens, clean public spaces"
	case value >= 60:
		return "adequate public facilities, functional healthcare buildings, organized public areas"
	case value >= 40:
		return "basic public services visible, some public facilities"
	case value >= 20:
		return "overcrowded public facilities, visible poverty, worn public infrastructure"
	default:
		return "homeless encampments, dilapidated public buildings, visible suffering, stark inequality"
	}
}

func describeSecurity(value int) string {
	switch {
	case value >= 80:
		return "clean well-lit streets, orderly traffic, peaceful atmosphere, safe-looking neighborhoods"
	case value >= 60:
		return "generally safe streets, security presence, maintained public order"
	case value >= 40:
		return "normal urban environment, standard security measures"
	case value >= 20:
		return "graffiti on walls, some areas looking neglected, security barriers visible"
	default:
		return "barred windows, security checkpoints, damaged buildings, tense atmosphere, visible decay"
	}
}

func describeEducation(value int) string {
	switch {
	case value >= 80:
		return "prestigious university campuses visible, modern school buildings, libraries, research facilities"
	case value >= 60:
		return "well-maintained schools, educational institutions present"
	case value >= 40:
		return "standard educational buildings, average school facilities"
	case value >= 20:
		return "older school buildings, limited educational infrastructure"
	default:
		return "neglected school buildings, closed libraries, lack of educational facilities"
	}
}

func describeHumanRights(value int) string {
	switch {
	case value >= 80:
		return "diverse crowds of people, street art and cultural expression, open public gatherings, vibrant street life"
	case value >= 60:
		return "mixed population visible, cultural venues, public expression"
	case value >= 40:
		return "typical urban population, some cultural elements"
	case value >= 20:
		return "uniform appearance, surveillance cameras visible, controlled public spaces"
	default:
		return "heavy surveillance infrastructure, restricted areas, conformist atmosphere, oppressive feeling"
	}
}

func describePolicies(policies []*entity.MasterPolicy) string {
	if len(policies) == 0 {
		return ""
	}

	var policyKeywords []string
	for _, p := range policies {
		// タイトルから主要なキーワードを抽出して英語に変換
		keywords := extractPolicyKeywords(p.Title)
		if keywords != "" {
			policyKeywords = append(policyKeywords, keywords)
		}
	}

	if len(policyKeywords) > 0 {
		return strings.Join(policyKeywords, ", ")
	}
	return ""
}

func extractPolicyKeywords(title string) string {
	// 政策タイトルからビジュアル要素を抽出
	keywordMap := map[string]string{
		// 既存の政策マスター
		"消費税":       "bustling shopping areas with many shoppers",
		"再生可能":      "wind turbines and solar panels on rooftops",
		"防犯カメラ":     "security cameras on poles and buildings",
		"ベーシックインカム": "content citizens relaxing in public spaces",
		"教育無償化":     "crowded school buildings with happy students",
		"ショッピングモール": "large shopping mall complexes",
		"公園":        "beautifully landscaped public parks",
		"緑地":        "abundant green spaces and trees",
		"警察":        "police officers patrolling streets",
		"IT企業":      "modern tech company headquarters",
		"高齢者":       "elderly people enjoying public amenities",
		"自然保護":      "protected natural areas and wildlife",
		"夜間外出規制":    "quiet empty streets at night with street lights",
		"起業":        "startup offices and co-working spaces",
		"市民農園":      "urban community gardens and farms",
		"情報公開":      "transparent government buildings with open design",
		// 追加キーワード
		"AI":   "futuristic tech buildings with digital displays",
		"軍事":   "military vehicles and personnel visible",
		"移民":   "culturally diverse population on streets",
		"原発":   "power plant cooling towers in distance",
		"医療":   "modern hospital complex with ambulances",
		"年金":   "elderly people peacefully in parks",
		"規制緩和": "construction cranes and development activity",
	}

	for keyword, visual := range keywordMap {
		if strings.Contains(title, keyword) {
			return visual
		}
	}
	return ""
}

func describeOverallAtmosphere(cityParams *entity.CityParams) string {
	// 全体的な雰囲気を計算
	avg := (cityParams.Economy + cityParams.Welfare + cityParams.Education +
		cityParams.Environment + cityParams.Security + cityParams.HumanRights) / 6

	switch {
	case avg >= 70:
		return "utopian prosperous city, hopeful atmosphere, bright future feeling"
	case avg >= 55:
		return "thriving modern city, optimistic atmosphere"
	case avg >= 45:
		return "typical modern city, neutral atmosphere"
	case avg >= 30:
		return "struggling city, somewhat gloomy atmosphere"
	default:
		return "dystopian cityscape, dark oppressive atmosphere, decline visible everywhere"
	}
}

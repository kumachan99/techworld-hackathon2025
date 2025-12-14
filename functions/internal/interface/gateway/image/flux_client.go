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

	// ベースプロンプト（遠景・航空視点のフォトリアルスタイル）
	baseStyle := "Photorealistic aerial view of a city, professional photography, golden hour lighting, ultra detailed, 8k resolution"

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
		return "luxury brand stores with elegant window displays, Tesla and BMW cars parked on street, business people in suits, gleaming glass storefronts, upscale cafes with outdoor seating"
	case value >= 60:
		return "busy shopping street with well-dressed pedestrians, modern retail stores, clean sidewalks, food delivery bikes, mix of local and chain stores"
	case value >= 40:
		return "ordinary shops and convenience stores, some vacant storefronts, mix of old and new buildings, average cars parked along street"
	case value >= 20:
		return "many closed shops with shutters down, for rent signs in windows, older worn buildings, few pedestrians, discount stores"
	default:
		return "boarded up storefronts with graffiti, broken windows, abandoned buildings, homeless people visible, trash on streets, very few cars"
	}
}

func describeEnvironment(value int) string {
	switch {
	case value >= 80:
		return "lush green street trees with full canopy, flower planters on sidewalks, solar panels visible on rooftops, crystal clear blue sky, bicycle lanes, electric vehicle charging stations"
	case value >= 60:
		return "healthy street trees, some planters with flowers, clean streets, recycling bins visible, partly cloudy sky"
	case value >= 40:
		return "sparse street trees, some litter on sidewalks, gray sky, mix of electric and gas cars"
	case value >= 20:
		return "bare or dying trees, visible smog in air, overflowing garbage bins, hazy brownish sky, no green spaces"
	default:
		return "dead trees with bare branches, thick smog obscuring buildings, garbage piled on corners, brown polluted sky, people wearing masks"
	}
}

func describeWelfare(value int) string {
	switch {
	case value >= 80:
		return "families with strollers on clean sidewalks, elderly people on benches smiling, children playing safely, accessible ramps and crosswalks, well-maintained public toilets"
	case value >= 60:
		return "mix of ages walking comfortably, bus stops with shelters, public benches in good condition, people waiting at crosswalks"
	case value >= 40:
		return "ordinary pedestrians of various ages, basic street furniture, some worn public facilities"
	case value >= 20:
		return "elderly struggling with bags, worn out bus stops, people sleeping on benches, visibly poor people"
	default:
		return "homeless people with cardboard shelters, beggars on corners, people in ragged clothes, abandoned shopping carts, tent encampments"
	}
}

func describeSecurity(value int) string {
	switch {
	case value >= 80:
		return "bright street lights, clean crosswalks, women walking alone safely, children on bicycles, no graffiti, security cameras discretely placed"
	case value >= 60:
		return "well-lit streets, police officer visible in distance, orderly parking, functioning traffic lights"
	case value >= 40:
		return "average street lighting, some graffiti on walls, normal pedestrian activity"
	case value >= 20:
		return "graffiti covering walls, broken street lights, bars on shop windows, people looking over shoulders"
	default:
		return "heavy graffiti everywhere, smashed windows, barbed wire on fences, people hurrying nervously, security shutters down, dark shadowy corners"
	}
}

func describeEducation(value int) string {
	switch {
	case value >= 80:
		return "bookstore with crowded display window, students with tablets and laptops at cafe, modern library building visible, tutoring center signs, cultural posters on walls"
	case value >= 60:
		return "bookshop visible, students with backpacks walking, public library sign, educational advertisement boards"
	case value >= 40:
		return "few students visible, basic convenience store, ordinary commercial signage"
	case value >= 20:
		return "no bookstores visible, mostly entertainment shops, pachinko parlor signs, few young people"
	default:
		return "gambling parlors and adult entertainment signs, no educational facilities visible, loitering youth, vandalized public signs"
	}
}

func describeHumanRights(value int) string {
	switch {
	case value >= 80:
		return "diverse crowd with different ethnicities and styles, rainbow flags visible, street musicians performing, political posters on walls, open air market with various vendors"
	case value >= 60:
		return "mix of people from different backgrounds, some street art, community bulletin board, outdoor cafe conversations"
	case value >= 40:
		return "mostly homogeneous crowd, standard urban population, neutral expressions"
	case value >= 20:
		return "surveillance cameras prominently placed, uniformly dressed people, no street art, controlled atmosphere"
	default:
		return "heavy CCTV cameras everywhere, propaganda posters, people avoiding eye contact, uniformed officials visible, no personal expression, oppressive atmosphere"
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
	// 政策タイトルからビジュアル要素を抽出（街中視点）
	keywordMap := map[string]string{
		// 既存の政策マスター
		"消費税":       "sale signs in shop windows, shoppers with many bags",
		"再生可能":      "solar panels on nearby rooftops, electric car charging station visible",
		"防犯カメラ":     "security cameras mounted on poles and building corners",
		"ベーシックインカム": "relaxed people at outdoor cafes, leisurely pedestrians",
		"教育無償化":     "students in uniforms walking happily, tutoring school signs",
		"ショッピングモール": "large shopping center entrance visible, escalators through glass",
		"公園":        "green park visible at intersection, children on playground",
		"緑地":        "flower beds along sidewalk, small garden plots visible",
		"警察":        "police officers walking beat, police box visible",
		"IT企業":      "tech company logos on buildings, people with laptops at cafe",
		"高齢者":       "elderly couples walking arm in arm, accessible benches",
		"自然保護":      "bird feeders on trees, wildlife crossing signs",
		"夜間外出規制":    "curfew notice boards, empty streets with patrol car",
		"起業":        "co-working space sign, startup logos in windows",
		"市民農園":      "community garden plots visible, people tending vegetables",
		"情報公開":      "public information boards, transparent glass government office",
		// 追加キーワード
		"AI":   "digital displays showing AI services, robot delivery on sidewalk",
		"軍事":   "military recruitment poster, uniformed personnel visible",
		"移民":   "diverse ethnic restaurants, multilingual signs",
		"原発":   "power line infrastructure prominent, energy company ads",
		"医療":   "pharmacy with green cross sign, ambulance passing",
		"年金":   "senior citizens center sign, elderly at cafe tables",
		"規制緩和": "construction scaffolding, new building going up",
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
		return "warm golden sunlight, people smiling and chatting, vibrant energetic mood, sense of prosperity and hope"
	case avg >= 55:
		return "pleasant sunny day, people walking with purpose, generally positive mood, clean and orderly"
	case avg >= 45:
		return "overcast day, neutral busy atmosphere, typical urban scene"
	case avg >= 30:
		return "gloomy gray light, people hurrying with heads down, tense uneasy mood, signs of neglect"
	default:
		return "dark oppressive atmosphere, harsh shadows, people looking fearful or desperate, sense of decay and hopelessness"
	}
}

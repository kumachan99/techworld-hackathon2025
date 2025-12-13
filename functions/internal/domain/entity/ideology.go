package entity

// MasterIdeology は思想マスターを表す
// パス: master_ideologies/{ideologyId}
// IdeologyID はドキュメントIDと同一
type MasterIdeology struct {
	IdeologyID   string             `json:"ideologyId" firestore:"ideologyId"`
	Name         string             `json:"name" firestore:"name"`
	Description  string             `json:"description" firestore:"description"`
	Coefficients map[string]float64 `json:"coefficients" firestore:"coefficients"` // スコア計算用係数
}

// CalculateScore は街の状態と思想から最終スコアを計算する
func (i *MasterIdeology) CalculateScore(cityParams *CityParams) int {
	score := 0.0
	cityMap := cityParams.ToMap()
	for param, value := range cityMap {
		if coef, ok := i.Coefficients[param]; ok {
			score += float64(value) * coef
		}
	}
	return int(score)
}

// GetDefaultIdeologies はデフォルトの思想マスターを返す
func GetDefaultIdeologies() []MasterIdeology {
	return []MasterIdeology{
		{
			IdeologyID:  "ideology_capitalist",
			Name:        "新自由主義者",
			Description: "経済成長こそが市民の幸福につながると信じる。規制緩和と市場原理を重視。",
			Coefficients: map[string]float64{
				"economy":     2.0,
				"welfare":     0.5,
				"education":   1.0,
				"environment": 0.5,
				"security":    1.0,
				"humanRights": 0.5,
			},
		},
		{
			IdeologyID:  "ideology_socialist",
			Name:        "社会民主主義者",
			Description: "全ての市民に平等な福祉を提供することが最優先。格差是正を目指す。",
			Coefficients: map[string]float64{
				"economy":     0.5,
				"welfare":     2.0,
				"education":   1.5,
				"environment": 1.0,
				"security":    0.5,
				"humanRights": 1.5,
			},
		},
		{
			IdeologyID:  "ideology_environmentalist",
			Name:        "環境保護主義者",
			Description: "持続可能な環境なくして未来はない。自然との共生を最重視。",
			Coefficients: map[string]float64{
				"economy":     0.5,
				"welfare":     1.0,
				"education":   1.0,
				"environment": 2.0,
				"security":    0.5,
				"humanRights": 1.0,
			},
		},
		{
			IdeologyID:  "ideology_authoritarian",
			Name:        "秩序重視派",
			Description: "安全な街こそが全ての基盤。強い統治による社会の安定を求める。",
			Coefficients: map[string]float64{
				"economy":     1.0,
				"welfare":     0.5,
				"education":   0.5,
				"environment": 0.5,
				"security":    2.0,
				"humanRights": 0.5,
			},
		},
		{
			IdeologyID:  "ideology_libertarian",
			Name:        "自由至上主義者",
			Description: "個人の自由と権利を何よりも尊重。政府の介入を最小限に。",
			Coefficients: map[string]float64{
				"economy":     1.5,
				"welfare":     0.5,
				"education":   1.0,
				"environment": 0.5,
				"security":    0.5,
				"humanRights": 2.0,
			},
		},
		{
			IdeologyID:  "ideology_technocrat",
			Name:        "テクノクラート",
			Description: "教育と科学技術の発展が社会を前進させる。知識こそ力。",
			Coefficients: map[string]float64{
				"economy":     1.0,
				"welfare":     1.0,
				"education":   2.0,
				"environment": 1.0,
				"security":    0.5,
				"humanRights": 1.0,
			},
		},
	}
}

// GetIdeologyByID は指定されたIDの思想を返す
func GetIdeologyByID(id string) *MasterIdeology {
	for _, ideology := range GetDefaultIdeologies() {
		if ideology.IdeologyID == id {
			return &ideology
		}
	}
	return nil
}

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
// 係数設計: 最重視 +2.0, 重視 +1.0, やや重視 +0.5, 中立 0.0, 対立 -0.5~-1.0
// 全思想の係数合計を約3.0〜3.5に統一してバランスを取る
func GetDefaultIdeologies() []MasterIdeology {
	return []MasterIdeology{
		{
			// 合計: 2.0 + 0.0 + 0.5 + (-0.5) + 1.0 + 0.5 = 3.5
			IdeologyID:  "ideology_capitalist",
			Name:        "新自由主義者",
			Description: "経済成長こそが市民の幸福につながると信じる。規制緩和と市場原理を重視。",
			Coefficients: map[string]float64{
				"economy":     2.0,  // 最重視
				"welfare":     0.0,  // 中立（大きな政府を嫌う）
				"education":   0.5,  // やや重視（人材育成）
				"environment": -0.5, // 対立（規制を嫌う）
				"security":    1.0,  // 重視（ビジネス環境の安定）
				"humanRights": 0.5,  // やや重視
			},
		},
		{
			// 合計: (-0.5) + 2.0 + 1.0 + 0.5 + (-0.5) + 1.0 = 3.5
			IdeologyID:  "ideology_socialist",
			Name:        "社会民主主義者",
			Description: "全ての市民に平等な福祉を提供することが最優先。格差是正を目指す。",
			Coefficients: map[string]float64{
				"economy":     -0.5, // 対立（経済優先を批判）
				"welfare":     2.0,  // 最重視
				"education":   1.0,  // 重視（公教育）
				"environment": 0.5,  // やや重視
				"security":    -0.5, // 対立（権力集中を警戒）
				"humanRights": 1.0,  // 重視
			},
		},
		{
			// 合計: (-1.0) + 0.5 + 1.0 + 2.0 + 0.0 + 1.0 = 3.5
			IdeologyID:  "ideology_environmentalist",
			Name:        "環境保護主義者",
			Description: "持続可能な環境なくして未来はない。自然との共生を最重視。",
			Coefficients: map[string]float64{
				"economy":     -1.0, // 対立（開発優先を批判）
				"welfare":     0.5,  // やや重視
				"education":   1.0,  // 重視（環境教育）
				"environment": 2.0,  // 最重視
				"security":    0.0,  // 中立
				"humanRights": 1.0,  // 重視
			},
		},
		{
			// 合計: 1.0 + 0.0 + 0.5 + 0.5 + 2.0 + (-0.5) = 3.5
			IdeologyID:  "ideology_authoritarian",
			Name:        "秩序重視派",
			Description: "安全な街こそが全ての基盤。強い統治による社会の安定を求める。",
			Coefficients: map[string]float64{
				"economy":     1.0,  // 重視（秩序ある経済）
				"welfare":     0.0,  // 中立
				"education":   0.5,  // やや重視
				"environment": 0.5,  // やや重視
				"security":    2.0,  // 最重視
				"humanRights": -0.5, // 対立（自由より秩序）
			},
		},
		{
			// 合計: 1.0 + (-0.5) + 0.5 + 0.5 + 0.0 + 2.0 = 3.5
			IdeologyID:  "ideology_libertarian",
			Name:        "自由至上主義者",
			Description: "個人の自由と権利を何よりも尊重。政府の介入を最小限に。",
			Coefficients: map[string]float64{
				"economy":     1.0,  // 重視（自由市場）
				"welfare":     -0.5, // 対立（政府介入を嫌う）
				"education":   0.5,  // やや重視
				"environment": 0.5,  // やや重視
				"security":    0.0,  // 中立（監視は嫌うが治安は必要）
				"humanRights": 2.0,  // 最重視
			},
		},
		{
			// 合計: 0.5 + 0.5 + 2.0 + 0.5 + 0.0 + 0.0 = 3.5
			IdeologyID:  "ideology_technocrat",
			Name:        "テクノクラート",
			Description: "教育と科学技術の発展が社会を前進させる。知識こそ力。",
			Coefficients: map[string]float64{
				"economy":     0.5, // やや重視（技術革新）
				"welfare":     0.5, // やや重視
				"education":   2.0, // 最重視
				"environment": 0.5, // やや重視（技術で解決）
				"security":    0.0, // 中立
				"humanRights": 0.0, // 中立
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

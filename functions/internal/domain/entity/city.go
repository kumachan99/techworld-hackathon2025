package entity

// CityParams は街のパラメータを表す（6項目）
// パス: rooms/{roomId} の cityParams フィールド
type CityParams struct {
	Economy     int `json:"economy" firestore:"economy"`         // 経済
	Welfare     int `json:"welfare" firestore:"welfare"`         // 福祉
	Education   int `json:"education" firestore:"education"`     // 教育
	Environment int `json:"environment" firestore:"environment"` // 環境
	Security    int `json:"security" firestore:"security"`       // 治安
	HumanRights int `json:"humanRights" firestore:"humanRights"` // 人権
}

// NewCityParams は初期状態の街パラメータを作成する（各パラメータ50）
func NewCityParams() CityParams {
	return CityParams{
		Economy:     50,
		Welfare:     50,
		Education:   50,
		Environment: 50,
		Security:    50,
		HumanRights: 50,
	}
}

// ApplyEffects は政策の効果を街に適用する
func (c *CityParams) ApplyEffects(effects map[string]int) {
	if v, ok := effects["economy"]; ok {
		c.Economy += v
	}
	if v, ok := effects["welfare"]; ok {
		c.Welfare += v
	}
	if v, ok := effects["education"]; ok {
		c.Education += v
	}
	if v, ok := effects["environment"]; ok {
		c.Environment += v
	}
	if v, ok := effects["security"]; ok {
		c.Security += v
	}
	if v, ok := effects["humanRights"]; ok {
		c.HumanRights += v
	}
}

// IsCollapsed はいずれかのパラメータが0以下になったかを判定する
func (c *CityParams) IsCollapsed() bool {
	return c.Economy <= 0 ||
		c.Welfare <= 0 ||
		c.Education <= 0 ||
		c.Environment <= 0 ||
		c.Security <= 0 ||
		c.HumanRights <= 0
}

// ToMap は CityParams を map に変換する（スコア計算用）
func (c *CityParams) ToMap() map[string]int {
	return map[string]int{
		"economy":     c.Economy,
		"welfare":     c.Welfare,
		"education":   c.Education,
		"environment": c.Environment,
		"security":    c.Security,
		"humanRights": c.HumanRights,
	}
}

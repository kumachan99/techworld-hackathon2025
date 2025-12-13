package entity

// PolicyCategory は政策カテゴリを表す
type PolicyCategory string

const (
	PolicyCategoryEconomy     PolicyCategory = "Economy"
	PolicyCategoryWelfare     PolicyCategory = "Welfare"
	PolicyCategoryEducation   PolicyCategory = "Education"
	PolicyCategoryEnvironment PolicyCategory = "Environment"
	PolicyCategorySecurity    PolicyCategory = "Security"
	PolicyCategoryHumanRights PolicyCategory = "HumanRights"
)

// MasterPolicy は政策カードマスターを表す
// パス: master_policies/{policyId}
type MasterPolicy struct {
	ID          string         `json:"id" firestore:"id"`
	Category    PolicyCategory `json:"category" firestore:"category"`
	Title       string         `json:"title" firestore:"title"`
	Description string         `json:"description" firestore:"description"`
	NewsFlash   string         `json:"newsFlash" firestore:"newsFlash"`
	Effects     map[string]int `json:"effects" firestore:"effects"` // ⚠️ クライアントに直接渡さない
}

// PolicyOption はクライアントに渡す政策情報（effects を除外）
// Room.CurrentOptions で使用
type PolicyOption struct {
	ID          string         `json:"id" firestore:"id"`
	Category    PolicyCategory `json:"category" firestore:"category"`
	Title       string         `json:"title" firestore:"title"`
	Description string         `json:"description" firestore:"description"`
}

// ToOption は MasterPolicy を PolicyOption に変換する
func (p *MasterPolicy) ToOption() PolicyOption {
	return PolicyOption{
		ID:          p.ID,
		Category:    p.Category,
		Title:       p.Title,
		Description: p.Description,
	}
}

// GetDefaultPolicies はデフォルトの政策カードマスターを返す
func GetDefaultPolicies() []MasterPolicy {
	return []MasterPolicy{
		{
			ID:          "policy_001",
			Category:    PolicyCategoryEconomy,
			Title:       "消費税廃止",
			Description: "消費者の負担を軽減し、消費を促進する大胆な経済政策",
			NewsFlash:   "【速報】消費税廃止法案が可決！商店街は歓喜に沸く一方、財政への懸念も",
			Effects:     map[string]int{"economy": 20, "welfare": -15, "education": -10, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_002",
			Category:    PolicyCategoryEnvironment,
			Title:       "再生可能エネルギー推進法",
			Description: "太陽光・風力発電への補助金を大幅に増額し、脱炭素社会を目指す",
			NewsFlash:   "【特報】再エネ推進で CO2 排出量が大幅減！しかし電気代上昇に市民から不満の声も",
			Effects:     map[string]int{"economy": -10, "welfare": 0, "education": 5, "environment": 25, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_003",
			Category:    PolicyCategorySecurity,
			Title:       "防犯カメラ設置義務化",
			Description: "全ての公共スペースに監視カメラを設置し、犯罪抑止を図る",
			NewsFlash:   "【速報】犯罪発生率が激減！一方でプライバシー侵害を訴える市民団体がデモ",
			Effects:     map[string]int{"economy": -5, "welfare": 0, "education": 0, "environment": 0, "security": 20, "humanRights": -15},
		},
		{
			ID:          "policy_004",
			Category:    PolicyCategoryWelfare,
			Title:       "ベーシックインカム導入",
			Description: "全市民に毎月一定額を支給し、最低限の生活を保障する",
			NewsFlash:   "【歴史的決定】BI 開始で貧困率が急低下！財源確保のため増税議論も",
			Effects:     map[string]int{"economy": -20, "welfare": 25, "education": 0, "environment": 0, "security": 0, "humanRights": 10},
		},
		{
			ID:          "policy_005",
			Category:    PolicyCategoryEducation,
			Title:       "教育無償化",
			Description: "幼稚園から大学まで、全ての教育費用を無償化する",
			NewsFlash:   "【朗報】教育無償化で進学率過去最高に！予算超過で他施策に影響も",
			Effects:     map[string]int{"economy": -15, "welfare": 10, "education": 30, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_006",
			Category:    PolicyCategoryEconomy,
			Title:       "大型ショッピングモール誘致",
			Description: "郊外に大型商業施設を誘致し、雇用と消費を創出する",
			NewsFlash:   "【経済】巨大モール開業で雇用 5000 人創出！周辺の自然破壊に環境団体が抗議",
			Effects:     map[string]int{"economy": 25, "welfare": 0, "education": 0, "environment": -20, "security": -5, "humanRights": 0},
		},
		{
			ID:          "policy_007",
			Category:    PolicyCategoryEnvironment,
			Title:       "公園緑地化プロジェクト",
			Description: "市内の空き地を公園に整備し、緑豊かな街づくりを推進",
			NewsFlash:   "【環境】緑地面積 30% 増！市民の満足度向上も、維持費が財政を圧迫",
			Effects:     map[string]int{"economy": -10, "welfare": 10, "education": 0, "environment": 20, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_008",
			Category:    PolicyCategorySecurity,
			Title:       "警察官増員計画",
			Description: "警察官を大幅に増員し、パトロールを強化する",
			NewsFlash:   "【治安】パトロール強化で体感治安が向上！過剰取り締まりへの批判も",
			Effects:     map[string]int{"economy": -15, "welfare": 0, "education": 0, "environment": 0, "security": 25, "humanRights": -5},
		},
		{
			ID:          "policy_009",
			Category:    PolicyCategoryEconomy,
			Title:       "IT企業優遇税制",
			Description: "IT企業への減税措置により、ハイテク産業の集積を目指す",
			NewsFlash:   "【経済】IT 特区誕生でスタートアップ続々！データセンター増設で電力消費に懸念",
			Effects:     map[string]int{"economy": 20, "welfare": 0, "education": 10, "environment": -10, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_010",
			Category:    PolicyCategoryWelfare,
			Title:       "高齢者医療費補助拡大",
			Description: "高齢者の医療費自己負担を軽減し、安心できる老後を実現",
			NewsFlash:   "【福祉】高齢者の受診率向上で健康寿命延伸！現役世代の負担増に反発も",
			Effects:     map[string]int{"economy": -15, "welfare": 20, "education": 0, "environment": 0, "security": 0, "humanRights": 5},
		},
		{
			ID:          "policy_011",
			Category:    PolicyCategoryEnvironment,
			Title:       "自然保護区域拡大",
			Description: "開発制限区域を拡大し、生態系の保全を強化する",
			NewsFlash:   "【環境】希少種の生息確認相次ぐ！開発業者からは反発の声",
			Effects:     map[string]int{"economy": -20, "welfare": 0, "education": 0, "environment": 25, "security": 0, "humanRights": 5},
		},
		{
			ID:          "policy_012",
			Category:    PolicyCategorySecurity,
			Title:       "夜間外出規制条例",
			Description: "深夜帯の外出を届出制にし、犯罪発生を抑制する",
			NewsFlash:   "【治安】夜間犯罪が激減！「自由の侵害」として違憲訴訟の動きも",
			Effects:     map[string]int{"economy": 0, "welfare": -5, "education": 0, "environment": 0, "security": 15, "humanRights": -25},
		},
		{
			ID:          "policy_013",
			Category:    PolicyCategoryEconomy,
			Title:       "起業支援ファンド設立",
			Description: "スタートアップへの投資を促進し、イノベーションを加速",
			NewsFlash:   "【経済】ユニコーン企業が誕生！一方で支援を受けられない中小企業から不満",
			Effects:     map[string]int{"economy": 20, "welfare": -10, "education": 5, "environment": 0, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_014",
			Category:    PolicyCategoryEnvironment,
			Title:       "市民農園整備事業",
			Description: "市民が気軽に農業体験できる農園を各地に整備する",
			NewsFlash:   "【暮らし】市民農園が大人気！食育効果も期待、予約は半年待ちに",
			Effects:     map[string]int{"economy": 0, "welfare": 10, "education": 5, "environment": 15, "security": 0, "humanRights": 0},
		},
		{
			ID:          "policy_015",
			Category:    PolicyCategoryHumanRights,
			Title:       "情報公開条例強化",
			Description: "行政の透明性を高め、市民の知る権利を保障する",
			NewsFlash:   "【政治】情報公開で行政の不正が次々発覚！捜査情報漏洩の懸念も",
			Effects:     map[string]int{"economy": 0, "welfare": 5, "education": 0, "environment": 0, "security": -10, "humanRights": 20},
		},
	}
}

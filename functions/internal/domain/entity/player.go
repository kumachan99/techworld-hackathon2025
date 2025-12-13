package entity

// Player はプレイヤーを表す
// パス: rooms/{roomId}/players/{userId}
//
// ⚠️ ideology, currentVote は Security Rules で本人以外読み取り禁止
// 投票状態は Room.Votes の keys で判断可能
type Player struct {
	// 🌐 公開情報
	DisplayName    string `json:"displayName" firestore:"displayName"`
	IsHost         bool   `json:"isHost" firestore:"isHost"`
	IsReady        bool   `json:"isReady" firestore:"isReady"`
	IsPetitionUsed bool   `json:"isPetitionUsed" firestore:"isPetitionUsed"`

	// 🔒 秘匿情報（本人のみ読み取り可）
	Ideology    *MasterIdeology `json:"ideology" firestore:"ideology"`
	CurrentVote string          `json:"currentVote" firestore:"currentVote"`
}

// NewPlayer は新しいプレイヤーを作成する
func NewPlayer(displayName string, isHost bool, ideology *MasterIdeology) *Player {
	return &Player{
		DisplayName:    displayName,
		IsHost:         isHost,
		IsReady:        false,
		IsPetitionUsed: false,
		Ideology:       ideology,
		CurrentVote:    "",
	}
}

// Vote は投票を行う
func (p *Player) Vote(policyID string) {
	p.CurrentVote = policyID
}

// ClearVote は投票をクリアする（次のターン用）
func (p *Player) ClearVote() {
	p.CurrentVote = ""
}

// CalculateScore はスコアを計算する
func (p *Player) CalculateScore(cityParams *CityParams) int {
	if p.Ideology == nil {
		return 0
	}
	return p.Ideology.CalculateScore(cityParams)
}

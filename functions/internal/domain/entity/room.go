package entity

import (
	"math/rand"
	"time"
)

// RoomStatus は部屋の状態を表す
type RoomStatus string

const (
	RoomStatusLobby    RoomStatus = "LOBBY"    // プレイヤー待機中
	RoomStatusVoting   RoomStatus = "VOTING"   // 投票フェーズ
	RoomStatusResult   RoomStatus = "RESULT"   // 結果発表フェーズ
	RoomStatusFinished RoomStatus = "FINISHED" // ゲーム終了
)

// Room はゲームルームを表す
// パス: rooms/{roomId}
type Room struct {
	HostID            string                   `json:"hostId" firestore:"hostId"`
	Status            RoomStatus               `json:"status" firestore:"status"`
	Turn              int                      `json:"turn" firestore:"turn"`
	MaxTurns          int                      `json:"maxTurns" firestore:"maxTurns"`
	CreatedAt         time.Time                `json:"createdAt" firestore:"createdAt"`
	CityParams        CityParams               `json:"cityParams" firestore:"cityParams"`
	IsCollapsed       bool                     `json:"isCollapsed" firestore:"isCollapsed"`
	CurrentPolicyIDs  []string                 `json:"currentPolicyIds" firestore:"currentPolicyIds"` // IDのみ
	DeckIDs           []string                 `json:"deckIds" firestore:"deckIds"`                   // 山札
	PassedPolicyIDs   []string                 `json:"passedPolicyIds" firestore:"passedPolicyIds"`   // 可決された政策の履歴
	Votes             map[string]string        `json:"votes" firestore:"votes"`                       // { userId: policyId }
	LastResult        *VoteResult              `json:"lastResult" firestore:"lastResult"`
	GeneratedPolicies map[string]*MasterPolicy `json:"generatedPolicies" firestore:"generatedPolicies"` // AI陳情で生成された政策
}

// VoteResult は投票結果を表す（RESULT フェーズで使用）
type VoteResult struct {
	PassedPolicyID    string            `json:"passedPolicyId" firestore:"passedPolicyId"`
	PassedPolicyTitle string            `json:"passedPolicyTitle" firestore:"passedPolicyTitle"`
	ActualEffects     map[string]int    `json:"actualEffects" firestore:"actualEffects"`
	NewsFlash         string            `json:"newsFlash" firestore:"newsFlash"`
	VoteDetails       map[string]string `json:"voteDetails" firestore:"voteDetails"`
	CityImage         string            `json:"cityImage,omitempty" firestore:"-"` // Base64エンコードされた街の画像（Firestoreには保存しない）
}

// NewRoom は新しい部屋を作成する
func NewRoom(hostID string) *Room {
	return &Room{
		HostID:            hostID,
		Status:            RoomStatusLobby,
		Turn:              0,
		MaxTurns:          10,
		CreatedAt:         time.Now(),
		CityParams:        NewCityParams(),
		IsCollapsed:       false,
		CurrentPolicyIDs:  make([]string, 0),
		DeckIDs:           make([]string, 0),
		PassedPolicyIDs:   make([]string, 0),
		Votes:             make(map[string]string),
		LastResult:        nil,
		GeneratedPolicies: make(map[string]*MasterPolicy),
	}
}

// CanStart はゲームを開始できるかどうかを判定する
func (r *Room) CanStart(playerCount int) bool {
	return playerCount >= 2 && r.Status == RoomStatusLobby
}

// Start はゲームを開始する
func (r *Room) Start() {
	r.Status = RoomStatusVoting
	r.Turn = 1
}

// IsGameOver はゲーム終了条件を満たしているかを判定する
// turn >= maxTurns（最終ターン完了後）または街が崩壊した場合に終了
func (r *Room) IsGameOver() bool {
	return r.Turn >= r.MaxTurns || r.IsCollapsed
}

// Finish はゲームを終了する
func (r *Room) Finish() {
	r.Status = RoomStatusFinished
}

// ApplyPolicyEffects は政策の効果を適用する
func (r *Room) ApplyPolicyEffects(effects map[string]int) {
	r.CityParams.ApplyEffects(effects)
	r.IsCollapsed = r.CityParams.IsCollapsed()
}

// NextTurn は次のターンに進める
func (r *Room) NextTurn() {
	r.Turn++
	r.Status = RoomStatusVoting
	r.CurrentPolicyIDs = make([]string, 0)
	r.Votes = make(map[string]string) // 投票リセット
	r.LastResult = nil
}

// AllPlayersVoted は全プレイヤーが投票したかを判定する
func (r *Room) AllPlayersVoted(playerCount int) bool {
	votedCount := 0
	for _, vote := range r.Votes {
		if vote != "" {
			votedCount++
		}
	}
	return votedCount >= playerCount
}

// GetGeneratedPolicy はAI生成の政策を取得する
func (r *Room) GetGeneratedPolicy(policyID string) *MasterPolicy {
	if r.GeneratedPolicies == nil {
		return nil
	}
	return r.GeneratedPolicies[policyID]
}

// AddGeneratedPolicy はAI生成の政策を追加する
func (r *Room) AddGeneratedPolicy(policy *MasterPolicy) string {
	if r.GeneratedPolicies == nil {
		r.GeneratedPolicies = make(map[string]*MasterPolicy)
	}
	// ユニークなIDを生成（generated_から始まる）
	policyID := "generated_" + randString(8)
	policy.PolicyID = policyID
	r.GeneratedPolicies[policyID] = policy
	return policyID
}

// randString はランダムな文字列を生成する
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// CountVotes は投票を集計し、最多得票の政策IDを返す
// 同数の場合はランダムに選択
func (r *Room) CountVotes() string {
	voteCount := make(map[string]int)
	for _, policyID := range r.Votes {
		if policyID != "" {
			voteCount[policyID]++
		}
	}

	// 最多得票数を求める
	maxVotes := 0
	for _, count := range voteCount {
		if count > maxVotes {
			maxVotes = count
		}
	}

	// 最多得票の政策を全て集める
	var candidates []string
	for policyID, count := range voteCount {
		if count == maxVotes {
			candidates = append(candidates, policyID)
		}
	}

	// 候補がない場合は空文字を返す
	if len(candidates) == 0 {
		return ""
	}

	// 同数の場合はランダムに選択
	if len(candidates) == 1 {
		return candidates[0]
	}

	return candidates[rand.Intn(len(candidates))]
}

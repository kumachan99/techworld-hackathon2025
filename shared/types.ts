/**
 * Firestore データモデル型定義
 *
 * フロントエンド・バックエンド間で共有される型定義です。
 * 設計書: docs/firestore-schema.md
 */

import { Timestamp } from 'firebase/firestore';

// =============================================================================
// 共通型
// =============================================================================

/** 街のパラメータ（6項目） */
export interface CityParams {
  economy: number;      // 経済
  welfare: number;      // 福祉
  education: number;    // 教育
  environment: number;  // 環境
  security: number;     // 治安
  humanRights: number;  // 人権
}

/** 政策の効果（CityParams と同じ構造） */
export type PolicyEffects = CityParams;

/** スコア計算用係数（CityParams と同じ構造） */
export type IdeologyCoefficients = CityParams;

// =============================================================================
// master_policies コレクション
// =============================================================================

/**
 * 政策カードマスター
 * パス: master_policies/{policyId}
 *
 * 各政策は effects で6パラメータ全てに影響を与えます。
 * policyId はドキュメントIDと同一
 */
export interface MasterPolicy {
  policyId: string;
  title: string;
  description: string;
  newsFlash: string;
  effects: PolicyEffects;  // ⚠️ 結果発表まで非公開
}

// =============================================================================
// master_ideologies コレクション
// =============================================================================

/**
 * 思想マスター
 * パス: master_ideologies/{ideologyId}
 * ideologyId はドキュメントIDと同一
 */
export interface MasterIdeology {
  ideologyId: string;
  name: string;
  description: string;
  coefficients: IdeologyCoefficients;
}

// =============================================================================
// rooms コレクション
// =============================================================================

/** ゲームステータス */
export type RoomStatus = 'LOBBY' | 'VOTING' | 'RESULT' | 'FINISHED';

/**
 * ゲームルーム
 * パス: rooms/{roomId}
 */
export interface Room {
  hostId: string;
  status: RoomStatus;
  turn: number;
  maxTurns: number;
  createdAt: Timestamp;
  cityParams: CityParams;
  isCollapsed: boolean;
  currentPolicyIds: string[];           // ★ IDのみ。マスターから引いて表示
  deckIds: string[];                    // 山札
  passedPolicyIds: string[];            // 可決された政策の履歴
  votes: Record<string, string | null>; // { userId: policyId | null }
  lastResult: VoteResult | null;
}

/** 投票結果（RESULT フェーズで設定） */
export interface VoteResult {
  passedPolicyId: string;
  passedPolicyTitle: string;
  actualEffects: PolicyEffects;  // ここで効果を開示
  newsFlash: string;
  voteDetails: Record<string, string>;  // { userId: policyId }
}

// =============================================================================
// rooms/{roomId}/players サブコレクション
// =============================================================================

/**
 * プレイヤー
 * パス: rooms/{roomId}/players/{userId}
 *
 * ⚠️ ideology, currentVote は Security Rules で本人以外読み取り禁止
 * 投票済みかは Room.votes の keys を監視して判断
 */
export interface Player {
  // 🌐 公開情報
  displayName: string;
  isHost: boolean;
  isReady: boolean;
  isPetitionUsed: boolean;

  // 🔒 秘匿情報（本人のみ読み取り可）
  ideology: MasterIdeology;      // 割り振られた思想
  currentVote: string | null;    // 投票先の政策ID
}

/** プレイヤー公開情報（他プレイヤーが見れる部分） */
export interface PlayerPublic {
  displayName: string;
  isHost: boolean;
  isReady: boolean;
  isPetitionUsed: boolean;
}

// =============================================================================
// Cloud Run API リクエスト/レスポンス型
// =============================================================================

// -----------------------------------------------------------------------------
// POST /api/rooms - 部屋作成
// -----------------------------------------------------------------------------

/** 部屋作成リクエスト */
export interface CreateRoomRequest {
  displayName: string;
}

/** 部屋作成レスポンス */
export interface CreateRoomResponse {
  roomId: string;
  status: RoomStatus;
  playerId: string;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/join - 部屋参加
// -----------------------------------------------------------------------------

/** 部屋参加リクエスト */
export interface JoinRoomRequest {
  displayName: string;
}

/** 部屋参加レスポンス */
export interface JoinRoomResponse {
  playerId: string;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/leave - 部屋退出
// -----------------------------------------------------------------------------

/** 部屋退出リクエスト */
export interface LeaveRoomRequest {
  playerId: string;
}

/** 部屋退出レスポンス */
export interface LeaveRoomResponse {
  success: boolean;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/ready - Ready状態トグル
// -----------------------------------------------------------------------------

/** Ready状態リクエスト */
export interface ReadyRequest {
  playerId: string;
}

/** Ready状態レスポンス */
export interface ReadyResponse {
  isReady: boolean;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/start - ゲーム開始
// -----------------------------------------------------------------------------

/** ゲーム開始リクエスト */
export interface StartGameRequest {
  playerId: string;
}

/** ゲーム開始レスポンス */
export interface StartGameResponse {
  status: RoomStatus;
  turn: number;
  currentPolicyIds: string[];
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/vote - 投票
// -----------------------------------------------------------------------------

/** 投票リクエスト */
export interface VoteRequest {
  playerId: string;
  policyId: string;
}

/** 投票レスポンス */
export interface VoteResponse {
  success: boolean;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/resolve - 投票集計
// -----------------------------------------------------------------------------

/** 投票集計レスポンス */
export interface ResolveVoteResponse {
  status: RoomStatus;
  lastResult: VoteResult;
  cityParams: CityParams;
  isGameOver: boolean;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/next - 次ターンへ
// -----------------------------------------------------------------------------

/** 次ターンレスポンス */
export interface NextTurnResponse {
  status: RoomStatus;
  turn: number;
}

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/petition - AI陳情
// -----------------------------------------------------------------------------

/** 陳情リクエスト */
export interface PetitionRequest {
  playerId: string;
  text: string;
}

/** 陳情レスポンス */
export interface PetitionResponse {
  approved: boolean;
  policyId?: string;   // 承認時のみ
  message: string;
}

// -----------------------------------------------------------------------------
// 共通エラーレスポンス
// -----------------------------------------------------------------------------

/** APIエラーレスポンス */
export interface ApiErrorResponse {
  error: string;
}

// =============================================================================
// スコア計算用の型
// =============================================================================

/** プレイヤースコア（ゲーム終了後に表示） */
export interface PlayerScore {
  userId: string;
  displayName: string;
  ideology: MasterIdeology;  // ゲーム終了後に公開
  score: number;
  rank: number;
}

/** スコア計算結果 */
export interface ScoreResult {
  scores: PlayerScore[];
  isCollapsed: boolean;
  finalCityParams: CityParams;
}

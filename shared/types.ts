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

/** 政策カテゴリ */
export type PolicyCategory =
  | 'Economy'
  | 'Welfare'
  | 'Education'
  | 'Environment'
  | 'Security'
  | 'HumanRights';

/**
 * 政策カードマスター
 * パス: master_policies/{policyId}
 */
export interface MasterPolicy {
  id: string;
  category: PolicyCategory;
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
 */
export interface MasterIdeology {
  id: string;
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
  votes: Record<string, string | null>; // { oderId: policyId | null }
  lastResult: VoteResult | null;
}

/** 投票結果（RESULT フェーズで設定） */
export interface VoteResult {
  passedPolicyId: string;
  passedPolicyTitle: string;
  actualEffects: PolicyEffects;  // ここで効果を開示
  newsFlash: string;
  voteDetails: Record<string, string>;  // { oderId: policyId }
}

// =============================================================================
// rooms/{roomId}/players サブコレクション
// =============================================================================

/**
 * プレイヤー
 * パス: rooms/{roomId}/players/{oderId}
 *
 * ⚠️ ideology, currentVote は Security Rules で本人以外読み取り禁止
 */
export interface Player {
  // 🌐 公開情報
  displayName: string;
  photoURL: string;
  isHost: boolean;
  isReady: boolean;
  hasVoted: boolean;
  isPetitionUsed: boolean;

  // 🔒 秘匿情報（本人のみ読み取り可）
  ideology: MasterIdeology;      // 割り振られた思想
  currentVote: string | null;    // 投票先の政策ID
}

/** プレイヤー公開情報（他プレイヤーが見れる部分） */
export interface PlayerPublic {
  displayName: string;
  photoURL: string;
  isHost: boolean;
  isReady: boolean;
  hasVoted: boolean;
  isPetitionUsed: boolean;
}

// =============================================================================
// Cloud Run API リクエスト/レスポンス型
// =============================================================================

// -----------------------------------------------------------------------------
// POST /api/rooms/{roomId}/start - ゲーム開始
// -----------------------------------------------------------------------------

/** ゲーム開始レスポンス */
export interface StartGameResponse {
  status: RoomStatus;
  turn: number;
  currentPolicyIds: string[];
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
// POST /api/rooms/{roomId}/petitions - AI陳情
// -----------------------------------------------------------------------------

/** 陳情リクエスト */
export interface PetitionRequest {
  text: string;
}

/** 陳情レスポンス */
export interface PetitionResponse {
  approved: boolean;
  policyId?: string;   // 承認時のみ
  message: string;
}

// =============================================================================
// Firestore 直接操作用の型（フロントエンド参照用）
// =============================================================================

/** 部屋作成時の初期データ */
export interface CreateRoomData {
  hostId: string;
  status: 'LOBBY';
  turn: 0;
  maxTurns: 10;
  createdAt: unknown;  // serverTimestamp()
  cityParams: CityParams;
  isCollapsed: false;
  currentPolicyIds: [];
  deckIds: [];
  votes: Record<string, null>;
  lastResult: null;
}

/** プレイヤー作成時の初期データ */
export interface CreatePlayerData {
  displayName: string;
  photoURL: string;
  isHost: boolean;
  isReady: false;
  hasVoted: false;
  isPetitionUsed: false;
  ideology: MasterIdeology;
  currentVote: '';
}

// =============================================================================
// スコア計算用の型
// =============================================================================

/** プレイヤースコア（ゲーム終了後に表示） */
export interface PlayerScore {
  oderId: string;
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

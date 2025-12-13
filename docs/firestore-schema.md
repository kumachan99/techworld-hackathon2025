# Firestore データモデル設計書

フロントエンド・バックエンド間でデータ構造の認識を合わせるための設計書です。

## アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                      │
└───────────────┬─────────────────────────┬───────────────────┘
                │                         │
                ▼                         ▼
┌───────────────────────────┐   ┌─────────────────────────────┐
│       Firestore 直接       │   │     Cloud Run (Go API)      │
│                           │   │                             │
│ • ルーム監視 (onSnapshot) │   │ • POST /start  (ゲーム開始)│
│ • プレイヤー監視          │   │ • POST /resolve(投票集計)  │
│ • マスターデータ取得      │   │ • POST /petition(AI陳情)   │
│ • 部屋作成/参加           │   │                             │
│ • 投票                    │   │                             │
│ • Ready状態更新           │   │                             │
└───────────────────────────┘   └─────────────────────────────┘
```

---

## コレクション構造

```
ROOT
├── 📁 master_policies      # 政策カードのマスターデータ
├── 📁 master_ideologies    # 思想のマスターデータ
└── 📁 rooms                # ゲームルーム
    └── 📁 players          # 参加者（サブコレクション）
```

---

## 1. master_policies（政策カードマスター）

**パス:** `master_policies/{policyId}`

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | 政策ID |
| category | string | `"Economy"` / `"Welfare"` / `"Education"` / `"Environment"` / `"Security"` / `"HumanRights"` |
| title | string | タイトル |
| description | string | 説明文 |
| newsFlash | string | 結果発表時のニュース |
| effects | map | 効果値 ⚠️**結果発表まで非公開** |

---

## 2. master_ideologies（思想マスター）

**パス:** `master_ideologies/{ideologyId}`

| フィールド | 型 | 説明 |
|-----------|-----|------|
| id | string | 思想ID |
| name | string | 思想名 |
| description | string | 説明 |
| coefficients | map | スコア計算用係数 |

---

## 3. rooms（ゲームルーム）

**パス:** `rooms/{roomId}`

| フィールド | 型 | 説明 |
|-----------|-----|------|
| hostId | string | ホストのUID |
| status | string | `"LOBBY"` / `"VOTING"` / `"RESULT"` / `"FINISHED"` |
| turn | number | 現在のターン数（1〜10） |
| maxTurns | number | 最大ターン数（10） |
| createdAt | timestamp | 作成日時 |
| cityParams | map | 街のパラメータ |
| isCollapsed | boolean | 街崩壊フラグ |
| currentPolicyIds | array | 提示中の政策ID（3つ） |
| deckIds | array | 山札（残りの政策ID） |
| votes | map | 投票状況 `{ oderId: policyId }` |
| lastResult | map / null | 前回の結果（RESULT時のみ） |

---

## 4. players（参加者）- サブコレクション

**パス:** `rooms/{roomId}/players/{oderId}`

| フィールド | 型 | アクセス | 説明 |
|-----------|-----|---------|------|
| displayName | string | 🌐 公開 | 表示名 |
| photoURL | string | 🌐 公開 | アイコンURL |
| isHost | boolean | 🌐 公開 | ホストか |
| isReady | boolean | 🌐 公開 | 準備完了か |
| hasVoted | boolean | 🌐 公開 | 投票済みか |
| isPetitionUsed | boolean | 🌐 公開 | 陳情権使用済みか |
| ideology | map | 🔒 本人のみ | 割り振られた思想 |
| currentVote | string | 🔒 本人のみ | 投票先の政策ID |

---

## ステータス遷移

```
LOBBY → VOTING → RESULT → VOTING → ... → FINISHED
```

| ステータス | 説明 | 次へ進む条件 |
|-----------|------|-------------|
| LOBBY | 待機中 | 2人以上 & 全員 isReady → **API: /start** |
| VOTING | 投票中 | 全員投票完了 → **API: /resolve** |
| RESULT | 結果発表 | ホストが次ターン開始（Firestore直接） |
| FINISHED | 終了 | - |

---

## フロントエンド実装パターン

### Firestore 直接操作

```typescript
// ========================================
// 部屋作成
// ========================================
const roomRef = doc(collection(db, 'rooms'));
const roomId = roomRef.id;

await setDoc(roomRef, {
  hostId: oderId,
  status: 'LOBBY',
  turn: 0,
  maxTurns: 10,
  createdAt: serverTimestamp(),
  cityParams: { economy: 50, welfare: 50, education: 50, environment: 50, security: 50, humanRights: 50 },
  isCollapsed: false,
  currentPolicyIds: [],
  deckIds: [],
  votes: {},
  lastResult: null,
});

// ========================================
// 部屋参加（プレイヤー作成）
// ========================================
// 思想をランダムに選ぶ（未使用のものから）
const ideology = await getRandomIdeology(roomId);

await setDoc(doc(db, 'rooms', roomId, 'players', oderId), {
  displayName: 'Alice',
  photoURL: '',
  isHost: false,
  isReady: false,
  hasVoted: false,
  isPetitionUsed: false,
  ideology: ideology,      // 🔒 本人のみ
  currentVote: '',         // 🔒 本人のみ
});

// votes に自分を追加
await updateDoc(doc(db, 'rooms', roomId), {
  [`votes.${oderId}`]: null
});

// ========================================
// Ready 状態更新
// ========================================
await updateDoc(doc(db, 'rooms', roomId, 'players', oderId), {
  isReady: true
});

// ========================================
// 投票
// ========================================
await updateDoc(doc(db, 'rooms', roomId, 'players', oderId), {
  currentVote: policyId,
  hasVoted: true
});
await updateDoc(doc(db, 'rooms', roomId), {
  [`votes.${oderId}`]: policyId
});

// ========================================
// 次ターンへ（ホストのみ、RESULTフェーズで）
// ========================================
await updateDoc(doc(db, 'rooms', roomId), {
  status: 'VOTING',
  turn: increment(1),  // ターン数をインクリメント
  // currentPolicyIds, votes, hasVoted は /resolve API で既に設定済み
});

// ========================================
// リアルタイム監視
// ========================================
// ルーム
onSnapshot(doc(db, 'rooms', roomId), (doc) => {
  const room = doc.data();
});

// プレイヤー一覧
onSnapshot(collection(db, 'rooms', roomId, 'players'), (snapshot) => {
  const players = snapshot.docs.map(d => ({ id: d.id, ...d.data() }));
});
```

---

## Cloud Run API（3つのみ）

### POST `/api/rooms/{roomId}/start`

ゲーム開始。山札作成、最初の3枚選出。

**リクエスト:** なし（roomId は URL から）

**処理:**
1. 全政策IDを取得してシャッフル → `deckIds`
2. 先頭3枚を `currentPolicyIds` に
3. `deckIds` から3枚を削除
4. `votes` を初期化（全プレイヤーID → `null`）
5. `status` を `VOTING` に、`turn` を `1` に
6. 全プレイヤーの `hasVoted` を `false` に

---

### POST `/api/rooms/{roomId}/resolve`

投票集計・結果反映。

**リクエスト:** なし

**処理:**
1. `votes` を集計して最多得票の政策を決定（同数の場合はランダム）
2. `master_policies` から `effects` を取得
3. `cityParams` に効果を適用
4. `isCollapsed` をチェック（いずれかのパラメータが 0 以下 or 100 以上）
5. `lastResult` を設定（`passedPolicyId`, `actualEffects`, `newsFlash`, `voteDetails`）
6. **次のターンの準備:**
   - `deckIds` から3枚を `currentPolicyIds` に移動
   - `votes` をリセット（全プレイヤーID → `""`）
   - 全プレイヤーの `hasVoted` を `false` に、`currentVote` を `""` に
7. `status` を `RESULT` に
8. ゲーム終了判定:
   - `turn >= maxTurns` または `isCollapsed == true` → `status` を `FINISHED` に

---

### POST `/api/rooms/{roomId}/petitions`

AI陳情。

**リクエスト:**
```json
{
  "text": "週休3日制を導入したい"
}
```

**処理:**
1. プレイヤーの `isPetitionUsed` を確認（使用済みならエラー）
2. OpenAI API で審査
3. 承認なら政策カードを `master_policies` に追加し、IDを `deckIds` に追加
4. プレイヤーの `isPetitionUsed` を `true` に

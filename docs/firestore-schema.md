# Firestore データモデル設計書

フロントエンド・バックエンド間でデータ構造の認識を合わせるための設計書です。

## アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Next.js)                      │
│                                                              │
│  • API呼び出し（全ての更新操作）                             │
│  • Firestore リアルタイム監視（読み取りのみ）                │
└───────────────────────────────┬──────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Cloud Run (Go API)                       │
│                                                              │
│  【部屋管理】                                                │
│  • POST /rooms              - 部屋作成                       │
│  • POST /rooms/:id/join     - 部屋参加                       │
│  • POST /rooms/:id/leave    - 部屋退出                       │
│                                                              │
│  【ゲーム進行】                                              │
│  • POST /rooms/:id/ready    - Ready状態トグル                │
│  • POST /rooms/:id/start    - ゲーム開始                     │
│  • POST /rooms/:id/vote     - 投票                           │
│  • POST /rooms/:id/resolve  - 投票集計                       │
│  • POST /rooms/:id/next     - 次ターンへ                     │
│  • POST /rooms/:id/petition - AI陳情                         │
└───────────────────────────────┬──────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────┐
│                        Firestore                             │
│                                                              │
│  ⚠️ フロントエンドからの直接更新は禁止                       │
│  ✅ 読み取り・監視のみ許可                                   │
└─────────────────────────────────────────────────────────────┘
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
| title | string | タイトル |
| description | string | 説明文 |
| newsFlash | string | 結果発表時のニュース |
| effects | map | 効果値（6パラメータ全てに影響）⚠️**結果発表まで非公開** |

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
| isPetitionUsed | boolean | 🌐 公開 | 陳情権使用済みか |
| ideology | map | 🔒 本人のみ | 割り振られた思想 |
| currentVote | string | 🔒 本人のみ | 投票先の政策ID |

> **Note:** 投票済みかどうかは `Room.votes` の keys を監視することで判断できます。

---

## ステータス遷移

```
LOBBY → VOTING → RESULT → VOTING → ... → FINISHED
```

| ステータス | 説明 | 次へ進む条件 |
|-----------|------|-------------|
| LOBBY | 待機中 | 2人以上 & 全員 isReady → `POST /start` |
| VOTING | 投票中 | 全員投票完了 → `POST /resolve` |
| RESULT | 結果発表 | ホストが `POST /next` |
| FINISHED | 終了 | - |

---

## Cloud Run API 仕様

### 共通仕様

- **ベースURL:** `/api`
- **認証:** Firebase Authentication（Bearer Token）
- **エラーレスポンス:**
  ```json
  {
    "error": "エラーメッセージ"
  }
  ```

---

### 部屋管理

#### POST `/api/rooms` - 部屋作成

新しいゲームルームを作成し、ホストとして参加する。

**リクエスト:**
```json
{
  "displayName": "プレイヤー名",
  "photoURL": "https://..."  // optional
}
```

**処理:**
1. 新しい roomId を生成
2. Room ドキュメントを作成（初期値設定）
3. ホストを players サブコレクションに追加
4. 思想をランダムに割り当て

**レスポンス:**
```json
{
  "roomId": "abc123",
  "status": "LOBBY"
}
```

---

#### POST `/api/rooms/{roomId}/join` - 部屋参加

既存のルームに参加する。

**リクエスト:**
```json
{
  "displayName": "プレイヤー名",
  "photoURL": "https://..."  // optional
}
```

**処理:**
1. ルームの存在・状態確認（LOBBY のみ参加可）
2. 既に参加済みでないか確認
3. 未使用の思想からランダムに割り当て
4. プレイヤーを追加
5. votes に追加

**レスポンス:**
```json
{
  "success": true
}
```

**エラー:**
- `404`: ルームが存在しない
- `400`: ゲームが既に開始している
- `400`: 既に参加済み
- `400`: 思想が足りない（最大6人）

---

#### POST `/api/rooms/{roomId}/leave` - 部屋退出

ルームから退出する。

**リクエスト:** なし

**処理:**
1. プレイヤーを削除
2. votes から削除
3. ホストが退出した場合、別のプレイヤーをホストに昇格（または部屋を削除）

**レスポンス:**
```json
{
  "success": true
}
```

---

### ゲーム進行

#### POST `/api/rooms/{roomId}/ready` - Ready状態トグル

準備完了状態を切り替える。

**リクエスト:** なし

**処理:**
1. LOBBY 状態であることを確認
2. `isReady` をトグル

**レスポンス:**
```json
{
  "isReady": true
}
```

---

#### POST `/api/rooms/{roomId}/start` - ゲーム開始

ゲームを開始する（ホストのみ）。

**リクエスト:** なし

**処理:**
1. リクエスト者がホストであることを確認
2. LOBBY 状態であることを確認
3. 2人以上 & 全員 Ready であることを確認
4. 全政策IDを取得してシャッフル → `deckIds`
5. 先頭3枚を `currentPolicyIds` に
6. `status` を `VOTING` に、`turn` を `1` に

**レスポンス:**
```json
{
  "status": "VOTING",
  "turn": 1,
  "currentPolicyIds": ["policy_001", "policy_005", "policy_012"]
}
```

**エラー:**
- `403`: ホストではない
- `400`: 条件を満たしていない

---

#### POST `/api/rooms/{roomId}/vote` - 投票

政策に投票する。

**リクエスト:**
```json
{
  "policyId": "policy_001"
}
```

**処理:**
1. VOTING 状態であることを確認
2. 有効な政策IDであることを確認（currentPolicyIds に含まれる）
3. プレイヤーの `currentVote` を更新
4. Room の `votes` を更新

**レスポンス:**
```json
{
  "success": true
}
```

---

#### POST `/api/rooms/{roomId}/resolve` - 投票集計

投票を集計し結果を反映する（ホストのみ）。

**リクエスト:** なし

**処理:**
1. リクエスト者がホストであることを確認
2. VOTING 状態であることを確認
3. 全員が投票済みであることを確認
4. `votes` を集計して最多得票の政策を決定（同数はランダム）
5. `master_policies` から `effects` を取得
6. `cityParams` に効果を適用
7. `isCollapsed` をチェック
8. `lastResult` を設定
9. 次のターンの準備:
   - `deckIds` から3枚を `currentPolicyIds` に移動
   - `votes` をリセット
   - 全プレイヤーの `currentVote` を `""` に
10. `status` を `RESULT` に
11. ゲーム終了判定: `turn >= maxTurns` or `isCollapsed` → `FINISHED`

**レスポンス:**
```json
{
  "status": "RESULT",
  "isGameOver": false,
  "lastResult": {
    "passedPolicyId": "policy_001",
    "passedPolicyTitle": "消費税廃止",
    "actualEffects": { "economy": 20, "welfare": -15, ... },
    "newsFlash": "【速報】...",
    "voteDetails": { "user1": "policy_001", "user2": "policy_001" }
  },
  "cityParams": { "economy": 70, ... }
}
```

---

#### POST `/api/rooms/{roomId}/next` - 次ターンへ

結果発表後、次のターンに進む（ホストのみ）。

**リクエスト:** なし

**処理:**
1. リクエスト者がホストであることを確認
2. RESULT 状態であることを確認
3. `turn` をインクリメント
4. `status` を `VOTING` に

**レスポンス:**
```json
{
  "status": "VOTING",
  "turn": 2
}
```

---

#### POST `/api/rooms/{roomId}/petition` - AI陳情

AIに新しい政策を提案する（1人1回）。

**リクエスト:**
```json
{
  "text": "週休3日制を導入したい"
}
```

**処理:**
1. プレイヤーの `isPetitionUsed` を確認
2. OpenAI API で審査
3. 承認なら政策カードを生成し `deckIds` に追加
4. `isPetitionUsed` を `true` に

**レスポンス:**
```json
{
  "approved": true,
  "policyId": "petition_xxx",
  "message": "政策が承認されました"
}
```

---

## フロントエンド実装パターン

### API クライアント

```typescript
// api/client.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL;

async function apiCall<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const token = await auth.currentUser?.getIdToken();
  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.error);
  }
  return res.json();
}

// 部屋作成
export const createRoom = (displayName: string) =>
  apiCall<{ roomId: string }>('/api/rooms', {
    method: 'POST',
    body: JSON.stringify({ displayName }),
  });

// 部屋参加
export const joinRoom = (roomId: string, displayName: string) =>
  apiCall<{ success: boolean }>(`/api/rooms/${roomId}/join`, {
    method: 'POST',
    body: JSON.stringify({ displayName }),
  });

// Ready
export const toggleReady = (roomId: string) =>
  apiCall<{ isReady: boolean }>(`/api/rooms/${roomId}/ready`, {
    method: 'POST',
  });

// ゲーム開始
export const startGame = (roomId: string) =>
  apiCall<StartGameResponse>(`/api/rooms/${roomId}/start`, {
    method: 'POST',
  });

// 投票
export const vote = (roomId: string, policyId: string) =>
  apiCall<{ success: boolean }>(`/api/rooms/${roomId}/vote`, {
    method: 'POST',
    body: JSON.stringify({ policyId }),
  });

// 投票集計
export const resolveVote = (roomId: string) =>
  apiCall<ResolveVoteResponse>(`/api/rooms/${roomId}/resolve`, {
    method: 'POST',
  });

// 次ターン
export const nextTurn = (roomId: string) =>
  apiCall<{ status: string; turn: number }>(`/api/rooms/${roomId}/next`, {
    method: 'POST',
  });

// 陳情
export const submitPetition = (roomId: string, text: string) =>
  apiCall<PetitionResponse>(`/api/rooms/${roomId}/petition`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  });
```

### リアルタイム監視

```typescript
// hooks/useRoom.ts
import { doc, collection, onSnapshot } from 'firebase/firestore';

export function useRoom(roomId: string) {
  const [room, setRoom] = useState<Room | null>(null);
  const [players, setPlayers] = useState<Player[]>([]);

  useEffect(() => {
    // ルーム監視
    const unsubRoom = onSnapshot(
      doc(db, 'rooms', roomId),
      (doc) => setRoom(doc.data() as Room)
    );

    // プレイヤー監視
    const unsubPlayers = onSnapshot(
      collection(db, 'rooms', roomId, 'players'),
      (snapshot) => {
        setPlayers(snapshot.docs.map(d => ({
          oderId: d.id,
          ...d.data()
        } as Player)));
      }
    );

    return () => {
      unsubRoom();
      unsubPlayers();
    };
  }, [roomId]);

  return { room, players };
}
```

---

## Security Rules

フロントエンドからの直接更新を禁止し、読み取りのみ許可する。

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    // マスターデータ: 誰でも読み取り可
    match /master_policies/{policyId} {
      allow read: if true;
      allow write: if false;
    }

    match /master_ideologies/{ideologyId} {
      allow read: if true;
      allow write: if false;
    }

    // ルーム: 認証済みユーザーのみ読み取り可
    match /rooms/{roomId} {
      allow read: if request.auth != null;
      allow write: if false;  // APIからのみ更新

      // プレイヤー: 認証済みユーザーのみ読み取り可
      // ただし ideology, currentVote は本人のみ
      match /players/{oderId} {
        allow read: if request.auth != null && (
          request.auth.uid == oderId ||
          !('ideology' in resource.data) ||
          !('currentVote' in resource.data)
        );
        allow write: if false;  // APIからのみ更新
      }
    }
  }
}
```

> **Note:** バックエンドは Admin SDK を使用するため、Security Rules をバイパスします。

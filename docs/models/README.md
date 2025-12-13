# Firestore データモデル（JSONサンプル）

各ファイルは Firestore に保存されるドキュメントのサンプルです。

## ファイル一覧

| ファイル | Firestoreパス | 説明 |
|---------|--------------|------|
| `room.json` | `rooms/{roomId}` | ゲームルーム |
| `player.json` | `rooms/{roomId}/players/{userId}` | プレイヤー（サブコレクション） |
| `master_policy.json` | `master_policies/{policyId}` | 政策カードマスター |
| `master_ideology.json` | `master_ideologies/{ideologyId}` | 思想マスター |

## ステータス遷移

```
LOBBY → VOTING → RESULT → VOTING → ... → FINISHED
```

## 注意事項

- `_path`, `_description`, `_comment_*` はドキュメント説明用のメタ情報で、実際のFirestoreには保存しません
- `player.ideology` と `player.currentVote` は本人のみ読み取り可能（ハッカソン用の簡易実装ではフロント側で非表示にする）
- `master_policy.effects` は結果発表まで非公開（フロント側で非表示にする）

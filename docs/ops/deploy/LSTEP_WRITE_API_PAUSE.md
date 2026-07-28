# Lステップ Write API 一時停止メモ

> **目的**: Lステップ Write API 一時停止の運用決定を記録する。
> **読者**: 開発者・PO。
> **タイミング**: Lステップ書込み再開を判断する時。

> **対象**: Lステップへのタグ付与・タグ解除・プロパティ更新の外部 API 呼び出し
> **状態**: 一時停止中。読み取り系 API とアプリ内 DB 更新は継続。

---

## 現在停止している操作

以下の `backend/internal/infra/lstep` 実装は、入力検証のみ行い、Lステップへの HTTP write を送信しない。

| ファイル | メソッド | 抑止している API |
|:---|:---|:---|
| `tag.go` | `AddTag` | `POST /contacts/{id}/tags` |
| `tag.go` | `RemoveTag` | `DELETE /contacts/{id}/tags` |
| `tag.go` | `AddTagBulk` | `POST /contacts/tags/bulk` |
| `user.go` | `SetProperty` | `POST /contacts/{id}/properties` |

各メソッドには `[DISABLED]` コメントを残している。呼び出し元サービスは通常どおり実行されるが、外部 Lステップ側の状態は変更されない。

---

## 停止の目的

- STG 初期公開前に誤配信・誤タグ付与を防ぐ
- タグ判定ロジック、配信対象抽出、監査ログ、手動操作 UI を先に検証する
- clinic ごとの token / channel / Lステップ設定が確定するまで外部 write を抑止する

---

## 再有効化の前提条件

再有効化は以下をすべて満たした後に行う。

1. STG の対象 clinic ごとに Lステップ token / channel / LIFF 設定が確認済み
2. `clinic_integrations` の値が対象 clinic の実アカウントと一致している
3. タグ名・タグコード・配信トリガーの master / seed が現行運用と一致している
4. STG で読み取り API と対象抽出ロジックの smoke test が完了している
5. 誤付与時の手動復旧手順が運用者に共有済み

---

## 再有効化手順

1. `backend/internal/infra/lstep/tag.go` と `user.go` の `[DISABLED]` 実装を git history から復元する
2. 空の `lineUserID` が `ErrUserNotFound` を返す既存挙動は維持する
3. write 系 API の unit test / integration smoke を追加または更新する
4. STG で少数のテスト LINE User ID に対して Add/Remove/SetProperty を確認する
5. Cloudflare Workers Logs / Containers のログと Lステップ管理画面で write 結果を照合する

---

## 注意

- `is_sync_enabled=false` の clinic は、再有効化後もサービス層で同期を抑止する。
- このメモは外部 write 停止の運用判断を記録するものであり、Lステップ連携機能全体の削除を意味しない。

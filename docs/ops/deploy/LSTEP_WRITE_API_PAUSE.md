# Lステップ Write API 一時停止メモ

> **目的**: Lステップ Write API の deploy-level kill switch と clinic 二重 gate の運用を記録する。
> **読者**: 開発者・PO・運用者。
> **タイミング**: Lステップ書込み再開を判断する時。

> **対象**: Lステップへのタグ付与・タグ解除・プロパティ更新の外部 API 呼び出し
> **状態**: **deploy gate 既定 OFF**（`LSTEP_WRITE_API_ENABLED` 未設定）。読み取り系 API とアプリ内 DB 更新は継続。

---

## 二重 gate（必須）

Write が外部へ送られるのは **両方** が true のときだけ。

| Gate | 場所 | 制御 | 既定 |
|:---|:---|:---|:---|
| **Deploy kill switch** | `backend/internal/infra/lstep`（`LSTEP_WRITE_API_ENABLED`） | 環境変数のみ。UI/API/seed/migration から変更不可 | 未設定・空・`false`・未知値 → **false**。exact `true` のみ有効 |
| **Clinic flag** | サービス層 `buildClient`（`is_sync_enabled`） | clinic 設定（既存 PATCH/UI） | clinic ごと。false なら client を生成せず HTTP 0 |

### Deploy gate 無効時の契約

- Write 4 メソッド（`AddTag` / `RemoveTag` / `AddTagBulk` / `SetProperty`）は **HTTP request 0**
- 戻り値は **`ErrWriteDisabled`**（`nil` 成功にしない）
- 上位の delivery fired / tag cache receipt 更新に進まない（エラー経路）

### Clinic flag 無効時の契約

- `buildClient` が `nil` client を返し、サービスは同期をスキップまたは設定エラーとする
- infra client まで到達しないため外部 write は 0

---

## 現在の Write メソッド

| ファイル | メソッド | API |
|:---|:---|:---|
| `tag.go` | `AddTag` | `POST /contacts/{id}/tags` |
| `tag.go` | `RemoveTag` | `DELETE /contacts/{id}/tags` |
| `tag.go` | `AddTagBulk` | `POST /contacts/tags/bulk` |
| `user.go` | `SetProperty` | `POST /contacts/{id}/properties` |

repo 上の write 実装は復元済み。送信可否は上記二重 gate が決める。
paused no-op（`return nil`）や `[DISABLED]` 経路は廃止済み。

---

## 停止の目的

- STG 初期公開前に誤配信・誤タグ付与を防ぐ
- タグ判定ロジック、配信対象抽出、監査ログ、手動操作 UI を先に検証する
- clinic ごとの token / channel / Lステップ設定が確定するまで外部 write を抑止する
- clinic flag だけを true にしても deploy gate が OFF なら外部 write は始まらない

---

## 再有効化の前提条件

再有効化は以下をすべて満たした後に行う（**USER 専権**。エージェントは実環境変数変更・実送信を行わない）。

1. STG の対象 clinic ごとに Lステップ token / channel / LIFF 設定が確認済み
2. `clinic_integrations` の値が対象 clinic の実アカウントと一致している
3. タグ名・タグコード・配信トリガーの master / seed が現行運用と一致している
4. STG で読み取り API と対象抽出ロジックの smoke test が完了している
5. 誤付与時の手動復旧手順が運用者に共有済み
6. IP 許可リスト等ベンダー側制約があれば確認済み

---

## 再有効化手順（USER 専権・順序固定）

1. **外部 API 有効化の承認**（先方・PO）
2. **Deploy gate 設定**: 対象環境に `LSTEP_WRITE_API_ENABLED=true` を設定して再デプロイ
   - それ以外の値・未設定は書き込み無効のまま
3. **対象 clinic の `is_sync_enabled=true`**（必要な clinic のみ）
4. **STG で少数のテスト LINE User ID** に対して Add/Remove/SetProperty を 1 回確認
5. Cloudflare Workers Logs / Containers のログと Lステップ管理画面で write 結果を照合
6. 問題時の **rollback**: `LSTEP_WRITE_API_ENABLED` を削除または `false` にして再デプロイ（clinic flag より先に全体停止できる）

---

## 注意

- `is_sync_enabled=false` の clinic は、deploy gate 有効後もサービス層で同期を抑止する。
- api key・request body・response body はログにも error 文字列にも出さない。
- このメモは外部 write 制御の運用判断を記録するものであり、Lステップ連携機能全体の削除を意味しない。
- 実送信の確認結果は USER が記録する（repo 準備完了 ≠ 外部有効化完了）。

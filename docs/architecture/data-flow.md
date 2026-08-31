# データフローとトレーサビリティ (Data Flow)

> **目的**: リクエスト追跡(Request ID)と非同期処理のデータフローを定義する。
> **読者**: 新規参加開発者。
> **タイミング**: リクエスト処理フローを具体的に理解したい時。

> **Animal Ekarte**: リクエストからレスポンス、バックグラウンド処理までの追跡
> **最新更新**: 2026-07-31

---

## 1. トレーサビリティとロギング

登録済み RequestID middleware を通る HTTP request は、request-completion log と response header で request ID を追跡できます。これは background work や downstream log 全体の自動 correlation を保証しません。

### Request ID の伝播フロー
1.  **生成**: リクエスト受信時、`middleware.RequestID()` が 8 文字のランダム hex 文字列（4 バイト）を生成。クライアントが英数字・ハイフン・アンダースコアのみ 64 文字以内の有効な `X-Request-ID` を送信した場合はそれを再利用。
2.  **格納**: `gin.Context` に `request_id` として保持。
3.  **返却**: レスポンスヘッダー `X-Request-ID` としてクライアントへ返却。
4.  **記録**: request-completion structured log に含めます。ID は `gin.Context` にだけ保持され、`c.Request.Context()` へ自動伝播しないため、downstream の全 `slog.*Context` log に常に含まれるわけではありません。出力先は現在設定された log sink / platform に従います。

---

## 2. 典型的な処理フロー (CRUD)

### 例：飼主一覧の取得 (GET /api/v1/owners)

[ADR-006](./adr/006-backend-domain-package-boundaries.md) の domain/capability-first modular monolith では、固定の `internal/handler`・`internal/service`・`internal/repository` ディレクトリを必須としない。小規模 resource は `internal/<domain>` 内に HTTP 境界・use case・persistence を同居させる。

1.  **Middleware (Auth)**:
    - `access_token` Cookie（または Bearer）から JWT を検証（`middleware.Auth`）。
    - トークンの `clinic_ids` はログイン時スナップショットであり **最終 authority ではない**。
    - 原則として毎リクエスト `current_access_service` で account / staff / clinic assignment を再解決し、信頼できる clinic scope を `gin.Context` に格納する。
    - 対象 clinic は `X-Clinic-ID`（未指定時は既定 clinic）で確定し、再解決済み所属集合との一致を必須とする。request-time authority lookup が利用不能な場合も 503 で fail closed とし、JWT snapshot へ fallback しない。
2.  **HTTP 境界（`internal/owner` 等）**:
    - ページング等の query を bind / 検証する。
    - ルート側の `RequirePermission` 等で RBAC を強制する。
3.  **Use case（同一 domain package）**:
    - 再解決済み clinic scope を使って一覧・件数を取得する（例: `clinic_id IN (?)`）。
4.  **Persistence（同一 domain package 内）**:
    - **テナント隔離**: clinic 述語を強制適用。
    - 総件数 (Total) とリストを一貫した状態で取得。

---

## 3. イベント駆動フロー（会計完了時の CPM タグ同期）

### 例：Lステップタグ自動付与 (会計完了時)

1.  **Event Trigger**: `internal/billing` の `accountingService.Create/Update` が会計完了（billing 確定）を検知（`accounting_service_core.go`）。
2.  **同期ディスパッチ**: レスポンス返却前に `syncCPMStageTag` → `tagSyncSvc.SyncCPMStageTag(ctx, clinicID, ownerID)` を**同期呼び出し**（goroutine ではない）。タグ同期が失敗してもエラーは記録のみで会計処理は継続（fail-open）。
3.  **Condition Judge**:
    - 累計売上、来院頻度、最終来院日を再計算。
    - CPM ステージを再算出（変動判定は行わず、旧ステージタグを全削除して新ステージタグを付与する冪等方式。`lstep_tag_sync_visit_cpm.go`）。
4.  **External API**: Lステップ API クライアント経由でタグを付与/解除する。Write 系は **二重ゲート**:
    - **Deploy kill switch**: 環境変数 `LSTEP_WRITE_API_ENABLED` が exact `true` のときのみ有効（未設定・空・`false`・未知値は無効）。無効時は HTTP を送らず `ErrWriteDisabled` を返す（成功 `nil` の silent noop ではない）。
    - **Clinic setting**: `is_sync_enabled` 等の医院設定も満たす必要がある。
    - Read 系（`GetUserTags` 等）は deploy gate の対象外で稼働し得る。運用詳細は [`LSTEP_WRITE_API_PAUSE.md`](../ops/deploy/LSTEP_WRITE_API_PAUSE.md)。
5.  **記録**: 処理結果は `slog` とタグキャッシュ（`lstep_tag_cache_repository.go` の UpsertTag/DeleteTag）に反映。API 失敗はエラーカウンターに記録し、閾値到達で `EXCL_カルテ連携エラー` タグを付与。タグ同期経路では `audit_logs` / `lstep_delivery_trigger_log` への記録は行わない（後者は自動配信トリガー専用ログ）。

---

## 4. エラーハンドリング体系

`RespondError(c, err)` による統一レスポンス。

| ステータス | 分類 | レスポンス内容 |
|:---|:---|:---|
| **400** | 不正な入力 | safe error string。endpoint が明示的に field detail を返す contract の場合だけ field 名を含む |
| **401/403** | 認証/認可エラー | 権限不足の明示 |
| **404** | リソース不在 | 他テナントへのアクセスも「不在」として扱い情報を隠蔽 |
| **409** | 整合性・衝突 | 使用中のマスタ削除、重複登録など |
| **500** | サーバーエラー | 詳細は隠蔽し `"internal server error"` を返却 |

---

## 5. マルチテナント隔離原則

clinic-scoped endpoint は以下の原則を満たす。

- **No Trust (request-time authority)**: クライアントが query/body で指定する `clinic_id` は信用しない。通常リクエストの最終 authority は JWT の `clinic_ids` スナップショットではなく、`current_access_service` による request-time 再解決結果である。対象 clinic は `X-Clinic-ID`（未指定時は既定 clinic）で選び、再解決済みの有効所属集合との一致を必須とする（system admin も active clinic のみ）。lookup availability failure も 503 で fail closed とし、token snapshot を continuity authority にしない。
- **Strict Isolation**: clinic-owned data の read/write/delete、join/preload/count/bulk/background path に clinic scope と ownership check を適用する。index だけを isolation の証明にしない。
- **Audit Trace (path-dependent)**: セキュリティ・資格情報・clinic 切替・臨床/会計上必須と定めた変更など、経路ごとに監査対象が決まる。すべての CUD が機械的に `audit_logs` へ入るわけではなく、タグ同期のように意図的に監査しない経路もある。必須監査は業務 write と同一 transaction で fail-closed にする（詳細は各 domain / [auth.md](./auth.md)）。

---

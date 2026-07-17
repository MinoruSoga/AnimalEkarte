# データフローとトレーサビリティ (Data Flow)

> **目的**: リクエスト追跡(Request ID)と非同期処理のデータフローを定義する。
> **読者**: 新規参加開発者。
> **タイミング**: リクエスト処理フローを具体的に理解したい時。

> **Animal Ekarte**: リクエストからレスポンス、バックグラウンド処理までの追跡
> **最新更新**: 2026-06-12

---

## 1. トレーサビリティとロギング

本システムは、すべての処理を一意の ID で追跡し、商用グレードの運用監視を実現しています。

### Request ID の伝播フロー
1.  **生成**: リクエスト受信時、`middleware.RequestID()` が 8 文字のランダム hex 文字列（4 バイト）を生成。クライアントが英数字・ハイフン・アンダースコアのみ 64 文字以内の有効な `X-Request-ID` を送信した場合はそれを再利用。
2.  **格納**: `gin.Context` に `request_id` として保持。
3.  **返却**: レスポンスヘッダー `X-Request-ID` としてクライアントへ返却。
4.  **記録**: `slog` 出力に常に含まれ、ログ基盤（CloudWatch等）での一括検索を可能にします。

---

## 2. 典型的な処理フロー (CRUD)

### 例：飼主一覧の取得 (GET /api/v1/owners)

1.  **Middleware (Auth)**:
    - `access_token` Cookie から JWT を検証。
    - Claims から `clinic_id` を抽出し `gin.Context` へ格納。
2.  **Handler (ListOwners)**:
    - `extractClinicID` でテナントを確定。
    - `parsePagination` でページ・リミットを解析。
3.  **Service (OwnerService.List)**:
    - `clinic_id` でスコープしたリポジトリ呼び出し（権限チェックは handler 層の `RequirePermission` ミドルウェアが担い、service 層では行わない）。
4.  **Repository (OwnerRepository.FindAll)**:
    - **テナント隔離**: `WHERE clinic_id = ?` を強制適用。
    - 総件数 (Total) とリストを単一トランザクションまたは一貫した状態で取得。

---

## 3. イベント駆動フロー（会計完了時の CPM タグ同期）

### 例：Lステップタグ自動付与 (会計完了時)

1.  **Event Trigger**: サービス層 `accountingService.Create/Update` が会計完了（billing 確定）を検知（`accounting_service_core.go`）。
2.  **同期ディスパッチ**: レスポンス返却前に `syncCPMStageTag` → `tagSyncSvc.SyncCPMStageTag(ctx, clinicID, ownerID)` を**同期呼び出し**（goroutine ではない）。タグ同期が失敗してもエラーは記録のみで会計処理は継続（fail-open）。
3.  **Condition Judge**: 
    - 累計売上、来院頻度、最終来院日を再計算。
    - CPM ステージを再算出（変動判定は行わず、旧ステージタグを全削除して新ステージタグを付与する冪等方式。`lstep_tag_sync_visit_cpm.go`）。
4.  **External API**: Lステップ API クライアント経由でタグを付与/解除。ただし **Write 系 API（AddTag/RemoveTag 等）はポリシーにより一時停止中**で、`internal/infra/lstep/tag.go` は HTTP 送信を抑止した noop（Read 系 `GetUserTags` は稼働）。
5.  **記録**: 処理結果は `slog` とタグキャッシュ（`lstep_tag_cache_repository.go` の UpsertTag/DeleteTag）に反映。API 失敗はエラーカウンターに記録し、閾値到達で `EXCL_カルテ連携エラー` タグを付与。タグ同期経路では `audit_logs` / `lstep_delivery_trigger_log` への記録は行わない（後者は自動配信トリガー専用ログ）。

---

## 4. エラーハンドリング体系

`RespondError(c, err)` による統一レスポンス。

| ステータス | 分類 | レスポンス内容 |
|:---|:---|:---|
| **400** | 不正な入力 | フィールド名を含むバリデーションエラー |
| **401/403** | 認証/認可エラー | 権限不足の明示 |
| **404** | リソース不在 | 他テナントへのアクセスも「不在」として扱い情報を隠蔽 |
| **409** | 整合性・衝突 | 使用中のマスタ削除、重複登録など |
| **500** | サーバーエラー | 詳細は隠蔽し `"internal server error"` を返却 |

---

## 5. マルチテナント隔離原則

全エンドポイントで以下の原則を徹底しています。

- **No Trust**: クライアントからの `clinic_id` 指定は一切信用せず、JWT から取得。
- **Strict Isolation**: 全クエリに `clinic_id` フィルタを適用し、他院のデータ混入を物理的に遮断。
- **Audit Trace**: 全てのデータ変更（Create/Update/Delete）について、実行者と対象院を監査ログに記録。

---

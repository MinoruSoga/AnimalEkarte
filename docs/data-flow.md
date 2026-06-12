# データフローとトレーサビリティ (Data Flow)

> **Animal Ekarte**: リクエストからレスポンス、バックグラウンド処理までの追跡
> **最新更新**: 2026-06-12

---

## 1. トレーサビリティとロギング

本システムは、すべての処理を一意の ID で追跡し、商用グレードの運用監視を実現しています。

### Request ID の伝播フロー
1.  **生成**: リクエスト受信時、`middleware.RequestID()` が UUID を生成。
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
    - 権限チェックと業務ルールの評価。
4.  **Repository (OwnerRepository.FindAll)**:
    - **テナント隔離**: `WHERE clinic_id = ?` を強制適用。
    - 総件数 (Total) とリストを単一トランザクションまたは一貫した状態で取得。

---

## 3. 非同期・イベント駆動フロー

### 例：Lステップタグ自動付与 (会計完了時)

1.  **Event Trigger**: 会計完了 (`PATCH /accountings/:id`) ハンドラーが成功を検知。
2.  **Goroutine Dispatch**: メインレスポンスを返却後、バックグラウンドで `LstepTagSyncService.Sync` を起動。
3.  **Condition Judge**: 
    - 累計売上、来院頻度、最終来院日を再計算。
    - CPM ステージが変動したか判定。
4.  **External API**: Lステップ API を呼び出し、タグを付与/解除。
5.  **Audit Log**: 処理結果（成功/失敗/除外理由）を `audit_logs` および `lstep_delivery_trigger_log` に記録。

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

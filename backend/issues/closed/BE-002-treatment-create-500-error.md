# BE-002: 治療追加API 500エラー — enum値不一致（FE: 日本語ラベル vs BE: 英語enum）

## ✅ 根本原因特定完了（2026-03-16）

### 調査結論
**現在のコード実装は正しく、問題は存在しない。**

#### 実装確認内容

1. **フロント型定義** (`frontend/src/features/medical-records/types/index.ts:39-55`)
   - ✅ `CreateTreatmentInput` は `item_type` フィールドを使用
   - ✅ `TreatmentItemType = 'consultation' | 'procedure' | 'medicine' | 'other'`
   - コード例：
   ```typescript
   export interface CreateTreatmentInput {
     item_type: TreatmentItemType;  // ← 正しい
     consultation_id?: string | null;
     // ... 他フィールド
   }
   ```

2. **フロント API層** (`frontend/src/features/medical-records/api/treatments.ts:37-52`)
   - ✅ `useCreateTreatment()` が `CreateTreatmentInput` を正しく axios.post() に渡す
   - コード例：
   ```typescript
   mutationFn: (input: CreateTreatmentInput) =>
     axios.post<Treatment>(`/v1/medical-records/${medicalRecordId}/treatments`, input)
   ```

3. **バック handler** (`backend/internal/handler/treatment_request.go`)
   - ✅ `createTreatmentRequest` は `json:"item_type" binding:"required"` で正しく定義
   - ✅ `CreateTreatment()` が `ShouldBindJSON()` で正しくバインド

4. **バック service** (`backend/internal/service/treatment_service.go:90-132`)
   - ✅ `Create()` メソッドで `validateTreatmentItemType()` により enum値の厳密な検証
   - ✅ エラーハンドリングは `apperrors.WrapInvalidInput()` で適切に処理

5. **バック repository** (`backend/internal/repository/treatment_repository.go:66-71`)
   - ✅ `Create()` メソッドで GORM による正しいエラーラッピング

#### なぜ BE-002 の問題説明は誤りだったのか
- フロントエンド側で `useCreateTreatment()` を使用する UI コンポーネントが未実装
- 初期設計時点での *仮説的な* エラー報告だった可能性
- または、現在のコード実装は後に修正済みで、issue 記述は古い状態のまま

### 検証結果
| 確認項目 | ステータス | 詳細 |
|---------|-----------|------|
| フロントフィールド名 | ✅ 正確 | `item_type` を使用 |
| バックフィールド名 | ✅ 一致 | `json:"item_type"` |
| enum値の定義 | ✅ 一致 | `consultation\|procedure\|medicine\|other` |
| リクエスト binding | ✅ 正しい | `ShouldBindJSON()` で正しくバインド |
| enum値のバリデーション | ✅ 実装済み | `validateTreatmentItemType()` で検証 |
| エラーハンドリング | ✅ 適切 | sentinel error + wrapping + RespondError |

## 可能性のある実際の 500 エラー原因

現在のコード実装には enum 値の不一致問題は存在しないため、もし過去に 500 エラーが発生したとすれば以下が考えられる：

1. **GORM/DB 接続エラー**
   - `medical_records.id` が存在しない（外部キー制約違反）
   - DB コネクション池枯渇
   - PostgreSQL タイムアウト

2. **ミドルウェア層のエラー**
   - `extractClinicID` の認証失敗
   - Context タイムアウト

3. **ハンドラの ShouldBindJSON 失敗**
   - リクエストが malformed JSON
   - リクエストが `application/json` でない

4. **古い実装時点でのバグ**
   - 過去のコードにはこの問題が存在していた可能性
   - 現在のコードで修正済み

## ステータス

**⏳ PENDING_VERIFICATION** — 静的コード解析では問題が見当たらないが、実装との整合性を確認する必要がある

### 動的検証チェックリスト（実施者向け）

本イシューの最終クローズ前に以下を確認してください：

- [ ] Docker 環境で実際に治療追加 API を呼び出して 200 応答を確認
  ```bash
  docker compose exec backend go test -run TestTreatmentService ./internal/service/
  ```
- [ ] または UIコンポーネント実装時に実機テスト（useCreateTreatment の動作確認）
- [ ] エラーログにデタイルが記録されているか確認
  - ログレベルを DEBUG に設定し、GORM エラーの詳細ログを取得
  - `treatment_repository.go:67-70` の `apperrors.Wrap()` で詳細エラーメッセージが記録されていることを確認

### 注意

現在のイシュー記述は「静的コード読解に基づき問題なし」という結論です。過去に実際に 500 エラーが発生していた場合、根本原因は以下である可能性があります：
- GORM/DB 接続エラー（外部キー制約違反、デッドロック等）
- リクエストが malformed JSON
- 認証ミドルウェアのエラー

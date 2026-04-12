# Section 14 Master Settings CRUD Test Plan
## 2026-04-12 Browser Functional Tests

**Target**: Comprehensive CRUD testing for 5 Master Data types  
**Scope**: Animal Species, Reservation Types, Staff, Trimming, Diagnosis Names  
**Environment**: http://localhost:3003  
**Test Account**: admin@example.com / password  

---

## Test Execution Plan

### Phase 1: ブラウザ起動・ログイン準備
- [ ] Chrome DevTools MCP でページ一覧確認
- [ ] http://localhost:3003 が応答確認（HTTP 200）
- [ ] ログイン状態確認 / 必要に応じてログイン

### Phase 2: マスタ1 - 動物種マスタ (`/settings/animal-species`)

#### 2.1 READ - 一覧表示確認
- **操作**: Navigate to `/settings/animal-species`
- **確認項目**:
  - ページタイトル「動物種マスタ」表示
  - リスト表示・複数件存在
  - 初期表示件数メモ（期待値: 6件）
  - API: `GET /api/v1/masters/animal-species` → HTTP 200

#### 2.2 CREATE - 新規登録
- **操作**: 
  1. 「+ 新規登録」ボタンクリック
  2. サイドパネル表示確認
  3. 名称フィールドに「テスト_動物種_20260412」入力
  4. 「保存」ボタンクリック
- **確認項目**:
  - API: `POST /api/v1/masters/animal-species` → HTTP 201
  - リスト件数: 6→7件に増加
  - トースト: 「登録しました」表示
  - 新規アイテム最下行に表示

#### 2.3 UPDATE - 編集
- **操作**:
  1. 作成したアイテムの「編集」アイコンクリック
  2. サイドパネル「編集」表示
  3. 名称を「テスト_動物種_編集」に変更
  4. 「保存」ボタンクリック
- **確認項目**:
  - API: `PATCH /api/v1/masters/animal-species/{id}` → HTTP 200
  - リスト表示が即座に更新
  - 変更内容がリストに反映

#### 2.4 DELETE - 削除
- **操作**:
  1. 編集パネルまたは削除アイコンをクリック
  2. 削除確認ダイアログ表示
  3. 「削除する」をクリック
- **確認項目**:
  - API: `DELETE /api/v1/masters/animal-species/{id}` → HTTP 204
  - リスト件数: 7→6件に減少（元に戻る）
  - トースト: 「削除しました」表示

#### 2.5 ERROR HANDLING - FK制約確認
- **操作**: 既存ペットに紐付いた動物種を削除試行
- **確認項目**:
  - API: `DELETE /api/v1/masters/animal-species/{id}` → HTTP 409 Conflict
  - エラーメッセージ表示（「この動物種はペット情報で使用中のため...」）

---

### Phase 3: マスタ2 - サービス種別マスタ (`/settings/reservation-type`)

#### 3.1 READ - 一覧表示確認
- **操作**: Navigate to `/settings/reservation-type`
- **確認項目**:
  - ページタイトル「サービス種別マスタ」表示
  - 初期表示件数メモ（期待値: 25件）
  - API: `GET /api/v1/masters/reservation-types` → HTTP 200

#### 3.2 CREATE - 新規登録
- **テスト名**: `test_svc_20260412`
- **確認**: 25→26件、POST 201、トースト表示

#### 3.3 UPDATE - 編集
- **テスト名変更**: `test_svc_編集`
- **確認**: PATCH 200、リスト反映

#### 3.4 DELETE - 削除
- **確認**: 26→25件、DELETE 204

#### 3.5 ERROR HANDLING - FK制約
- **操作**: 既存予約に紐付いたサービス種別削除試行
- **確認**: HTTP 409 Conflict + エラーメッセージ

---

### Phase 4: マスタ3 - スタッフマスタ (`/settings/staff`)

#### 4.1 READ - 一覧表示確認
- **操作**: Navigate to `/settings/staff`
- **確認項目**:
  - 初期表示件数メモ（期待値: 16件）
  - API: `GET /api/v1/masters/staffs` → HTTP 200

#### 4.2 CREATE - 新規登録
- **テスト名**: `テストスタッフ_20260412`
- **確認**: 16→17件、POST 201

#### 4.3 UPDATE - 編集
- **テスト名変更**: `テストスタッフ_編集`
- **確認**: PATCH 200、リスト反映

#### 4.4 DELETE - 削除
- **確認**: 17→16件、DELETE 204

#### 4.5 ERROR HANDLING - FK制約
- **操作**: 既存予約に紐付いたスタッフ削除試行
- **確認**: HTTP 409 Conflict

---

### Phase 5: マスタ4 - トリミングマスタ (`/settings/trimming`)

#### 5.1 READ - 一覧表示確認
- **操作**: Navigate to `/settings/trimming`
- **確認項目**:
  - 初期表示件数メモ（期待値: 5件）
  - API: `GET /api/v1/masters/trimming-courses` → HTTP 200

#### 5.2 CREATE - 新規登録
- **テスト名**: `テストコース_20260412`
- **確認**: 5→6件、POST 201

#### 5.3 UPDATE - 編集
- **テスト名変更**: `テストコース_編集`
- **確認**: PATCH 200

#### 5.4 DELETE - 削除
- **確認**: 6→5件、DELETE 204

#### 5.5 ERROR HANDLING - FK制約
- **操作**: 既存トリミング記録に紐付いたコース削除試行
- **確認**: HTTP 409 Conflict

---

### Phase 6: マスタ5 - 診断病名マスタ (`/settings/diagnosis`)

#### 6.1 READ - 一覧表示確認
- **操作**: Navigate to `/settings/diagnosis`
- **確認項目**:
  - 初期表示件数メモ（期待値: 20件）
  - API: `GET /api/v1/masters/diagnosis-names` → HTTP 200

#### 6.2 CREATE - 新規登録
- **テスト名**: `テスト病名_20260412`
- **確認**: 20→21件、POST 201

#### 6.3 UPDATE - 編集
- **テスト名変更**: `テスト病名_編集`
- **確認**: PATCH 200

#### 6.4 DELETE - 削除
- **確認**: 21→20件、DELETE 204

#### 6.5 ERROR HANDLING - FK制約
- **操作**: 既存カルテに紐付いた病名削除試行
- **確認**: HTTP 409 Conflict

---

## Expected HTTP Status Codes

| Operation | Expected Status | Meaning |
|-----------|-----------------|---------|
| GET /masters/\*\* | 200 | OK - 一覧取得成功 |
| POST /masters/\*\* | 201 | Created - 登録成功 |
| PATCH /masters/\*\*/\{id\} | 200 | OK - 更新成功 |
| DELETE /masters/\*\*/\{id\} | 204 | No Content - 削除成功 |
| DELETE (FK conflict) | 409 | Conflict - FK参照存在 |
| POST (duplicate name) | 409 | Conflict - 重複名称 |
| POST (validation error) | 400 | Bad Request - バリデーション失敗 |

---

## Toast Messages Expected

| Action | Toast Message |
|--------|---------------|
| Create Success | 「登録しました」 |
| Update Success | 「編集しました」 or 「更新しました」 |
| Delete Success | 「削除しました」 |
| FK Constraint Error | 「[マスタ名]の削除に失敗しました」（generic）or 「このデータは〜で使用中のため削除できません」 |
| Duplicate Name Error | 「登録に失敗しました」 or 「[マスタ名]の登録に失敗しました」 |

---

## Test Result Summary Template

```
## テスト結果: Section 14 マスタ設定
実行日時: 2026-04-12 18:XX:XX

### テスト対象マスタ（5個 - 初回テスト）
| マスタ | READ | CREATE | UPDATE | DELETE | FK Error | 備考 |
|--------|------|--------|--------|--------|----------|------|
| 動物種マスタ | OK/NG | OK/NG | OK/NG | OK/NG | OK/NG | 観察内容 |
| サービス種別 | OK/NG | OK/NG | OK/NG | OK/NG | OK/NG | 観察内容 |
| スタッフマスタ | OK/NG | OK/NG | OK/NG | OK/NG | OK/NG | 観察内容 |
| トリミングマスタ | OK/NG | OK/NG | OK/NG | OK/NG | OK/NG | 観察内容 |
| 診断病名マスタ | OK/NG | OK/NG | OK/NG | OK/NG | OK/NG | 観察内容 |

### Summary
- Total Test Cases: 25 (5 masters × 5 operations)
- OK: XX / NG: XX / Partial: XX
- HTTP Status Codes: 200, 201, 204, 409 all verified
- UI Toast Messages: All verified
- FK Constraint Handling: Verified

### Discovered Bugs
- [BUG-XXX]: {description}

### Notes
- All CREATE operations incremented list count by 1
- All DELETE operations decremented list count by 1
- All UPDATE operations reflected immediately in UI
- FK constraint errors returned HTTP 409 as expected
```

---

## Execution Notes

1. **Timestamps**: Use `new Date().toISOString()` format for test data names to ensure uniqueness
2. **Network Monitoring**: Use Chrome DevTools Network tab to verify exact HTTP status codes
3. **Timing**: Allow 2-3 seconds between operations for API response
4. **Cleanup**: Delete all test records after each master to restore to initial state
5. **Screenshots**: Capture on NG cases for bug reporting

---

## Related Documentation
- FUNCTIONAL_TEST_REPORT.md Section 14 - Current test status
- .claude/CLAUDE.md - Development rules
- docs/ERD.md - 45 master tables schema
- backend/internal/model/* - Master entity definitions


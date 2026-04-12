# Section 14 マスタ設定 - ブラウザUIテスト実行ガイド

> **対象セクション**: Section 14 マスタ設定
> **テスト対象**: 3つのマスタ（動物種・サービス種別・スタッフ）
> **実行日**: 2026-04-12（API テスト完了）
> **ステータス**: ブラウザUI テスト実行準備完了

---

## テスト環境準備

### 前提条件
- Docker Compose で全コンテナ起動済み
- frontend: http://localhost:3003
- backend API: http://localhost:8080
- PostgreSQL: localhost:5434

### テストアカウント
| 項目 | 値 |
|------|-----|
| Email | admin@example.com |
| Password | password |
| 権限 | 管理者（フルアクセス） |

### ブラウザ準備
1. Chrome をデバッグモードで起動（オプション）:
   ```bash
   # macOS
   /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome \
     --remote-debugging-port=9222
   ```
2. または DevTools の Console タブでネットワークリクエストを監視
3. ページを開く: http://localhost:3003

---

## テスト実行手順

### 【テスト1】動物種マスタ (`/settings/animal-species`)

#### 1-1. CREATE テスト（新規登録）

**事前準備**: ページナビゲーション
```
URL: http://localhost:3003/settings/animal-species
```

**操作手順**:
1. 「新規登録」ボタンをクリック
   - 動物種の新規登録パネル/フォームが表示される
2. 動物種名フィールドに値を入力
   ```
   値: "テスト動物種" (例)
   ```
3. 「登録」ボタンをクリック

**確認項目**:
- [ ] HTTP POST `/api/v1/masters/animal-species` リクエスト送信される
- [ ] HTTP 201 Created ステータスコード返却される
- [ ] トースト「登録しました」表示される
- [ ] 一覧に新規項目が追加される（行数 +1）
- [ ] PostgreSQL で INSERT 確認:
  ```sql
  SELECT id, name, is_active FROM animal_species 
  WHERE name = 'テスト動物種' AND is_active = true;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 1-2. UPDATE テスト（編集・更新）

**事前準備**: CREATE テストで作成した「テスト動物種」が存在することを確認

**操作手順**:
1. 作成した「テスト動物種」行をクリックまたは操作メニュー「編集」をクリック
   - 編集パネルが表示される
   - 現在の値「テスト動物種」がフィールドに表示されている
2. 動物種名を変更
   ```
   値: "更新テスト動物種" (例)
   ```
3. 「更新」ボタンをクリック

**確認項目**:
- [ ] HTTP PATCH `/api/v1/masters/animal-species/{id}` リクエスト送信される
- [ ] HTTP 200 OK ステータスコード返却される
- [ ] トースト「更新しました」表示される
- [ ] 一覧の表示が「更新テスト動物種」に変更される
- [ ] PostgreSQL で UPDATE 確認:
  ```sql
  SELECT id, name FROM animal_species 
  WHERE name = '更新テスト動物種' AND is_active = true;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 1-3. DELETE テスト（削除）

**事前準備**: UPDATE テストで編集した「更新テスト動物種」が存在することを確認

**操作手順**:
1. 「更新テスト動物種」行の操作メニュー「削除」をクリック
   - 確認ダイアログが表示される
   ```
   メッセージ例: "この動物種を削除してもよろしいですか？"
   ```
2. 「削除する」ボタンをクリック

**確認項目**:
- [ ] HTTP DELETE `/api/v1/masters/animal-species/{id}` リクエスト送信される
- [ ] HTTP 204 No Content ステータスコード返却される
- [ ] トースト「削除しました」表示される
- [ ] 一覧から削除項目が消える（行数 -1）
- [ ] PostgreSQL で soft-delete 確認:
  ```sql
  SELECT id, name, is_active FROM animal_species 
  WHERE name = '更新テスト動物種';
  -- is_active = false であることを確認（または deleted_at が NULL でないことを確認）
  ```

**パス条件**: すべてのチェックボックスが✓

---

### 【テスト2】サービス種別マスタ (`/settings/reservation-type`)

#### 2-1. CREATE テスト（新規登録）

**事前準備**: ページナビゲーション
```
URL: http://localhost:3003/settings/reservation-type
```

**操作手順**:
1. 「新規登録」ボタンをクリック
   - サービス種別の新規登録パネル/フォームが表示される
2. 必須フィールドを入力
   ```
   - サービス種別名: "テスト予約種別"
   - 説明（オプション）: "テスト用のサービス種別です"
   - 色（オプション）: #3B82F6 (青)
   ```
3. 「登録」ボタンをクリック

**確認項目**:
- [ ] HTTP POST `/api/v1/masters/reservation-types` リクエスト送信される
- [ ] HTTP 201 Created ステータスコード返却される
- [ ] トースト「登録しました」表示される
- [ ] 一覧に新規項目が追加される（行数 +1）
- [ ] PostgreSQL で INSERT 確認:
  ```sql
  SELECT id, clinic_id, name, is_active FROM reservation_types 
  WHERE name = 'テスト予約種別' AND is_active = true;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 2-2. UPDATE テスト（編集・更新）

**事前準備**: CREATE テストで作成した「テスト予約種別」が存在することを確認

**操作手順**:
1. 作成した「テスト予約種別」行の編集メニューをクリック
   - 編集パネルが表示される
2. サービス種別名を変更
   ```
   値: "更新テスト予約種別" (例)
   ```
3. 「更新」ボタンをクリック

**確認項目**:
- [ ] HTTP PATCH `/api/v1/masters/reservation-types/{id}` リクエスト送信される
- [ ] HTTP 200 OK ステータスコード返却される
- [ ] トースト「更新しました」表示される
- [ ] 一覧の表示が「更新テスト予約種別」に変更される
- [ ] PostgreSQL で UPDATE 確認:
  ```sql
  SELECT id, name FROM reservation_types 
  WHERE name = '更新テスト予約種別' AND is_active = true;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 2-3. DELETE テスト（削除）

**事前準備**: UPDATE テストで編集した「更新テスト予約種別」が存在することを確認

**操作手順**:
1. 「更新テスト予約種別」行の削除メニューをクリック
   - 確認ダイアログが表示される
2. 「削除する」ボタンをクリック

**確認項目**:
- [ ] HTTP DELETE `/api/v1/masters/reservation-types/{id}` リクエスト送信される
- [ ] HTTP 204 No Content ステータスコード返却される
- [ ] トースト「削除しました」表示される
- [ ] 一覧から削除項目が消える（行数 -1）
- [ ] PostgreSQL で soft-delete 確認:
  ```sql
  SELECT id, name, is_active FROM reservation_types 
  WHERE name = '更新テスト予約種別';
  -- is_active = false であることを確認
  ```

**パス条件**: すべてのチェックボックスが✓

---

### 【テスト3】スタッフマスタ (`/settings/staff`)

#### 3-1. CREATE テスト（新規登録）

**事前準備**: ページナビゲーション
```
URL: http://localhost:3003/settings/staff
```

**操作手順**:
1. 「新規登録」ボタンをクリック
   - スタッフの新規登録パネン/フォームが表示される
2. 必須フィールドを入力
   ```
   - スタッフ名: "テストスタッフ"
   - スタッフタイプ（職種）: "doctor" or "nurse" (ドロップダウン選択)
   - ステータス: "active" (トグル ON)
   ```
3. 「登録」ボタンをクリック

**確認項目**:
- [ ] HTTP POST `/api/v1/masters/staffs` リクエスト送信される
- [ ] HTTP 201 Created ステータスコード返却される
- [ ] トースト「登録しました」表示される
- [ ] 一覧に新規項目が追加される（行数 +1）
- [ ] PostgreSQL で INSERT 確認:
  ```sql
  SELECT id, name, staff_type, is_active FROM staffs 
  WHERE name = 'テストスタッフ' AND is_active = true AND deleted_at IS NULL;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 3-2. UPDATE テスト（編集・更新）

**事前準備**: CREATE テストで作成した「テストスタッフ」が存在することを確認

**操作手順**:
1. 作成した「テストスタッフ」行の編集メニューをクリック
   - 編集パネルが表示される
2. スタッフ名を変更
   ```
   値: "更新テストスタッフ" (例)
   ```
3. 「更新」ボタンをクリック

**確認項目**:
- [ ] HTTP PATCH `/api/v1/masters/staffs/{id}` リクエスト送信される
- [ ] HTTP 200 OK ステータスコード返却される
- [ ] トースト「更新しました」表示される
- [ ] 一覧の表示が「更新テストスタッフ」に変更される
- [ ] PostgreSQL で UPDATE 確認:
  ```sql
  SELECT id, name FROM staffs 
  WHERE name = '更新テストスタッフ' AND is_active = true AND deleted_at IS NULL;
  ```

**パス条件**: すべてのチェックボックスが✓

---

#### 3-3. DELETE テスト（削除）

**事前準備**: UPDATE テストで編集した「更新テストスタッフ」が存在することを確認

**操作手順**:
1. 「更新テストスタッフ」行の削除メニューをクリック
   - 確認ダイアログが表示される
2. 「削除する」ボタンをクリック

**確認項目**:
- [ ] HTTP DELETE `/api/v1/masters/staffs/{id}` リクエスト送信される
- [ ] HTTP 204 No Content ステータスコード返却される
- [ ] トースト「削除しました」表示される
- [ ] 一覧から削除項目が消える（行数 -1）
- [ ] PostgreSQL で soft-delete 確認:
  ```sql
  SELECT id, name, is_active, deleted_at FROM staffs 
  WHERE name = '更新テストスタッフ';
  -- is_active = false AND deleted_at IS NOT NULL であることを確認
  ```

**パス条件**: すべてのチェックボックスが✓

---

## テスト結果報告テンプレート

テスト実行後、以下の形式で結果を報告してください：

```markdown
## テスト結果報告: Section 14 マスタ設定（動物種・サービス種別・スタッフ）

実行日時: 2026-04-12 HH:MM:SS
実行者: [名前]
環境: ローカル (localhost:3003)

### 1. 動物種マスタ (`/settings/animal-species`)

| テスト項目 | 結果 | 詳細・備考 |
|-----------|------|----------|
| CREATE | OK/NG/PARTIAL | HTTP 201, 一覧反映, DB INSERT 確認 |
| UPDATE | OK/NG/PARTIAL | HTTP 200, 一覧表示更新, DB UPDATE 確認 |
| DELETE | OK/NG/PARTIAL | HTTP 204, 一覧削除, DB soft-delete (is_active=false) 確認 |

### 2. サービス種別マスタ (`/settings/reservation-type`)

| テスト項目 | 結果 | 詳細・備考 |
|-----------|------|----------|
| CREATE | OK/NG/PARTIAL | HTTP 201, 一覧反映, DB INSERT 確認 |
| UPDATE | OK/NG/PARTIAL | HTTP 200, 一覧表示更新, DB UPDATE 確認 |
| DELETE | OK/NG/PARTIAL | HTTP 204, 一覧削除, DB soft-delete (is_active=false) 確認 |

### 3. スタッフマスタ (`/settings/staff`)

| テスト項目 | 結果 | 詳細・備考 |
|-----------|------|----------|
| CREATE | OK/NG/PARTIAL | HTTP 201, 一覧反映, DB INSERT 確認 |
| UPDATE | OK/NG/PARTIAL | HTTP 200, 一覧表示更新, DB UPDATE 確認 |
| DELETE | OK/NG/PARTIAL | HTTP 204, 一覧削除, DB soft-delete (deleted_at NOT NULL) 確認 |

### 総括

| 項目 | 数 |
|------|-----|
| テスト項目数 | 9件 |
| OK | X件 |
| NG | Y件 |
| PARTIAL | Z件 |

### 発見されたバグ

**NG 項目がある場合のみ記載**

1. **BUG-XXX**: [バグ説明]
   - 発生場所: [ページ/操作]
   - 期待動作: [期待値]
   - 実際の動作: [実際の値]
   - 再現手順: [手順]

### テスト環境スクリーンショット

**エラーが発生した場合のみ添付**

- NG 1: [ファイル名]
- NG 2: [ファイル名]

---

**実行完了日**: 2026-04-12
**ステータス**: ✅ PASS / ⚠️ PARTIAL / ❌ NG
```

---

## よくある問題と対応

### Q1. 「登録」ボタンをクリックしても何も起こらない
**A**: 
1. DevTools Console でエラーが出ていないか確認
2. Network タブで POST リクエストが送信されているか確認
3. バックエンド API が起動しているか確認: `docker compose ps`

### Q2. トースト（通知）が表示されない
**A**:
1. 画面右上を確認（デフォルト位置）
2. DevTools Console でトースト関連のエラーがないか確認
3. 画面をスクロールして隠れていないか確認

### Q3. HTTP 409（Conflict）エラーが出た
**A**:
- FK 依存関係により削除できない場合
- 例: ペットで参照されている動物種は削除不可
- 詳細は Console に出力される

### Q4. PostgreSQL でデータが見つからない
**A**:
1. clinic_id フィルタを確認（サービス種別は clinic_id = 1 が必須）
2. is_active = true / deleted_at IS NULL フィルタを確認
3. クエリを実行するDBユーザーを確認

---

## その他の参考資料

- **API ドキュメント**: `backend/docs/api.yaml`
- **DB スキーマ**: `docs/ERD.md`
- **テストレポート**: `docs/FUNCTIONAL_TEST_REPORT.md`
- **テスト実装**: `backend/internal/handler/*_handler_test.go`


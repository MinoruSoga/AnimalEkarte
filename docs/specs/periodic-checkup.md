# 定期健診 (Periodic Checkup) — 仕様書

> 作成: 2026-05-21 / Issue #53 対応

---

## 1. 用語定義

| 用語 | 定義 |
|------|------|
| **定期健診** | `checkups` テーブルの 1 レコード。`medical_records` のサブリソース。ワクチン接種や検査とは別テーブル |
| **健診種別** | `checkup_types` マスタ（クリニック単位）。名称・価格・推奨インターバル・対象年齢・階層構造を持つ |
| **次回予定日** | `checkups.next_date`（nullable）。健診アラート (`GetAlerts`) の判定基準 |
| **健診アラート** | `next_date < today` → Overdue / `today ≤ next_date ≤ today+N` → Upcoming |
| **CheckupSync** | 別機能。「次回健診予定者」に一括 LINE 配信するバッチ (`/lstep/checkup-sync`)。健診登録そのものとは独立 |

---

## 2. 登録エントリーポイント

### Path A — カルテ経由（メイン）

```
/medical-records/:id → "定期健診"タブ → CheckupsTab
```

- 既存カルテに checkup サブリソースを追加 / 編集 / 削除
- API: `GET /v1/medical-records/{id}/checkups`
- API: `POST /v1/medical-records/{id}/checkups`
- API: `PATCH /v1/medical-records/{id}/checkups/{checkupId}`
- API: `DELETE /v1/medical-records/{id}/checkups/{checkupId}`

**現状の入力項目:**

| フィールド | 必須 | 現状 |
|-----------|------|------|
| 健診種別 | ✅ | セレクト（checkup_types マスタ） |
| 実施日 | ✅ | 日付ピッカー |
| 次回予定日 | ❌ | 日付ピッカー |
| 担当医 | ❌ | **AddForm: なし / EditRow: `-` 固定（⚠ 入力不可）** |
| 結果 | ❌ | テキストエリア |

### Path B — 独立健診登録

```
/checkups/new → CheckupForm → useCheckupForm
```

- カルテが存在しない状態から健診を登録する際に使用
- 内部で **カルテを自動生成**してから checkup を登録する 2 ステップ処理

```typescript
// Step 1: カルテ自動生成（visit_type なし）
const { data: medicalRecord } = await axios.post("/v1/medical-records", {
  pet_id: pet.id,
  owner_id: pet.ownerId,
  visit_date: formData.date,
  // ⚠ visit_type が設定されない
  // ⚠ chief_complaint が空
});

// Step 2: 健診記録登録
await axios.post(`/v1/medical-records/${medicalRecord.id}/checkups`, { ... });
```

**現状の入力項目:**

| フィールド | 必須 | 現状 |
|-----------|------|------|
| ペット（+ 飼主） | ✅ | 検索セレクト |
| 健診種別 | ✅ | セレクト |
| 実施日 | ✅ | 日付ピッカー |
| 次回予定日 | ❌ | 日付ピッカー |
| 担当医 | ❌ | セレクト（⚠ Path A と不整合） |
| 結果 | ❌ | テキストエリア |

---

## 3. DB スキーマ（関連テーブル）

```sql
-- checkups
CREATE TABLE checkups (
  id               bigserial PRIMARY KEY,
  clinic_id        bigint    NOT NULL,
  medical_record_id bigint   NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
  pet_id           bigint,              -- nullable（Path A から疑わしい）
  checkup_type_id  bigint    NOT NULL REFERENCES checkup_types(id)   ON DELETE RESTRICT,
  date             date      NOT NULL,
  next_date        date,                -- nullable
  doctor_id        bigint,              -- nullable
  result           text      NOT NULL DEFAULT '',
  created_at, updated_at, deleted_at
);

-- checkup_types (クリニック単位マスタ)
CREATE TABLE checkup_types (
  id          bigserial PRIMARY KEY,
  clinic_id   bigint    NOT NULL,
  name        text      NOT NULL,
  price       bigint,                  -- nullable
  is_active   boolean   NOT NULL DEFAULT true,
  description text      NOT NULL DEFAULT '',
  interval    text      NOT NULL DEFAULT '',   -- 例: "1年", "6ヶ月"
  target_age  text      NOT NULL DEFAULT '',   -- 例: "1歳以上"
  parent_id   bigint REFERENCES checkup_types(id), -- 階層サポート
  sort_order  int       NOT NULL DEFAULT 0
);

-- medical_records（is_periodic_checkup フラグなし）
-- 定期健診かどうかは checkups サブリソースの有無でのみ判断可能
```

---

## 4. API 仕様（現状）

### 健診一覧・取得
```
GET /v1/medical-records/{medicalRecordId}/checkups
```

### 健診登録
```
POST /v1/medical-records/{medicalRecordId}/checkups

Body:
{
  "checkup_type_id": number,    // 必須
  "date":           "YYYY-MM-DD", // 必須
  "next_date":      "YYYY-MM-DD" | null, // 任意
  "doctor_id":      number | null,        // 任意
  "result":         string               // 任意
}
```

### 健診更新
```
PATCH /v1/medical-records/{medicalRecordId}/checkups/{checkupId}

Body: 同上（全フィールド任意）
```

### 健診削除
```
DELETE /v1/medical-records/{medicalRecordId}/checkups/{checkupId}
```

### 健診アラート（一覧画面用）
```
GET /v1/clinics/{clinicId}/checkup-alerts?days=30
```

### クリニック横断健診一覧
```
GET /v1/clinics/{clinicId}/checkups
フィルタ: status(overdue/upcoming), start_date, end_date
ソート: date, owner_name, pet_name, checkup_type, next_date
```

---

## 5. Lステップ連携

### TriggerCheckupFollowUp（健診作成後フォローアップ）

- `checkup_service.go Create()` 内で goroutine 起動（非同期・非致命的）
- **現在スタブ状態**（FEAT-383 Phase 2 として切り出し済み）
- 将来: 健診後 N 日でフォローアップ LINE 配信予定

### CheckupSync（定期健診バッチ配信）

- 別機能。`/lstep/checkup-sync` エンドポイント
- 「次回健診予定者」を条件フィルタして一括 LINE 配信
- checkup 登録とは独立（登録トリガーではなくバッチ処理）

---

## 6. 入力→保存→表示の例

### 例: ワンちゃんの年1回健康診断（Path B）

1. `/checkups/new` を開く
2. ペット "ポチ"、健診種別 "年間健康診断"、実施日 "2026-05-21"、次回予定日 "2027-05-21" を入力
3. 保存
4. **内部処理**:
   - `POST /v1/medical-records` → `medical_record.id = 100`（visit_type なし）
   - `POST /v1/medical-records/100/checkups`
5. `/checkups` 一覧で表示（次回予定日まで 365 日 → Upcoming）
6. 2027-05-22 以降 → Overdue アラート

### 例: 既存カルテへの追加（Path A）

1. `/medical-records/100` → "定期健診"タブ
2. "健診を追加" → 種別・日付・結果を入力
3. `POST /v1/medical-records/100/checkups`
4. 同カルテに複数健診を紐付け可能（重複防止バリデーションなし）

---

## 7. 現状の実装不整合・未決事項

### [Q1] 担当医フィールドの不整合（HIGH）

- Path A (CheckupsTab): AddForm に担当医なし、EditRow は `-` 固定（入力不可）
- Path B (CheckupForm): 担当医入力可能
- **決定が必要**: 担当医入力を Path A でも有効化するか、または「担当医は不要」として Path B から削除するか

### [Q2] Path B のカルテ自動生成に `visit_type` がない（HIGH）

- 自動生成カルテの `visit_type` が NULL → カルテ一覧でどう表示/区別するか不明
- **推奨**: `visit_type = "checkup"` または `"health_check"` を固定設定する
- または、自動生成カルテを「定期健診専用カルテ」として `is_periodic_checkup` フラグを `medical_records` に追加する

### [Q3] `finalized` カルテでの健診タブ編集ロック（MEDIUM）

- 現在 `CheckupsTab` は `finalized` カルテでも編集・削除ボタンが表示される
- 会計確定後の健診追加・削除は会計金額と乖離する可能性
- **推奨**: `medical_record.status === "finalized"` 時は追加・編集・削除ボタンを非表示または disabled

### [Q4] 同一カルテへの同一健診種別の重複登録（LOW）

- 現状バリデーションなし（同一 `checkup_type_id` を同カルテに複数登録可能）
- 意図的な仕様（例: 6ヶ月に1回を同日2回記録したい場合）か、バグ防止が必要か

### [Q5] `next_date` のクリア方法（LOW）

- `buildCheckupUpdate` は pointer nil 判定のみ
- フロントから `next_date: null` を送信して null 化できるか確認が必要

### [Q6] `medical_records` の「定期健診」識別（MEDIUM）

- `is_periodic_checkup` フラグなし
- カルテ一覧で「このカルテは定期健診」と表示する UI 要件があるか確認が必要

### [Q7] `pet_id` の nullable（LOW）

- `checkups.pet_id` が nullable だが、実際に NULL になるユースケースが不明
- `medical_records.pet_id` は NOT NULL なので、checkup 登録時は必ず pet_id を持てるはず

### [Q8] CheckupsList の表示データソース（LOW）

- `/checkups` 一覧は `checkups` テーブルの直接取得か、`checkup_sync` 経由のキャッシュか
- 現状: `checkup_service.GetByClinic()` による直接取得（checkup_sync とは独立）

---

## 8. 推奨仕様案（最小変更・既存互換）

**原則**: 既存データ構造を変えず、UI の不整合を修正する。

### A. 担当医フィールド整合 → Path A に追加

- `CheckupsTab` の AddForm と EditRow に担当医セレクトを追加
- DB 変更なし（`checkups.doctor_id` は既に存在）

### B. Path B カルテ自動生成に `visit_type` 設定

- `useCheckupForm.ts` の `POST /v1/medical-records` に `visit_type: "health_check"` を追加
- これにより「健診由来カルテ」をフィルタ可能になる（将来の一覧UI改善に対応）

### C. `finalized` カルテの健診タブ編集ロック

- `MedicalRecordForm.tsx` から `medicalRecord.status` を `CheckupsTab` に渡す
- `status === "finalized"` 時は追加・編集・削除ボタンを `disabled` / 非表示

### D. `is_periodic_checkup` フラグ追加は **見送り**

- `checkups` サブリソースの有無で「定期健診カルテ」の識別は可能
- migration コストに対してメリットが小さい

---

## 9. 実装タスク（follow-up）

| Issue | 内容 | 優先度 |
|-------|------|--------|
| TBD-A | [MedicalRecords] Add doctor field to CheckupsTab (Path A) | MEDIUM |
| TBD-B | [Checkups] Set visit_type on auto-generated medical record in Path B | MEDIUM |
| TBD-C | [MedicalRecords] Lock checkups tab when medical record is finalized | LOW |

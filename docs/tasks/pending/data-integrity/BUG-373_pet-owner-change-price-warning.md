# BUG-373: 飼主変更時の金額変動警告ポップアップ

**作成日**: 2026-04-14
**Status**: Pending（仕様未確定）
**Priority**: Medium（データ整合性・UX）
**Affects**: ペット編集モーダル / カルテ詳細画面の飼主変更機能

**依頼元（原文）**:

> カルテ入力
> 複製する際にポップアップなどで金額に関する注意がでるか
> コピー時にポップアップなどで金額が変わっている可能性があるので注意、などのアナウンスが可能か

**ユーザー補足（原文）**:

> 複製というよりは飼主の変更です。
> ペット情報にて、飼主変更する処理はあるでしょ？

---

## 概要

ペットの飼主を変更する機能（2箇所）では、変更後に紐づく会計・見積の金額が変動する可能性があるにもかかわらず、確認モーダルや警告が一切表示されない。変更者が金額変動リスクを認識しないまま操作できてしまう。

## 飼主変更機能の実装箇所（2箇所）

| 画面 | 実装ファイル | 現状の動作 |
|------|------------|----------|
| ペット編集モーダル（飼主詳細画面内） | `frontend/src/features/owners/routes/OwnerForm.tsx:527-543` + `frontend/src/features/owners/components/PetEditModal.tsx:272-280` | `OwnerSearchModal` で飼主選択 → 確認なしで `updatePetMutate` 実行 → トースト「飼主を ◯◯ に変更しました」 |
| カルテ詳細画面 | `frontend/src/features/medical-records/hooks/use-medical-record-form.ts:258-277` + `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:556-566` | `OwnerSearchModal` で飼主選択 → 確認なしで `updateMutation` 実行（owner_id のみ更新） → トースト「飼主を ◯◯ に変更しました」 |

## 金額変動の根拠

飼主ごとに以下のフィールドが異なる可能性がある:

| フィールド | 影響 | 参照 |
|----------|------|-----|
| `Owner.discount_rate` | 値引率が異なる（社販運用含む） | `backend/internal/model/owner.go`, `owner_service.go:413` |
| `Owner.membership_type` | 会員/非会員で価格体系が変わる | `backend/internal/model/owner.go` |

飼主変更 = これらの値が切り替わる = 今後の会計・見積の金額が変動する可能性。

## 仕様未確定事項（Q1-Q4）

### Q1: 対象範囲
- (A) **両方**（ペット編集モーダル + カルテ画面の両方に警告追加）
- (B) ペット編集モーダルのみ
- (C) カルテ画面のみ

### Q2: 警告条件
- (A) **常時警告**（飼主選択直後、全ケースで確認モーダル表示）
- (B) **条件付き**（旧飼主と新飼主の `discount_rate` または `membership_type` が異なる場合のみ警告）

### Q3: モーダルの操作
- (A) 「続行 / キャンセル」の二択確認モーダル（続行しないと変更されない）
- (B) 変更は即実行、事後に注意トースト表示のみ

### Q4: 警告文言（任意）
- デフォルト案: 「飼主によって値引率や会員区分が異なるため、今後の会計金額が変動する可能性があります。変更を続行してよろしいですか?」

### Q5: 既存会計への影響
- 変更対象のペットに既に会計レコードが存在する場合、過去の会計の飼主は変わるか？ それとも今後の会計のみ影響するか？

---

## 仕様確認後のサブタスク（想定）

| # | サブタスク | 領域 | 依存 |
|---|----------|------|------|
| 1 | `ConfirmDialog` をペット編集モーダル `PetEditModal.tsx:272-280` のボタン onClick にラップ | FE | - |
| 2 | `ConfirmDialog` をカルテ画面 `MedicalRecordForm.tsx:556-566` の `handleChangeOwner` にラップ | FE | - |
| 3 | Q2=B の場合: 新飼主の `discount_rate`/`membership_type` を `OwnerSearchModal` の選択結果から取得 | FE | - |

想定影響範囲:
- BE 変更不要（`OwnerSearchModal` の onSelect payload に `discount_rate` / `membership_type` を含めるだけで対応可。既存 API レスポンスに既に含まれている）
- FE のみの変更で完結する見込み

## 再開条件

Q1-Q3 の回答が得られ次第、`docs/tasks/open/data-integrity/` に移動し FE-251 を起票する。

## 関連メモリー

- `memory/domain_employee_internal_sale.md` — 社販は `Owner.discount_rate` で運用（本 BUG の金額変動要因の一つ）

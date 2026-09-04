# ペット選択 仕様書

## 概要
- **画面の目的**: 診療記録、入院、トリミング、会計等の各機能において、新規データ作成の対象となるペットを検索・特定するための中間ページ。
- **URLパターン**: `/:feature/select-pet`
  - 例: `/medical-records/select-pet`, `/hospitalization/select-pet` 等
  - 対象 feature: medical-records / hospitalization / trimming / examinations / checkups / accounting / vaccinations
- **アクセス権限**: 通常は遷移元機能の`<Resource>:create`権限。例外としてcheckupsは`ResourceMedicalRecords:create` **かつ** `ResourceMedicalRecords:edit`を要求する（`RequirePermission`）。

---

## 画面構成

### 1. 検索契約（実装正本）

- **入力**: ラベル「検索（ペット名・飼主名・よみ・電話）」の単一フリーテキスト `search`、任意の「飼主No」（`owner_id` = owners.id）、種別はマスタ id の select（種別名テキスト部分一致ではない）。住所フィールドは無い（BUG-451）。
- **取得**: サーバ側 page（20 件）+ 300ms debounce（テナント全ペット初期ロードではない）。死亡ペットも一覧には出す（`includeDeceased: true`）が選択不可。危険度高は Popover。
- **結果列**: 飼主No／飼主名／ペット番号／ペット名／生死／種／生年月日／体重／環境／前回来院／操作。
- **権限**: 通常は遷移元機能の`<Resource>:create`。checkupsは`ResourceMedicalRecords:create` **かつ** `ResourceMedicalRecords:edit`。


# BUG-367: 領収書が適格請求書の法的要件を満たさない — 明細兼領収書への統合

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: CRITICAL（法的コンプライアンス）
**Affects**: `features/accounting` — AccountingDocument, AccountingDetail

## 実装結果（2026-04-14）

### 完了
- BE: `MeClinicInfo` に `StandardTaxRate` / `ReducedTaxRate` を追加（`auth_response.go`）
- BE: `buildMeResponse` で税率を設定（`auth_handler.go`）
- BE: `make codegen` 実行
- FE: `AuthClinic` に `standardTaxRate` / `reducedTaxRate` 追加（`auth/types/index.ts`）
- FE: zod schema + `mapMeClinicInfo` に税率 2 フィールド追加（`auth/api/transforms.ts`）
- FE: `AccountingDocument.tsx` を **明細兼領収書（A4 単一帳票）** に書き換え
  - `type` prop 廃止（receipt / statement 2 帳票統合）
  - `approxEqual` で軽減税率判定（`item.taxRate === clinic.reducedTaxRate`）
  - 品目テーブルに税率列 + 軽減税率品目に「※」マーク
  - 税率別内訳: 標準税率 + 軽減税率（対象あれば）
  - 登録番号未設定時の警告（AC-6）+ 帳票内に「登録番号: 未設定」明示
  - 適格請求書要件 ①〜⑦ すべて網羅
- FE: `AccountingDetail.tsx`
  - `previewType` state 削除
  - `handlePrint` 引数削除（単一帳票化）
  - 印刷ボタン 2 個 → 1 個（「明細兼領収書」）
  - Dialog タイトル固定化
- build / lint: 成功（エラー 0）

### 設計上の注意
- `invoiceRegistrationNumber` は `companies.invoice_registration_number`（会社マスタ）から取得（既存動作維持）
- `Math.abs(a - b) < 0.0001` で税率比較の浮動小数誤差を吸収

### 未完了（別イシュー化候補）
- 80mm サーマル運用確認（全病院 A4 プリンタ有無）
- 病院設定で登録番号必須バリデーション追加
- 品目側への税区分保存（過去データ判定保全）
**依頼元（原文）**:

> 領収書の内容では税に関する記載がないため法的に問題がある　明細兼領収書にしてインボイスの番号を入れるようにしてほしい

---

## 概要

現在の領収書（80mm サーマル印刷）は、品目明細・税率別内訳・税額の記載がなく、**適格請求書等保存方式（インボイス制度）の法的要件を満たしていない**。動物病院は簡易適格請求書対象業種ではないため、通常の適格請求書の全要件が必要。要望に従い、既存の領収書（80mm）と診療明細書（A4）を廃止し、**「明細兼領収書」A4 1種類に統合**する。

## 現状のコード

`frontend/src/features/accounting/components/AccountingDocument.tsx:54-106`

現状の領収書には以下が**欠落**している:

```typescript
// 領収書 (receipt) 80mm
<div className="flex justify-between border-b border-dotted pb-1">
  <span>診療費等合計</span>
  <span>¥{paymentInfo.totalAmount.toLocaleString()}</span>
</div>
// ... 税率別内訳・税額・品目明細すべてなし
{clinic?.invoiceRegistrationNumber ? <p>登録番号: {clinic.invoiceRegistrationNumber}</p> : null}
```

`frontend/src/features/accounting/routes/AccountingDetail.tsx:883`

```typescript
const [previewType, setPreviewType] = useState<"receipt" | "statement">("receipt");
```

## 適格請求書の法的要件と現状ギャップ

| 要件（消費税法第 57 条の 4） | 現状の領収書 | 現状の明細書 |
|----------------------------|------------|------------|
| ① 発行事業者の氏名・**登録番号** | ✅ | ✅ |
| ② 取引年月日 | ✅ | ✅ |
| ③ 取引内容（品目名） | ❌ 欠落 | ✅ |
| ④ 軽減税率対象品である旨（※記号） | ❌ 欠落 | ❌ 欠落 |
| ⑤ 税率ごとに区分した対価の合計額 | ❌ 欠落 | ✅（10% / 8%） |
| ⑥ 税率ごとの消費税額 | ❌ 欠落 | ✅ |
| ⑦ 書類交付先の氏名・名称（宛名） | ✅（owner名） | ✅ |

**領収書が③④⑤⑥を満たしていない** → 単独では適格請求書として無効。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 帳票形態 | (A) 明細兼領収書 A4 1種類に統合（既存 領収書80mm / 明細書A4 は廃止） |
| 2 | 軽減税率の税区分設定 | `clinics.standard_tax_rate`(0.10) / `clinics.reduced_tax_rate`(0.08) がクリニック単位で既存。品目側は `billing_items.tax_rate` 等の数値フィールドのみ。軽減税率判定は `item.tax_rate === clinic.reduced_tax_rate` で行う |
| 3 | 宛名 | 既存 `ownerName`（○○様）で問題なし |

## 受入条件（Acceptance Criteria）

- [ ] **AC-1**: 会計詳細画面で「明細兼領収書」ボタンをクリックすると、A4 1枚の統合帳票がプレビュー表示される
- [ ] **AC-2**: 帳票に以下がすべて含まれる
  - タイトル「明細兼領収書」
  - 発行日、発行番号
  - 発行事業者名・住所・電話番号
  - **登録番号（T + 13桁）** ※`clinics.invoice_registration_number` が空の場合は未表示ではなくエラー表示（下記 AC-6）
  - 宛名（○○様）+ ペット名
  - 品目明細テーブル（項目名・単価・数量・金額・税率マーク）
  - **軽減税率対象品（tax_rate = reduced_tax_rate）の品目名横に「※」**
  - 税率別内訳: 「10%対象 ¥XXX（内 消費税 ¥XXX）」「8%対象 ¥XXX（内 消費税 ¥XXX、※軽減税率）」
  - 合計金額（税込）
  - 保険適用額（あれば）
  - 請求金額
  - お預かり額・お釣り
- [ ] **AC-3**: 旧「領収書」「診療明細書」ボタンは UI から削除される（`previewType` state 廃止）
- [ ] **AC-4**: 軽減税率対象品が 0 件の場合、税率別内訳セクションは「10%対象」行のみ表示（8% 行は非表示）
- [ ] **AC-5**: `window.print()` で実際の用紙サイズ A4 で印刷され、レイアウトが崩れない
- [ ] **AC-6**: `clinic.invoice_registration_number` が未設定の場合、画面に警告「登録番号が未設定です。病院設定から登録してください」を表示し、印刷ボタンは押下可能だが文書には「登録番号: 未設定」と明示（インボイス無効となるため運用者に気づかせる）
- [ ] **AC-7**: 税額計算は現状の `Math.floor`（税率別合計額 × 税率 → 切り捨て）を継続

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 帳票タイプ管理 | 単一帳票（type prop 削除） | 仕様確認 Q1 で (A) を選択 | receipt/statement 併存 |
| 軽減税率判定ロジック | `item.tax_rate === clinic.reduced_tax_rate` で判定 | 既存データモデルに影響を与えない。clinic 単位で税率が設定済 | 新規に `is_reduced_tax_rate` フラグ追加 → DB変更必要 |
| 税額計算 | `Math.floor`（既存踏襲） | 既存 `AccountingDocument.tsx:49-50` と整合性維持 | 四捨五入変更 → 既存データとの整合性が崩れる |
| 用紙サイズ | A4 | 仕様確認 Q1 で統一帳票に決定 | 80mm 継続 → 品目明細を載せるスペースなし |

## 影響範囲

### Frontend
- `frontend/src/features/accounting/components/AccountingDocument.tsx` — 全面書き換え（`type` prop 削除、A4統合レイアウト実装、軽減税率マーク追加）
- `frontend/src/features/accounting/routes/AccountingDetail.tsx:883` — `previewType` state 削除、`setPreviewType` 呼び出し除去、ダイアログタイトル固定化、印刷ボタン2つを1つに統合
- `frontend/src/features/accounting/types/` or `frontend/src/features/hospital-settings/` — Clinic 型に `reducedTaxRate` / `standardTaxRate` が既にあるか確認し、`ClinicInfo` prop に追加

### Backend
- **変更なし**。`clinics.standard_tax_rate` / `clinics.reduced_tax_rate` は既存。`clinic_response.go:23-24` に既に露出済み。`companies.invoice_registration_number` は既存。

### DB
- **変更なし**

## 参照実装

- `frontend/src/features/accounting/components/AccountingDocument.tsx:109-194` — 現状の statement (A4) のレイアウトをベースに拡張する
- `frontend/src/features/hospital-settings/` — `clinic.reducedTaxRate` / `clinic.standardTaxRate` の取得方法（既存のどこで取得されているか確認）

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 80mm サーマルプリンタで運用中の病院が A4 印刷に切り替えできない | **高** | 運用確認必要。A4 プリンタが全店舗にあるかユーザーに確認。無ければ80mmレイアウトの拡張版も検討 |
| 軽減税率判定を `tax_rate === reduced_tax_rate` の数値比較で行うと浮動小数誤差リスク | 中 | `Math.abs(a - b) < 0.0001` 形式で比較。将来的には `is_reduced_tax_rate` boolean 列の追加を別 issue で検討 |
| 過去の会計データで `clinic.reduced_tax_rate` が変更されると、過去分の判定が狂う | 中 | 本 issue では対応外。将来的にアイテム側に税区分を保存する設計変更を別 issue で起票 |
| 登録番号未設定での印刷運用 | 高 | AC-6 で警告表示。別途、病院設定画面で必須バリデーションを検討（別 issue） |

## 未解決事項

- [ ] 運用確認: 80mm サーマルプリンタでの運用継続が必要な病院があるか
- [ ] 軽減税率対象品が実際の運用でどれだけ存在するか（ペットフード処方等）— 対象が 0 件なら AC-4 により 8% 行は非表示となる

## 実装順序

1. `ClinicInfo` prop に `standardTaxRate` / `reducedTaxRate` を追加
2. `AccountingDocument.tsx` を単一帳票レイアウトに書き換え（type prop 削除）
3. 軽減税率判定ロジックと「※」マーク表示
4. `AccountingDetail.tsx` の previewType state / 切替 UI 削除
5. 登録番号未設定時の警告表示（AC-6）
6. 印刷確認（A4 実機）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] デザイントークン `C`, `STYLE` 使用（Hex 直指定禁止）
- [ ] `memo()` 維持
- [ ] `useMemo` で税計算キャッシュ維持

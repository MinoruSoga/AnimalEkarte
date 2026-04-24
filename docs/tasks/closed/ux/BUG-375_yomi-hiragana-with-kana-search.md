# BUG-375: 飼主・ペットの「よみ」をひらがな運用に統一（検索はひらがな・カタカナ両対応）

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: Medium（UX 改善・運用入力負荷低減）
**Affects**: `owners.name_kana`, `pets.name_kana`, owner/pet 検索, 飼主/ペット入力フォーム

**依頼元（原文）**:

> 飼主情報のフリガナですが、カタカナじゃなくてひらがなにしてください
>
> 補足: 検索においてはひらがな、カタカナ両方機能するようにしてください。
>       飼主名よみ、ペット名よみ　の項目ですね

---

## 概要

飼主・ペット双方の `name_kana` の運用文字種をカタカナからひらがなに変更する。

- UI ラベルを「**飼主名よみ / ペット名よみ**」に変更（旧「カナ」廃止）
- placeholder と既存データはひらがな化
- **検索はひらがな・カタカナ両方でヒットする**（`translate()` で正規化比較）
- 入力時の文字種制約は **設けない**（漢字・英数字も許可）

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 既存データの扱い | **(Y) 一括ひらがな変換マイグレーション実行** |
| 2 | 対象範囲 | **飼主 + ペット 両方** |
| 3 | 入力バリデーション | **(C) 文字種制約なし**（漢字・英数字も許可） |
| 4 | UI ラベル文言 | **「飼主名よみ」「ペット名よみ」** |
| 5 | DB カラム名 | **`name_kana` のまま維持** |
| 6 | placeholder | ひらがな例（「はやし ふみあき」等） |
| 7 | 検索 | **ひらがな・カタカナ両対応**（PostgreSQL `translate()` で正規化比較） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | 既存 `name_kana` のカタカナ→ひらがな一括変換マイグレーション | DB | BE-113 | - | [ ] |
| 2 | owner / pet 検索の ひらがな⇔カタカナ正規化（`translate()` 適用、owners.name_kana 検索対象追加） | BE | BE-114 | - | [ ] |
| 3 | UI ラベル「カナ」→「よみ」変更 + placeholder + エラーメッセージ更新（owner / pet / 検索フォーム） | FE | FE-251 | - | [ ] |

## 受入条件（Acceptance Criteria）

### DB（マイグレーション）

- [ ] **AC-1**: マイグレーション適用後、`owners.name_kana` のカタカナ文字（U+30A1〜U+30F6）が **すべてひらがな** (U+3041〜U+3096) に変換されている
- [ ] **AC-2**: `pets.name_kana` も同様に変換済み
- [ ] **AC-3**: マイグレーションは冪等（複数回実行しても結果が変わらない）
- [ ] **AC-4**: シードデータ `003_seed_demo.sql` の `name_kana` 値もひらがなに更新済み

### Backend（検索正規化）

- [ ] **AC-5**: `GET /v1/owners?search=ハヤシ` でカタカナ検索しても、`name_kana='はやし'` の飼主がヒットする
- [ ] **AC-6**: `GET /v1/owners?search=はやし` でひらがな検索しても、過去にカタカナで入力されたデータがあった場合（テスト用）でもヒットする
- [ ] **AC-7**: owner 検索対象に `name_kana` を追加（現状は `name`, `phone`, `email` のみ）
- [ ] **AC-8**: pet 検索の `pets.name_kana` も同様に正規化適用 + `owners.name_kana` を検索対象に追加

### Frontend（UI）

- [ ] **AC-9**: 飼主編集フォームのラベル: **「飼主名(カナ) *」 → 「飼主名よみ *」**
- [ ] **AC-10**: ペット編集フォームのラベル: 「ペット名(カナ)」→「ペット名よみ」
- [ ] **AC-11**: placeholder: 「ハヤシ フミアキ」→「はやし ふみあき」（owner / pet / 検索フォーム すべて）
- [ ] **AC-12**: バリデーションエラー文言: 「飼主名（カナ）を入力してください」→「飼主名よみを入力してください」
- [ ] **AC-13**: 入力欄に カタカナ・ひらがな・漢字・英数字 いずれを入力しても保存できる（バリデーション拒否なし）
- [ ] **AC-14**: PetSelectionSearchForm / PatientSelectionTable / 予約フォーム等の検索フォーム placeholder も ひらがな化

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 既存データ変換 | **PostgreSQL `translate()` 1 文字単位マッピング** | カタカナ U+30A1〜U+30F6 (86 文字) ↔ ひらがな U+3041〜U+3096 (86 文字) を 1 対 1 で対応可。標準関数のみで完結 | アプリ層変換 → マイグレーションランチャ複雑化 |
| 検索正規化 | **DB 側 `translate(column, 'カナ', 'かな') ILIKE translate(query, 'カナ', 'かな')`** | 検索クエリも DB 側で正規化することで対称比較が成立。インデックス効かず Seq Scan になる点はマスタサイズ的に許容 | アプリ層で query 変換 → DB 側既存値と不整合（「ハヤシ」入力で「はやし」レコードヒットしない） |
| カラム名変更 | **しない** （`name_kana` 維持） | 「カナ」は「カタカナ」ではなく一般的に「ふりがな」を指す。改名コストが過大 | `name_yomi` へリネーム → 全層改修 |
| バリデーション | **文字種制約なし**（C 採用） | ユーザー指示。柔軟性最大化 | ひらがなのみ正規表現拒否 |
| 検索パフォーマンス | **インデックス無し（Seq Scan 許容）** | owner/pet マスタは数千件〜数万件規模。translate を関数インデックスで cover も可能だがマイグレーション含めて別 issue | functional index `(translate(name_kana, ...))` 追加 |

## 影響範囲

### DB
- `owners.name_kana`（既存データ変換）
- `pets.name_kana`（既存データ変換）
- `backend/migrations/004_convert_kana_to_hiragana.sql`（新規）
- `backend/migrations/003_seed_demo.sql`（既存シードもひらがな化）

### Backend
- `backend/internal/repository/owner_repository.go:35-61` — FindAll 検索に `name_kana` 追加 + translate 正規化
- `backend/internal/repository/pet_repository.go:31-60` — FindAll 検索に translate 正規化（既存対象 `name_kana`）+ `owners.name_kana` 追加
- バリデーション: 変更なし（model/handler/service の変更なし）

### Frontend
- `frontend/src/features/owners/routes/OwnerForm.tsx:361-374` — ラベル + placeholder
- `frontend/src/features/owners/components/PetEditModal.tsx` — ペット名よみ ラベル + placeholder（要確認）
- `frontend/src/features/owners/hooks/use-owner-form.ts:126` — エラーメッセージ
- `frontend/src/components/shared/PetSelection/PetSelectionSearchForm.tsx:30` — 検索 placeholder
- `frontend/src/components/shared/ReservationFormModal/PatientSelectionTable.tsx:26` — 検索 placeholder
- 他、`カナ`/`kana` を含む全 UI 文言（要 grep）

### codegen
- 不要（型定義変更なし）

## 参照実装

- `backend/internal/repository/owner_repository.go:42` — escapeLike + ILIKE パターン
- `backend/internal/repository/pet_repository.go:43` — JOIN + 検索パターン
- `frontend/src/features/owners/routes/OwnerForm.tsx:361` — Label + Input パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 既存ユーザーが「カナ」運用に慣れている | 低 | UI ラベル「よみ」+ placeholder ひらがな化で誘導。バリデーション緩いので強制せず |
| 検索 translate がインデックス無効化 → Seq Scan | 中 | owner/pet は数千件規模。許容。性能問題発生時に functional index 追加で対応（別 issue） |
| マイグレーション漏れ（一部レコードがカタカナ残存） | 中 | translate は冪等。再実行で網羅 |
| 半角カナ（ﾊﾔｼ）入力時の扱い | 低 | スコープ外。半角カナはほぼ運用されないため対象外 |
| 既存テストデータ・モックがカタカナ前提 | 中 | テストは新規ひらがなで再生成 or 検索両対応で動作確認 |

## 未解決事項

- なし

## 実装順序

1. BE-113: マイグレーション 004 作成 + シード 003 のカタカナ→ひらがな変換 + 適用確認
2. BE-114: 検索正規化（owner_repository + pet_repository）+ 動作確認（ひらがな入力 → カタカナ DB / カタカナ入力 → ひらがな DB 双方ヒット）
3. FE-251: UI ラベル + placeholder + エラーメッセージ全箇所変更（owner / pet / 検索フォーム）
4. ブラウザ E2E 検証（BUG-374 と統合 or 別途追記）

## 関連イシュー

- BE-113: name_kana カタカナ→ひらがな 一括変換マイグレーション
- BE-114: owner / pet 検索の ひらがな⇔カタカナ正規化
- FE-251: 飼主・ペット「よみ」UI ラベル統一（カナ → よみ + placeholder ひらがな化）

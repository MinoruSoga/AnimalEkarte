# FE-refactor.md — frontend/ リファクタリング実行計画（第 3 期）

- **作成日**: 2026-07-11 / **基準 HEAD**: `a8347f61`
- **対象読者**: この計画書とリポジトリのコード以外の文脈を持たない実行者（AI/人間）
- **性質**: 全項目が**挙動保存**（画面表示・通信・ユーザー操作の結果を変えない。値同一のトークン化・型レベル変更・定数化のみ）。挙動変更は「§5 やらないこと」で禁止。

> **過去エピックとの関係**: 第 1 期（FD1〜12 / R-F1〜25、完了 `2a1ef3ad`）と第 2 期（FE-R1〜R17、完了 `48053b7e`・knip CI ゲート化まで済み）は**全て実行済み**。本書は第 3 期 = ①第 2 期実行結果の全体掃き出しで検出した取りこぼし、②これまで未監査だった次元（監査対象外面の色直値・クエリキャッシュ直値・手書き型 vs tygo 生成型）の実測結果、を計画化したものである。旧バックログ 9 件は本書の作業項目または §4 別トラックに全件吸収した。過去の完了記録は git 履歴が正本。

> **⚠️ 実行前にオーナーが解決すべき既知の問題（本計画のスコープ外）**:
> origin/main の CI は `5568a0f0` 時点で **Backend Lint と Codegen Sync が RED**（backend 側 lint 指摘 + `make codegen` 未実行の generated/models.ts drift）。いずれも backend 起因で本計画では触らない。また未 push のローカルコミットが複数存在する。**FE 実行者はこの赤を直そうとしてはならない**（§5-9）。

---

## §1 現状理解（構造マップ）

### 1.1 全体像

動物病院向け電子カルテ SaaS のフロントエンド。React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query。
**1 つの `frontend/vite.config.ts` で 3 アプリを multi-entry ビルド**（`build.rollupOptions.input`）:

| アプリ | エントリ | 用途 |
|---|---|---|
| main | `src/main.tsx` | 院内向け電子カルテ SPA 本体（26 feature・約 1,180 ファイル） |
| liff | `liff/src/main.tsx` | LINE LIFF ペット健康手帳 |
| line-reserve | `line-reserve/src/main.tsx` | LINE 予約ウィザード |

`src/shared-liff/` は liff/line-reserve から import される共有モジュール（ビルドエントリではない）。

### 1.2 main アプリ src/ 構成

- `app/` — react-router v7。ルート定義は `app/routes/*-routes.tsx`。
- `features/` — 26 feature。各 feature は `api/ components/ hooks/ types/ index.ts`（barrel）構成。
- `components/ui/` — shadcn/ui 原始部品（既にプロジェクト用にカスタマイズ済み）。`components/shared/` — feature 非依存の共有部品。
- `hooks/` — グローバル横断フック + クロス feature の Query キャッシュ共有フック。
- `lib/` — `axios.ts`（cookie 認証・`X-Clinic-ID` 付与・401 自動リフレッシュ）、`design-tokens.ts`（1,029 行・色/スタイル定数の単一正本。`C`/`STYLE`/`PALETTE` を export）、`react-query.ts`（`QUERY_STALE_TIMES`/`QUERY_GC_TIMES` 定数）、`query-keys.ts`、`calc-age.ts`、日付 utils。
- `types/` — 手書きドメイン型。`types/generated/models.ts` は Go モデルから tygo 自動生成（545 型・**編集禁止**・CI codegen-check が drift 検知）。**手書き型と生成型の関係は監査済み**: 機械置換可能な重複は 0。`Owner`/`Pet`/`Medicine` は意図的な camelCase 変換層、リクエスト DTO 群は生成型からの Omit/Pick 派生（良いパターン）。残る問題は本計画 FE3-3〜6 参照。

### 1.3 必須規約（正本: frontend/CLAUDE.md・frontend/src/features/CLAUDE.md）

1. **Feature Indexing**: feature 外からの import は barrel（`@/features/x`）経由。deep import 禁止。
2. **React 19**: フォームは `useActionState`。`React.FC`/`forwardRef` 禁止。
3. **条件レンダリング**: `{cond ? <X/> : null}` のみ（`&&` 禁止・機械強制なし）。
4. **Design Tokens**: 色は `@/lib/design-tokens` の定数。hex 直書き禁止。
5. **命名**: コンポーネント `PascalCase.tsx`、hooks/utils/API `kebab-case.ts`。
6. **型**: `any` 禁止（ESLint error）。

### 1.4 CI ゲート（壊してはならない）

type-check / lint / build / test:coverage の 4 ゲート + `design-system-audit.mjs`（routes/pages 面の hex 等・zero-tolerance）+ eslint-disable ratchet + filename ratchet + coverage ratchet + **knip（第 2 期でゲート化済み・検出 0 件維持必須）** + codegen-check（tygo drift）。

### 1.5 検証コマンド規約（Docker 必須）

- ローカル pnpm 直接実行禁止。必ず `docker compose exec frontend ...`。
- **scoped テスト**: `npx vitest run <path>`（`pnpm test:run -- <path>` は全件実行になる罠）。
- フル実行クラス（`pnpm run type-check` / `lint` / `unused` / `design-audit`）は各項目の完了条件に明記されたもののみ実行。実行環境に拒否されたらコマンドを提示して人間に依頼し、結果を得てから判定する。
- ルートレンダー系テストは `vi.mock("@/hooks/use-auth", ...)` が必要な場合がある。同一 Docker 上の vitest 並行実行はタイムアウト多発 — 疑わしい失敗は単体再実行。

---

## §2 項目 R0: 安全網の構築（最初に必ず実行）

### R0-1. 前提確認

```
docker compose ps                     # frontend コンテナ Up
git rev-parse HEAD                    # 記録
git status --porcelain -- frontend/  # 空であること。空でなければ中断・報告
```

backend/ や `.claude/`・`BE-refactor.md` の dirty は無視してよい（触らない・ステージしない）。

### R0-2. ベースライン記録（全て green が着手条件）

```
docker compose exec frontend pnpm run type-check   # エラー 0
docker compose exec frontend pnpm run lint         # エラー 0
docker compose exec frontend pnpm run unused       # 検出 0 件・exit 0（knip はゲート済み）
docker compose exec frontend pnpm run design-audit # PASS
```

1 つでも赤ならベースライン不成立 — 着手せず中断・報告。**origin/main の CI が Backend Lint / Codegen Sync で赤なのは既知**（backend 起因・本計画のスコープ外）。上のローカル FE ゲートが green であれば作業してよい。

### R0-3. コミット規約

- **1 項目 = 1 コミット**。メッセージは `<type>(frontend): <説明> (FE3-<n>)`（type: refactor/chore/test/ci/docs）。
- `Co-Authored-By` を含めない。**push しない**（完了報告でハッシュ一覧提示）。
- 戻し方: 直前項目は `git reset --hard HEAD~1`。それより前は `git revert`。

---

## §3 作業項目リスト（この順に実行する）

> 完了条件を満たせない項目は変更を破棄し、SKIP/BLOCKED として理由を記録して**依存されていない次の項目へ進む**（依存先が SKIP なら連鎖 SKIP）。
> 変更中にテスト・型チェックが「実は使われている/実は挙動が変わる」ことを示したら、その対象だけスキップして報告する。実装バグを発見しても直さない — BLOCKED 記録して報告する。

### — Phase A: design token の完全化（監査面の外に残った色直値 12 行） —

### FE3-1. rgba 直値 12 行の値同一トークン化（8 ファイル） — ✅ DONE (8093b582)

- **対象**: 下表の 8 ファイル + `frontend/src/lib/design-tokens.ts`（トークン追加）
- **問題**: design-system-audit.mjs の監査面（routes/pages）の外に rgb/rgba 直値が 12 行残存。うち 4 ファイル（ui 原始部品）の focus 色 `rgba(35,131,226,…)` は**廃止済みレガシーアクセント #2383E2 の rgb 表記**で、現行監査の regex では検出不能。
- **変更内容**（全て**値同一**の置換。レンダリング結果は 1px も変えない）:

| 箇所 | 直値 | 対応 |
|---|---|---|
| `components/ui/searchable-select.tsx:125` / `ui/select.tsx:32` / `ui/input.tsx:15-16` / `ui/textarea.tsx:11-12` | `hover:bg-[rgba(242,241,238,0.5)]` + `focus:border-[rgba(35,131,226,0.57)]` + `focus:shadow-[0_0_0_1px_rgba(35,131,226,0.35)]` | hover は既存 `PALETTE.hoverBgInput`（design-tokens.ts:215、同値）を参照。focus 2 種は新トークン `PALETTE.focusBorderLegacyAccent` / `PALETTE.focusRingLegacyAccent` を**現在の値のまま**新設して参照（名前に Legacy を含め、値の変更は別途デザイン判断であることを JSDoc に明記） |
| `features/reservations/components/week-view-grid-constants.ts:14` | `boxShadow: "0 10px 30px rgba(0,0,0,0.15)"` | 新トークン `STYLE.dragPreviewShadowLarge`（既存 dragOverlayShadow とは値が異なるため統合しない） |
| `features/reception/components/AppointmentCard.tsx:150` | `shadow-[0px_1px_3px...rgba(0,0,0,0.1)...]` | 新トークン `STYLE.cardShadow` |
| `features/shifts/components/ShiftTemplateSettingsParts.tsx:226` | `shadow-[0_1px_2px_rgba(0,0,0,0.1)]` | 同一シャドウ文字列が design-tokens.ts:954 の STYLE プリセット内に埋め込み済み — 新トークン `STYLE.pillShadow` として抽出し、:954 側と :226 の両方から参照 |
| `components/shared/Layout/Layout.tsx:26` | `shadow-[0_0_8px_rgba(3,139,148,0.5)]` | 新トークン `STYLE.brandGlow`（#038B94 のブランドグロー） |

  ※ `src/hooks/use-reservation-type-color-map.ts:126,128` の `rgba(${r},...)` は実行時動的生成のため**対象外**（FE3-2 で allowlist）。Tailwind の arbitrary value 内で定数参照する形式（テンプレートリテラル or 既存 STYLE パターン）は、design-tokens.ts 内の既存プリセットの書き方に合わせること。
- **完了条件**: `git diff` で置換前後の CSS 値が同一であることを目視確認（diff に現れる値の変化はトークン名のみ）。`pnpm run type-check` → 0。`pnpm run unused` → 検出 0（**注意**: 新トークンは新規 export ではなく既存の `PALETTE` / `STYLE` オブジェクトへのキー追加として定義すること — `components/ui/**` は knip 解析対象外のため、ui 配下でしか使わない新規 export は未使用と誤検知される）。`npx vitest run src/components/ui src/features/reservations src/features/reception src/features/shifts src/components/shared/Layout` → PASS。`grep -rnE 'rgba?\(' frontend/src --include='*.tsx' --include='*.ts' | grep -v test | grep -v design-tokens | grep -v use-reservation-type-color-map | grep -v generated` → 0 件。
- **リスク / 戻し方**: ui 原始部品は全画面に波及するため、値の write ミスは目視差分確認で防ぐ。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-2. design-system-audit.mjs のスコープ全 src 化 + rgb/rgba/hsl ルール新設 — ✅ DONE (ba4561e0)

- **対象**: `frontend/scripts/design-system-audit.mjs`、`frontend/scripts/design-system-audit.test.mjs`、JSDoc 2 行（`src/components/shared/Form/SubmitButton.tsx:8`、`PrimaryButton.tsx:6`）
- **問題**: 監査面が routes/pages 限定のため、FE3-1 のような直値が components/hooks 面に無検知で蓄積し得る。また C1 regex（`C\.accent`）は正当な `C.accentBrand` に前方一致誤検知する欠陥があり、rgb/rgba/hsl 表記の色は現行ルールの完全な盲点。
- **変更内容**:
  1. C1 regex を単語境界化（`C\.accent\b` 等 — `CheckupSyncPreviewTable.tsx:27,160` の `C.accentBrand` が誤検知しないこと）。
  2. C3（引用 hex）の走査対象を `src/**`・`liff/src/**`・`line-reserve/src/**` 全体へ拡大。事前に JSDoc 2 行のバッククォート付き `#038B94` 表記を文言変更（例: 「ブランドティール」等、hex を含まない表現）し、`design-tokens.ts` を allowlist に追加。
  3. 新ルール C6: `rgba?\(|hsla?\(` の直値検知を全 src 対象で追加。allowlist: `src/lib/design-tokens.ts`（定義正本）と `src/hooks/use-reservation-type-color-map.ts`（実行時動的生成・JSDoc 根拠あり）の 2 ファイルのみ。
  4. `design-system-audit.test.mjs` に C6 と境界修正のテストケースを追加（既存テストの流儀に合わせる）。
- **完了条件**: `docker compose exec frontend pnpm run design-audit` → PASS（検出 0 件）。`node --test frontend/scripts/design-system-audit.test.mjs`（コンテナ内）→ PASS。FE3-1 を revert した状態なら C6 が fail することをローカルで一時確認（確認後元に戻す）。
- **リスク / 戻し方**: 誤検知による CI 赤 → allowlist 追加ではなくまず regex を疑う。`git reset --hard HEAD~1`。
- **依存**: **FE3-1**（直値ゼロが前提）

### — Phase B: 型ドリフトの封じ込め（手書き型 vs tygo 生成型監査の帰結） —

### FE3-3. 生リテラル union 4 件を生成定数由来（`typeof`）に置換 — ⏭️ SKIP（実装前検証で全4件が型安全性後退と判明）

> **SKIP 理由**: tygo 生成側は `export type X = string; export const XFoo: X = "literal";` 形式（`VisitType`/`ReservationStatus`/`TreatmentItemType`/`BodyWeightUnit` 全て確認済み）。この形式では `typeof XFoo` は明示アノテーションにより widen 済みの `X`（= `string`）に評価され、リテラル型は保持されない。実測確認: 参照実装とされた `shifts/types/index.ts` の `ShiftType`（既に `typeof ShiftTypeFull | ...` 形式）に対し、和集合に存在しない任意文字列を代入する一時スクラッチファイルを作成して `tsc --noEmit` を実行したところ、型エラーなくコンパイルが通った（= 既に `string` へ degrade 済み。スクラッチファイルは検証後に削除、`frontend/` は porcelain 空を確認済み）。
> 本計画 §3 FE3-3 自身が明記する回避規定「union の広がり（`string` 化）に注意 — `typeof` 定数が `string` 型で宣言されていたら literal が失われ型安全性が落ちる。その場合はスキップして報告。」に該当するため、4 件（`VisitType`・`ReservationStatus`・`TreatmentItemType`・`BodyWeightUnit`）を SKIP する。参照実装（shifts/accounting）自体が既に同じ問題を抱えている点は本計画のスコープ外の既存債務として報告のみに留める（別途 BE-refactor 側での tygo 生成方式見直しが必要）。

- **対象**: `frontend/src/types/index.ts` の `VisitType`・`ReservationStatus`、`frontend/src/features/medical-records/types/`（barrel 実体ファイル）の `TreatmentItemType`・`BodyWeightUnit`
- **問題**: 生成側 `types/generated/models.ts` は値定数（例: `ReservationStatus*` 8 定数）を export しているのに、手書き union が生リテラルで再宣言されている。backend が値を追加すると手書き側が黙って古くなる（accounting/shifts feature は既に `typeof` 定数 union パターンで解決済み — その横展開）。
- **変更内容**（型レベルのみ・実行時値の変化なし）: 各 union を accounting の既存パターンに合わせて書き換える。スケッチ（ReservationStatus の例。生成定数名は generated/models.ts を読んで正確に使う）:
  ```ts
  // 変更前
  export type ReservationStatus = "reserved" | "checked_in" | ... /* 8 literals */;
  // 変更後
  import type * as gen from "@/types/generated/models";
  export type ReservationStatus = (typeof RESERVATION_STATUS_VALUES)[number];
  // RESERVATION_STATUS_VALUES を generated の定数群から構成する（値・順序は現状維持）
  ```
  既存の `RESERVATION_STATUS_VALUES` / ラベル map との整合を保つ（値集合は現状と同一 8 件 — 変えない）。`VisitType`（2 値）・`TreatmentItemType`（4 値）・`BodyWeightUnit`（2 値）も同様。生成側に対応する定数が無い場合はその union をスキップして報告（generated を編集して定数を作るのは禁止）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/medical-records src/features/reservations` → PASS。diff が型宣言のみ（実行時コードの変化なし）であること。
- **リスク / 戻し方**: union の広がり（`string` 化）に注意 — `typeof` 定数が `string` 型で宣言されていたら literal が失われ型安全性が落ちる。その場合はスキップして報告。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-4. `MembershipType` の名前衝突解消（UI ラベル union のリネーム） — ✅ DONE (caea4bfa)

- **対象**: `frontend/src/features/owners/types/`（barrel 実体）で定義される `MembershipType`（値 = `"非会員" | "会員" | "退亡者" | "他診/準"` の日本語ラベル union）と、その全 importer
- **問題**: 生成型 `MembershipType`（wire 値 `non_member | member | deceased | transferred`）と**同名で値域が完全に異なる**型が feature 側に存在する。import 補完で取り違えると型は通るのに実行時に不整合を起こす、最も危険なクラスの名前衝突。
- **変更内容**（機械的リネーム・値と表示は不変）: feature 側の型名を `MembershipTypeLabel` に変更。`grep -rn 'MembershipType' frontend/src --include='*.ts' --include='*.tsx'` で全参照を列挙し、**owners feature の日本語ラベル union を参照している箇所だけ**を新名称に書き換える（generated 由来の参照は触らない）。barrel の export 名も更新。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/owners` → PASS。`grep` で旧名参照が generated 由来のみになっていること。
- **リスク / 戻し方**: リネーム漏れは型チェックで検出される。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-5. `{ ids: number[] }` Reorder DTO 5 重複の統合 — ✅ DONE (fa78d101)

- **対象**: `ReorderDiagnosisTypeRequest` / `ReorderDiagnosisNameRequest`（types/diagnosis.ts）、`ReorderMedicinesRequest`（types/medicine.ts）、`ReorderReservationTypeRequest`（types/reservation-type.ts）、`ReorderTreatmentRequest`（types/treatment.ts）
- **問題**: 同一形状 `{ ids: number[] }` の型が 4 ファイルに 5 個。
- **変更内容**: `src/types/form.ts`（または index.ts — 既存の共有型の置き場に合わせる）に `export interface ReorderRequest { ids: number[] }` を 1 つ定義し、5 型を `export type ReorderXxxRequest = ReorderRequest;` のエイリアスに置換（**既存の型名は import 互換のため残す**。importer の書き換えは不要）。
- **完了条件**: `pnpm run type-check` → 0。diff が types/ 配下のみ。
- **リスク / 戻し方**: なし（構造的同値のエイリアス化）。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-6. Update 系 DTO の Create 派生化 2 件 — ✅ DONE (a22cacef)

- **対象**: `frontend/src/types/owner.ts`（`UpdateOwnerRequest`）、`frontend/src/types/trimming.ts`（`UpdateTrimmingRequest`）
- **問題**: `UpdateOwnerRequest` は `Partial<Omit<CreateOwnerRequest, "clinic_id">>` とフィールド単位で同値、`UpdateTrimmingRequest` は `CreateTrimmingRequest` から 3 フィールドを除いた形 — 手書き二重定義で Create 側の変更が Update に伝播しない。
- **変更内容**: 派生型に書き換える（変更前後でフィールド集合・optionality・型が**完全一致**することを両定義の突合で確認してから置換。1 フィールドでも差があればスキップして差分を報告）:
  ```ts
  export type UpdateOwnerRequest = Partial<Omit<CreateOwnerRequest, "clinic_id">>;
  export type UpdateTrimmingRequest = Omit<CreateTrimmingRequest, "appointment_id" | "reservation_type_id" | "reservation_route">;
  ```
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/owners src/features/trimming` → PASS。
- **リスク / 戻し方**: optionality の微差を見落とすと型エラーで顕在化（それが検出装置）。`git reset --hard HEAD~1`。
- **依存**: R0

### — Phase C: 散在直値の定数化（クエリキャッシュ・ページネーション） —

### FE3-7. staleTime 直値 11 箇所の定数化 — ✅ DONE (73913467)

- **対象**: `frontend/src/lib/react-query.ts`（定数追加）+ 以下 11 サイト
- **問題**: 約 180 サイトが `QUERY_STALE_TIMES` 定数で統一されている中、11 サイトだけ直値が残る。うち 5 つは既存定数 `MEDIUM` と同値。
- **変更内容**（全て値同一）:
  1. `5 * 60 * 1000` → `QUERY_STALE_TIMES.MEDIUM` に置換（5 件）: `aggregation/api/get-cpm-stage-counts.ts:55`、`aggregation/api/get-aggregations.ts:105`、`lstep/api/get-lstep-delivery-stats.ts:27`、`lstep/api/get-lstep-tag-summary.ts:28`、`lstep/api/get-lstep-visit-conversion.ts:32`
  2. `react-query.ts` に既存命名規則に合わせて 2 定数を追加（値は現状の直値と同一): 30 秒 tier と 1 分 tier（名称は既存の REALTIME/MEDIUM 等の語彙に整合させる。例 `SHORT: 30_000` / `MINUTE: 60_000`）。
  3. `30_000` → 新 30 秒定数（3 件）: `accounting/hooks/use-accounting-detail-state.ts:66`、`accounting/api/get-ungrouped-items.ts:33`、`accounting/api/get-discount-suggestions.ts:30`。`60 * 1000` → 新 1 分定数（3 件）: `lstep/api/get-lstep-delivery-trigger-summary.ts:31`、`get-lstep-delivery-trigger-logs.ts:50`、`get-lstep-csv-imports.ts:30`
  4. **触らない**: `lstep/api/get-checkup-sync-preview.ts:84`（`staleTime: 0`・意図的）、`auth/api/get-me.ts:27`（`10 * 1000`・JSDoc で根拠文書化済み）
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/aggregation src/features/lstep src/features/accounting` → PASS。`grep -rn 'staleTime: [0-9]' frontend/src --include='*.ts' | grep -v react-query.ts | grep -v test` の残存が「触らない」2 件のみ。
- **リスク / 戻し方**: なし（値同一）。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-8. 履歴取得上限 100 の共有定数化 + 冗長 pageSize 指定の削除 — ✅ DONE (cdd3035f)

- **対象**: cap=100 の 5 サイト: `trimming/api/get-trimming.ts:28`、`trimming/api/get-trimmings.ts:16`、`owner-report/api/get-pet-treatment-history.ts:100`、`owner-report/api/get-pet-trimming-history.ts:26`、`lstep/api/get-lstep-tag-owners.ts:47`。冗長 `pageSize: 20` の 3 サイト: `accounting/routes/AccountingList.tsx:166`、`estimates/routes/EstimateList.tsx:172`、`checkups/routes/CheckupsList.tsx:119`
- **問題**: 「履歴系フェッチの上限 100」という同一セマンティクスの直値が 5 ファイルに散在。また `usePagination` のデフォルトが既に 20（`src/hooks/use-pagination.ts:39`）なのに 3 ルートが同値を明示指定している。
- **変更内容**: (1) `src/config/` に定数ファイル（既存 config ファイルの命名に合わせる。例 `fetch-limits.ts` の `HISTORY_FETCH_LIMIT = 100`）を新設し 5 サイトを置換（`perPage` 変数名の違いは値の参照のみ差し替え）。(2) 3 ルートの `pageSize: 20` 引数を削除（デフォルト 20 に一致することを use-pagination.ts で確認済み — 削除は値同一）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/trimming src/features/owner-report src/features/lstep src/features/accounting src/features/estimates src/features/checkups` → PASS。
- **リスク / 戻し方**: なし（値同一）。`git reset --hard HEAD~1`。
- **依存**: R0

### — Phase D: 第 2 期実行の取りこぼし是正 —

### FE3-9. 年齢計算 3 重実装の統合（calcAgeAt 統一の完遂） — ✅ DONE (1bf0c284)

- **対象**: `frontend/src/lib/calc-age.ts`（共有ヘルパ追加）、`frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx:13-32`（私的 `calcAge`）、`frontend/src/features/owner-report/lib/pet-age.ts:15`（`formatPetAge`）
- **問題**: `14849cdc` の calcAgeAt 統一は PetDeceased 系 2 ファイルのみで、年齢計算の独立実装が 2 つ残った。3 実装はエッジケースの挙動（1 歳未満の表記・未来日ガード・クランプ）が微妙に異なる。
- **変更内容**（**各呼び出し元の現在の出力を 1 文字も変えない**ことが絶対条件）:
  1. `lib/calc-age.ts` に年月コンポーネント版 `calcAgePartsAt(birthdate, at): { years: number; months: number }` を追加（クランプ・ガードなしの生値。既存 `calcAgeAt` は変更しない）。
  2. PatientContextHeader の私的 `calcAge` を `calcAgePartsAt` ベースに置換。**表示フォーマット（"Xヶ月"/"X歳"）と未来日ガード無しの現挙動はコンポーネント側に残す**。
  3. `formatPetAge` も同様に計算部のみ共有化し、`"0歳Mヶ月"` 表記・null ガードは現状のまま残す。
  4. 置換前に両呼び出し元の現挙動を固定する特性テストを追加: `PatientContextHeader.test.tsx`（または新規）に「誕生日 2026-01-15・基準日 2026-07-11 → "5ヶ月"」「誕生日 2020-07-10 → "6歳"」等、`pet-age` テストに「1 歳未満 → "0歳Mヶ月"」「無効日付 → null」を、**リファクタ前に RED→GREEN 確認してから**計算部を置換する。
- **完了条件**: `npx vitest run src/lib/calc-age.test.ts src/components/shared/PatientContextHeader src/features/owner-report` → PASS（新規特性テスト含む）。`grep -rn 'getFullYear' frontend/src --include='*.tsx' --include='*.ts' | grep -v test | grep -v lib/` で年齢計算の独立実装が残っていないこと。
- **リスク / 戻し方**: 月数計算の日付境界（月末生まれ等）で挙動差が出るリスク → 特性テスト先行で防ぐ。差が出たらスキップして報告。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-10. RefundSection の PAYMENT_METHOD_LABELS 複製統合（dedupe の完遂） — ✅ DONE (9f94b963)

- **対象**: `frontend/src/features/accounting/components/RefundSection.tsx:21-26`、正本: `components/daily-accounting-utils.ts:4`
- **問題**: `fd355b92` は PaymentCard のみ統合し、RefundSection の値同一コピーが残った。
- **変更内容**: 置換前に両定義を diff し**値がバイト同一**であることを確認 → RefundSection のローカル定義を削除して import に置換。任意の強化: 共有側の型を `Record<string, string>` から `Record<PaymentMethod, string>`（accounting/types の union）へ厳格化 — 型チェックが通る場合のみ。通らなければ import 置換だけ行い、型厳格化はスキップして報告。**AccountingListTable.tsx:25 の分岐済みコピーは触らない**（§4）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/accounting` → PASS。accounting/components 内の `PAYMENT_METHOD_LABELS` 定義が daily-accounting-utils.ts と AccountingListTable.tsx の 2 つだけになること。
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-11. accounting-detail-model のテストをモジュールの隣へ移動 — ✅ DONE (99c26a83, 4eb87c64)

- **対象**: `frontend/src/features/accounting/routes/__tests__/accounting-detail-model.test.ts`
- **問題**: 第 2 期 FE3 でモデル本体は `components/` へ移動したが、テストが旧レイヤ `routes/__tests__/` に残り `../../components/...` を import している。
- **変更内容**: `git mv` で `components/__tests__/accounting-detail-model.test.ts`（accounting feature 内の既存テスト配置慣行に合わせる — 同 feature に `__tests__/` が無く co-located 形式なら `components/accounting-detail-model.test.ts`）へ移動し、import 相対パスを修正。
- **完了条件**: `npx vitest run src/features/accounting` → PASS（対象テストが新パスで実行されること）。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: FE3-10（同 feature の連続変更のため直列）

### FE3-12. CarePlanTab/badges.tsx の命名規則是正 — ✅ DONE (403fd6bf)

- **対象**: `frontend/src/features/hospitalization/components/CarePlanTab/badges.tsx`（`TypeIcon` / `StatusBadge` / `TimingBadges` の 3 コンポーネントを export）
- **問題**: features/CLAUDE.md はコンポーネントファイルに `PascalCase.tsx` を義務付けるが、第 2 期のコミット `f4ed72ad` が kebab-case でコンポーネントファイルを作った。
- **変更内容**: `git mv badges.tsx CarePlanBadges.tsx` し、importer（`grep -rn "from.*CarePlanTab/badges\|from.*./badges" frontend/src/features/hospitalization` で列挙）の import パスを更新。ファイル内容は不変。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/hospitalization` → PASS。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### FE3-13. knip.json の $schema バージョン整合 — ✅ DONE (f56186f9)

- **対象**: `frontend/knip.json:2`
- **問題**: `$schema` が `knip@5` の URL を指すが devDependency は `knip ^6.25.0`。スキーマ検証が旧バージョン仕様で行われる。また `src/types/generated/**` ignore の根拠がどこにも記録されていない。
- **変更内容**: `$schema` を `https://unpkg.com/knip@6/schema.json` に更新。コミットメッセージに generated ignore の根拠（tygo 自動生成・codegen-check が正本ゲート・knip の解析対象外とする）を記録する。
- **完了条件**: `docker compose exec frontend pnpm run unused` → 検出 0・exit 0（挙動不変）。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### — Phase E: 分割残差 —

### FE3-14. OwnerInfoFieldSections.tsx（441 行）の 3 セクション分割 — ✅ DONE (c984b765)

- **対象**: `frontend/src/features/owners/components/OwnerInfoFieldSections.tsx`。importer は `OwnerInfoSection.tsx` の 1 件のみ
- **問題**: 第 1 期 R-F19 の分割産物自身が 441 行で残った（移動 > 分解だった）。内部は独立した 3 セクション（`OwnerBasicFields` :63 / `OwnerAddressFields` :237 / `OwnerMembershipFields` :365）+ 私的 memo 子（`MembershipTypeButtons` :21）。
- **変更内容**（逐語移動・第 2 期 FE-R14 の PetEditModalFieldSections 分割と同じ手順）:
  1. `OwnerBasicFields.tsx` / `OwnerAddressFields.tsx` / `OwnerMembershipFields.tsx` を新設し各セクションを逐語移動。`MembershipTypeButtons` は唯一の利用側（membership セクション）のファイルへ同居させる。
  2. 共有定数・型があれば `owner-info-field-shared.ts` に抽出（使うものだけ export）。
  3. `OwnerInfoSection.tsx` の import を 3 ファイル直接参照へ書き換え、元ファイルを `git rm`（互換 re-export は作らない）。
  4. 新設ファイル全てが 400 行未満であることを wc -l で確認。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/owners` → PASS。`pnpm run unused` → 検出 0（export 過剰に注意）。新設ファイル全て < 400 行。
- **リスク / 戻し方**: 中（JSX 移動量）。knip ゲートと型チェックが検出装置。`git reset --hard HEAD~1`。
- **依存**: FE3-4, FE3-6（owners の型変更を先に確定させ、分割との衝突を避ける）

---

## §4 別トラック（本計画では実行しない・記録のみ）

実行者はこのセクションに手を付けない。完了報告に「未着手・別トラック」として転記するだけでよい。

| 項目 | 内容 | 実行しない理由 |
|---|---|---|
| **CI 赤の解消（最優先・オーナー対応）** | origin/main の Backend Lint 赤（`openapi_route_drift_test.go:672` builtinShadow、`billing_item_trimming_test.go` ×4 ほか）+ Codegen Sync 赤（`make codegen` 未実行の models.ts drift）+ 未 push ローカルコミット群 | backend 起因。FE 実行者が触ると事故る |
| AccountingListTable の支払方法表記 | `credit_card: "クレジットカード"`（AccountingListTable.tsx:25）vs 他画面 "カード" | どちらが正かは PO 判断のユーザー可視変更 |
| `{cond && <JSX>}` の機械強制 | `eslint-plugin-react`（jsx-no-leaked-render）等の導入。現状違反 0 件で導入即 green | devDependency 追加はオーナー判断 |
| `dangerouslySetInnerHTML` / PrintPortal XSS 監査 | セキュリティレビュー | 別途レビュー枠 |
| React Query `clinic_id` キャッシュ境界 | クリニック切替時のキャッシュ分離検証 | 調査タスク（結果次第で挙動変更） |
| FE zod ↔ Backend バリデーション二重管理 | 設計判断 | architect 判断 |
| MedicalRecordForm.tsx（411 行）の再分割 | モーダル JSX 群の抽出には大量の prop 引き回し判断が必要 | 800 行上限内・効果がリスクに見合わない |
| シフト休憩時刻 Input 4 箇所の aria-label | `ShiftFormDialog.tsx:304,311`・`ShiftTemplateSidePanelFields.tsx:138,145`（可視ラベルは「〜」区切りのみ） | ラベル文言の新規決定（例「休憩N 開始時刻」）が必要 = 文言発明は仕様追加 |
| medical-records の `Treatment` UI 型が #201 に未追随 | 生成型の `dose_*` 5 フィールド等を持たない stale twin | #201 の FE UI 実装（OPEN）側で対応すべき機能作業 |
| `TreatmentPlan`（src/types/index.ts）の stale twin 解消 | 生成型と同名・別形状。transform 層に寄せるかリネームか | transform 化は挙動リスク、リネームだけでは本質未解決 — 設計判断 |
| `BackendTrimming`（types/trimming.ts）の codegen 化 | 手書きミラー DTO（ファイル冒頭に手動同期の警告あり） | tygo がハンドラ DTO を出力しない設計に関わる |
| ForgotPasswordPage の anti-enumeration と toast の矛盾 | エラー時 `{status:"sent"}` を返しつつ handleApiError が toast する（列挙防止意図を toast が部分的に破る） | 挙動変更（セキュリティ文脈）— 別判断 |
| タイマー直値 2 件のコメント付与 | `use-medical-record-manual-errors.ts:35`（50ms）・`use-medical-record-form-modals.ts:28`（100ms） | コメントのみの nit。次に当該ファイルを触るコミットに同乗させる |
| `models.ts`（tygo 生成）・`design-tokens.ts` の分割 | 自動生成/定数カタログ | 対象外（確定済み）。なお旧バックログの「手書き models.ts」は**存在しない**ことを実測確認済み（誤記だった） |

## §5 やらないことリスト（禁止事項）

1. **挙動変更の一切**: 表示文言・通信・画面遷移・CSS 計算値を変えない。FE3-1 は値同一置換、FE3-9 は特性テスト先行が条件。バグ発見時は BLOCKED 記録（テストを曲げるのも禁止）。
2. **§4 別トラック項目への着手**。特に CI 赤（Backend Lint / Codegen Sync）を直そうとしない — `make codegen` の実行も backend ファイルの変更も禁止。
3. **依存ライブラリの追加・更新・削除**（package.json / pnpm-lock.yaml に触らない。eslint プラグイン追加も禁止）。
4. **`src/types/generated/**` の編集**、および **`components/ui/**` の FE3-1 で指定した行以外の変更**。
5. **旧計画書・進捗文書の編集**: 進捗は本ファイル（FE-refactor.md）の項目見出しへ `✅ DONE (hash)` を追記する形でのみ記録。`BE-refactor.md` / `.claude/` / backend/ に触らない。
6. **design-tokens.ts の値の変更**: FE3-1/FE3-2 で行うのは**定数の追加**と参照置換のみ。既存トークンの値・名前を変えない。レガシーアクセント focus 色の「正しい色への修正」もしない（それはデザイン判断）。
7. **Feature Indexing の再設計・barrel の増減**（FE3-4 の export 名変更と FE3-14 の分割に伴う必要最小限を除く）。
8. **e2e / lighthouse / CI ワークフローの変更**（本計画に CI 変更項目はない）。
9. **フル検証コマンドの無断省略・無断実行**: 完了条件に明記されたコマンドのみ実行。拒否されたら人間に依頼し、結果を得ずにコミットしない。`pnpm test:run`（全件）・`pnpm install` は実行しない。
10. **push・PR 作成・外部書き込み・コミットの squash/rebase/amend 禁止**。

## §6 実行者への指示文（このままコピペして渡す）

```
あなたはこのリポジトリのフロントエンドリファクタリング実行者である。
リポジトリルートの FE-refactor.md が唯一の作業指示書である。以下を厳守せよ。

1. まず FE-refactor.md を全文読む。次に frontend/CLAUDE.md と
   frontend/src/features/CLAUDE.md を読む（規約と検証コマンドの正本）。
2. §2 の R0（安全網）を最初に実行する。ローカル FE ゲート（type-check / lint /
   knip / design-audit）が 1 つでも赤なら着手せず報告して終了。
   origin/main の CI が Backend Lint / Codegen Sync で赤なのは既知・スコープ外 —
   直そうとするな。
3. §3 の作業項目を FE3-1 から FE3-14 まで番号順に、1 項目ずつ実施する。
   - 1 項目 = 1 コミット。コミット後に次へ。並行作業禁止。
   - 各項目の「完了条件」のコマンドを全て実行し、期待結果を満たさない限りコミットしない。
   - フル実行コマンドが権限で拒否されたら、コマンドを提示して人間に実行を依頼し、
     結果を得てから判定する。結果なしでコミットしない。
   - 完了条件を満たせない場合: 変更を破棄し、SKIP/BLOCKED として理由を記録して
     依存されていない次の項目へ進む。依存先が SKIP なら連鎖 SKIP。
4. 値同一が条件の項目（FE3-1/7/8/10）では、置換前後の値の同一性を diff で確認した
   証拠を報告に残す。FE3-9 は特性テストを先に書き、RED→GREEN を確認してから
   計算部を置換する。
5. §4（別トラック）と §5（やらないこと）に該当する作業は、必要だと感じても実行するな。
6. push するな。コミットはローカルに残す。
7. 全項目終了後、以下を含む完了報告を書く:
   - 項目ごとの DONE/SKIP/BLOCKED とコミットハッシュ
   - FE3-1 の置換対応表（直値 → トークン名）
   - FE3-3〜6 でスキップした型とその理由（あれば）
   - FE3-9 の特性テスト一覧
   - ベースラインと完了時点の 4 ゲート結果比較
```

---

## 付録: 実行順の依存関係まとめ

```
R0 ─┬─ FE3-1 ─ FE3-2                     （design token: 直値ゼロ化 → 監査拡大）
    ├─ FE3-3, FE3-4, FE3-5, FE3-6        （型: 相互独立・番号順推奨）
    ├─ FE3-7, FE3-8                      （定数化: 相互独立）
    ├─ FE3-9, FE3-12, FE3-13             （独立）
    ├─ FE3-10 ─ FE3-11                   （accounting 直列）
    └─ (FE3-4, FE3-6 完了後) FE3-14      （owners 分割は型確定後）
```

- FE3-2 は FE3-1 の直値ゼロ化が前提（新 rgba ルールが即 fail するため）。
- FE3-10 → FE3-11 は同 feature（accounting）の連続変更のため直列。
- FE3-14 は owners の型変更（FE3-4/FE3-6）を先に確定させてから。
- それ以外は独立だが、コンフリクト回避のため番号順の直列実行を推奨する。


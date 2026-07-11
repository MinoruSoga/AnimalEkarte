# FE-refactor.md — frontend/ リファクタリング実行計画（第 4 期）

- **作成日**: 2026-07-11 / **基準 HEAD**: `e53a6b2b`
- **対象読者**: この計画書とリポジトリのコード以外の文脈を持たない実行者（AI/人間）
- **性質**: 原則**挙動保存**。ただし今期は実測で確定した**実バグ 3 クラスタ**を含むため、挙動変更を伴う項目は `fix:` コミットとして明示分離する（該当: FE4-6・FE4-9 のみ。他は全て挙動保存）。

> **過去エピックとの関係**: 第 1 期（R-F1〜25）・第 2 期（FE-R1〜17）・第 3 期（FE3-1〜14、完了記録 `904da793`）は実行済み。本書は第 4 期 = ①第 3 期で SKIP された FE3-3 の根本原因（tygo 生成定数の型 widen）から発見された**退化 union 6 件**、②未監査次元（日付ユーティリティ地形・ラベルマップ・クエリキー・テスト基盤）の実測結果、③第 3 期実行の掃き出し結果、を計画化したもの。過去の完了記録は git 履歴（`git log --grep='FE3-\|FE-R\|R-F'`）が正本。

> **⚠️ オーナー対応事項（本計画のスコープ外・実行者は触るな)**: origin/main へ約 26 コミットが未 push（BE 実行分 + FE3 全量）。backend/ に別セッションの作業中 dirty ファイルあり。

---

## §1 現状理解（構造マップ）

### 1.1 全体像

動物病院向け電子カルテ SaaS のフロントエンド。React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query。
**1 つの `frontend/vite.config.ts` で 3 アプリを multi-entry ビルド**: main（`src/main.tsx`・26 feature）、liff（`liff/src/main.tsx`・LINE 健康手帳）、line-reserve（`line-reserve/src/main.tsx`・LINE 予約）。**3 アプリは同一 tsconfig/vite 設定を共有し、liff/line-reserve からも `@/`（= `src/`）への import が可能**（`src/shared-liff/` がその確立パターン。「別アプリだから共有不可」はもはや誤り）。

### 1.2 レイヤ構成（main アプリ）

- `app/` — react-router v7。`features/` — 各 feature は `api/ components/ hooks/ types/ index.ts`（barrel）。
- `components/ui/`（shadcn・knip 解析対象外）、`components/shared/`（横断部品）。
- `lib/` — `axios.ts`・`design-tokens.ts`（1,059 行。`C`/`STYLE`/`PALETTE` を export）・`react-query.ts`（`QUERY_STALE_TIMES` 等）・`query-keys.ts`（**registry はまだ accountings/masters/me のみ** — 大半のフックは inline queryKey）・`jst-date.ts`（日付の事実上の標準・60 ファイルが import）・`iso-date.ts`・`calc-age.ts`。
- `types/` — 手書きドメイン型 + `types/generated/models.ts`（tygo 自動生成・545 型・**編集禁止**）。
- `src/testing/` — vitest setup + MSW（**共有 render ヘルパは未整備** — CODING_RULES が `src/testing/utils.tsx` を記載しているが実体が無い）。

### 1.3 今期の中心問題: tygo 生成定数の型 widen と「退化 union」

tygo は `export type ShiftType = string; export const ShiftTypeFull: ShiftType = "full";` 形式で出力する。**定数に `: ShiftType`（= string）の明示アノテーションが付くため、`typeof ShiftTypeFull` は literal `"full"` ではなく `string` に評価される**。したがって `typeof 生成定数` で組んだ union は**全て `string` と等価**であり、任意の文字列を型エラーなく受け入れる（第 3 期実行者が tsc スクラッチ検証で実証済み）。この形式の union が 6 件現存し、`Record<退化Union, string>` のラベルマップも連鎖的に `Record<string, string>` 化してキー網羅性検査を失っている。

**修理方針**（生成ファイルは編集禁止のため）: 手書き literal union に戻して型安全性を回復し、**生成定数のランタイム値との一致を vitest の drift テストで機械固定**する（コンパイル時: `const _check: ShiftType = ShiftTypeFull;` が生成側 string 化で無意味になるため、**ランタイム値集合の完全一致アサーション**を正本とする）。

### 1.4 CI ゲート（壊してはならない）

type-check / lint / build / test:coverage + `design-system-audit.mjs`（**第 3 期で全 src 化 + C6 rgba/hsla ルール済み**・allowlist は design-tokens.ts と use-reservation-type-color-map.ts の 2 ファイルのみ）+ eslint-disable ratchet + filename ratchet + coverage ratchet + **knip（ゲート化済み・検出 0 維持必須**・`components/ui/**` と `types/generated/**` は解析対象外）+ codegen-check。

### 1.5 検証コマンド規約（Docker 必須）

- 必ず `docker compose exec frontend ...`。scoped テストは `npx vitest run <path>`（`pnpm test:run -- <path>` は全件実行の罠）。
- フル実行クラス（`pnpm run type-check` / `lint` / `unused` / `design-audit`）は完了条件に明記されたもののみ。権限拒否時はコマンドを人間に提示して実行依頼し、結果を得てから判定（結果なしでコミット禁止）。
- ルートレンダー系テストは `vi.mock("@/hooks/use-auth", ...)` 必須の場合あり。vitest 並行実行はタイムアウト多発 — 疑わしい失敗は単体再実行。

---

## §2 項目 R0: 安全網の構築（最初に必ず実行）

### R0-1. 前提確認

```
docker compose ps                     # frontend コンテナ Up
git rev-parse HEAD                    # 記録
git status --porcelain -- frontend/  # 空であること。空でなければ中断・報告
```

backend/・`BE-refactor.md`・`.claude/` の dirty は無視（触らない・ステージしない）。

### R0-2. ベースライン記録（全て green が着手条件）

```
docker compose exec frontend pnpm run type-check   # エラー 0
docker compose exec frontend pnpm run lint         # エラー 0
docker compose exec frontend pnpm run unused       # 検出 0・exit 0
docker compose exec frontend pnpm run design-audit # PASS
```

1 つでも赤なら着手せず中断・報告。

### R0-3. コミット規約

- **1 項目 = 1 コミット**。メッセージは `<type>(frontend): <説明> (FE4-<n>)`。挙動変更項目（FE4-6/FE4-9）は type を必ず `fix` にする。
- `Co-Authored-By` を含めない。**push しない**。
- 戻し方: 直前項目は `git reset --hard HEAD~1`、それより前は `git revert`。

---

## §3 作業項目リスト（この順に実行する）

> 完了条件を満たせない項目は変更を破棄し、SKIP/BLOCKED として理由を記録して依存されていない次の項目へ進む（依存先 SKIP は連鎖 SKIP）。実装バグを新たに発見したら直さず BLOCKED 記録。

### — Phase A: 退化 union の修理（型安全性の回復 + drift テスト） —

**共通仕様**: 対象 union を手書き literal union に置換し、同一コミットで drift テストを追加する。drift テストの形（各 feature の `types/` 近傍に `*-union-drift.test.ts`）:
```ts
import * as gen from "@/types/generated/models";
// 値集合の完全一致（欠落・過剰の双方向で fail）
expect(new Set<string>(SHIFT_TYPE_VALUES)).toEqual(
  new Set([gen.ShiftTypeFull, gen.ShiftTypeMorning, gen.ShiftTypeAfternoon, gen.ShiftTypeOff, gen.ShiftTypePaidLeave]),
);
```
literal 化で `Record<Union, string>` が厳格化され、**キー不足の型エラーが出た場合**: 意図的な部分集合マップは `Partial<Record<Union, string>>` へ、全域マップはキー追加ではなく**エラー内容を報告して当該 Record の型注釈のみ現状維持**（値・キーの実体は 1 文字も変えない）。

### FE4-1. accounting の退化 union 3 件（AccountingStatus / PaymentMethod / ItemCategory） ✅ DONE (d1e75641)

- **対象**: `frontend/src/features/accounting/types/index.ts:30-56`、新規 `frontend/src/features/accounting/types/union-drift.test.ts`
- **問題**: 3 union とも `typeof 生成定数` 形式で `string` に退化済み（§1.3）。会計ドメイン（金銭）の状態・支払方法・品目カテゴリが任意文字列を受け入れる。
- **変更内容**: literal union へ置換（値は生成定数のランタイム値と同一: AccountingStatus = waiting/pending/completed/cancelled、PaymentMethod = cash/credit_card/electronic_money/bank_transfer、ItemCategory = examination/test/procedure/surgery/medicine/food/goods/other/vaccine/trimming/hotel/training）+ 3 union ぶんの drift テスト（参照する生成定数名: `BillingStatus*`(models.ts:41-45)・`PaymentMethod*`(:46-50)・`ItemCategory*`(:51-63)）。
- **完了条件**: `npx vitest run src/features/accounting` → PASS（drift テスト含む）。`pnpm run type-check` → 0。diff に文字列値の変化が無いこと。
- **リスク / 戻し方**: ラベル Record の厳格化で consumer に型エラー → 共通仕様の手順で対処。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-2. estimates の退化 union 2 件（EstimateStatus 本体 + transforms.ts のファイル内重複） ✅ DONE (1e7bde53)

- **対象**: `frontend/src/features/estimates/types/index.ts:13-17`、`frontend/src/features/estimates/api/transforms.ts:3-13`、新規 drift テスト
- **問題**: EstimateStatus が退化 union。さらに `transforms.ts:9-13` に**同内容のファイルローカル重複定義**があり DRY 違反。
- **変更内容**: types/index.ts 側を literal union（draft/sent/approved/rejected）化 + drift テスト（`EstimateStatus*` models.ts:1148-1152）。transforms.ts のローカル定義を削除し feature 型を import（`../types` 相対）。
- **完了条件**: `npx vitest run src/features/estimates` → PASS。`pnpm run type-check` → 0。`grep -c 'type EstimateStatus' frontend/src/features/estimates -r` → 1。
- **リスク / 戻し方**: 極小。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-3. shifts の退化 union（ShiftType）+ 虚偽 JSDoc の是正 ✅ DONE (6b1495a1)

- **対象**: `frontend/src/features/shifts/types/index.ts:13-19`（union + JSDoc）、`:50-54`（SHIFT_TYPE_LABELS の型が連鎖回復することを確認）、新規 drift テスト
- **問題**: ShiftType が退化 union。しかも JSDoc `:13` が「型安全性のため union 維持」と**事実に反する説明**を掲げている。
- **変更内容**: literal union（full/morning/afternoon/off/paid_leave）化 + drift テスト（`ShiftType*` models.ts:3027-3032）+ JSDoc を実態（literal union + drift テストで生成値に固定）へ書き換え。
- **完了条件**: `npx vitest run src/features/shifts` → PASS。`pnpm run type-check` → 0。
- **リスク / 戻し方**: ShiftCell.tsx:15-19 等の Record 厳格化 → 共通仕様で対処。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-4. 既存 literal union 4 件への drift テスト追加（FE3-3 の別解完遂） ✅ DONE (2b56b254)

- **対象**: 新規テストのみ（実装ファイル無変更）: `src/types/index.ts` の `VisitType`（"first"/"revisit" ↔ `VisitType*` models.ts:2737-2739）と `ReservationStatus`（8 値 ↔ `ReservationStatus*` :2728-2736）、`src/features/medical-records/types/index.ts` の `TreatmentItemType`（4 値 ↔ :3133-3137）と `BodyWeightUnit`（"Kg"/"g" ↔ :3349-3351）
- **問題**: これらは literal union で型安全性は健在だが、backend が値を追加すると黙って古くなる（FE3-3 が typeof 化を試みて SKIP した本来の動機）。drift テストなら型安全性を落とさず追随漏れだけ検知できる。
- **変更内容**: `src/types/union-drift.test.ts` と `src/features/medical-records/types/union-drift.test.ts` を新設し、Phase A 共通仕様の値集合一致アサーションを 4 union ぶん記述。
- **完了条件**: `npx vitest run src/types src/features/medical-records` → PASS。実装 diff 0 行。
- **リスク / 戻し方**: なし（テスト追加のみ）。`git reset --hard HEAD~1`。
- **依存**: R0

### — Phase B: クエリキーの実バグ是正 —

### FE4-5. 死んだ invalidation 2 行の削除（挙動保存） ✅ DONE (e8129170)

- **対象**: `frontend/src/features/accounting/hooks/use-accounting-completion-action.ts:159-160`
- **問題**: `["accounting", id]`・`["accounting-detail", id]` を invalidate しているが、detail クエリの実キーは `queryKeys.accountings.detail(id)` = `["accountings", id]`（get-accounting.ts:17。"accounting-detail" は旧リファクタ前のキーと query-keys.ts:20 のコメントが明言）。**両行ともどのクエリにもマッチしない no-op** で、:158 の `["accountings"]` prefix invalidation が偶然実挙動を担保している。
- **変更内容**: :159-160 の 2 行を削除（:158 の prefix invalidation が `["accountings", id]` を包含するため実行時の再取得挙動は完全に不変）。
- **完了条件**: `npx vitest run src/features/accounting` → PASS。`grep -rn '"accounting-detail"\|\["accounting",' frontend/src --include='*.ts*' | grep -v test | grep -v query-keys` → 0 件。
- **リスク / 戻し方**: なし（no-op 削除）。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-6. `fix:` detail クエリの単複キー不一致 3 件（更新後 stale の解消 — **挙動変更**） ✅ DONE (cd02be43)

- **対象**: `frontend/src/features/vaccinations/api/update-vaccination.ts:31` 付近、`frontend/src/hooks/use-update-reservation.ts:49` 付近、`frontend/src/hooks/use-update-examination.ts:38` 付近
- **問題**: detail クエリのキーが単数形（`["vaccination", id]` get-vaccination.ts:15 / `["reservation", id]` get-reservation.ts:14 / `["examination", id]` get-examination.ts:14）なのに、update mutation は複数形リストキーしか invalidate しない。**更新成功後も詳細画面が古いデータを表示し続ける実バグ**。正しい先例: `update-reservation-route.ts:27-28` は両方を invalidate している。
- **変更内容**: 各 mutation の onSuccess に detail キーの invalidation を追加（先例 update-reservation-route.ts:27-28 の形式をそのまま踏襲。キー文字列は各 get 側の実キーをコピー）。**これは挙動変更（正しい再取得が発生するようになる）のため type=fix のコミットとする。**
- **完了条件**: `npx vitest run src/features/vaccinations src/hooks` → PASS。可能なら各 mutation テストに「detail キーが invalidate される」アサーションを追加。
- **リスク / 戻し方**: 再取得が 1 回増える以外の影響なし。`git reset --hard HEAD~1`。
- **依存**: FE4-5（クエリキー整合の前提を先に掃除）

### — Phase C: 日付ユーティリティ地形の整備 —

（背景: `src/lib/jst-date.ts`（60 消費者）が事実上の標準で健全。負債は端部の重複・シム・inline 再実装と、**"YYYY-MM-DD"→Date の parse 契約が「UTC 深夜」(line-reserve) と「ローカル正午」(DatePickerModel) の 2 系統併存**している点。）

### FE4-7. 日付シム・死に export の掃除 ✅ DONE (a3240d1f)

- **対象**: `frontend/src/features/checkups/lib/today-iso.ts`（2 行の純 re-export・消費者は CheckupsList.tsx の 1 件）、`frontend/src/lib/jst-date.ts` の `formatJSTOffsetDateTime`（外部消費者 0・内部の `jstNowISOString` からのみ参照）
- **問題**: `today-iso.ts` は `todayISODate` へ 2 ホップの間接層。`formatJSTOffsetDateTime` は export される必要がない。
- **変更内容**: CheckupsList.tsx の import を `@/lib/iso-date` 直接参照に書き換えて `today-iso.ts` を `git rm`。`formatJSTOffsetDateTime` の `export` キーワードを除去（関数本体は不変）。
- **完了条件**: `pnpm run type-check` → 0。`pnpm run unused` → 検出 0。`npx vitest run src/features/checkups src/lib` → PASS。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-8. 和式日付表示の共有層整備（特性テスト先行・出力同一） ✅ DONE (2a0f7f04)

- **対象**: `frontend/src/utils/format/date.ts`（関数追加）、`frontend/src/components/shared/DatePicker/DatePickerModel.ts:formatIso/formatDisplay`（委譲化）、`frontend/src/features/hospitalization/components/DailyRecordsTab/DailyDateNav.tsx:53-54`（inline 再実装の置換）
- **問題**: 「Date → `Y年M月D日（曜）`」表示が DatePickerModel・DailyDateNav に個別実装され、`formatIso` は `lib/jst-date.ts` の `formatJSTWallDate` と意味同一の重複。DatePickerModel にはテストが無い。
- **変更内容**（出力を 1 文字も変えないことが絶対条件）:
  1. **特性テスト先行**: `DatePickerModel.test.ts` を新設し、現実装の `formatIso`/`formatDisplay`/`formatShort`/`parseLocalDate` を入出力で固定（例: `parseLocalDate("2026-07-11")` → ローカル正午 Date、`formatDisplay` → `"2026年7月11日（土）"`）。RED→GREEN 確認。
  2. `utils/format/date.ts` に `formatJapaneseDate(date: Date): string`（`Y年M月D日（曜）`・ローカル getter 実装 — 現 DatePickerModel.formatDisplay の逐語移設）を追加。
  3. DatePickerModel の `formatDisplay` を 2. への委譲に、`formatIso` を `formatJSTWallDate`（lib/jst-date.ts）への委譲に置換（特性テストが緑のまま）。
  4. DailyDateNav.tsx:53-54 の inline 実装を `formatJapaneseDate` 呼び出しに置換（同じくローカル getter 由来のため出力同一）。
  5. `lib/jst-date.ts` と DatePickerModel に **parse 契約の JSDoc** を追記: 「"YYYY-MM-DD"→Date の解釈は 2 契約が併存する（line-reserve=UTC 深夜 / DatePicker=ローカル正午）。相互交換は日付ズレを起こすため禁止」。
- **完了条件**: `npx vitest run src/components/shared/DatePicker src/utils/format src/features/hospitalization` → PASS（特性テスト含む）。`pnpm run type-check` → 0。
- **リスク / 戻し方**: 曜日配列・パディングの写経ミス → 特性テストが検出。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-9. `fix:` browser タイムゾーン依存の日付表示 3 件（**挙動変更 — 非 JST ブラウザでの誤日付を修正**） ✅ DONE (0cc29efa・実際の修正対象は PetDeceasedBanner の1件のみ。下記メモ参照)

**実行時メモ**: ShiftFormDialog.tsx:194 / ClinicHolidayModal.tsx:63 は `toLocaleDateString("ja-JP", { timeZone: "Asia/Tokyo", ... })` に既に `timeZone` が明示指定されており、コンテナ内で `TZ=America/New_York` に変えて実測した結果 TZ 非依存(バグ無し)と確認。実際にバグがあったのは PetDeceasedBanner.tsx（ローカル getter 使用）の1箇所のみで、そこだけ `toJSTWallDate` 経由に修正した。

- **対象**: `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:194`、`frontend/src/features/shifts/components/ClinicHolidayModal/ClinicHolidayModal.tsx:63`、`frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedBanner.tsx:15-19`
- **問題**: 前 2 者は `new Date(\`${date}T00:00:00+09:00\`).toLocaleDateString("ja-JP", ...)` — JST 深夜の instant を**ブラウザローカル TZ** で整形するため、非 JST ブラウザでは前日が表示される。PetDeceasedBanner は `new Date(deceasedAt)` + ローカル getter で同クラス。
- **変更内容**: 3 箇所を JST 固定の整形へ置換 — 前 2 者は `parseLocalDate`（DatePickerModel 系のローカル正午 parse）+ `formatJapaneseDate`（FE4-8 で共有化済み）の組合せ、または `toJSTWallDate` + `formatJapaneseDate`。PetDeceasedBanner は `toJSTWallDate(deceasedAt)` を通してから整形。**JST ブラウザでは出力同一・非 JST ブラウザでのみ表示が正しくなる = fix コミット**。曜日表示の有無など現在のフォーマット要素は完全維持。
- **完了条件**: `npx vitest run src/features/shifts src/components/shared/PetDeceasedRecordButton` → PASS。各箇所にユニットテスト追加（JST 相当の入力で現行出力と同一になること。可能なら `TZ=America/New_York` の分岐検証はコメントで方針記載に留める — vitest プロセス TZ 切替はスコープ外）。
- **リスク / 戻し方**: 極小（3 行規模 ×3）。`git reset --hard HEAD~1`。
- **依存**: **FE4-8**（formatJapaneseDate が前提）

### FE4-10. line-reserve の jst-date 双子ファイルを shared-liff へ統合 ✅ DONE (f34632ff)

- **対象**: `frontend/line-reserve/src/lib/jst-date.ts`（46 行・テスト無し・消費者 5 ファイル）→ 新規 `frontend/src/shared-liff/jst-date.ts`
- **問題**: `+9h シフト` アルゴリズムと `addDaysISO` が main 側 `lib/jst-date.ts` と重複。「別アプリだから」という理由は既に虚偽（line-reserve は `@/shared-liff/use-fetch-state` を 7 ファイルで import 済み）。当モジュールだけ日付系で唯一テストが無い。
- **変更内容**: (1) `src/shared-liff/jst-date.ts` を新設し `parseISODate`/`addDaysISO`/`getJSTToday`/`formatJapaneseDate`/`formatJSTApplicationDate` を**逐語移動**（UTC 深夜 parse 契約はそのまま維持 — FE4-8 の JSDoc 契約に従い変換しない）。(2) shared-liff の既存テスト流儀に合わせ `jst-date.test.ts` を新設（各関数の入出力固定・最低 1 ケースずつ + addDaysISO の月跨ぎ）。(3) line-reserve の 5 消費者の import を `@/shared-liff/jst-date` へ書き換え、旧ファイルを `git rm`。line-reserve の `formatJapaneseDate` は UTC getter 実装のため main 側 `formatJapaneseDate`（FE4-8・ローカル getter）とは**統合しない**（契約が異なる — 名前衝突を避けるため shared-liff 側は現名のまま）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run line-reserve/src src/shared-liff` → PASS。`pnpm run unused` → 0。
- **リスク / 戻し方**: import パス書き換え漏れは型チェックで検出。`git reset --hard HEAD~1`。
- **依存**: FE4-8（JSDoc 契約の記載が先）

### — Phase D: ラベル・パス・クエリキーの機械統一 —

### FE4-11. 曜日ラベル 3 複製の共有定数化 ✅ DONE (166a9c04)

- **対象**: `frontend/src/features/master/components/ReservationTypeUnavailableTimesSection.tsx:33`、`ReservationTypeAvailableSlotsSection.tsx:33`、`frontend/src/features/closing-settings/components/StandardClosingTimeSection.tsx:14`
- **問題**: `{0:"日",...,6:"土"}` のバイト同一マップが 3 ファイルに複製（うち 1 つは react-refresh 制約を理由にローカル定義を正当化するコメント付き — 非コンポーネントの共有モジュールに置けば解消する）。
- **変更内容**: `frontend/src/constants/day-of-week.ts` を新設し `export const DAY_OF_WEEK_LABELS: Record<number, string>` を 1 定義。3 ファイルのローカル定義を削除して import（値バイト同一を置換前に diff 確認）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/master src/features/closing-settings` → PASS。`pnpm run unused` → 0。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-12. lstep 配信ステータスラベル 2 複製の統一 ✅ DONE (2298a234)

- **対象**: `frontend/src/features/lstep/components/lstep-analytics-model.ts:25`（`STATUS_LABELS`・消費者 1）→ 正本: `frontend/src/features/lstep/constants/trigger-types.ts:4`（`TriggerStatusLabels`・消費者 2）
- **問題**: scheduled/fired/excluded/failed → 予定/送信済/除外/失敗 の内容同一マップが同一 feature 内に 2 つ。
- **変更内容**: 内容同一を diff 確認 → lstep-analytics-model.ts のローカル定義を削除し `TriggerStatusLabels` を import（参照名は既存コードの流儀に合わせる）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/lstep` → PASS。
- **リスク / 戻し方**: なし。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-13. インラインルートパス 21 件の paths.ts 定数化 ✅ DONE (7de77e79)

- **対象**: `frontend/src/config/paths.ts`（定数 1 件追加）+ 以下 21 箇所: `app/routes/settings-routes.tsx:37-47`（レガシーリダイレクト 11 件・全て既存 `paths.settings.*` に対応）と `:238`（`/settings/shift-templates` — **定数が無いため `paths.settings.shiftTemplates` を新設**）、`features/aggregation/components/AggregationOwnerTableColumns.tsx:68` と `features/lstep/components/TagOwnerListDrawer.tsx:79`（→ `paths.owners.detail.getHref`）、`features/accounting-reports/routes/AccountingReportsPage.tsx:64,261`、`features/manual/routes/ManualPage.tsx:97`、`features/shifts/routes/ShiftTemplateSettings.tsx:154`、`lib/axios.ts:131,157`（→ `paths.auth.login` + 既存クエリ組み立て維持）、`components/errors/RouteErrorBoundary.tsx:51`（→ `paths.home`）
- **問題**: paths.ts が正本として確立しているのに 21 箇所だけ文字列直書き（リンク切れ・リネーム漏れの温床）。
- **変更内容**: 各箇所を対応する `paths.*` 定数/`getHref` へ置換（**生成される URL 文字列が完全同一**であることを置換毎に確認。クエリパラメータ付きは既存の getHref シグネチャに従う）。`paths.settings.shiftTemplates` は既存エントリの形式を踏襲して追加。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/app src/features/aggregation src/features/lstep src/features/accounting-reports src/features/manual src/features/shifts src/lib src/components/errors` → PASS。`grep -rn 'navigate("/\|to="/\|href="/' frontend/src --include='*.tsx' | grep -v test | grep -v paths.ts` の残存が 0 件（または新設不能な例外のみ・報告）。
- **リスク / 戻し方**: URL 文字列の同一性確認を怠ると画面遷移が壊れる → 置換毎の同一性確認を厳守。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-14. query-keys registry の既存範囲への統一（accountings / masters） ✅ DONE (408cfee3・skipリストはコミットメッセージ参照)

- **対象**: inline `["accountings"]` 5 箇所（`use-accounting-settlement-actions.ts:41`、`use-accounting-completion-action.ts:138,158`、`get-accountings.ts:42`、`billing-confirmation.ts:50,78`）と、`queryKeys.masters.category` と**構造同一**の inline `["masters", <cat>]` 箇所（`use-treatment-master.ts` / `use-staffs.ts:41` / `use-animal-species.ts:19` / `use-reservation-type-color-map.ts:77,89` / `get-diagnosis-options.ts` / `get-chief-complaint-types.ts:21` / `get-insurances.ts:13` — 各箇所で registry 出力とキー配列が完全一致する場合のみ）
- **問題**: 同一キーが registry 形式と inline 形式で混在し、キー変更時の drift 事故（FE4-5 で実証済みのクラス）を再生産する。
- **変更内容**: registry 関数の返す配列と inline literal が**構造・値とも完全一致**する箇所のみ、registry 参照へ機械置換。一致しない形（追加要素があるキー等）は**触らずスキップ**して報告（registry の拡張は §4 の方針判断待ち）。
- **完了条件**: `pnpm run type-check` → 0。`npx vitest run src/features/accounting src/hooks` → PASS。置換箇所リストと skip リストを報告。
- **リスク / 戻し方**: キー配列の不一致置換はキャッシュミスを生む → 完全一致確認を厳守。`git reset --hard HEAD~1`。
- **依存**: FE4-5（同一ファイル `use-accounting-completion-action.ts` を触るため直列）

### — Phase E: design-tokens の死にキー削除 —

### FE4-15. PALETTE/STYLE の未使用キー 70 件の削除 ✅ DONE (43f8f65f・実削除は55件。乖離理由はコミットメッセージ参照)

- **対象**: `frontend/src/lib/design-tokens.ts`（153 キー中 70 キーが全 3 アプリで参照 0 — knip はオブジェクトメンバーを検査しないため機械ゲートの盲点。例: `focusBorderInput`、`dot*` 10 件、`btnPrimary`、`warning*` 4 件、`statusGreen/Purple/Gray*` 6 件）
- **問題**: 「単一正本」を名乗る定数カタログの半分近くが死にキーで、読者に「使われている」と誤認させる。1,059 行の过半はこの死荷重。
- **変更内容**: **キー毎に** `grep -rn '<キー名>' frontend/src frontend/liff frontend/line-reserve --include='*.ts*'` で参照 0 を確認してから削除（スプレッド/分割代入/動的アクセスが無いことは監査済みだが、削除直前の再確認を必須とする）。1 件でも参照が見つかったキーは残して報告。削除のみ・値の変更なし。
- **完了条件**: `pnpm run type-check` → 0。`pnpm run design-audit` → PASS。`pnpm run unused` → 0。`npx vitest run src/lib src/components` → PASS。削除キーの全数リストをコミットメッセージに記載。
- **リスク / 戻し方**: 近い将来使う予定のキーを消す可能性 → git revert で完全復元可能（コミットメッセージの全数リストが復元の手引き）。`git reset --hard HEAD~1`。
- **依存**: R0（FE4-1〜3 とファイル無関係・独立）

### — Phase F: テスト基盤 —

### FE4-16. 共有テストラッパー `src/testing/utils.tsx` の新設と機械採用 ✅ DONE (f469a846・実採用27件。skipリストはコミットメッセージ参照)

- **対象**: 新規 `frontend/src/testing/utils.tsx` + 私設 wrapper を持つ 28 テストファイルのうち**変種 (a) 素の QueryClientProvider / (b) +MemoryRouter / (c) +initialEntries に該当する約 20 ファイル**（例: `settings/__tests__/LstepSettingsForm.test.tsx:101`、`aggregation/__tests__/AggregationDashboardPage.test.tsx:47`、`medical-records/api/get-medical-records.test.tsx:25`）
- **問題**: `new QueryClient({ defaultOptions: { queries: { retry: false } } })` + Provider の同型 wrapper が 28 ファイルに私設定義されている。CODING_RULES §8.1 は `src/testing/utils.tsx` の存在を明記しているが実体が無い。
- **変更内容**: (1) `createTestWrapper(opts?: { initialEntries?: string[]; router?: boolean })` を utils.tsx に実装（最頻出変種の逐語一般化。AuthContext 注入など複雑な変種はスコープ外）。(2) 変種 (a)〜(c) のファイルを import 置換（**カスタム Routes / AuthContext.Provider / hoisted spy を含む wrapper のファイルは触らない** — 対象判定に迷ったらスキップして報告）。
- **完了条件**: `npx vitest run <置換した全ファイル>` → PASS（全件・置換前後でテスト数不変）。`pnpm run type-check` → 0。
- **リスク / 戻し方**: テストのみ・ランタイム無風。`git reset --hard HEAD~1`。
- **依存**: R0

### FE4-17. CODING_RULES §8.1 のテスト規約を実態同期 ✅ DONE (4707ad98)

- **対象**: `frontend/CODING_RULES.md` §8.1
- **問題**: §8.1 は「`__tests__/` ディレクトリは使用しない（co-located 必須）」と規定するが、**実態は 194 中 53 ファイルが `__tests__/` 配下**で、規定は死文化している。また存在しない `src/testing/server/` を記載（実体は `src/testing/mocks/`）。`src/testing/utils.tsx` の記載は FE4-16 で実体化済みになる。
- **変更内容**（ドキュメントのみ）: §8.1 を実態に同期 — (1) 配置規定を「新規テストは co-located を推奨。既存の `__tests__/` 配置は許容（一括移動は行わない）」へ改定（53 ファイルの現実を記録した上での規定緩和であることを変更理由として本文に 1 行残す）。(2) `server/` → `mocks/` に訂正。(3) utils.tsx の記述が実体と一致していることを確認。
- **完了条件**: ドキュメントのみのためランタイム検証不要。`grep -n 'testing/server' frontend/CODING_RULES.md` → 0 件。
- **リスク / 戻し方**: なし。`git revert`。
- **依存**: FE4-16（utils.tsx が実在してから記述を確定）

### FE4-18. 800 行超テストファイル 2 件の describe 単位分割 ✅ DONE (7bbbffce)

- **対象**: `frontend/src/features/accounting/components/OwnerAccountingHistory.test.tsx`（867 行）、`frontend/src/features/medical-records/hooks/use-medical-record-form.test.ts`（821 行）
- **問題**: ファイルサイズ上限（800 行ハード）超過はテストファイルも例外ではない（プロジェクトローカル規約に例外規定なし）。
- **変更内容**: 各ファイルを describe ブロック境界で 2 ファイルに機械分割（例: `OwnerAccountingHistory.test.tsx` + `OwnerAccountingHistory.pagination.test.tsx` のように対象サブ機能名を付す）。テストケースの中身・アサーションは 1 文字も変えない。共有 setup（レンダーヘルパ・fixture）は分割後の両ファイルが import できる形（同居ヘルパファイル or FE4-16 の utils）へ逐語移動。
- **完了条件**: `npx vitest run src/features/accounting/components src/features/medical-records/hooks` → PASS かつ**総テスト数が分割前と同一**（分割前に `npx vitest run <対象> --reporter=verbose | tail` でケース数を記録して突合）。両ファイルとも 800 行未満。
- **リスク / 戻し方**: setup 分割ミスはテスト失敗で即検出。`git reset --hard HEAD~1`。
- **依存**: FE4-16（共有 wrapper があれば setup 移動が単純化する）

---

## §4 判断待ちバックログ — 意思決定パック（2026-07-11 実調査済み）

FE4-1〜18 完了後、旧「別トラック」全項目を実調査した結果。**🟢 = 決定不要・そのまま実行依頼できる Ready 項目**（次回実行ラウンドの候補）、**🟡 = 下記の 1 決定が下りれば Ready**、**🔴 = 実バグ（最優先対応推奨）**、**⚪ = 現状維持で確定**。

### 4.1 調査で決着した項目

#### 🔴 zod 調査の副産物: 実バグ 2 件（「二重管理」という項目自体は前提消滅）

zod は全 3 アプリで **8 スキーマのみ・全てレスポンス検証**（リクエスト側の FE/BE 二重管理はゼロ、`min/max/regex/enum` も皆無）。「二重管理」は解消済みの前提だった。代わりに調査が実バグを検出した:

- **M1（HIGH・crash 経路）**: BE `MeResponse.Clinics` が `json:"clinics,omitempty"`（`auth_response.go:18`）なのに FE スキーマは `clinics: z.array(...)` **必須**（`features/auth/api/transforms.ts:59`）。**所属クリニック 0 件のスタッフで `/me` の parse が throw し認証フローが落ちる**。修正はどちらか 1 行: FE に `.default([])`（推奨・FE 完結）or BE の omitempty 除去。 ✅ DONE (8f04bedd・FE5-1・`.default([])` + 回帰テスト2件)
- **M3（HIGH・機能不全）**: liff の `fetchHealthCard` が叩く `GET /v1/liff/health-card`（`liff/src/api/liff-api.ts:51`）は **backend に実装が存在しない**（`health-card` grep 0 件・api.yaml にも無し・自前 grep で再確認済み）。`PetHealthPage` は本番で常にエラー経路に落ちる。**プロダクト判断が必要**: エンドポイント実装（機能作業）か、ページの一時撤去か。 ✅ DONE (8cfb49e2・4b62a38e・7a999264・FE8-1/2/3・エンドポイント実装採用)。`GET /api/liff/:clinicId/health-card` を LiffAuth 配下に新設し FE を追随。実装過程で3独立レビュー（go-reviewer/security-reviewer/clinic-isolation-auditor）が同一の既存クロステナントIDOR（`LineCustomer.Owner` preload に clinic_id 述語欠落、`LinkOwner` 書き込み時未検証）を検出したため、read側の防御的修正を同時実施 (43927aeb)。書き込み側の根本修正（`LinkOwner` の ownerID 所属検証）は別チケット追跡。
- M2（MED）: FE/openapi が期待する `avatar_url` を BE `MeResponse` が持たない三方向 drift（常に null 動作・実害小）。M4（設計注意）: `meClinicInfoSchema` の 17 フィールド `.default()`（税率 0.1 含む）は BE リネーム時に**黙って既定値へ差し替わる**逆方向の失敗モード — 将来の税率系変更時に注意。
- 🟢 再発ゲート（Ready）: tygo.yaml に handler レスポンス struct のパッケージを追加生成し、FE 側で `backendMeResponseSchema satisfies z.ZodType<MeResponse>` を書けば **`pnpm type-check` がフルシェイプ drift ゲートになる**（union-drift テストより強い・依存追加なし・~0.5 日）。

#### 🔴 React Query clinic_id キャッシュ境界 → 調査完了: SAFE-BY-ACCIDENT + マルチタブに実在する穴

単一タブの全切替経路は安全（Sidebar 切替 → `window.location.reload()`（use-auth.tsx:105・回帰テストで固定済み）、logout → `queryClient.clear()`（:92）、401 → フルリロード）。ただし:

- 安全性は**キー設計ではなく reload 1 行に全依存**。コアドメインのクエリキー（accountings/medical-records/pets/owners）は clinic id を含まないため、将来誰かが切替を SPA 化した瞬間に旧クリニックの患者データが 5 分キャッシュから漏れる。`isSwitchingClinic` 定数 false（:59）は SPA 切替設計の痕跡で回帰圧力は現実的。
- **マルチタブの穴（現存）**: `storage` イベント監視が無いため、タブ A で切替してもタブ B は旧クリニック画面のまま。axios はリクエスト毎に localStorage を読む（axios.ts:46）ので、**以後のタブ B の書き込みは全て新クリニックの X-Clinic-ID で永続化される** — カルテ・会計の誤テナント書き込みが成立し得る（同一ユーザーの所属クリニック間に限定）。
- 🟢 対策 2 件（Ready・いずれも挙動保存・小）: ① `CURRENT_CLINIC_STORAGE_KEY` の `storage` イベントで他タブ変更を検知したら `window.location.reload()`。 ✅ DONE (b5bca9a4・FE5-2) ② `switchClinic` 内の reload 直前に `queryClient.clear()`（将来の SPA 化への最後の防壁）。 ✅ DONE (19e6088a・FE5-3) 中期の構造修正（クエリキーに clinic id をルート付与）は query-keys registry 全面採用（4.3）とセットで設計すべき（未着手）。

#### 🟢 dangerouslySetInnerHTML / PrintPortal XSS 監査 → **CLOSED（該当なし）**

`dangerouslySetInnerHTML` は 3 アプリ + テストで **0 箇所**（唯一の既存箇所は `b6f99200`(2026-04-10) で死にコードとして削除済み — 本項目は起票時点で既に陳腐化していた）。PrintPortal は `createPortal` の純 JSX（エスケープ済み・#187 は CSS 特異度バグで injection ではない）。manual は react-markdown で **rehype-raw 不使用**（raw HTML はパース時に落ちる）+ `getSafeMarkdownHref` は decode→制御文字除去→スキーム allowlist の堅牢実装。`document.write`/`innerHTML`/`eval` 系 0 件。CSP `script-src 'self'` がバックストップ。総合リスク **LOW**。
🟢 フォロー（Ready・依存追加不要）: `eslint.config.js` に core ルールだけで再発ガードを追加 — `no-restricted-syntax`（`JSXAttribute[name.name='dangerouslySetInnerHTML']`）+ `no-restricted-properties`（innerHTML/outerHTML/insertAdjacentHTML/document.write）。現状 0 件なので導入即 green。 ✅ DONE (7dc170ea・FE5-4)

### 4.2 1 つの決定で Ready になる項目

#### 🟡 tygo 生成定数の型 widen 解消 → **上流が解決済み・設定 1 行**（決定者: オーナー） ✅ DONE (d8c03478・c3d61fde・FE6-1/FE6-2・下記 FE6 ラウンド参照)

tygo v0.2.20+ の **`enum_style: "union"`** が正にこの機能: literal 定数 + エイリアス自体が literal union になる。全 64 型グループの Go ソース監査済みで検出条件を満たす（例外: `lstep_delivery_trigger_log.go` の 1 const ブロックが 2 型混在で対象外に残る — Go 側でブロック分割すれば解消）。FE 影響調査済み: **型エラー想定ゼロ**（string→エイリアス代入は全て cast 経由・ラベル Record は FE4-1〜3 の手書き union キー・値集合は drift テストが同一性を機械保証）。工数 ~0.5 日。
決定事項: ① `enum_style: "union"` 採用可否（**推奨: 採用**。post-process スクリプト案 B は不要、現状維持 C はドリフト検知がランタイム止まり）② 同時に **tygo バージョン pin**（現状 `@latest` が docker-compose.yml:94 / ci.yml:625 の 2 箇所 — 出力形式が load-bearing になるため再現性穴を塞ぐべき）。
着地後の後続: 手書き union → `import type` 生成型へ移行し drift テスト削除（コンパイル時検知に昇格）。注意: `make codegen` の大 diff を 1 コミットで（codegen-check が赤くなるため分割不可）・models.ts を触る並行セッション禁止。

#### 🟡 ラベル表記の分岐（決定者: PO — 下表を見て表記を 1 つずつ決めるだけ）

| キー | 画面 A | 画面 B | 追加の非対称 |
|---|---|---|---|
| `credit_card` | 会計一覧（AccountingListTable:27,66）+ クレジット訂正（CreditCorrectionDialog:32）=「クレジットカード」 | 日次会計・支払カード・返金（daily-accounting-utils:7 経由）=「カード」 | — |
| ItemCategory 3 値 | 会計明細（AccountingItemRow:23-25）=「処方/フード/物販」 | 見積明細（EstimateLineItems:13-15）=「薬剤/食事/物品」 | 見積側は trimming キー欠落 |
| AccountingStatus | 会計一覧（AccountingListTable:33-35）= waiting **と pending の両方**を「会計待ち」に潰す | 飼主会計履歴（OwnerAccountingHistoryParts:214-217）=「未精算/保留/精算済/取消」 | **表記以前に「一覧で pending を区別するか」の状態粒度の決定が先** |

決定後の作業は FE4-11 と同じ手順の機械 dedup（各 ~S）。line-reserve の ReservationStatus ラベル（`確定/キャンセル済/完了` vs main `予約確定/キャンセル/完了`）は飼主向け文言としての意図差の可能性があり、統一するかも含めて PO 判断。

#### 🟡 `{cond && <JSX>}` の機械強制（決定者: オーナー — devDependency 追加の可否のみ） ✅ DONE (48d83cbc・FE6-8)

決定が下りれば作業は: `docker compose exec frontend pnpm add -D eslint-plugin-react` → `eslint.config.js` に `plugins: { react }` + `rules: { "react/jsx-no-leaked-render": "error" }`（このルールだけ有効化・recommended セットは入れない）。現状違反 0 件で導入即 green・~30 分。代替の `@eslint-react/eslint-plugin` は型情報連携が強みだが本件 1 ルールには過剰。

#### 🟡 シフト休憩時刻 Input の aria-label（決定者: PO — 文言 1 組の承認のみ) ✅ DONE (0154eaf4・FE6-7)

対象 4 箇所（ShiftFormDialog.tsx:304,311 / ShiftTemplateSidePanelFields.tsx:138,145）は `breaks.map((b, i) =>` ループ内。**文言案: `休憩${i + 1} 開始時刻` / `休憩${i + 1} 終了時刻`**（可視の「休憩」セクション見出しと整合）。承認されれば aria-label 4 行追加のみ（~15 分）。

#### 🟡 ForgotPasswordPage の anti-enumeration 矛盾（決定者: PO/セキュリティ） ✅ DONE (3a6d84d9・FE6-6)

現コード確認済み: catch 内で `handleApiError(err, ...)` が **toast を出してから** `{ status: "sent" }` を返す（ForgotPasswordPage 内・「アドレス不存在でも成功として扱う」コメント付き）。ネットワーク断でも toast が出るため列挙防止は部分的にしか破れないが、意図と実装が矛盾。**推奨: catch 内の handleApiError 呼び出しを削除**（1 行・完全な anti-enumeration に整合）。UX 上「失敗が見えない」ことを許容するかだけ決定が必要。

#### 🟡 MedicalRecordForm.tsx（411 行）の再分割（決定者: オーナー — 優先度判断のみ） ✅ DONE (71c5d10a・FE6-9・411→381行)

やるなら具体手順は確定済み: `:339-390` のモーダル塊（MedicalRecordDeleteDialog / Suspense+VitalsModal / Suspense+StaffSelectionModal / Suspense+OwnerSearchModal / ConfirmDialog）を `MedicalRecordFormModals.tsx` へ抽出。props は塊内の自由識別子を列挙して逐語 pass-through（判断不要の機械手順）。効果は 411→~330 行で、依然 LOW 推奨のまま。

### 4.3 現状維持・長期（新情報のみ追記）

| 項目 | 状態 |
|---|---|
| **未 push コミット群** | **44 コミットに増加**（オーナー対応・全計画より優先を推奨） |
| query-keys registry 全面採用（inline ~200 vs registry 14） | 4.1 の clinic-id ルートキー化と**同一の設計判断**なのでセットで別チケット化を推奨（registry 拡張時にキー先頭へ clinicId を含める設計を一度に決める） |
| medical-records `Treatment` の #201 未追随 | 現 interface（medical-records/types/index.ts:23）に `dose_*` 0 件を再確認。#201 FE UI 実装（OPEN）の一部として対応 |
| `TreatmentPlan` stale twin | ✅ DONE (3cb794fd・FE6-4)。`HospitalizationTreatmentPlan` へリネーム。実消費者は hospitalization の3ファイルのみだった（medical-records は同名変数/関数名のみで型 import なし）。master/* の TreatmentPlan* コンポーネント名は無関係のため不変 |
| `BackendTrimming` の codegen 化 | ✅ DONE (75406a78・1805103f・9de59742・FE7-1/FE7-2/FE7-3)。`trimmingResponse` 等3型を export し tygo 生成対象化、FE の `BackendTrimming` を生成型 `TrimmingResponse` extends へ移行。pet/staff の2フィールドのみ生成型が tstype:"-" で除外するため手書き維持（petSummaryResponse/staffSummaryResponse は11+18箇所で共有される非公開型のため今回は export せず） |
| `src/lib/iso-date.ts` 統合 / design-tokens.ts 分割 | 現状維持で確定（YAGNI / 定数カタログ） |

### 4.4 推奨着手順（オーナー向けサマリ）

1. **push + CI 確認**（44 コミット→本ラウンドの FE5-1〜4 の 4 コミットを加えて 48 コミット。全ての前提。未実施）
2. 🔴 M1（1 行）・M3（プロダクト判断）・マルチタブ穴の対策 2 件 — 実害系。**M1・マルチタブ対策 2 件は ✅ DONE（FE5-1/2/3・下記参照）。M3 も ✅ DONE（FE8-1/2/3・下記参照）**
3. 🟡 tygo `enum_style: "union"` + バージョン pin — 型ガード全体が恒久化し、後続の型整理（手書き union 撤去・TreatmentPlan・BackendTrimming・zod satisfies ゲート）が一本道になる（未着手・オーナー決定待ち）
4. 🟢 XSS 再発 ESLint ガード（依存追加なし・15 分） ✅ DONE（FE5-4・下記参照）
5. 🟡 PO 決定 3 件（ラベル表記・aria-label 文言・ForgotPassword）— 決定さえあれば全て S サイズ（未着手・PO 決定待ち）

#### FE5 ラウンド実行結果（2026-07-11・未 push）

§4.1 の決定不要 Ready 項目 + M1 推奨修正を実施。1 項目 = 1 コミット、TDD（RED→GREEN）で実施。

| 項目 | 内容 | 状態 | コミット |
|---|---|---|---|
| FE5-1 | `/me` clinics omitempty crash 是正（`.default([])`） | ✅ DONE | 8f04bedd |
| FE5-2 | マルチタブ storage イベント検知 → reload | ✅ DONE | b5bca9a4 |
| FE5-3 | `switchClinic` reload 前に `queryClient.clear()` | ✅ DONE | 19e6088a |
| FE5-4 | XSS 再発防止 ESLint ガード（依存追加なし） | ✅ DONE | 7dc170ea |

**除外（本ラウンド未着手・残件）**:
- 🟢 codegen 系（tygo に handler レスポンス struct 追加 + `satisfies z.ZodType<MeResponse>` ゲート）— §5.3 codegen 禁止と衝突するため除外。別チケット化推奨。
- 🔴 M3（`GET /v1/liff/health-card` 未実装）— ✅ DONE（別ラウンド・FE8-1/2/3・下記 FE8 ラウンド節参照）。
- 🟡 全 6 項目（tygo `enum_style`、ラベル表記 3 件、`jsx-no-leaked-render` dep 追加、休憩 aria-label、ForgotPassword anti-enumeration、MedicalRecordForm 再分割）— いずれもオーナー/PO 決定待ちのため未着手。
- §4.3 長期項目（query-keys registry 全面、Treatment #201、TreatmentPlan、BackendTrimming、iso-date/design-tokens 分割）— 未着手。
- push / PR — 未実施（本ラウンドの禁止事項）。ローカルに未 push コミットが積み増しされている。

**4 ゲート比較**: R0 ベースラインと完了時点で type-check 0 / lint 0 errors（既存 warning 26 件で不変）/ knip 0 / design-audit PASS 0 — 全て不変（green 維持）。

#### FE6 ラウンド実行結果（2026-07-11・Grill-me Q1-A/Q2-A/Q3-A/Q4-B 承認済み・未 push）

§4.2 の全 🟡 項目（Q3-A で一括採用）+ Q4-B で追加承認された TreatmentPlan リネーム/BackendTrimming codegen を実施。FE6-1→5 は codegen 依存のため直列、FE6-6〜9 は番号順。

| 項目 | 内容 | 状態 | コミット |
|---|---|---|---|
| FE6-1 | tygo `enum_style: "union"` + v0.2.21 pin + lstep_delivery_trigger_log const ブロック分割 | ✅ DONE | d8c03478 |
| FE6-2 | `make codegen` 実行 + union 化で顕在化した型安全性の穴8箇所の是正 | ✅ DONE | c3d61fde |
| FE6-3 | 手書き literal union 9 件を生成型 re-export へ移行・drift テスト5件削除 | ✅ DONE | 02d6af53 |
| FE6-4 | `TreatmentPlan` → `HospitalizationTreatmentPlan` リネーム（型のみ・master/* コンポーネント名は不変） | ✅ DONE | 3cb794fd |
| FE6-5 | MeResponse 追加生成 + BE 契約ゲート導入。BackendTrimming は当時 BLOCKED（**FE7-1〜3 で解消済み・下記参照**） | ✅ DONE（MeResponse 部分） | f57289fb |
| FE6-6 | ForgotPasswordPage anti-enumeration 矛盾是正（回帰テスト付） | ✅ DONE | 3a6d84d9 |
| FE6-7 | シフト休憩時刻 Input 4 箇所へ aria-label 追加 | ✅ DONE | 0154eaf4 |
| FE6-8 | `eslint-plugin-react` 追加 + `jsx-no-leaked-render` 導入（実違反9件を coerce パターンで是正） | ✅ DONE | 48d83cbc |
| FE6-9 | MedicalRecordForm.tsx モーダル塊抽出（411→381行） | ✅ DONE | 71c5d10a |

**当初想定との差分（実行時に判明・すべて Run Summary の「保守側を選ぶ」原則で対応）**:
- **FE6-1**: `lstep_delivery_trigger_log.go` の 2 型混在 const ブロックは「対象外の可能性」ではなく実際に `TriggerStatus` 側が union 化されないことを実測確認。原因は tygo の `detectEnumGroup` が 1 const ブロックにつき先頭定数の型のみでグループ化するため（ソース調査で特定）。ブロック分割（`type X = string` エイリアスは変更しない・純粋に構文分割のみ）で解消し `go build ./internal/model/...` で確認。
- **FE6-1**: `lab_report.go` の3型（LabExamReportSummary/Detail/ResultItem）は P7 準拠で json タグが handler 側 non-public response struct へ既に移設済みのため、model パッケージから素朴に生成すると実際の wire 形式（snake_case）と一致しない PascalCase 型が生成されることを実測発見（`enum_style` 変更前の純粋な version-pin diff でも既に顕在化していた既存 drift）。`exclude_files` で除外（FE 消費者 0 件を確認済み）。
- **FE6-1**: 副次的に AuditLog.ip_address の nullable 化漏れ・SharedFile の幽霊フィールド deleted_at という無関係な既存 drift も modelパッケージから拾われた（いずれも FE 消費者 0 件・実害なしを確認し許容）。
- **FE6-2**: ドキュメントの「型エラー想定ゼロ」という FE 影響予測は誤りだった。union 化で 8 箇所（Resource の "" sentinel・shadcn Select 経由の DosageForm/MedicineUnit/ItemCategory/StaffType/PetGender/AcquisitionType/DangerLevel）が型エラーになった。全て実行時は Select の選択肢や意図的な sentinel に限定される安全な値のみのため、型キャストまたは REVERSE_MAP の値型 narrow で是正（意図通りの「型安全性の穴の可視化」効果）。
- **FE6-5**: 当初案の `backendMeResponseSchema satisfies z.ZodType<MeResponse>` は実装直後に3件の不一致で失敗（avatar_url 三方向 drift・occupation/clinic への意図的な `.nullable()` 防御的許容）。全て「BE より寛容」な方向の不一致でありバグではないため、代わりに `MeResponse extends z.input<typeof schema>`（BE の実際の契約はこのスキーマで必ず parse できる、の方向のみを保証）を採用。M1 相当の退行が再発すれば検知することを実測確認済み（clinics を一時的に必須へ戻し型エラーを確認 → 復元）。
- **FE6-5 BackendTrimming**: 当時 BLOCKED（`trimmingResponse` 非公開のため backend 編集が必要で本ラウンド範囲外）。**FE7-1〜3（2026-07-11・別ラウンド）で解消済み** — 詳細は本ファイル末尾の FE7 ラウンド節参照。
- **FE6-8**: 「現状違反 0 件」は誤りで実際は9件（NavigationBlocker の `when` prop・OwnersListRow の `can*` prop、いずれも boolean 型）。原因は `jsx-no-leaked-render` が JSX children だけでなく JSXAttribute の `prop={a && b}` にも反応する仕様のため（rule ソース確認済み）。ESLint 標準の `--fix`（ternary 戦略）は `boolean | null` を生む型崩壊を起こすため使わず、coerce 戦略（`!!x && y`）を手動適用。
- **FE6-9**: 411→381行（目安の~330より多い）。モーダル抽出コンポーネントへの呼び出しが19個の named prop になったため、インライン JSX 展開より若干行数が増えた。可読性は同等でロジック変更なし。

**除外（本ラウンド未着手・残件）**:
- 🔴 M3（`GET /v1/liff/health-card` 未実装）— ✅ DONE（別ラウンド・FE8-1/2/3・下記 FE8 ラウンド節参照）。
- ラベル表記分岐（`credit_card`/`ItemCategory`/`AccountingStatus`）— PO 判断待ち（Q2-A・引き続き対象外）。
- **BackendTrimming の codegen 化**（FE6-5 の一部）— ✅ DONE（FE7-1/FE7-2/FE7-3・下記 FE7 ラウンド参照）。
- §4.3 長期項目（query-keys registry 全面、Treatment #201、iso-date/design-tokens 分割）— Q4-A 相当のため除外（Q4-B は TreatmentPlan リネームのみ追加承認・実施済み）。
- push / PR — 未実施。ローカルに 48 コミット（FE5 まで）+ 本ラウンド9コミット = 57 コミットが積み増しされている。

**4 ゲート比較**: R0 ベースラインと完了時点で type-check 0 / lint 0 errors（既存 warning 26 件で不変・jsx-no-leaked-render 新規違反 0）/ knip 0 / design-audit PASS 0 — 全て不変（green 維持）。

#### FE7 ラウンド実行結果（2026-07-11・BackendTrimming BLOCKED 解消・未 push）

FE6-5 で BLOCKED だった BackendTrimming の codegen 化を、backend 本番コード編集の許可を得て実施。

| 項目 | 内容 | 状態 | コミット |
|---|---|---|---|
| FE7-1 | `trimmingResponse`/`trimmingCourseSummaryResponse`/`trimmingOptionSummaryResponse` を export（Go 型名のみ・json タグ不変） | ✅ DONE | 75406a78 |
| FE7-2 | tygo.yaml の handler include_files に trimming_response.go 追加 + `make codegen` | ✅ DONE | 1805103f |
| FE7-3 | FE `BackendTrimming` を生成型 `TrimmingResponse` extends へ移行（契約ゲート） | ✅ DONE | 9de59742 |

**実行時に判明した差分**:
- petSummaryResponse/staffSummaryResponse（TrimmingResponse.Pet/.Staff が参照）は 11+18 箇所で共有される非公開型で、petSummaryResponse はさらに animalSpeciesSummaryResponse/ownerSummaryResponse を参照する連鎖的依存を持つ。export を連鎖させるとスコープが「trimming_response.go」を大きく超えるため見送った。
- tygo は同一パッケージ内の無資格型参照（`petSummaryResponse` のような識別子）を type_mappings で置換できない（ソース確認: type_mappings は `pkg.Type` 形式の SelectorExpr にのみ適用され、無資格 Ident はそのまま出力される）。放置すると生成 TS が存在しない型を参照しコンパイルエラーになるため、tygo 公式機能の `tstype:"-"` 構造体タグで Pet/Staff の2フィールドのみ生成対象から除外した。JSON wire 形状・タグは無変更（実際のレスポンスには pet/staff は引き続き含まれる。生成型のみこの2フィールドを欠く）。
- FE 側 `BackendTrimming` は MeResponse（zod スキーマあり）と異なり zod 検証層を持たないため、`extends z.input<...>` ではなく TypeScript interface の `extends TrimmingResponse` で契約ゲートを実現した。RED→GREEN で実測確認（生成型から `status` フィールドを一時的に削除 → `transforms/trimming.ts` で型エラー発生 → 復元）。

**4 ゲート比較**: type-check 0 / lint 0 errors / knip 0 / design-audit PASS 0 — 全て不変（green 維持）。trimming 関連 vitest 5ファイル 40件 PASS。

#### FE8 ラウンド実行結果（2026-07-11・M3 解消・未 push）

M3（`GET /v1/liff/health-card` 未実装・PetHealthPage 常時エラー）をエンドポイント実装で解消（Q1-A/Q2-A 採用）。1 項目 = 1 コミット、TDD（RED→GREEN）で実施。Workflow オーケストレーション（backend service → handler/route/openapi → frontend の順に実装、完了後 go-reviewer/security-reviewer/clinic-isolation-auditor の3並列独立レビュー）で実行。

| 項目 | 内容 | 状態 | コミット |
|---|---|---|---|
| FE8-1 | LiffService.GetHealthCard 新設（Owner/Pets/Vaccinations 集約、clinic_id スコープ、Owner未紐付時は空pets） | ✅ DONE | 8cfb49e2 |
| FE8-2 | `GET /api/liff/:clinicId/health-card` を LiffAuth 配下に追加。P7 `toLiffHealthCardResponse`、api.yaml 追記、401/200/500 テスト | ✅ DONE | 4b62a38e |
| FE8-3 | FE `fetchHealthCard` を新ルートへ追随（clinicId をパスへ、`X-Clinic-ID` ヘッダ廃止）。MSW テストも追随、旧パス参照 0 件 | ✅ DONE | 7a999264 |
| FE8-4 | 本ファイル M3 を DONE 化 | ✅ DONE | 本コミット |

**独立レビューで検出・同時対応した既存バグ（このラウンドの副産物）**:
go-reviewer / security-reviewer / clinic-isolation-auditor の3エージェントが独立に同一のクロステナントIDORを検出した — `LineCustomerRepository.FindByID/FindAll` の `Preload("Owner", "deleted_at IS NULL")` に `clinic_id` 述語が無く、`LineCustomerService.LinkOwner` も書き込み時に `ownerID` の所属クリニックを検証していない（`line_customers.owner_id` が他クリニックの Owner を指せる）。**既存の `GetLiffProfile`（本番稼働中）と共有する pre-existing バグ**だが、`GetHealthCard` が `breed`/`last_visit_date` を追加露出させ露出面が広がるため、read 側の防御的修正を本ラウンド内で実施（43927aeb・`Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID)` + クロステナント回帰テスト追加）。**書き込み側の根本修正**（`LinkOwner` に `ownerRepo.FindByID(ctx, clinicID, *ownerID)` の事前検証を追加）は `lineCustomerService` への新規依存注入を伴う別スコープのため、追跡 Issue 化が必要（未実施）。

**除外（本ラウンド未着手・残件）**:
- MEDIUM: `liff_response.go` の pet_id 変換が `fmt.Sprintf("%d", ...)` — プロジェクト標準の `strconv.FormatUint` からの逸脱（go-reviewer指摘、非ブロッキング）。
- MEDIUM: `PetHealthCard.NextRecommendedVisitDate` は常に nil（次回推奨来院日の算出ロジック未実装・設計上の Assumption どおり、捏造ロジック禁止のため意図的に未実装）。
- 🔴 `LineCustomerService.LinkOwner` の cross-clinic ownerID 書き込み検証欠如（根本修正・別チケット）。
- push / PR — 未実施。

**4 ゲート相当（backend scoped test）**: `go test ./internal/service/...`（LiffService パッケージ全体）/ `go test ./internal/handler/...`（Handler パッケージ全体）/ `go test ./internal/repository/... -run TestLineCustomerRepository` / `TestPreloadClinicScope` — 全て PASS。frontend: `npx vitest run liff/src/pages/PetHealthPage.test.tsx` 4件 PASS。

## §5 やらないことリスト（禁止事項）

1. **FE4-6・FE4-9 以外での挙動変更**。この 2 項目も指定した 3+3 箇所以外に fix を広げない（同類を見つけたら報告のみ）。
2. **§4 別トラック項目への着手**。特にラベル表記分岐の「統一」（どちらかの文言を勝たせる行為）は PO 判断を先取りするため厳禁。
3. **`src/types/generated/models.ts` と backend/・tygo.yaml の編集**（退化 union の根本解決は BE 別トラック）。
4. **依存ライブラリの追加・更新・削除**（package.json / pnpm-lock.yaml 不可侵）。
5. **`components/ui/**` の編集**（本計画に該当項目なし）。
6. **design-tokens.ts の既存キーの値変更・リネーム**（FE4-15 は参照 0 キーの削除のみ）。
7. **クエリキーの「改善」**（形の変更・階層の再設計）— FE4-14 は registry と完全一致する置換のみ。
8. **`__tests__/` 53 ファイルの一括移動**（FE4-17 は規定側を実態に合わせる）。
9. **フル検証コマンドの無断省略・無断実行**。完了条件記載分のみ。拒否時は人間に依頼し結果必須。`pnpm test:run`（全件）・`pnpm install`・`make codegen` 禁止。
10. **push・PR・外部書き込み・squash/rebase/amend 禁止**。進捗は本ファイルの項目見出しに `✅ DONE (hash)` を追記する形でのみ記録。

## §6 実行者への指示文（このままコピペして渡す）

```
あなたはこのリポジトリのフロントエンドリファクタリング実行者である。
リポジトリルートの FE-refactor.md が唯一の作業指示書である。以下を厳守せよ。

1. まず FE-refactor.md を全文読む。次に frontend/CLAUDE.md と
   frontend/src/features/CLAUDE.md を読む。
2. §2 の R0（安全網）を最初に実行する。ローカル FE ゲート（type-check / lint /
   knip / design-audit）が 1 つでも赤なら着手せず報告して終了。
3. §3 の作業項目を FE4-1 から FE4-18 まで番号順に、1 項目ずつ実施する。
   - 1 項目 = 1 コミット。FE4-6 と FE4-9 は必ず fix: type でコミットする。
   - 各項目の「完了条件」のコマンドを全て実行し、満たさない限りコミットしない。
   - フル実行コマンドが権限で拒否されたら人間に実行を依頼し、結果を得てから判定。
   - 満たせない場合: 変更を破棄し SKIP/BLOCKED 記録、依存されていない次の項目へ。
4. 値・出力の同一性が条件の項目（FE4-8/10/11/12/13/14）では、同一性を確認した証拠
   （diff・grep 結果）を報告に残す。FE4-8 は特性テストの RED→GREEN を先に確認する。
5. Phase A で Record の型エラーが出たら、共通仕様の手順（Partial 化 or 注釈現状維持）
   に従い、値とキーの実体は 1 文字も変えない。
6. §4（別トラック）と §5（やらないこと）に該当する作業は実行するな。push するな。
7. 全項目終了後の完了報告に含める:
   - 項目ごとの DONE/SKIP/BLOCKED とコミットハッシュ
   - Phase A の drift テスト一覧と、型エラー対処を行った Record の一覧
   - FE4-13 の置換対応表 / FE4-14 の置換・skip リスト
   - FE4-15 の削除キー全数 / FE4-18 のテスト数突合結果
   - ベースラインと完了時点の 4 ゲート結果比較
```

---

## 付録: 実行順の依存関係まとめ

```
R0 ─┬─ FE4-1, FE4-2, FE4-3, FE4-4        （Phase A: 相互独立・番号順推奨）
    ├─ FE4-5 ─┬─ FE4-6                    （B: クエリキー是正）
    │          └─ FE4-14                  （registry 統一は同一ファイルのため FE4-5 後）
    ├─ FE4-7                              （独立）
    ├─ FE4-8 ─┬─ FE4-9                    （C: 表示層整備 → fix）
    │          └─ FE4-10                  （契約 JSDoc 記載後に双子統合）
    ├─ FE4-11, FE4-12, FE4-13             （D: 相互独立）
    ├─ FE4-15                             （E: 独立）
    └─ FE4-16 ─ FE4-17, FE4-18            （F: wrapper 新設 → 規約同期・分割）
```

- FE4-6/FE4-9 は fix: コミット（挙動変更を含む唯一の 2 項目）。
- FE4-14 は FE4-5 と同一ファイルを触るため直列。
- FE4-9/FE4-10 は FE4-8 の共有関数・契約 JSDoc が前提。
- FE4-17/FE4-18 は FE4-16 の utils.tsx 実在が前提。


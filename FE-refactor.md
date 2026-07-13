# FE-refactor.md — フロントエンド リファクタリング計画書（第6期）

- **作成日**: 2026-07-13
- **基準コミット**: `e3f243ef`(main)。行番号はずれたら**シンボル名で再特定**する。
- **性格**: 本書は実行計画の正本。判断できない事態は**中断して報告**。本書とコード以外の文脈を前提にしない。
- **進捗**: 未着手。各項目完了時に見出しの `[ ]` を `[x]` に更新してよい（同一コミットに `FE-refactor.md` を含めてよい）。

---

## 1. 現状理解（実行者への文脈共有）

### 1.1 全体像

動物病院向け電子カルテ（EMR）のフロントエンド。**単一の Vite プロジェクト（`frontend/`）に3アプリが同居**する（`vite.config.ts` の `build.rollupOptions.input` でマルチエントリ）:

| アプリ | エントリ | 利用者 | データ層 |
|---|---|---|---|
| メインEMR | `frontend/index.html` → `src/main.tsx` | 病院スタッフ | axios(`src/lib/axios.ts`) + TanStack Query v5 |
| LIFF | `frontend/liff/index.html` → `liff/src/main.tsx` | 飼主(LINE内) | 独自 `liff/src/api/liff-api.ts` + `shared-liff/use-fetch-state` |
| LINE予約 | `frontend/line-reserve/index.html` → `line-reserve/src/main.tsx` | 飼主(LINE内) | 独自 `line-reserve/src/api/liff-api.ts` + 同上 |

liff / line-reserve はメインアプリの axios・TanStack Query を**使わない**（認証系統が別）。両者の共有コードは `src/shared-liff/`（`use-liff.ts` / `use-fetch-state.ts` / `handle-fetch-error.ts` / `Spinner.tsx` / `jst-date.ts`）に集約されている。

### 1.2 メインアプリの層構造

- `src/app/` — ルーティング（`createBrowserRouter`、全ルート `lazy` 分割、`errorElement: <RouteErrorBoundary />` を全ツリーに配線済み）。`app/pages/*.tsx`（4ファイル、12〜91行）は**複数 feature を合成する薄い合成層**で、意図されたパターン（CODING_RULES §1.2）。
- `src/features/`（25ディレクトリ） — Feature Indexing。外部からは必ず `@/features/xxx`（barrel `index.ts`）経由。構成: `index.ts` / `api/` / `components/` / `hooks/` / `routes/` / `types/`。**純関数の計算ロジックは feature 直下**に置く前例あり（`cash-register/closing-summary.ts` 等）。
- `src/hooks/` — feature 横断のグローバルフック置き場（`src/hooks/CLAUDE.md` に全量ドキュメント化）。**cross-feature で使うデータ取得フックの正位置**。feature 側は再export で互換維持するパターンが確立済み（例: `features/owners/hooks/use-animal-species.ts` は `@/hooks/use-animal-species` の1行再export）。
- `src/components/shared/` — feature 非依存の共有コンポーネント（`components/shared/CLAUDE.md` が「feature 非依存」を明文化）。`src/components/ui/` は shadcn/ui 生成物（**編集禁止**）。
- `src/lib/` — `axios.ts`（単一クライアント、401リフレッシュ・GETリトライ・X-Clinic-ID付与）/ `react-query.ts`（`QueryCache.onError` → `handle-api-error.ts` → sonner toast で **GET失敗は全てトースト表示される**。FE5-16で是正済み）/ `query-keys.ts` / `design-tokens.ts`（`C`/`STYLE`/`Z` 等。hex直書き禁止）/ `jst-date.ts`（**JST日時演算の正本**）/ `transforms/`（BE⇔FE 変換の正本）。
- `src/config/` — `fetch-limits.ts`（`HISTORY_FETCH_LIMIT=100`、FE3-8で直値集約済み）等。`src/constants/` — `day-of-week.ts`（`DAY_OF_WEEK_LABELS`、0=日〜6=土。FE4-11で3複製を統合済み）等。
- テスト: **対象ファイル隣接配置**（`Foo.tsx` → 同ディレクトリ `Foo.test.tsx`）。`__tests__/` は FE5-23 で全廃済み・新設禁止。src/ に220、line-reserve に12、liff に2ファイル。テスト基盤は `src/testing/`（MSW + `createTestWrapper`）。

### 1.3 既存の機械ゲート（壊してはならない）

CI（`.github/workflows/ci.yml`）: `pnpm type-check` / `pnpm lint` / **knip（未使用コード検出、findingsがあれば fail）** / `pnpm build` / `pnpm test:coverage` + **coverage-ratchet（baseline 43.78% statements、下回ると fail）** / eslint-disable rationale ガード。ローカル専用: `scripts/design-system-audit.mjs` / `scripts/check-feature-filename-convention.mjs`。

**含意**: knip が gating 済みのため「ファイル/exportまるごとのデッドコード」はほぼ残っていない。本計画のデッドコード項目は knip が検出**できない**種類（オブジェクトのキー・恒久デッドブランチ）に限られる。テストを削る変更は coverage-ratchet を割るリスクがあるため、**テスト削除を伴う項目はない**（追加のみ）。

### 1.4 第5期までの到達点（本計画が「やり直さない」こと）

hex直書き・cross-feature import・`__tests__/`・camelCaseフック名・孤立ファイルは過去期で解消済みであることを今回全数再監査で確認済み。本計画は残存する「knip不可視のデッドコード・ローカル再実装による正本乖離・実バグ2件・層逆転1群」に絞る。

---

## 2. 実行者が守る規約

### 2.1 検証コマンド（Dockerスコープ限定）

- テストは **`docker compose exec frontend npx vitest run <path...>`** を使う。`pnpm test:run -- <path>` は**全件実行になる罠**があるため使用禁止。
- フルリポジトリの `pnpm lint` / `pnpm type-check` / `pnpm build` / `pnpm test:run` / `pnpm install` は**自動実行禁止**。必要なら完了報告時にコマンドをユーザーへ提示して手動実行を依頼する。
- `docker compose exec frontend node scripts/design-system-audit.mjs` はスコープ検証として実行してよい。

### 2.2 Git

- main 直作業。**push / PR 禁止**。`Co-Authored-By` を入れない。
- コミットは **1項目=1コミット**。メッセージは `fix(frontend): ...` / `refactor(frontend): ...` / `docs(frontend): ...`（各項目に指定あり）。
- `git add` は**ファイル指定のみ**（`git add -A` / `git add .` 禁止）。

### 2.3 dirty ファイル（触らない）

以下は別ワークストリームの未コミット変更。**変更・コミット・stash 禁止**:

- `backend/` 配下全て（`.coverage-baseline`、`internal/repository/*_test.go` 群）
- `docs/coverage-policy.md` / `q&a.html`
- 例外: `FE-refactor.md`（本書）の進捗欄更新のみ可。

---

## 3. FE6-0 [ ] 安全網の構築（最初に実行）

**目的**: 変更前の green 状態を記録し、各項目の完了条件の基準にする。

1. `git status --porcelain` を実行し、dirty が §2.3 のリスト＋本書のみであることを確認。**想定外の dirty があれば中断して報告**。
2. `git rev-parse --short HEAD` を記録（以後の戻し先）。
3. ベースラインテストを1回実行し、**全PASSであることを確認・記録**する:

```bash
docker compose exec frontend npx vitest run \
  src/components/shared/OwnerSearchModal \
  src/components/shared/ReservationFormModal \
  src/features/auth/hooks/use-auth-clinic-switch.test.tsx \
  src/features/accounting/components/AccountingDocument.test.tsx \
  src/features/accounting/components/UnpaidTab.test.tsx \
  src/features/settings/components/LstepTagConfigSection.test.tsx \
  src/features/shifts/components/ShiftCalendar/ShiftCalendar.test.tsx \
  src/lib/jst-date.test.ts \
  src/features/trimming src/features/medical-records/hooks \
  liff/src line-reserve/src
```

4. 1件でも FAIL する場合は**着手せず中断して報告**（既存赤の上に積まない）。
5. **各項目に着手する直前**にも、その項目の「完了条件」に書かれた vitest コマンドを**変更前に1回実行**する。変更前から赤の場合はその項目をスキップし、理由を最終報告に含める（上記ベースラインは代表であり全項目の対象を網羅しない）。
6. FE6-0 でのコミットは不要（変更がないため）。以後、各項目の「戻し方」は原則 `git revert <該当コミット>`（コミット済みの場合）または `git checkout -- <対象ファイル>`（未コミットの場合）。

> 特性テストの新設は各項目内に組み込んである（FE6-1: REDテスト先行 / FE6-2: 失敗系ケース追加 / FE6-4: バウンダリテスト新設 / FE6-8: ドリフトガード / FE6-14: 税計算の現挙動固定）。FE6-0 での追加作業はない。

---

## 4. 作業項目（実行順）

### Phase 1 — fix: 実バグとエラー処理の穴

#### FE6-1 [ ] fix: 飼主検索モーダルの飼主名が常に空表示になるバグ

- **対象**: `src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx:18-46,66-69` / `OwnerSearchModal.test.tsx:46`
- **問題**: `GET /v1/owners` のレスポンスは `ownerResponse` Go struct（`backend/internal/handler/owner_response.go:60` — `json:"owner_name"`）だが、本ファイルはローカル定義の `transformOwner` で generated `Owner` 型（`json:"name"`）を前提に `o.name` を読むため、**検索結果・確認ダイアログの飼主名が常に空文字になる**。正本 `src/lib/transforms/owner.ts` の `transformOwner`（`owner_name` を読む）を再利用していないローカル再実装が原因。既存テストのMSWフィクスチャも `name: "山田 太郎"` を返しており（`OwnerSearchModal.test.tsx:46`）、実BEと乖離したモックがバグを固定している。呼び出し元は `PetEditModal` / medical-records のモーダル群＝**飼主変更フロー**。
- **どう変える**（テスト先行）:
  1. `OwnerSearchModal.test.tsx` のフィクスチャを実BE準拠に変更: `name: "山田 太郎"` → `owner_name: "山田 太郎"`（`phone`/`address1` 等は `ownerResponse` のフィールド名のまま）。検索結果に「山田 太郎」が表示されることを assert するテストがなければ追加する（`テスト名: 検索結果に飼主名が表示される`）。この時点でテストが **RED** になることを確認。
  2. 実装をローカル変換の廃止＋正本委譲に変更:

```tsx
// Before（抜粋）
import type { Owner as BackendOwner } from "@/types/generated/models";
import { MEMBERSHIP_TYPE_FROM_API } from "@/lib/transforms/owner";
function transformOwner(o: BackendOwner): OwnerSummary {
  return { id: String(o.id ?? 0), name: o.name ?? "", ... };
}

// After（抜粋）
import { transformOwner, type OwnerApiResponse } from "@/lib/transforms/owner";
function toOwnerSummary(o: OwnerApiResponse): OwnerSummary {
  const owner = transformOwner(o);
  return {
    id: owner.id,
    name: owner.ownerName,
    phone: owner.phone,
    address: [owner.address1, owner.address2].filter(Boolean).join(" "),
    discountRate: owner.discountRate,
    membershipType: owner.membershipType,
  };
}
// axios.get の型引数も { data: OwnerApiResponse[] } に変更
```

- **完了条件**: `docker compose exec frontend npx vitest run src/components/shared/OwnerSearchModal` 全PASS（新assert含む）。
- **リスク / 戻し方**: 低（表示専用モーダル。API・保存経路の変更なし）。失敗時 `git revert`。
- **コミット**: `fix(frontend): 飼主検索モーダルの飼主名欠落を正本transformOwner委譲で修正`
- **依存**: FE6-0

#### FE6-2 [ ] fix: クリニック切替時の localStorage 書込失敗が無音で旧クリニック継続になる

- **対象**: `src/features/auth/hooks/use-auth.tsx:18-27`（`saveClinicToStorage`）、同 `switchClinic`（114-131行付近） / `src/features/auth/hooks/use-auth-clinic-switch.test.tsx`
- **問題**: `saveClinicToStorage` は書込失敗を DEV 限定 `console.warn` で握りつぶす。`switchClinic` は失敗でも `queryClient.clear()` → `window.location.reload()` を続行するため、リロード後に**旧クリニックIDのまま復帰し、ユーザーは切替成功と誤認する**。マルチテナントEMRで「誤クリニック状態が無音で継続する」経路は放置不可。
- **どう変える**:

```tsx
// saveClinicToStorage: 成否を返す
function saveClinicToStorage(clinicId: string): boolean {
  try {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, clinicId);
    return true;
  } catch (error) {
    if (import.meta.env.DEV) console.warn("[auth] failed to save clinic to localStorage", error);
    return false;
  }
}

// switchClinic: 失敗時はトーストを出して reload しない
if (!saveClinicToStorage(clinicId)) {
  toast.error("クリニックの切替に失敗しました。ブラウザのストレージ設定を確認してください。");
  return;
}
queryClient.clear();
window.location.reload();
```

  `toast` は `sonner` から import（`handle-api-error.ts` と同じ流儀）。
  テスト追加（`use-auth-clinic-switch.test.tsx` に、既存のモック流儀を踏襲して）: `テスト名: localStorage書込が失敗した場合はreloadせずエラートーストを表示する` — `localStorage.setItem` を throw するようスタブし、`window.location.reload` 未呼び出しと `toast.error` 呼び出しを assert。
- **完了条件**: `docker compose exec frontend npx vitest run src/features/auth/hooks/use-auth-clinic-switch.test.tsx` 全PASS（新ケース含む）。
- **リスク / 戻し方**: 低（失敗パスのみ挙動変更、成功パスは不変）。失敗時 `git revert`。
- **コミット**: `fix(frontend): クリニック切替のlocalStorage書込失敗を無音継続からトースト+中断に変更`
- **依存**: FE6-0

#### FE6-3 [ ] refactor: liff / line-reserve の ErrorPage を shared-liff に統合

- **対象**: `frontend/liff/src/pages/ErrorPage.tsx`（23行） / `frontend/line-reserve/src/pages/ErrorPage.tsx`（23行） / 両 `App.tsx` の参照箇所
- **問題**: 両ファイルは構造・文言・props が同一で、差分は Tailwind 色クラス（`liff-brand*` vs `noah-teal*`）のみの95%コピペ。shared-liff への集約から漏れた唯一のページコンポーネント。
- **どう変える**: `src/shared-liff/ErrorPage.tsx` を新設。既存2実装の共通シグネチャに**色クラスを差し込む prop を1つだけ追加**する（例: `accentClassName: string`。既存の文言・DOM構造・その他のクラスは一切変えない）。両アプリの `App.tsx` 等の参照を `@/shared-liff/ErrorPage` に差し替え、各アプリのローカル `ErrorPage.tsx` を削除。呼び出し側は現行の自アプリ色クラスをそのまま渡す。
- **完了条件**: `rg -l "ErrorPage" liff/src line-reserve/src src/shared-liff` で実装ファイルが `src/shared-liff/ErrorPage.tsx` の1つのみ。`docker compose exec frontend npx vitest run line-reserve/src liff/src` 全PASS。
- **リスク / 戻し方**: 低（見た目同一の移設）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): liff/line-reserveのErrorPage重複をshared-liffへ統合`
- **依存**: FE6-0

#### FE6-4 [ ] fix: liff / line-reserve に React エラーバウンダリを導入

- **対象**: `frontend/liff/src/main.tsx` / `frontend/line-reserve/src/main.tsx`（+ 新規 `src/shared-liff/ErrorBoundary.tsx`）
- **問題**: 両アプリとも `createRoot(...).render(<StrictMode><App/></StrictMode>)` のみで、エラーバウンダリも `onUncaughtError` もない（`rg "componentDidCatch|getDerivedStateFromError" liff/src line-reserve/src` → 0件）。飼主向け画面でレンダー時例外が1件でも起きると**白紙のまま復旧手段なし**。メインアプリは全ルートに `RouteErrorBoundary` 配線済みで、この2アプリだけ防御層がない。
- **どう変える**: 外部パッケージは**追加しない**。`src/shared-liff/ErrorBoundary.tsx` を自前クラスコンポーネントで新設:

```tsx
import { Component, type ReactNode } from "react";

interface Props { fallback: ReactNode; children: ReactNode }
interface State { hasError: boolean }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };
  static getDerivedStateFromError(): State { return { hasError: true }; }
  componentDidCatch(error: unknown): void {
    if (import.meta.env.DEV) console.error("[shared-liff] uncaught render error", error);
  }
  render(): ReactNode { return this.state.hasError ? this.props.fallback : this.props.children; }
}
```

  両 `main.tsx` で `<App/>` を `<ErrorBoundary fallback={<ErrorPage ...現行の自アプリ色クラスと固定文言... />}>` で包む。fallback の文言は「エラーが発生しました。お手数ですが、もう一度開き直してください。」で固定（ErrorPage の既存 props 契約に合わせて渡す）。
  テスト新設 `src/shared-liff/ErrorBoundary.test.tsx`: `テスト名: 子がthrowした場合にfallbackを表示する` / `テスト名: 正常時はchildrenを表示する`（throwする子コンポーネントを渡して assert。`vi.spyOn(console, "error")` でノイズ抑制）。
- **完了条件**: `docker compose exec frontend npx vitest run src/shared-liff` 全PASS。`rg -n "ErrorBoundary" liff/src/main.tsx line-reserve/src/main.tsx` で両エントリに配線済み。
- **リスク / 戻し方**: 低（正常系パススルー）。失敗時 `git revert`。
- **コミット**: `fix(frontend): liff/line-reserveにエラーバウンダリを導入し白紙画面を防止`
- **依存**: FE6-3（fallback に統合後の ErrorPage を使う）

### Phase 2 — refactor: 重複の統合

#### FE6-5 [ ] refactor: formatJSTDate / padDatePart の3重実装を lib/jst-date.ts に統合

- **対象**: `src/features/trimming/hooks/trimming-form-utils.ts:3,26-33` / `src/features/medical-records/hooks/use-medical-record-form-model.ts:8,27-34`（正本: `src/lib/jst-date.ts:16`）
- **問題**: `JST_OFFSET_MS` + `padDatePart` + `formatJSTDate` が正本とバイト同一ロジックで2ファイルに再実装されている。将来 lib 側に修正が入っても追従しない（ドリフトリスク）。
- **どう変える**: 両ファイルでローカル `formatJSTDate` の**定義を削除**し、`import { formatJSTDate } from "@/lib/jst-date";` + 既存の外部利用者向けに `export { formatJSTDate };` を残す（trimming 側は `use-trimming-form.ts` が、medical-records 側は `use-medical-record-form.ts` / `use-medical-record-auto-create.ts` が本シンボルを import しているため再export必須）。**注意**: `use-medical-record-form-model.ts` のローカル `formatJSTDateTime` / `createDateAtCurrentJSTTime`（36-47行）は `padDatePart`/`JST_OFFSET_MS` を使い続けるため、この2定数・関数は**モデル側にローカルのまま残す**（lib に秒精度 `+09:00` の同一フォーマッタは存在しない — `rg "export" src/lib/jst-date.ts` で確認済み）。trimming 側は `padDatePart`/`JST_OFFSET_MS` の残利用が消えるなら定義ごと削除。
- **完了条件**: `rg -n "function formatJSTDate" src/features` → 0件。`docker compose exec frontend npx vitest run src/features/trimming src/features/medical-records/hooks src/lib/jst-date.test.ts` 全PASS。
- **リスク / 戻し方**: 低（同一ロジックへの委譲）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): formatJSTDateの3重実装をlib/jst-dateへ統合`
- **依存**: FE6-0

#### FE6-6 [ ] refactor: UnpaidTab の JST 日付演算を lib/jst-date.ts に移設

- **対象**: `src/features/accounting/components/UnpaidTab.tsx:23-34`（`daysSince` / `currentJSTYearMonth`）
- **問題**: JSTオフセットの手動計算（`Date.now() + 9*60*60*1000` 等）がコンポーネント内に再実装されており、JST演算の正本 `lib/jst-date.ts` と実装が割れている。
- **どう変える**: `daysSince(dateString: string): number` と `currentJSTYearMonth(): string` を **`src/lib/jst-date.ts` に移設**（実装は現行のまま移す。挙動を変えない）し、`src/lib/jst-date.test.ts` に境界ケースを追加（`テスト名: daysSinceは日付跨ぎをJST基準で数える` / `テスト名: currentJSTYearMonthはJSTの年月を返す` — `vi.setSystemTime` でUTC 15:00前後=JST日付跨ぎを固定して assert）。`UnpaidTab.tsx` は import に差し替え。
- **完了条件**: `docker compose exec frontend npx vitest run src/lib/jst-date.test.ts src/features/accounting/components/UnpaidTab.test.tsx` 全PASS。`rg -n "9 \* 60 \* 60 \* 1000" src/features/accounting` → 0件。
- **リスク / 戻し方**: 低。失敗時 `git revert`。
- **コミット**: `refactor(frontend): UnpaidTabのJST日付演算をlib/jst-dateへ移設`
- **依存**: FE6-5（jst-date.ts への変更を直列化して衝突回避）

#### FE6-7 [ ] refactor: 曜日ラベルの重複3箇所を DAY_OF_WEEK_LABELS 由来に統一

- **対象**:
  - `src/features/master/components/ReservationTypeAvailableSlotsSection.tsx:21-31`
  - `src/features/master/components/ReservationTypeUnavailableTimesSection.tsx:22-32`
  - `src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx:203`
- **問題**: FE4-11 で共有化された `DAY_OF_WEEK_LABELS`（`src/constants/day-of-week.ts`）が存在し、master の2ファイルは**それを import 済みなのに**、`<SelectItem value="0">日曜日</SelectItem>` … という同一JSXブロック `DAY_OF_WEEK_ITEMS` を各自ハードコードしている（統合後に放置された再発コピペ）。ShiftCalendar も `["日","月",...][dayOfWeek]` をインライン再導入している。
- **どう変える**:
  1. `src/features/master/components/day-of-week-select-items.tsx` を新設し、両セクションの `DAY_OF_WEEK_ITEMS` 定義を1箇所に統合:

```tsx
import { SelectItem } from "@/components/ui/select";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";

export const DAY_OF_WEEK_SELECT_ITEMS = Object.entries(DAY_OF_WEEK_LABELS).map(
  ([value, label]) => (
    <SelectItem key={value} value={value}>{label}曜日</SelectItem>
  ),
);
```

  2. 両セクションのローカル `DAY_OF_WEEK_ITEMS` を削除し `DAY_OF_WEEK_SELECT_ITEMS` を import（生成結果が既存JSXと同一 value/文言になることを目視確認。`DAY_OF_WEEK_LABELS` は 0=日〜6=土）。
  3. `ShiftCalendar.tsx:203` のインライン配列を `DAY_OF_WEEK_LABELS[dayOfWeek]` に置換（このファイルの dayOfWeek も 0=日曜。既存表示と同一になる）。
- **完了条件**: `rg -n "日曜日</SelectItem>" src/features/master` のヒットが新設ファイル1箇所のみ。`rg -n '\["日", "月"' src` → 0件。`docker compose exec frontend npx vitest run src/features/shifts/components/ShiftCalendar src/features/master` 全PASS。
- **リスク / 戻し方**: 低（表示同一の置換）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 曜日ラベル重複3箇所をDAY_OF_WEEK_LABELS由来に統一`
- **依存**: FE6-0

#### FE6-8 [ ] test: Pet属性ラベルの二重定義にドリフトガードを追加

- **対象**: `src/features/owners/types/index.ts:10-15`（`PET_GENDER_VALUES` / `ACQUISITION_TYPE_VALUES` / `DANGER_LEVEL_VALUES`） vs `src/lib/transforms/pet.ts:21-58`（`PET_GENDER_MAP` / `ACQUISITION_TYPE_MAP` / `DANGER_LEVEL_MAP`）
- **問題**: 同じ日本語ラベル群（雄/雌/不明、購入/譲渡/保護/その他、低/中/高）が「UI選択肢の配列」と「BE⇔FE変換マップ」に手打ちで二重定義されている。目的が違うため**単純統合は不可**（配列はリテラル型の源泉、マップはEN⇔JA変換）。片方だけ変更されると無音で乖離する。
- **どう変える**: 実装は変更せず、**乖離したらテストが落ちるガード**を追加する。`src/features/owners/types/pet-label-consistency.test.ts` を新設:

```ts
import { PET_GENDER_MAP, ACQUISITION_TYPE_MAP, DANGER_LEVEL_MAP } from "@/lib/transforms/pet";
import { PET_GENDER_VALUES, ACQUISITION_TYPE_VALUES, DANGER_LEVEL_VALUES } from "./index";

test("owners の Pet 属性選択肢は transforms の変換マップと一致する", () => {
  expect(new Set(PET_GENDER_VALUES)).toEqual(new Set(Object.values(PET_GENDER_MAP)));
  expect(new Set(ACQUISITION_TYPE_VALUES)).toEqual(new Set(Object.values(ACQUISITION_TYPE_MAP)));
  expect(new Set(DANGER_LEVEL_VALUES)).toEqual(new Set(Object.values(DANGER_LEVEL_MAP)));
});
```

  MAP 側が export されていない場合のみ `export` を付ける（純追加）。
- **完了条件**: `docker compose exec frontend npx vitest run src/features/owners/types` 全PASS。
- **リスク / 戻し方**: なし（テスト追加のみ）。失敗時はテストファイル削除。
- **コミット**: `test(frontend): Pet属性ラベルの二重定義にドリフトガードを追加`
- **依存**: FE6-0

#### FE6-9 [ ] refactor: line-reserve TrimmingOptionSelectPage の通貨整形を formatCurrency に委譲

- **対象**: `line-reserve/src/pages/TrimmingOptionSelectPage.tsx:15-18`
- **問題**: 兄弟ページ `TrimmingCourseSelectPage.tsx:16-19` は共通 `formatCurrency`（`@/utils/format/number`）へ委譲済みだが、本ファイルの同名 `formatPrice` は `` `+¥${price.toLocaleString()}` `` を独自再実装しており、通貨整形の将来変更に追従しない。
- **どう変える**: `` const formatPrice = (price: number) => `+${formatCurrency(price)}`; `` に変更（「+」プレフィックスは加算表示の仕様なので維持。`formatCurrency` が `¥` を含むことを `src/utils/format/number.ts` で確認してから置換）。
- **完了条件**: `docker compose exec frontend npx vitest run line-reserve/src/pages/TrimmingOptionSelectPage.test.tsx` 全PASS。
- **リスク / 戻し方**: 低。失敗時 `git revert`。
- **コミット**: `refactor(frontend): TrimmingOptionSelectPageの通貨整形をformatCurrencyへ委譲`
- **依存**: FE6-0

### Phase 3 — refactor: 直値・デッドコード

#### FE6-10 [ ] refactor: 一覧ルート5箇所の `pageSize: 20` 直書きを削除

- **対象**: `src/features/owners/routes/OwnersList.tsx:166` / `inventory/routes/InventoryList.tsx:121` / `vaccinations/routes/VaccinationList.tsx:104` / `hospitalization/routes/HospitalizationList.tsx:206` / `examinations/routes/ExaminationsList.tsx:117`
- **問題**: `usePagination`（`src/hooks/use-pagination.ts:39`）のデフォルトが既に `pageSize = 20` なのに、5ルートが同値を再指定している。デフォルト変更時に5箇所が追従しない直値散在。
- **どう変える**: 5箇所の `pageSize: 20,` 行を**削除**し、デフォルトに依存させる（他のオプション指定は不変）。
- **完了条件**: `rg -n "pageSize: 20" src/features` → 0件。`docker compose exec frontend npx vitest run src/features/owners/routes src/features/inventory src/features/vaccinations src/features/hospitalization src/features/examinations` 全PASS。
- **リスク / 戻し方**: 低（挙動同一）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 一覧ルートのpageSize直書きをusePaginationデフォルトに集約`
- **依存**: FE6-0

#### FE6-11 [ ] refactor: `limit: 100` 直書き6箇所を fetch-limits 定数に集約

- **対象**: `src/hooks/use-get-reservations.ts:32` / `src/features/reception/api/get-reception.ts:26` / `owner-report/api/get-pet-examinations.ts:14` / `medical-records/api/get-record-examinations.ts:38` / `medical-records/api/get-diagnosis-options.ts:33,40`
- **問題**: `src/config/fetch-limits.ts` に `HISTORY_FETCH_LIMIT = 100`（FE3-8で集約済み）が存在するのに、同セマンティクスの直値 100 が6箇所で再発している。
- **どう変える**: `src/config/fetch-limits.ts` に以下を追記し、各サイトを置換（**値は全て100のまま。挙動変更なし**）:

```ts
// 予約/受付の日別ビューを1リクエストで取り切るための上限（BUG #82 の経緯は use-get-reservations 参照）
export const DAY_VIEW_FETCH_LIMIT = 100;
// 選択肢マスタ（診断オプション等）の取得上限
export const OPTIONS_FETCH_LIMIT = 100;
```

  - `use-get-reservations.ts:32` / `get-reception.ts:26` → `DAY_VIEW_FETCH_LIMIT`（BUG #82 の既存コメントは保持）
  - `get-pet-examinations.ts:14` / `get-record-examinations.ts:38` → 既存 `HISTORY_FETCH_LIMIT`
  - `get-diagnosis-options.ts:33,40` → `OPTIONS_FETCH_LIMIT`
  - ※ `get-lstep-csv-imports.ts` の `limit = 20` は関数デフォルト引数であり対象外。
- **完了条件**: `rg -n "limit: 100" src` → 0件（コメント内の言及は除く）。`docker compose exec frontend npx vitest run src/hooks src/features/reception src/features/owner-report src/features/medical-records/api` 全PASS。
- **リスク / 戻し方**: 低（値不変）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 一括取得limit直書き6箇所をfetch-limits定数へ集約`
- **依存**: FE6-0

#### FE6-12 [ ] refactor: design-tokens の死にキー11個を削除

- **対象**: `src/lib/design-tokens.ts` の `LAYOUT` 配下: `sidebar.expandedPx` / `sidebar.collapsedPx` / `propertyRow.minH` / `propertyRow.labelWPx` / `touch.sm` / `touch.row` / `touch.tableHead` / `touch.iconBtn` / `touch.badge` / `modal.lg` / `pageIcon.size`
- **問題**: 全リポジトリ横断 grep で参照0件のキー（knip はオブジェクトのキー単位を検出できないため残存）。`touch.md` のみ生存（5箇所で使用）なので**巻き込み削除禁止**。`PALETTE`/`C`/`BADGE`/`ICON`/`STYLE`/`Z`/`Z_CLASS`/`TABLE_STYLES` は全キー生存確認済みで対象外。
- **どう変える**: 上記11キーを削除する。**削除前に1キーずつ** `rg -n "LAYOUT\.<parent>\.<key>" src liff line-reserve` を実行して0件を再確認（計画作成時点から使用が増えている可能性に備える。1件でもヒットしたキーは**残して報告**）。親オブジェクトが空になった場合（例: `propertyRow`）は親ごと削除。
- **完了条件**: 削除した各キーの rg が0件。`docker compose exec frontend node scripts/design-system-audit.mjs` が exit 0。`docker compose exec frontend npx vitest run src/lib` 全PASS。
- **リスク / 戻し方**: 低（未参照の削除）。型エラーが出た場合は参照が実在した証拠なので**即 revert して報告**。
- **コミット**: `refactor(frontend): design-tokens LAYOUTの未使用キー11個を削除`
- **依存**: FE6-0

#### FE6-13 [ ] refactor: 恒久デッドブランチ isSwitchingClinic を削除

- **対象**: `src/features/auth/hooks/use-auth.tsx:62-63,158,168` / `src/types/auth.ts:78` / `src/components/shared/Layout/Layout.tsx:8,33-38` / これらを参照するテストのモック
- **問題**: `const isSwitchingClinic = false;` がハードコードされており（コメント「クリニック切替はフルリロードで行うため常に false」）、`Layout.tsx:33-38` の切替中オーバーレイJSXは**実行時に到達不能**。将来のSPA切替用の足場だが、現設計（フルリロード方式、FE5-3コメント参照）が正である以上 YAGNI。git 履歴に残るため復元可能。
- **どう変える**: `rg -n "isSwitchingClinic" src`（テスト含む）で全参照を列挙し、`AuthContextValue`（`types/auth.ts:78`）のフィールド・`use-auth.tsx` の定義とcontext値・`Layout.tsx` の分岐とオーバーレイJSX・テストモックの該当プロパティを**すべて削除**する。
- **完了条件**: `rg -c "isSwitchingClinic" src` → 0件。`docker compose exec frontend npx vitest run src/features/auth src/components/shared/Layout` 全PASS。
- **リスク / 戻し方**: 低〜中（AuthContext 型変更がテストモックに波及。rg 全数列挙で対処）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 到達不能なisSwitchingClinic切替オーバーレイを削除`
- **依存**: FE6-2（use-auth.tsx の変更を直列化して衝突回避）

### Phase 4 — refactor: 構造（M規模）

#### FE6-14 [ ] refactor: AccountingDocument の消費税内訳計算を純関数に抽出（特性テスト先行）

- **対象**: `src/features/accounting/components/AccountingDocument.tsx:89-104`（`taxBreakdown` useMemo）
- **問題**: 標準/軽減税率の内訳計算（`Math.floor`/`Math.round` の丸め規則を含む、領収書・明細書に印字される法定金額）がコンポーネント内 useMemo にベタ書きされ、**単体テストの死角**になっている（`AccountingDocument.test.tsx:11-12` 自身が「税率自体はどのテストも検証しない」と明記）。cash-register では同種の金額計算を feature 直下の純関数に抽出するパターンが確立済み（`closing-summary.ts` 等）。
- **どう変える**（特性テスト先行）:
  1. `src/features/accounting/tax-breakdown.ts` を新設し、useMemo の中身を**一字一句変えずに**純関数 `calcTaxBreakdown(...)` として持ち出す（引数は現 useMemo が閉じ込めている値をそのまま引数化。戻り値の形も現行のまま）。
  2. `src/features/accounting/tax-breakdown.test.ts` を新設し、**現挙動を固定する特性テスト**を書く: `テスト名: 標準税率のみの明細で内訳を計算する` / `軽減税率のみ` / `標準・軽減混在` / `端数はMath.floorで切り捨てる` / `明細0件で全て0を返す`。期待値は移設した現行ロジックから手計算で導出（= 現挙動の固定であり、正しさの再判定はしない）。
  3. `AccountingDocument.tsx` の useMemo を `calcTaxBreakdown` 呼び出しに差し替え。
- **完了条件**: `docker compose exec frontend npx vitest run src/features/accounting/tax-breakdown.test.ts src/features/accounting/components/AccountingDocument.test.tsx` 全PASS。
- **リスク / 戻し方**: 低（ロジック移設のみ、挙動固定テスト付き）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 消費税内訳計算をtax-breakdown純関数へ抽出し特性テストを追加`
- **依存**: FE6-0

#### FE6-15 [ ] refactor: LstepTagConfigSection の3セクション複製を汎用コンポーネントに統合

- **対象**: `src/features/settings/components/LstepTagConfigSection.tsx`（391行。`AutoManagedPrefixesSection`:49行〜 / `ConditionTagMappingsSection`:164行〜 / `SendPurposeTagPrefixesSection`:271行〜）
- **問題**: 3関数（各約110行）がフィールド名以外ほぼ同一構造（2つの `useState` → 追加ハンドラ（trim+必須チェック+toast+create mutation）→ 削除ハンドラ → loading/空/一覧レンダリング → 追加フォームJSX）。素朴な3重コピペで、修正が常に3箇所必要。
- **どう変える**: **フックを props で渡さない**（Rules of Hooks 事故防止）。プレゼンテーション部分だけを汎用化し、3セクションは自前のフック呼び出しを保持する薄いコンテナに変える:

```tsx
// 同ファイル内（または settings/components/TagConfigListSection.tsx に分離）
interface TagConfigListSectionProps<TItem> {
  title: string;
  description: string;
  items: TItem[] | undefined;
  isLoading: boolean;
  getId: (item: TItem) => string | number;
  renderRow: (item: TItem) => ReactNode;   // 行の表示部分（削除ボタン以外）
  onDelete: (item: TItem) => void;
  form: ReactNode;                          // 追加フォーム（入力構成はセクションごとに差異があるため slot 注入）
}
function TagConfigListSection<TItem>(props: TagConfigListSectionProps<TItem>) { /* 共通の一覧+空+loading描画 */ }
```

  3セクションは「useState・mutation・handleAdd/handleDelete・フォームJSX」を保持したまま、一覧描画部分を `TagConfigListSection` に委譲する。handleAdd/handleDelete の共通骨格（trim → 必須チェック → toast）が素直に関数抽出できる場合のみ、同ファイル内のローカルヘルパ（非フック）として1つに畳む。**既存テスト `LstepTagConfigSection.test.tsx` は変更しない**（挙動パリティのゲートとして使う）。
- **完了条件**: `docker compose exec frontend npx vitest run src/features/settings/components/LstepTagConfigSection.test.tsx` を**無変更のまま**全PASS。
- **リスク / 戻し方**: 中（JSX構造の変化で既存テストのクエリが落ちる可能性 → その場合は表示構造が変わった証拠なので実装側を直す。テストの書き換えで逃げない）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): LstepTagConfigSectionの3セクション複製を汎用一覧コンポーネントへ統合`
- **依存**: FE6-0

#### FE6-16 [ ] refactor: cross-feature データフック3本を src/hooks/ の正位置へ移設

- **対象**:
  - `src/features/reservations/api/get-reservation.ts`（21行、`useGetReservation`）→ `src/hooks/use-get-reservation.ts`
  - `src/features/master/api/reservation-type-unavailable-times.ts`（122行）のうち `useGetUnavailableTimes` とその fetch 関数・`ReservationTypeUnavailableTime` 型 → `src/hooks/use-reservation-type-unavailable-times.ts`
  - `src/features/reservations/api/update-reservation-route.ts`（`useUpdateReservationRoute`）→ `src/hooks/use-update-reservation-route.ts`
- **問題**: `components/shared/ReservationFormModal/`（feature 非依存が規約の層）が `@/features/reservations` / `@/features/master` / `@/features/owners` に依存している（層逆転、8 import。`components/shared/CLAUDE.md` 違反）。一方でプロジェクトには「feature 横断で使うデータフックは `src/hooks/` に本体を置き、feature 側は再export で互換維持」という確立済みパターンがある（実例: `features/owners/hooks/use-animal-species.ts` は `@/hooks/use-animal-species` の1行再export。`use-get-reservations` / `use-update-reservation` も既に global）。
- **どう変える**（機械的な移設。ロジック変更なし）:
  1. 上記3ファイルのフック本体・fetch 関数・専用型を `src/hooks/` の新ファイルへ**そのまま移す**（queryKey・staleTime 等は不変）。
  2. 元ファイルは再exportのみに変える（例: `export { useGetReservation } from "@/hooks/use-get-reservation";`）。feature barrel（`index.ts`）の export は不変 — **feature 内外の既存 import は全て無変更で動く**。
  3. `reservation-type-unavailable-times.ts` に master 画面専用の CRUD mutation 等が同居している場合、**それらは master に残す**（移すのは `useGetUnavailableTimes` 系のみ）。
  4. `src/hooks/CLAUDE.md` のフック一覧に3本を追記（既存エントリの体裁に合わせる）。
- **完了条件**: `docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal src/features/reservations src/features/master src/hooks` 全PASS。`rg -n "useGetReservation|useGetUnavailableTimes|useUpdateReservationRoute" src/hooks | rg "export"` で3本とも global 化。
- **リスク / 戻し方**: 低〜中（移設先の import 解決漏れ。テストで検出可能）。失敗時 `git revert`。
- **コミット**: `refactor(frontend): 予約系cross-featureフック3本をsrc/hooksへ移設`
- **依存**: FE6-0

#### FE6-17 [ ] refactor: ReservationFormModal 群の層逆転を解消（@/features import 0件化）

- **対象**: `src/components/shared/ReservationFormModal/` 配下の全8 import:
  - `ReservationFormModal.tsx:14-15`（`useGetReservation` / `NewOwnerFormData`）
  - `ReservationFormModalPanels.tsx:9-10`（`ReservationRouteSelect` / `NewOwnerFormData`, `ReservationRoute`）
  - `ReservationFormFields.tsx:5`（`useGetUnavailableTimes`）
  - `NewOwnerInlineForm.tsx:8-9`（`useAnimalSpecies` / `NewOwnerFormData`）
  - `reservation-time-utils.ts:2`（`ReservationTypeUnavailableTime` 型）
- **問題**: FE6-16 の通り。この項目で shared 層からの `@/features` 依存を完全にゼロにする。
- **どう変える**:
  1. **フック import の差し替え**: `useGetReservation` → `@/hooks/use-get-reservation`、`useGetUnavailableTimes`・`ReservationTypeUnavailableTime` 型 → `@/hooks/use-reservation-type-unavailable-times`、`useAnimalSpecies` → `@/hooks/use-animal-species`（既存 global。1行変更）。
  2. **`NewOwnerFormData` 型の正位置移動**: この型は予約モーダル内の新規飼主インラインフォームの形であり、CODING_RULES §3.6（型定義は `src/types/` に置き、feature は re-export のみ）に従って `src/features/reservations/types/index.ts:39` から `src/types/reservation-form.ts`（新設）へ**定義を移動**。`features/reservations/types/index.ts` は `export type { NewOwnerFormData } from "@/types/reservation-form";` の re-export に変える（barrel 経由の既存利用は無変更で動く）。モーダル側は `@/types/reservation-form` から import。`ReservationRoute` 型も同ファイルの import 元を確認し、`src/types/` 側に正本があればそちらへ、なければ同様に移動する。
  3. **`ReservationRouteSelect` の移設**: `src/features/reservations/components/ReservationRouteSelect.tsx`（65行。依存は design-tokens・ui・`useUpdateReservationRoute` のみ＝FE6-16 で global 化済み）を `src/components/shared/ReservationRouteSelect/ReservationRouteSelect.tsx` へ移動し、import を `@/hooks/use-update-reservation-route` に変更。`features/reservations` 側は元パスに re-export を残し、barrel export も不変とする（feature 内の既存利用は無変更で動く）。
- **完了条件**: **`rg -n "from \"@/features/" src/components/shared/` → 0件**（本計画の主目的ゲート）。`docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal src/components/shared/ReservationRouteSelect src/features/reservations` 全PASS（`ReservationFormModal.test.tsx` は無変更で通ること）。
- **リスク / 戻し方**: 中（移動対象のimport解決・型経路の付け替え漏れ。テスト＋rgゲートで検出）。失敗時 `git revert`（本項目は1コミットに閉じる）。
- **コミット**: `refactor(frontend): ReservationFormModal群の層逆転を解消しshared層のfeature依存を0件化`
- **依存**: FE6-16

### Phase 5 — docs

#### FE6-18 [ ] docs: 規約文書と実態の乖離3点を是正

- **対象と変更**:
  1. `frontend/CODING_RULES.md:1798` 付近 — 「既存の `__tests__/` 配置（FE4-17 時点で 53 ファイル）は許容する」の段落を削除し、「`__tests__/` は FE5-23 で全廃済み。テストは対象ファイル隣接配置とし、`__tests__/` の新設は禁止（正本: `frontend/CLAUDE.md`）」に置換。実測: `src`/`liff`/`line-reserve` に `__tests__/` は0件。
  2. `frontend/CLAUDE.md` §1.1 のディレクトリツリー例 — `usePagination.ts` / `useReducedMotion.ts` / `useSortableList.ts` という camelCase 表記を実ファイル名（`use-pagination.ts` / `use-reduced-motion.ts` / `use-sortable-list.ts`）に修正。
  3. `frontend/CODING_RULES.md` §3.6 — `src/types/index.ts:17-24` が7 feature の型を逆方向に re-export している「FA9 パターン」（transform の ReturnType から型を調達する意図的設計。コード側コメントに FA9 と記載あり）が §3.6 の文言と矛盾しているため、§3.6 に「例外(FA9): transform 関数の ReturnType を型の正本とする場合に限り、`src/types/index.ts` から feature barrel の型を re-export してよい」を追記して実態を明文化。
- **完了条件**: 上記3点の grep（`rg -n "FE4-17 時点" frontend/CODING_RULES.md` → 0件 / `rg -n "usePagination.ts" frontend/CLAUDE.md` → 0件 / `rg -n "FA9" frontend/CODING_RULES.md` → 1件以上）。コード変更なしのため実行時検証は不要。
- **リスク / 戻し方**: なし。
- **コミット**: `docs(frontend): テスト配置・hooks命名・FA9型例外の規約文書を実態に同期`
- **依存**: FE6-17（FA9 記述は型移動後の最終状態を反映するため最後に行う）

---

## 5. やらないこと（実行者はこれらに手を出さない）

1. **機能追加・仕様変更・UI文言変更**（本書に明記されたトースト/フォールバック文言の追加を除く）。
2. **依存ライブラリの追加・更新・削除**。`react-error-boundary` 等のパッケージ導入も禁止（FE6-4 は自前実装）。Sentry 等のテレメトリ導入も禁止（判断待ち、§6参照）。
3. `RECEPTION_TELEMETRY_PHASE2_ENABLED`（`use-reception-telemetry.ts:21`）の削除・変更。恒久 true に見えるがキルスイッチとしての存置判断が未確定（§6）。
4. `line-reserve/src/components/Calendar.tsx` の `DAYS_OF_WEEK`（月曜始まり）と `LineReservationSettingsFormSections.tsx` の `WEEKDAYS`（月〜土）の統合。`DAY_OF_WEEK_LABELS`（0=日始まり）と**契約が異なる**ため FE6-7 の対象外。
5. `src/features/owners/components/pet-edit-field-shared.tsx` のリネーム・移動（JSX定数を含むため `.ts` 化は不可。監査時の誤検出）。
6. `src/components/ui/`（shadcn生成物）・`src/types/generated/`（tygo生成物）の編集。
7. `use-*-form` 系フック（vaccination/trimming/examination等）の共通スケルトン抽象化。ドメインロジックが実質的に異なり、抽象化は害と判定済み。
8. `ChangePasswordDialog` / `ShiftFormDialog` / `TreatmentRow` の構造リファクタ（§6 に次期送り）。
9. `types/index.ts` の FA9 構造自体の変更（FE6-18 でドキュメント明文化のみ）。
10. バックエンド・`api.yaml`・migration・seed への変更（FE6-1 はFE側のマッピング修正で完結する）。
11. テストの削除・skip 化（coverage-ratchet を割るため）。既存テストの変更は本書が明示した箇所のみ。
12. knip 設定・CI workflow・eslint 設定の変更。
13. push / PR 作成 / フルリポジトリ検証コマンドの自動実行。

---

## 6. 次期監査への引き継ぎ（実行者は無視してよい）

- **テレメトリ不在（MEDIUM）**: 本番のレンダー例外・API失敗を運用チームが観測する手段がない（Sentry等未導入）。導入はコスト・依存追加を伴うためPO判断。
- **`RECEPTION_TELEMETRY_PHASE2_ENABLED` の扱い**: 恒久 true のフラグ。キルスイッチとして残すか、フラグとfalse分岐テストを削除するか要判断。
- **OwnerSearchModal の React Query 化**: FE6-1 はバグ修正に留めた。`useState`+`useTransition` の素朴fetchを feature 側 `useSearchOwners` フックに置き換える構造改善は次期。
- **`ShiftFormDialog` の `use-shift-form.ts` 抽出 / `TreatmentRow` の EditableCell 化 / `ChangePasswordDialog` の api 層整理**: いずれも実害なしの一貫性改善。
- **liff / line-reserve の `index.html` に CSP メタタグがない**（メインアプリのみ設定済み）。セキュリティ観点の追加検討。
- **`src/lib/` と `src/utils/` の役割分担が不文律**（両方にフォーマット系が分散）。規約明文化候補。
- **z-index の中間スケール**（sticky/dropdown 用）が未整理。`Z.overlay` 以外は Tailwind 標準スケールのまま（FE5-4 の意図的スコープ限定）。
- **export されているが外部参照のない型シンボル約15件**（`CPMStageOption` 等）: `export` キーワード除去のみの薄い掃除。knip が型exportを gating しない設定か確認の上で次期にまとめて。
- **`.filename-baseline`（値23）** の ratchet を 0 に向けて下げる余地。
- **FE6-8 の単一ソース化**: 二重定義はガードテストでの乖離検知に留めた（PRODUCT_PHILOSOPHY ②「二重管理禁止」との緊張）。`Object.values(MAP)` 由来でリテラル型を保ったまま片方を導出する単一ソース化（`as const satisfies` 等）を次期に検討。

---

## 7. 実行者への指示文（このままコピペして渡す）

```
あなたは AnimalEkarte のフロントエンド実行者です。FE-refactor.md（第6期）を実行してください。

1. frontend/CLAUDE.md と FE-refactor.md 全文を読む。本書とコード以外の文脈は存在しない前提で作業する。
2. FE6-0（安全網）から着手し、以後 FE6-1 → FE6-18 を番号順に、1項目ずつ実施する。
3. 各項目は「変更 → 完了条件のコマンドを実行 → 全PASS確認 → 指定メッセージでコミット（git add はファイル指定のみ）」の順。完了条件を満たせない場合は、その項目の変更を revert し、中断して状況を報告する。勝手な代替実装をしない。
4. 1項目=1コミット。複数項目をまとめない。本書の進捗欄 [x] 更新は同一コミットに含めてよい。
5. §2.3 の dirty ファイルに触らない。§5「やらないこと」に列挙された変更を行わない。
6. push・PR作成・フルリポジトリの lint/type-check/build/test は実行しない。
7. 全項目完了後、以下をユーザーに提示して手動実行を依頼する:
   $ docker compose exec frontend pnpm type-check
   $ docker compose exec frontend pnpm lint
   $ docker compose exec frontend pnpm test:run
   $ docker compose exec frontend pnpm unused
8. 最後に、各項目のコミットハッシュ一覧と、中断・スキップした項目があればその理由を報告する。
```

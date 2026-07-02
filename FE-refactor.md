# フロントエンド リファクタリング計画

- **作成日**: 2026-07-02
- **対象**: `frontend/` の**3アプリ全体** — `src/`（main・919 srcファイル）＋ `liff/`（523行）＋ `line-reserve/`（2,695行）。src/ だけを対象にすると LINE 系2アプリが漏れるため明示する。
- **スタック**: React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui / TanStack Query
- **性格**: 全項目 **behavior-preserving（挙動保存）** を原則とする負債返済計画。振る舞いを変える修正（bug.md の H-1 FE側・M-5 等）は**バグ修正トラックであり本計画の対象外**（相互参照のみ）。BE-refactor.md と対を成す。
- **根拠**: 本日の実測（下表）で FE は BE 同様に極めて健全と確認。全面リファクタは不要であり、**実在する負債7件（FD1-FD7）に的を絞る**。投機的リファクタは行わない。

---

## 1. 現状評価（2026-07-02 実測）

### 健全な点（リファクタ不要と判断する根拠）

| 観点 | 実測値（3アプリ合算） | 評価 |
|---|---|---|
| TODO/FIXME/HACK | **0件**（src/ 0・liff/ 0・line-reserve/ 0） | 放置負債なし |
| 型逃げ | `as any` 実質0（唯一のヒットは旧実装を説明するコメント）・`: any` 0・`@ts-ignore` 0 | 「any 禁止」規約が完全に守られている |
| console.log | 0件（全3アプリ・テスト除く） | デバッグ残骸なし |
| Feature Indexing | 全26 features に index.ts あり・**feature 横断の直接 import（index 迂回）0件**（正規表現で再計測し確定） | 規約完全準拠 |
| 最大コンポーネント | `DailyAccountingTab.tsx` 636行（上限800内）。上位20全て800行未満 | 分割不要 |
| CI ゲート | Frontend job は type-check / test:coverage / lint / build の**4ゲート実行中**（ci.yml:323-348。過去の「lint+build のみ」は #186 系で解消済み — 陳腐化していたメモリを本日訂正） | ゲート構成は健全（ただし type-check の対象範囲に穴 → FD8） |
| E2E 基盤 | Playwright spec **47本**＋専用ワークフロー `e2e.yml`（paths-filter 起動・自前 JWT でスタック起動） | 主要フロー（owners/examinations/accounting/hospitalization 等）の E2E は整備済み |
| lint 対象範囲 | `eslint .` + `files: **/*.{ts,tsx}` で **3アプリ全てをカバー**（ignores は dist/generated 等のみ） | 健全 |

### 残存する負債（本計画の対象）

| # | 負債 | 証拠（現HEADで確認済み） | リスク |
|---|---|---|---|
| FD1 | lstep 系の API 層規約逸脱 | `features/settings/integrations/lstep/hooks/` の4ファイル（useTriggerPriorities / useLstepTagCodeMappings / useLstepSettings / useLstepTagConfig）＋ `features/lstep/components/LstepCsvImportSection.tsx` が `@/lib/axios` を直接使用。他 feature は全て `features/*/api/` 層経由 | データ取得規約の例外。クエリキー・エラー処理・変換の実装ドリフト |
| FD2 | react-hooks 系 eslint-disable 23件 | `react-hooks/exhaustive-deps` 15件＋`react-hooks/set-state-in-effect` 8件。**このプロジェクトは render 中 setState + useActionState の stale closure 本番バグ前歴あり**（6月・修正済み） | 抑制の中に同型バグが潜んでいても lint では二度と検出されない |
| FD3 | design-tokens 逸脱の生 hex カラー | features/ + components/ の tsx に 8件（テスト除く） | トークン変更（テーマ調整）が波及しない箇所が残る |
| FD4 | line-reserve のテスト実質ゼロ | 2,695行・**テストファイル1件のみ**。顧客（飼主）向け LINE 予約導線という最も外部露出の高い画面群。LINE 予約の実機 E2E も未実施（6/11 の明示的積み残し） | 予約フロー回帰が検出できない。障害＝顧客直撃 |
| FD5 | FE カバレッジゲート不在 | CI は test:coverage 実行＋artifact 保存のみで非ゲート（bug.md H-7 と同根） | カバレッジ低下が検出されない |
| FD6 | PrintPortal 全 inactive 時の白紙化が構造未保証 | bug.md #187 LOW: 全ポータル `active={false}` だと印刷が白紙化しうる（JSDoc で呼び出し側責務と明記済み） | 規約は文書のみ。将来の呼び出し側ミスを実行時に検出できない |
| FD7 | その他の lint 抑制 7件 | `react-hooks/preserve-manual-memoization` 3・`react-refresh/only-export-components` 2・`no-control-regex` 2（＋liff 2・line-reserve 1） | 軽微。FD2 の監査に同梱して根拠コメントの有無だけ揃える |
| FD8 | **type-check が liff/ と line-reserve/ を対象外** | `tsconfig.json:24` `include: ["src"]` のみ。liff/・line-reserve/ に個別 tsconfig も無し（現HEADで確認済み）。Vite build は esbuild で型を剥がすだけのため、**LINE系2アプリ（約3,200行）の型エラーは検出手段ゼロ** | CI の type-check ゲートが実質 1/3 アプリしかカバーしていない。顧客向け予約導線の型破損が本番まで素通りし得る |

**方針**: FD8 が「検査されていると思われていたものが検査されていない」穴で即効性が最も高い。FD2 が唯一「潜在バグ」に直結するため精査価値が高い。FD1/FD3 は規約統一（機械的・低リスク）。FD4/FD5/FD6 はガードレール。

---

## 2. フェーズ計画

規模: S=半日以内 / M=1日 / L=2-3日。各項目は独立コミット。

### Phase 1: 規約統一（低リスク・即効）

#### R-F1. lstep 系データ取得の api 層移設（FD1）— 規模 M

- **現状**: lstep 統合設定まわりの hooks 4本とコンポーネント1本が `@/lib/axios` を直接呼び、`features/*/api/`（fetch 関数 + useQuery/useMutation ラッパー + transform の3点セット）という他 feature 共通の構造に従っていない。
- **あるべき姿**: `features/settings/integrations/lstep/api/` （または `features/lstep/api/`）に get/update 関数を切り出し、hooks は api 層を消費するだけにする。クエリキーは既存の feature 慣行（配列キー + clinicId スコープ）に合わせる。
- **手順**:
  1. 5ファイルの axios 呼び出しを列挙し、エンドポイント単位で `api/` ファイルへ移設（URL・パラメータ・レスポンス変換を**そのまま**移す。改善したい点があっても本コミットでは変えない）。
  2. hooks 側は import の付け替えのみ。呼び出し元コンポーネントは無変更。
  3. 既存の lstep テスト（あれば）GREEN 維持＋移設した api 関数のリクエスト形状を pin する msw テストを最低1本/エンドポイント。
- **検証**: `npx vitest run src/features/settings/integrations/lstep src/features/lstep`（`--` 付きは全件実行になる既知罠に注意）＋ lstep 設定画面の手動疎通（Lステップ Write API は現在 noop 化中のため read 系のみ確認）。

#### R-F2. 生 hex カラーの design-tokens 置換（FD3）— 規模 S

- **現状**: features/ + components/ に生 hex 8件。プロジェクト規約はデザイントークン（`lib/design-tokens.ts` 971行に集約済み）参照。
- **手順**: 8件を列挙 → 各色に対応する既存トークンへ置換。**対応トークンが無い色は追加せず、その色が意図的（印刷用・外部ブランド色等）かを判定して根拠コメントを付けて残す**（トークンの投機的追加はしない）。視覚変化ゼロが原則 — トークン値と hex 値が一致する場合のみ置換し、一致しない場合は「どちらが正か」を個別判断（視覚が変わるなら fix: に分離）。
- **検証**: 該当画面のスナップショット/表示確認。type-check + scoped vitest。

### Phase 2: hooks 健全化（最重要・慎重に）

#### R-F3. eslint-disable 30件の全件監査と分類処理（FD2/FD7）— 規模 M

- **現状**: react-hooks 系 23件＋その他 7件の lint 抑制。stale closure 本番バグ（render 中 setState + useActionState）の前歴があるプロジェクトで、`exhaustive-deps` 15件と `set-state-in-effect` 8件は**同型バグの温床になり得る**。
- **手順（監査 → 分類 → 処理の3段階。一括修正はしない）**:
  1. 30件を台帳化（ファイル:行・抑制ルール・対象コード・抑制理由の有無）。
  2. 各件を3分類する:
     - **(a) 正当**（意図的な deps 省略・mount 時のみ実行等）→ 抑制行の直上に**根拠コメントを必須化**して残す。現状コメント無しの抑制はこれを機に全て理由付きにする。
     - **(b) 機械的に解消可能**（useCallback/useMemo 化や deps 追加で挙動不変が証明できる）→ 抑制を外して修正。挙動不変は該当コンポーネントのテストで pin。
     - **(c) バグ隠蔽**（deps 追加で挙動が変わる＝現在の挙動が壊れている疑い）→ **本計画から fix: として分離**し、個別に再現テスト→修正（stale closure 前歴と同じ手順: effect 同期化等）。
  3. `set-state-in-effect` 8件は特に (c) の可能性を優先的に疑う（render/effect フェーズの状態同期は前歴バグと同カテゴリ）。
  4. 完了後、**「新規 eslint-disable は根拠コメント必須」を eslint ルール（`eslint-comments/require-description` 相当）で機械強制**できるか確認し、可能なら導入（これが本項のガードレール成果物）。
- **検証**: 分類 (b) の各修正ごとに `npx vitest run <対象パス>` + 対象画面の手動確認。全体 type-check。
- **注意**: React 19 + React Compiler 環境で `preserve-manual-memoization` の3件は Compiler の最適化前提と衝突していないかを個別確認。

### Phase 3: テスト・ガードレール整備

#### R-F7. type-check の3アプリ化（FD8）— 規模 S・**Phase 3 だが最初に着手すべき**

- **現状**: `tsconfig.json` の `include: ["src"]` により、`pnpm type-check`（tsc --noEmit）が liff/・line-reserve/ を検査していない。lint と build は3アプリをカバーしているのに、型検査だけ main アプリ限定。
- **手順**:
  1. `tsconfig.json` の include に `"liff/src"` / `"line-reserve/src"` を追加する（両アプリが main と同じ compilerOptions で通るならこれが最小）。エイリアスや環境型が異なる場合は各アプリに tsconfig を置き、ルートを project references 化する。
  2. **追加した瞬間に既存の型エラーが露出する可能性がある** — 露出分は同一 PR で解消する（今まで検査されていなかったことの実害証明そのもの。件数が多ければ台帳化して分割）。
  3. CI は既存の `pnpm run type-check` ステップのまま自動で3アプリ対象になる。
- **検証**: `docker compose exec -T frontend pnpm type-check` が3アプリのファイルを対象に 0 errors（意図的に liff 内へ型エラーを一時注入して検出されることを確認 — temp-revert 実証）。
- **順序上の注意**: R-F4（line-reserve テスト整備）の前提。型も検査されていないコードにテストだけ足すのは順序が逆。

#### R-F4. line-reserve の重要フローテスト整備（FD4）— 規模 L

- **現状**: 顧客向け LINE 予約アプリ 2,695行に対しテスト1件。3アプリ中で「外部露出が最も高く、テストが最も薄い」逆転状態。実機 E2E も未実施のまま積み残し。
- **手順（全網羅ではなく重要フロー優先）**:
  1. ユニット/コンポーネント（vitest + msw）: 予約作成フローの主要ステップ（枠選択→顧客情報入力→確認→送信）、入力バリデーション、API 失敗時のエラー表示、`CustomerInfoPage.tsx`（328行・最大）のフォーム挙動。
  2. カバレッジ計測対象に line-reserve/ と liff/ が含まれているかを vite.config の coverage 設定で確認し、漏れていれば含める（**計測外は FD5 のゲートからも漏れる**）。
  3. 実機 E2E（LINE アプリ内 LIFF 起動）は環境依存のため本計画では**手順書化まで**（`docs/testing/` に確認手順を追記）とし、自動化は別トラック。
  4. liff/（523行）は規模が小さく健全なため、リンク導線の smoke テスト1本で足りる。
- **検証**: `npx vitest run line-reserve liff` GREEN。カバレッジレポートに両アプリが出現すること。

#### R-F5. FE カバレッジ ratchet ゲート（FD5）— 規模 S

- **現状**: CI は計測と artifact 保存のみ。bug.md H-7 / BE-refactor.md R3-5 と同一施策の FE 側。
- **手順**: coverage-policy.md の Phase 1-2 に沿って「現状値を下回ったら warn→fail」の ratchet を frontend job に追加。ベースライン実数の記録（coverage-policy.md:49 のプレースホルダ解消）も同時に行う。BE 側 R3-5 と同一 PR でまとめて入れるのが効率的。
- **検証**: カバレッジを下げる一時変更で CI が warn/fail。

#### R-F6. PrintPortal 全 inactive の実行時検出（FD6）— 規模 S

- **現状**: 複数 PrintPortal 同居時に全て `active={false}` だと印刷が白紙化する制約が JSDoc 文書のみで、実行時には何も起きない。
- **手順**: 印刷実行時（beforeprint イベントまたは印刷用スタイル注入時）に「document 内に data-print-portal 要素が存在するのに active なものが1つも無い」状態を検出したら **dev ビルド限定で console.warn**（本番挙動は不変＝挙動保存）。テストは #187 の既存 JSDOM 構造検証に1ケース追加。
- **検証**: `npx vitest run src/components/shared`（PrintPortal テスト）。

---

## 3. 非対象（明示的にやらないこと）

| 項目 | 理由 |
|---|---|
| bug.md の挙動修正（H-1 の FE 側 year/month→start/end 切替・ページネーション UI、M-5 カルテ同居ペット表示） | バグ修正トラック。behavior-preserving でないため本計画に混ぜない |
| `DailyAccountingTab.tsx`（636行）等の大ファイル分割 | 全ファイル800行上限内。分割の実益なし（YAGNI） |
| `lib/design-tokens.ts`（971行）の分割 | トークン定義の集約ファイル。凝集が正。BE の transform.go keep 判断と同型 |
| React 19 Action パターンへの全面書き換え | 新規実装の規約であり、動作中の既存フォームを規約準拠のためだけに書き換えるのは回帰リスク純増 |
| 状態管理・ルーティング等のアーキテクチャ変更 | 現行構成（TanStack Query + URL state + feature 構造）で問題が出ていない |
| visual regression / a11y 自動化の新規導入 | 負債返済ではなく能力追加。価値はあるが別トラックで提案すべき投資判断（E2E 基盤は Playwright 47 spec + e2e.yml 導入済みのため、やるなら owner-report 等の印刷系から） |
| `.env.production` / `VITE_SHOW_DEMO_ACCOUNTS` の扱い | シークレット/環境管理トラック（bug.md M-10）。ビルド設定の変更であり behavior-preserving リファクタではない |
| #201 薬量自動計算の FE UI | 未完の feature 開発。リファクタではない |
| liff/ の拡充 | 523行・健全・変更頻度低。smoke テスト1本（R-F4-4）以上は不要 |

---

## 3.5 観測事項（2026-07-02 フルビルドで確認・いずれも既存事象・優先度 LOW）

| 事象 | 内容 | 扱い |
|---|---|---|
| CSS @import 順序警告 | Google Fonts の `@import` が他ルールの後に出力され最適化警告 | フォント読み込みは動作している。触るなら `<link>` preload 化（performance ルール準拠）と同時に |
| INEFFECTIVE_DYNAMIC_IMPORT ×2 | `features/auth/index.ts`（router 等から静的 import 済み）と `MasterSelectModal`（TrimmingLeftColumn から静的 import 済み）への dynamic import が chunk 分割効果なし | 実害なし（重複ロードではない）。該当箇所を触る際に dynamic import を外して整理 |
| manual チャンク 522KB（>500KB 警告） | マニュアル機能（スクリーンショット多数・markdown レンダラ同梱）が単一チャンク | 遅延ロード済みルートであり初期表示に影響なし。分割は実測で問題が出てから |

## 4. 実施ルール

1. **挙動保存の証明を各コミットに含める**: 既存テスト GREEN 維持＋触る箇所に pin テストが無ければ先に書く。**唯一の例外は R-F3 分類(c)** — これは fix: として本計画から分離する。
2. **FE 固有の検証罠（既知）を踏まない**:
   - tsc の PostToolUse hook は偽陽性がある — 型判定の正は `docker compose exec -T frontend pnpm type-check`
   - vitest にパスを渡すとき `--` を挟むと全件実行になる — scoped は `npx vitest run <path>`
   - Radix Select の option 閉鎖は fireEvent では再現しない — `user.click` を使う
   - render 中 setState + useActionState は stale closure — 状態同期は effect で行う
3. **検証は scoped で自走**し、フルの `pnpm lint` / `pnpm test:run` / `pnpm build` / `pnpm type-check`（全体）はプロジェクトルールに従いユーザー手動（完了報告時にコマンド提示）。
4. **コミット粒度は 1 項目 1 コミット**（R-F 番号をメッセージに含める）。commit 前に HEAD 確認・パス限定 stage（並行セッション対策）。
5. **subagent・grep の結果は再検証してから採用**（本計画の策定中にも「feature 横断 import 13件」が grep の誤マッチで、再計測により 0 件と確定した実例がある）。

## 5. 全体見積もりと完了条件

| フェーズ | 項目数 | 規模合計 |
|---|---|---|
| Phase 1（規約統一） | 2 | M+S ≒ 1.5日 |
| Phase 2（hooks 健全化） | 1 | M（分類(c)の fix は別途） ≒ 1日 |
| Phase 3（テスト・ガードレール） | 4 | S+L+S+S ≒ 4日 |

**推奨着手順**: R-F7（type-check 3アプリ化 — 検査の穴を先に塞ぐ）→ Phase 1 → Phase 2 → Phase 3 残り。

**完了条件**:
- **`pnpm type-check` が3アプリ全て（src/・liff/・line-reserve/）を検査**（型エラー注入で検出されることを実証済み）
- `@/lib/axios` の直接使用が `features/*/api/`・`lib/` 以外に存在しない
- 生 hex カラーが 0 件（または根拠コメント付き）
- eslint-disable 全件に根拠コメントがあり、可能なら require-description 相当の lint で機械強制
- line-reserve の重要フロー（予約作成・バリデーション・エラー表示）にテストがあり、カバレッジ計測対象に 3 アプリ全てが含まれる
- FE カバレッジ ratchet が CI で有効（BE 側 R3-5 と同時）
- PrintPortal の全 inactive 状態が dev ビルドで警告される

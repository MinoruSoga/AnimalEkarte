# フロントエンド リファクタリング計画

- **作成日**: 2026-07-07
- **更新日**: 2026-07-10（FINAL-DOC） — R-F3・R-F9〜R-F25の全17項目が完了（CLOSED）した。R-F3は着手前に実装済みであることが判明（下記参照）。本文はタスク定義を完了ログに置き換えて再構成した。R-F1〜R-F8の履歴は git で参照（commit `2acd7854`以前）。
- **エピック完了ステータス**: **全項目CLOSED（未対応0件）**。R-F17のみ「挙動保存フェーズ」完了・「統合フェーズ」は意図的に別`fix:`チケットへ分離（§3参照）。R-F23は方針(b)（react-query不採用、`useFetchState`共通フックへ集約）で実装完了。
- **対象**: `frontend/` の3アプリ全体 — `src/`（メインアプリ・1113 .ts/.tsxファイル・26 feature）＋ `liff/`（11ファイル）＋ `line-reserve/`（30ファイル）
- **スタック**: React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui / TanStack Query
- **性格**: 全項目 **behavior-preserving（挙動保存）** を原則とする負債返済計画である。振る舞いを変える修正が必要な項目（一部のバリデーション追加等）は該当箇所に明記し、別コミット（`fix:`）として分離する。本計画自体はコード変更を行わない設計書であり、実装は本計画をもとに別途着手する。
- **根拠**: 2026-07-07、13軸の並列コード監査（feature indexing / cross-feature import / design tokens / React 19パターン / 型安全性 / 条件レンダリング / ファイルサイズ / ディレクトリ構造・命名規則 / hooks配置 / テストカバレッジ質的分析 / アクセシビリティ / liff・line-reserve規約 / 未使用コード検出基盤+パフォーマンス）を実施し、既知の機械監査済み範囲（design-system-audit.mjs、eslint-disable根拠コメントratchet、frontend/.coverage-baseline、docs/UI_DESIGN_COMPLIANCE.md§2）を除外した上で約150件の個別指摘を確認した。前回のFE-refactor.md（R-F1〜R-F7、PR #218で完了・アーカイブ済み）が対象としたlstep API層・design-tokens・eslint-disable監査・PrintPortal・カバレッジratchet・line-reserveテスト整備・type-check3アプリ化とは別の観点であり、重複しない。

---

## 1. 現状評価（2026-07-07 実測、CLOSED分は2026-07-09/10完了）

### 健全な点（是正不要と判断する根拠）

| 観点 | 実測値 | 評価 |
|---|---|---|
| Feature Indexing（deep import） | `@/features/xxx/(api\|components\|hooks\|routes\|types\|loaders)` 形式の直接import 0件（52件のbarrel経由importを全数確認、静的・動的・相対パス迂回・エイリアス抜け道いずれも無し） | 規約完全準拠 |
| React 19パターン逸脱 | `FC<`/`React.FC`/`forwardRef(`/フォーム手動loading管理（`setIsLoading`等）いずれも0件。36ファイルで`useActionState`を正しく使用 | 規約完全準拠 |
| 条件レンダリング（`&&`）アンチパターン | `{cond && <JSX>}`形式0件。過去の包括是正（commit b7a5a342で68→0件、以降複数の個別是正コミット）が定着し再発なし | 規約完全準拠。ただしESLintでの機械強制（`react/jsx-no-leaked-render`相当）は未導入で手動規律に依存 |
| any型使用 | 明示的`any`（`: any`/`<any>`/`as any`/`any[]`/`Record<string, any>`）はアプリケーションコードに実質0件。唯一の出現は`src/types/generated/models.ts`（tygo自動生成、19箇所、eslint.config.js ignoreで対象外） | ほぼ完全準拠 |
| design-system-audit.mjs対象範囲（`src/features/**/routes/**`・`**/pages/**`のhex直書き・legacy accent・colorVariant） | 2026-07-06監査時点0件、CI zero-tolerance gateで新規混入を検知 | 機械監査運用中 |
| eslint-disable根拠コメント | R-F3（既存PR #218）で33件監査・分類済み。`frontend/scripts/check-eslint-disable-rationale.mjs`のratchetで新規増加のみ検知 | 運用中。本計画では再監査しない |
| UI Design Compliance（84リーフルート） | `docs/UI_DESIGN_COMPLIANCE.md`§2で2026-07-06監査済み・83準拠/1対象外 | 運用中。本計画では再監査しない |
| frontendカバレッジratchet | `frontend/.coverage-baseline`（43.78%、2026-07-05 arm済み）で低下をCI検知 | 運用中 |
| R-F1〜R-F8（FD1・FD3・FD4・FD5・FD6・FD11・FD12行メモ化） | 全項目CLOSED（完了日2026-07-09/10） | 完了。詳細は git 履歴（commit `2acd7854`以前）参照 |

### 残存する負債

| FD# | 負債 | 規模の目安 | リスク |
|---|---|---|---|
| FD2 | ディレクトリ構造・命名規則逸脱 | 約101件（*Model.ts等57件、hooks配置ミス11件、feature構造逸脱3件、その他） | 単発ミスでなく定着した「非公式ローカル規約」化。新規参加者・AIエージェントが誤って模倣するリスク |
| FD7 | ファイル・コンポーネントサイズ超過 | 400-800行帯13ファイル | 複数責務が単一関数/コンポーネントに平坦に同居。プロジェクト自身のCODING_RULES.md基準にも抵触 |
| FD8 | テストカバレッジの質的ギャップ | 11件（CRITICAL1・HIGH多数） | 「テストがある/ない」の粗い比率とリスクの高低が一致しない逆転現象あり。過去に複数回バグ修正された箇所が無防備 |
| FD9 | アクセシビリティ逸脱 | 53件（代表列挙） | 共有コンポーネント経由で多数画面に伝播する構造的パターン。受付ボードという日常業務中核画面にも波及 |
| FD10 | liff/line-reserveアプリ固有の規約逸脱 | 8件 | mainアプリで既に修正済みの障害クラス（BUG-067）が別アプリで再現しうる実害あり |
| FD12 | パフォーマンスパターン欠如（残: useDeferredValue・lazy。行メモ化領域はR-F8で解消済み） | 代表10件中 useDeferredValue3件・lazy3件が未着手 | 主要一覧画面は模範的だが、横展開されていない周辺領域に集中 |

---

## 2. 完了ログ（全17項目 CLOSED）

規模: S=半日以内 / M=1日 / L=2-3日。各R-F項目は独立コミット。commit hashは `git log --oneline` で参照可能（このセクションは要約のみ）。

| R-F | 内容 | 規模 | Status | commit |
|---|---|---|---|---|
| R-F3 | ディレクトリ構造・命名規則の是正（FD2） | L | **CLOSED（着手前に完了済みと判明）** | `4b0118fb`〜`483a6aa2`（R-F3-S1〜S8、本エピック着手前の既存コミット）。settings/lstep/aggregation構造化・PascalCase非コンポーネントファイル0件・available-slot-options.tsxリネームを確認済み。FE-refactor.md旧版が本項目を「未対応」と誤記載していた（実態乖離） |
| R-F9 | useDeferredValue欠如の是正（FD12） | S | CLOSED | `30b0d19a` |
| R-F10 | 非lazy化コンポーネントのlazy化（FD12） | S | CLOSED | `d5c54d1c` |
| R-F11 | PropertyRowのlabel関連付け是正（FD9） | M | CLOSED | `4afa47b6`（実装は当初案のcloneElement+useIdではなく、22箇所全てを機械的に是正できる**implicit `<label>`ラッピング方式**に変更。Radix Select/PropertyInput等がid propを転送しない構造的制約のため） |
| R-F12 | div onClick疑似ボタンのbutton化（FD9） | M | CLOSED | `e8e45545`（7ファイル対応。InterviewHistory/ImageGalleryGroupは内部に別のbuttonが既にネストされているため`<button>`化ではなく`role="button"`+`tabIndex`+`onKeyDown`で対応。WeekViewDayColumnは計画通り対象外） |
| R-F13 | inputのaria-label付与（FD9） | S | CLOSED | `c0a879c5` |
| R-F14 | 【CRITICAL】vaccinations次回接種日計算のテスト追加（FD8） | M | CLOSED | `66149700` |
| R-F15 | 【HIGH】master共有状態機械のテスト追加（FD8） | M | CLOSED | `b9cb5809` |
| R-F16 | 【HIGH】MedicineSettingsModel純粋関数テスト追加（FD8） | S | CLOSED | `8aabaf1f` |
| R-F17 | 【HIGH】pets calcAge重複解消とテスト追加（FD8） | S | **CLOSED（挙動保存フェーズのみ）** | `b403b664`。統合フェーズ（`calcAgeAt`共通化）は挙動変更を伴うため計画通り別`fix:`チケットに分離（未着手） |
| R-F18 | 【MEDIUM/LOW】残りのテストギャップ埋め（FD8） | M | CLOSED（6/6対応） | `6b697559` |
| R-F19 | 400-800行帯13ファイルの分割（FD7） | L | CLOSED（12ファイル、ファイル毎独立commit） | `74eb8818` `f4ed72ad` `017f13a4` `9a6f0076` `aaa76c20` `76694a1f` `d26f9d71` `bc939b67` `da2d412f` `19fe44a6` `91cb9809` `0591bccf` |
| R-F20 | line-reserve axiosへのNULLバイト対策共有化（FD10） | M | CLOSED（`fix:`） | `dca1acfb` |
| R-F21 | use-liff.ts重複解消（FD10） | M | CLOSED（`fix:`） | `9c869c92` |
| R-F22 | エラーハンドリング統一とリトライ導線追加（FD10） | L | CLOSED（`fix:`） | `582beab7`（R-F23と対象範囲が重複するため統合実装。コミットは1つにまとめた） |
| R-F23 | react-query導入 or 共通フェッチフック統一（FD10） | L | **CLOSED（方針(b)で実装）** | `582beab7`。**方針決定: (b) react-query不採用**。`frontend/src/shared-liff/use-fetch-state.ts`に軽量共通フックとして集約し、12箇所の手書きfetchロジックを置換。バンドル増を回避 |
| R-F24 | liffデザイントークン導入（FD10） | M | CLOSED | `b0089a50` |
| R-F25 | その他liff/line-reserve軽微是正（FD10） | S | CLOSED | `c7f6daee`（ネイティブダイアログ→インライン確認UI/バナー）, `91bc6832`（CSS重複解消・zodレスポンス検証） |

### 実装時の主な逸脱・判断ログ（計画書との差異）

- **R-F3**: 計画書は「未対応」としていたが、実際は本エピック着手前に完了済み（履歴汚染）。着手前のgit調査で発覚し、STRUCTレーンの実装を丸ごと省略した。
- **R-F11**: cloneElement+useId方式は構造的に機能しない（Radix Select・PropertyInput等がidを転送しない）ため、22箇所すべてを個別対応せず解決できるimplicit `<label>`ラッピング方式に変更。
- **R-F12**: InterviewHistory.tsx / ImageGalleryGroup.tsxは内部に別のインタラクティブ要素（`<button>`）を含むため、`<button>`化すると無効なHTMLネストになる。`role="button"`+`tabIndex`+`onKeyDown`（`e.target !== e.currentTarget`ガード付き）で対応した。
- **R-F16 / R-F17**: 計画書記載のファイルパス（`MedicineSettingsModel.ts`等のPascalCase）は既にR-F3（S1）でkebab-caseにリネーム済みだったため、実ファイル名（`medicine-settings-model.ts`等）に合わせて対応した。
- **R-F18**: `StaffSettings.tsx`/`PermissionGroupSettings.tsx`のテスト追加中に、`useMasterSave.handleSave`のvalidate失敗時エラーが呼び出し元で握りつぶされ画面に一切表示されないサイレント失敗を発見。現状挙動を回帰テストで固定し、修正は別チケット送りとした（behavior-preserving原則を優先）。
- **R-F19**: 分割時、`staffs.ts`は他8ファイルからの既存deep import経路を壊さないため、分割後3ファイルを末尾でre-exportする設計とした（`STAFFS_QUERY_KEY`経由の循環参照が生じるが、遅延評価のためモジュール評価時には問題なし。vitestで実証済み）。
- **R-F22 / R-F23**: 対象範囲（line-reserve/liffの7+2ページ）が完全に重複するため、2項目を1回の実装で統合し、`useFetchState`フック内にR-F22のステータス別エラーメッセージ・再試行ロジックを組み込んだ。コミットは1つにまとめた（メッセージに両R-F番号を明記）。
- **R-F25**: `window.confirm`は単純廃止せず、`docs/PRODUCT_PHILOSOPHY.md`の「確認ダイアログ禁止・インラインUIで解決」原則に沿ってカード内インライン確認UI（はい/いいえボタン）に置換。確認ステップ自体は維持した。

## 3. 非対象（明示的にやらないこと）

| 項目 | 理由 |
|---|---|
| `dangerouslySetInnerHTML`・PrintPortal生成HTMLのXSS監査 | セキュリティレビュー観点であり本リファクタ計画の範囲外。`security-reviewer`エージェント等で別途実施すべき |
| クリニック切替時のReact Queryキャッシュキーへの`clinic_id`包含有無 | マルチテナント境界のフロントエンド防御に関わる論点だが、本監査では未検証（**要調査**）。切替直後に前クリニックのデータが残留表示されるリスクの有無は別途調査が必要 |
| FE zodスキーマとBackend Goバリデーションの二重管理・乖離 | `docs/PRODUCT_PHILOSOPHY.md`が二重管理を戒めているが、解消には設計判断（どちらを正本にするか、共有スキーマ生成の要否）を要するため本behavior-preserving計画には含めない。別途architect判断が必要な設計課題として切り出すべき |
| `src/types/generated/models.ts`（3370行）の分割 | tygo自動生成ファイル。手動分割は次回codegenで上書きされ差分が消えるため有害。行数上限の対象外として明示的に除外する |
| `src/lib/design-tokens.ts`（1029行）の分割 | 7つの独立定数オブジェクト（PALETTE/C/BADGE/ICON/LAYOUT/STYLE/TABLE_STYLES）の集約であり、ロジックではなく純粋な定数カタログ。分割は機械的に可能だが凝集の価値が高く、明確な実害（差分の追いにくさ等）が出るまでは対象外とする |
| R-F23（react-query導入 or 共通フェッチフック統一）の方針決定そのもの | 本計画は実装手順を示すが、react-query導入の可否（バンドルサイズトレードオフ）はPO/architect判断が必要であり、決定を待ってから着手する |
| use-postal-code-lookup.ts等の機能追加・挙動改善 | 配置修正のみを行い、ロジック自体の改善（例: エラー時のリトライ）は別トラック |
| eslint-disable根拠コメント・design-system-audit機械監査済み範囲・UI Design Compliance既監査ルート | 既に運用中のガードレールがあり、本計画では再監査しない（「1. 現状評価」の健全な点を参照） |
| R-F17（pets calcAge統合フェーズ）の実装 | 挙動保存フェーズ（境界値テスト追加）のみ本計画に含み、統合自体は挙動変更を伴うため別`fix:`チケットとする。§2誤記修正: 旧版は誤って「R-F7 vaccinations calcAge」と記載していたが、正しくは本項目（R-F17 pets calcAge） |

---

## 4. 実施ルール

1. **挙動保存の証明を各コミットに含める**: 既存テストGREEN維持＋触る箇所にpinテストが無ければ先に書く。**唯一の例外はR-F20/R-F21/R-F22**（liff/line-reserveのバグ修正的性格を持つ項目）— これらは`fix:`として扱い、挙動変更を明記する。
2. **FE固有の検証罠（既知）を踏まない**:
   - tscのPostToolUse hookは偽陽性がある — 型判定の正は`docker compose exec frontend pnpm run type-check`
   - vitestにパスを渡すとき`pnpm test:run -- <path>`は罠 — `--`以降が全件実行になる。scoped検証は必ず`docker compose exec frontend npx vitest run <path>`を使う
   - Radix Selectのoption閉鎖はfireEventでは再現しない — `user.click`を使う
   - render中setState + useActionStateはstale closureの原因になりうる — 状態同期はeffectで行う
3. **検証はscopedで自走**し、フルの`pnpm lint`/`pnpm test:run`/`pnpm build`/`pnpm type-check`（全体）はプロジェクトルールに従いユーザー手動（完了報告時にコマンド提示）。
4. **コミット粒度は1項目1コミット**（R-F番号をメッセージに含める）。commit前にHEAD確認・パス限定stage（並行セッション対策）。
5. **subagent・grepの結果は再検証してから採用する**。本計画の策定時にも、13軸監査結果のうち複数件（knip未運用・PropertyRow label未関連付け・CODING_RULES.md自己矛盾・AppointmentCardキーボード操作不能）を実コード読み合わせで裏付け確認した上で採用している。実装時も同様に、着手前に該当ファイルを実際にReadしてから修正すること。
6. **R-F3は独立して着手可能**: cross-feature import解消（R-F2）は完了済みのため、R-F3（ディレクトリ構造是正）はファイルパス依存の懸念なく単独で着手できる。
7. **memo化の参照安定性は3段階チェック**（コールバックの`useCallback`化／Setなど参照が変わりやすいコレクションをdepsに直接入れない／カスタムhookの戻り値オブジェクト自体が毎レンダー新規生成されていないか）で確認する（R-F8の教訓）。
8. **feature間で共有フック/型を昇格する際は、他feature配下の関連テストファイルのimportも`rg`で横断確認する**（R-F2-S18の教訓。production側のimport付け替えだけではテストが旧経路の参照に取り残されるケースを機械的に検知できない）。

---

## 5. エピック完了サマリー

| 旧Phase | 項目 | 項目数 | Status |
|---|---|---|---|
| Phase 1（ディレクトリ構造） | R-F3（1項目） | 1 | CLOSED（着手前に完了済みと判明） |
| Phase 3（パフォーマンス） | R-F9・R-F10（2項目） | 2 | CLOSED |
| Phase 4（アクセシビリティ） | R-F11〜R-F13（3項目） | 3 | CLOSED |
| Phase 5（テストカバレッジ） | R-F14〜R-F18（5項目） | 5 | CLOSED（R-F17は挙動保存フェーズのみ、統合は別チケット） |
| Phase 6（ファイルサイズ） | R-F19（12ファイル、独立コミット） | 1 | CLOSED |
| Phase 7（liff/line-reserve） | R-F20〜R-F25（6項目） | 6 | CLOSED（R-F22/R-F23は対象範囲重複のため統合実装・単一コミット） |

**完了条件（達成済み）**:
- `*Model.ts`等のPascalCase非コンポーネントファイル命名違反が0件、settings/lstep/aggregationが標準構成に揃っている（R-F3）
- PropertyRow経由の22ファイルでlabel-input関連付けが機能している（R-F11、implicit `<label>`ラッピング方式）
- 受付ボード（AppointmentCard）を含む主要な疑似ボタンがキーボード操作可能（R-F12）
- vaccinations次回接種日計算・master共有CRUD状態機械・薬剤マスタ純粋関数にユニットテストが存在する（R-F14〜R-F16）
- line-reserveの予約作成フローがNULLバイト対策済みaxiosインスタンスを使用し、API失敗時に再試行導線を持つ（R-F20・R-F22）
- liff/line-reserveのuse-liff.tsが単一実装（`frontend/src/shared-liff/use-liff.ts`）に統合されている（R-F21）
- liff/line-reserveの手書きfetchロジック12箇所が`useFetchState`共通フックに集約されている（R-F23、方針(b)）

**未着手（意図的に別チケット分離、本エピックの範囲外）**:
- R-F17統合フェーズ（`calcAgeAt`共通化、挙動変更を伴うため別`fix:`）
- R-F18で発見したmaster設定ページのvalidationErrorサイレント握りつぶし（別チケット推奨、本エピックでは現状挙動を回帰テストで固定するに留めた）

**運用上の教訓（次回エピックへの申し送り）**:
- 本エピック着手前のgit調査で、計画書の「未対応」記載（R-F3）が実態と18項目分乖離していることが判明した。計画書は実装完了後に必ず陳腐化しうるため、着手前に対象ファイルの現況をgitで確認するステップを標準化すべき。
- 複数subagentが同一Dockerコンテナに対して並行でvitestを実行すると、ワーカー起動タイムアウト（"Failed to start forks worker"）や個別テストの5秒ハードタイムアウトが多発する。原因はコード側の回帰ではなくコンテナのCPU/プロセス枯渇であり、疑わしい失敗は必ず単体ファイルでの再実行によりinfra起因かどうか切り分けること。`--pool=threads`への切り替えも有効な回避策として確認された。

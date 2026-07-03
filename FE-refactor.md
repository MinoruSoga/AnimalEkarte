# フロントエンド リファクタリング計画

- **作成日**: 2026-07-02
- **最終更新**: 2026-07-04
- **対象**: `frontend/` の3アプリ全体（`src/`・`liff/`・`line-reserve/`）
- **状態**: 実装完了（PR #218 レビュー中・`main` 未マージ）

## 結論

本計画で扱っていた FE-refactor 項目（R-F1〜R-F7）は PR #218（ブランチ `feature/fe-refactor-r-f1-r-f7`）で実装完了。全項目 **behavior-preserving**（UI/API仕様変更なし）。

完了済み項目の詳細な実施手順・現状評価・見積もりは、重複した履歴情報としてこの文書から削除した。実装内容の正本は現在のコード・テスト・PR #218 のコミットログとする。

主な完了コミット:

- `c7d1d627` `fix(fe): R-F7 type-check を liff/line-reserve 含む3アプリに拡大`
- `47eaae9a` `fix(fe): R-F2 生hexカラーをdesign-tokensへ置換`
- `0a34dd35` `fix(fe): R-F1 lstep 系データ取得を api/ 層へ移設`
- `6edf5d7f` `fix(fe): R-F3 eslint-disable 33件を監査・分類`
- `f2c88db3` `fix(fe): R-F6 PrintPortal 全inactive状態をdevビルドで検出`
- `9cc7da80` `feat(fe): R-F5 FEカバレッジ ratchet ゲート導入`
- `5bbec7b4` `test(fe): R-F4 line-reserve/liff の重要フローテスト整備`
- `3fcbcaf4` `fix(fe): 独立レビュー指摘2件を是正（R-F5 CIステップ順序 / R-F1 clinicId解決タイミング）`

最終検証（PR #218記載・Docker経由・全て実測）:

- `pnpm run type-check`: 0 errors（3アプリ対象）
- `pnpm lint`: 0 errors、14 warnings（すべて本PR対象外ファイルの既存警告）
- `pnpm build`: 成功（既知警告のみ。3.5節参照）
- 開発中 scoped vitest: 581 + 18 + 135 tests 全PASS

## 現在残すフォローアップ

以下は R-F1〜R-F7 の未対応ではなく、PR #218 が完了後もなお残るとして明示的に切り出した項目。

### FE カバレッジ ratchet の実測値 arm（P1）

`frontend/.coverage-baseline` はプレースホルダ `0` のまま。R-F5 で導入した ratchet は現状 warn-only で、fail ゲートとして機能していない。

次回 CI run の `frontend-coverage` artifact から `coverage-summary.json` の `total.statements.pct` を取得し、`.coverage-baseline` に転記する別コミットが必要（backend 側と同じ2段階導入方式）。

### CI step順序ガードレール（ハーネス改善・P1）

過去に複数回、ratchet/test 系ステップが Lint/Build より前に配置され、fail-fast により後続の Lint/Build がスキップされたまま green に見えるバグを作り込んでいる（[[feedback_ci_step_order_masks_lint]]。直近では PR #218 でも同型が発生し `3fcbcaf4` で手動是正）。

再発防止の自動チェック（`ci.yml` の job 内 step 順序を検証する lint/スクリプト）はまだ存在しない。都度レビューで見つける運用のまま残っている。導入する場合は、既存の `P3.1 Preload Clinic-Scope Lint` 等と同様の go/node 製 CI ゲートとして追加するのが本プロジェクトの慣行に合う。

### eslint-disable 根拠コメントの機械強制（任意・P2）

R-F3 で既存33件を監査・分類し、新規追加分は根拠コメント付きにした。ただし今後新規に追加される eslint-disable に根拠コメントを義務付ける仕組み（`eslint-comments/require-description` 相当）はまだ導入していない。

未導入の理由は新規 pnpm 依存追加が必要なため。導入する場合はユーザーによる `pnpm add -D` 実行が前提（`docker compose exec frontend pnpm install` は自動実行禁止コマンド）。

### 既知の非対応事項（対応不要・再指摘防止の参考）

- `src/hooks/` 配下の `@/lib/axios` 直接使用（`use-pet.ts` 等）は `src/hooks/CLAUDE.md` で定義済みの意図的な cross-feature 共有パターン。R-F1 完了後もこれを規約逸脱として再指摘しないこと。
- `tsconfig.node.json`（`vite.config.ts` 専用 project reference）を単独 type-check すると無関係な既存エラーが出る（vite の型が vitest の `test` 拡張を認識しない、本PR以前から存在）。`pnpm run type-check` は `tsc --noEmit` 実行のため対象外であり実害なし。R-F7 の対象外。

## 観測事項（LOW優先度・未解消のまま）

R-F1〜R-F7 とは無関係な既存事象で、いずれも実害なし。触る際の参考として残す。

| 事象 | 内容 | 扱い |
|---|---|---|
| CSS @import 順序警告 | Google Fonts の `@import` が他ルールの後に出力され最適化警告 | フォント読み込みは動作している。触るなら `<link>` preload 化と同時に |
| INEFFECTIVE_DYNAMIC_IMPORT ×2 | `features/auth/index.ts` と `MasterSelectModal` への dynamic import が静的 import 済みのため chunk 分割効果なし | 実害なし。該当箇所を触る際に dynamic import を外して整理 |
| manual チャンク 522KB（>500KB 警告） | マニュアル機能が単一チャンク | 遅延ロード済みルートであり初期表示に影響なし。分割は実測で問題が出てから |

## 今後の扱い

この文書は完了済み計画のアーカイブとして残す。新しい FE リファクタ項目が発生した場合は、この文書へ完了済み項目を再追加せず、別のタスク文書または issue として起票する。

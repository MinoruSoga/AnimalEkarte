# FE-refactor.md — フロントエンド リファクタリング計画書

- **作成日**: 2026-07-13 / **更新日**: 2026-07-15
- **第6期**: **完了**（FE6-0〜FE6-18 全19項目）。詳細は git 履歴を参照。
- **本書の役割**: 次期監査への引き継ぎのみ。新規の第6期作業はない。
- **別台帳**: プロジェクト横断 TODO = `todo.md` / 受付テレメトリ完了ポインタ = `change-ui.md`

---

## 次期監査への引き継ぎ

- **テレメトリ不在（MEDIUM）**: 本番のレンダー例外・API失敗を運用チームが観測する手段がない（Sentry等未導入）。導入はコスト・依存追加を伴うためPO判断。
- **`RECEPTION_TELEMETRY_PHASE2_ENABLED` の扱い**: 恒久 true のフラグ（`use-reception-telemetry.ts`）。キルスイッチとして残すか、フラグと false 分岐テストを削除するか要判断。
- **OwnerSearchModal の React Query 化**: FE6-1 はバグ修正に留めた。`useState`+`useTransition` の素朴 fetch を feature 側 `useSearchOwners` フックに置き換える構造改善は次期。
- **`ShiftFormDialog` の `use-shift-form.ts` 抽出 / `TreatmentRow` の EditableCell 化 / `ChangePasswordDialog` の api 層整理**: いずれも実害なしの一貫性改善。
- **liff / line-reserve の `index.html` に CSP メタタグがない**（メインアプリのみ設定済み）。セキュリティ観点の追加検討。
- **`src/lib/` と `src/utils/` の役割分担が不文律**（両方にフォーマット系が分散）。規約明文化候補。
- **z-index の中間スケール**（sticky/dropdown 用）が未整理。`Z.overlay` 以外は Tailwind 標準スケールのまま（FE5-4 の意図的スコープ限定）。
- **export されているが外部参照のない型シンボル約15件**（`CPMStageOption` 等）: 次期にまとめて掃除。
- **`.filename-baseline`（値23）** の ratchet を 0 に向けて下げる余地。
- **Pet属性ラベルの単一ソース化**: FE6-8 は二重定義＋ガードテストでの乖離検知に留めた。単一ソース化を次期に検討。
- **曜日ラベル契約の統合**: `line-reserve` の月曜始まり契約と `DAY_OF_WEEK_LABELS`（0=日始まり）は契約が異なる。統合は契約設計が必要。

---

## 第6期で確定した「やらない」判断（次期でも踏襲推奨）

- `use-*-form` 系フックの共通スケルトン抽象化は、ドメインロジックが実質的に異なり害と判定済み。
- `src/features/owners/components/pet-edit-field-shared.tsx` のリネーム・`.ts` 化は不可（JSX 定数を含む）。
- `src/components/ui/`（shadcn 生成物）・`src/types/generated/`（tygo 生成物）は編集しない。
- `types/index.ts` の FA9 構造自体の変更はしない（FE6-18 でドキュメント明文化のみ実施済み）。

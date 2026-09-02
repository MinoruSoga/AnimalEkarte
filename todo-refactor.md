# リファクタ台帳（FE コード規約準拠）

更新日: 2026-09-02（FE 規約ゲート完了）

| 項目 | 値 |
|------|-----|
| **範囲** | `frontend/src` · `frontend/liff/src` · `frontend/line-reserve/src` の **production TS/TSX** |
| **除外** | `*.test.*` · `src/types/generated/**` · `src/components/ui/**` · `frontend/e2e/**` |
| **目的** | 挙動を変えずに `frontend/CLAUDE.md` と ESLint 機械ガードへ寄せる |
| **ブランチ** | `main`。クレームは末尾。エージェントは削除しない |
| **検証** | Docker scoped `npx vitest run <path>`。full `pnpm lint` / `test:run` / `type-check` / `build` はユーザー手動 |
| **正本** | ①実装 → ② `frontend/CLAUDE.md` → ③ `frontend/CODING_RULES.md` |

BE フェーズは `ad63bdf28` で完了。BE 全文はそのコミットの本ファイル。

スキャン: 関数行数は `function` / `memo(function` の開き `{` から対応 `}`。2026-09-02 終了時、生産の 150 行超関数は **0**。800 行超ファイルは `design-tokens.ts` のみ。

---

## 対象外（やらなかった）

台帳初版の対象外表を維持する。要点:

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割なし
- `utils/` 再作成なし、generated/models 262 件の一括移行なし（TASK-444）
- 50 行までの機械分割なし、200–399 行ファイルの「薄くするためだけ」の切断なし
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）
- トリミングフォームの権限・死亡ガード追加は **既存欠落**（今回の抽出で新たに無くしていない）。別 TASK
- full プロジェクト lint/test/build は未実行

---

## 完了したカテゴリ

### FE-ARCH

- **FE-ARCH-001**: CheckupsTab は `@/features/checkups` barrel。deep import 0
- **FE-ARCH-002**: `app/pages` 合成と owners `loaders.ts` 例外はそのまま
- **FE-ARCH-003**: `CODING_RULES.md` を 28 features + `lib/` に合わせた
- **FE-ARCH-004**: `hooks/CLAUDE.md` の queryKey 例を `queryKeys.pets.detail`
- **FE-ARCH-005**: generated/models allowlist は増やしていない（移行は別トラック）

### FE-QUERY

- **FE-QUERY-001**: `queryKeys.accounting.unpaidAll()`。`CreditCorrectionDialog` の配列リテラルを除去

### FE-REACT

- **FE-REACT-001**: `ReservationFormModal` は `useActionState` の pending。手動 `setIsSubmitting` なし

### FE-CONFIRM

- **FE-CONFIRM-001**: 生産の `window.confirm` 0。`useSidePeekDirty` は `runWithDiscardCheck` + `discardDialog`。`confirmDiscard` は mock 互換（dirty 時は false）

### FE-STYLE

- 印刷面と予約エラー帯を既存 `C.*` へ。hex 直書きはトークンファイル以外 0

### FE-SIZE

- 150 行超の生産関数 **0**（brace 対応、`memo(function` 含む）
- 800 行超ファイルは `design-tokens.ts` のみ
- フォームフック・巨大コンポーネント・一覧/設定ページを sibling helper に抽出。権限 ref / 死亡 sentinel / `useActionState` は各本体に残した

### FE-DOC

- CODING_RULES と hooks/CLAUDE.md のドリフトを直した

---

## 検証（Docker / scoped）

親 tree の `docker compose exec -T frontend npx vitest run`:

- 第1束（barrel / unpaid / modal / side-peek / form hooks / TreatmentRow / ExamType / App）: **16 files / 400 passed**
- 第2束（VaccinationList / MedicalRecord* / Hospitalization* / InventoryList / LoginForm / CashRegisterHistory / TreatmentPlanMaster / StaffSettings / side-peek）: **11 files / 85 passed**

隔離 worktree 側でも各エージェントが vitest を実行済み（compose が main を mount するため worktree は `docker run -v`）。

full `pnpm lint` / `test:run` / `type-check` / `build` は禁止コマンドのため未実行。

---

## 完了条件チェック

- [x] CheckupsTab から `@/features/checkups/...` の deep import が 0
- [x] 生産の `queryKey: [...]` リテラルが `query-keys.ts` 以外 0（preview の factory spread は残置）
- [x] ReservationFormModal に手動 `setIsSubmitting` が無い
- [x] 生産の `window.confirm` が 0
- [x] 印刷面の palette クラスがトークン化
- [x] 150 行超関数が、意図的残置以外 0
- [x] 800 行超ファイルが `design-tokens.ts` 以外 0
- [x] queryKey タプル文字列を変えていない
- [x] 死亡 sentinel と permission ref を分割で落としていない（トリミングの元からの欠落は未改修）
- [x] generated/models の allowlist を増やしていない
- [x] `utils/` と `__tests__/` を再作成していない
- [x] scoped vitest のみ

---

## クレーム（ユーザー解放）

エージェントは `git branch -D` しない。main に載ったあと、人間が削除する。

現存（この作業で残っているもの）:

- `claim/TODO-REFACTOR-FE`
- `claim/FE-CONFIRM-001`
- `claim/FE-SIZE-HELPER-EXTRACT`
- `claim/FE-SIZE-FORM-PANELS`
- `claim/FE-SIZE-LIST-PAGES`
- `claim/FE-150-PAGE-EXTRACT`

他セッションの claim（未操作）: `claim/CODEX-W4-CSV-FRONTEND-DEPLOY`、`claim/codex-w3-billing-audit`、`claim/csf_*`

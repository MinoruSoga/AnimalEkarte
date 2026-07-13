# BE-refactor.md — バックエンド リファクタリング計画書（第6期）

- **作成日**: 2026-07-13
- **基準コミット**: `e13a2987`(main)。行番号はずれたら**シンボル名で再特定**する。
- **性格**: 本書は実行計画の正本。判断できない事態は**中断して報告**。
- **別台帳**(本書と重複させない): PERF・SEED残 = `BE_todo.md` / 任意検証 = `BE-pending.md`
- **第6期進捗**: A-1〜F-5・C-4 は完了済み（コミット済み）。**残件は F-6 のみ**。

---

## 1. 実行者が守る規約（要約）

- 各層 `CLAUDE.md`（`backend/internal/{handler,service,repository}/CLAUDE.md`）を遵守。
- 機械強制ゲート（preload / dbortx / audit_tx / master_fk / routes snapshot / apicontract / coverage-ratchet）を無断で赤くしない。
- 検証は Docker スコープ限定。フル `go test ./...` / `golangci-lint run ./...` / `gofmt -w ./...` 禁止。
- main 直作業。**push しない**。`Co-Authored-By` を入れない。メッセージ形式は `docs(backend):`（本残件）。
- `git add` はファイル指定のみ（`git add -A` / `git add .` 禁止）。

### 1.3 dirty ファイル（触らない）

別ワークストリームの未コミットファイルは**変更・コミット・stash 禁止**（代表: liff 系・`api.yaml`・`FE-refactor.md` 等）。

**例外**: F-6 の `BE_todo.md` への**純追加**のみ。追記前に `git diff BE_todo.md` で既存 dirty と分離できることを確認し、分離不能なら中断して報告。

---

## 2. 残作業項目

### F-6. 恒久 skip された既知バグテストの台帳化（コード変更なし）

- **対象**: `t.Skip("known production bug — …")`（件数は実行時 `rg` 実測を正とする。計画当初の「17 件」は古い可能性あり — 2026-07-13 時点の再測は 12 件）。
- **問題**: skip されたテストはバグが直っても・悪化しても fail しない。台帳化されていない。
- **変更**: 全件を列挙し、`BE_todo.md` 末尾に「### 既知バグの skip テスト台帳(2026-07-13 棚卸し)」として `ファイル:行 / テスト名 / skip メッセージ要約` の表を**純追加**する。**テストコードは変更しない**。
- **完了条件**: 表が追記され、件数 = grep 実測。コミットは `docs(backend):`。
- **リスク**: 低。**依存**: 0（`BE_todo.md` が分離可能なときのみ着手）。

代表パス（実測で消えているものもある — 必ず再 grep）:

- `hospitalization_repository_test.go` / `medicine_repository_test.go` / `trimming_repository_test.go`
- `pet_chronic_condition_repository_test.go` / `hospitalization_plan_repository_test.go` / `payment_method_master_repository_test.go`

```bash
rg -n 't\.Skip\("known production bug' backend/internal --glob '*.go'
git diff --stat BE_todo.md
git diff BE_todo.md | rg '^@@'   # 単一 hunk の全面 rewrite なら分離不能 → 中断
```

---

## 3. やらないこと（残件 F-6 向け）

1. 機能追加・仕様変更・API / `api.yaml` / `go.mod` 変更禁止。
2. migration / seed / DB スキーマに触らない。
3. `BE-pending.md` の項目に着手しない。
4. §1.3 dirty ファイルに触らない（F-6 の `BE_todo.md` 純追加を除く）。
5. 他セッションの `BE_todo.md` rewrite を代行コミットしない（ユーザー明示指示がある場合のみ例外）。
6. skip 解除・テスト修正・バグ修正はしない（台帳化のみ）。
7. push / PR 禁止。フル検証コマンド禁止。

---

## 4. 次期監査への引き継ぎ（実行者は無視してよい）

- C-4 で検出: `LineCustomer.owner_name` / `ShiftEntry`（`staff_name` 含む）の api.yaml 未文書化・スキーマ乖離。
- その他未検証領域: service interface オーファンメソッド再走査、slog レベル誤用、helper 経由ページネーション、`backend/worker/`。

---

## 5. 実行者への指示（F-6）

```
あなたは AnimalEkarte のバックエンド実行者です。BE-refactor.md の残件は F-6 のみ。

1. backend/CLAUDE.md と必要なら各層 CLAUDE.md を読む。
2. git diff BE_todo.md で分離可否を確認。entangled（全面 rewrite）なら変更せず中断報告。
3. rg で known production bug skip を列挙し、BE_todo.md 末尾に表を純追加。
4. テストコードは触らない。git add BE_todo.md のみ。docs(backend): でコミット。
5. push しない。完了後に件数と hash を報告する。
```

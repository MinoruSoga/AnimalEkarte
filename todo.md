# タスク台帳 — Linear が正本

統合日: 2026-09-08。最終GitHub照合: 2026-09-11 / `origin/main` とローカル `main` はともに `f0e238f10`。`main` → `staging` PR #388 は OPEN / UNSTABLE（Backend check FAILURE を含む）。2026-09-11 の Linear 読み取り結果と検索限界は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) を正本とする。UAT・STG/PROD・go-live は再判定していない。

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **本ファイルの範囲** | repo と強く結び付く **開発タスク**、確認済み製品 FAIL、PO 入口、維持制約 |

状態・Done は Linear を正本とする。行値・秘密は書かない。対応済み項目と完了証跡は本ファイルから除き、Git 履歴を参照する。

入口: [開発タスク](#development-tasks) · [製品 FAIL](#product-bugs) · [PO / 人間レーン](#human-lane) · [FE 維持制約](#refactor-constraints)

横断・性能・認証の検証順は [todo-verification.md](todo-verification.md) を参照する。開発・検証・運用の実行先を混在させない。

性能の技術的な状態は [todo-performance.md](todo-performance.md)、測定・受入は [統合検証TODO](todo-verification.md#perf-stg-login) を参照する。

運用・外部環境に関する停止条件は [todo-operations.md](todo-operations.md) を参照する。push / dispatch / Linear Done / 秘密変更は明示承認が必要。

claim は ID ごとに初回編集前に確認・取得する。作成者別の削除条件は [AGENTS.md](AGENTS.md#branch-deletion-by-creator-mandatory) を正本とする。ユーザー作成は AI による削除禁止。AI 作成は統合・明示終了・成果を保全した引き継ぎと未使用を確認して削除可能。過去セッションの claim 記録は historical snapshot として読み、現在の保有状態は新規着手時に `git branch --list 'claim/<TASK-ID>'` で再確認する。本 META 追加照合に再着手する場合は `claim/META-LINEAR-APPLY` を確認・取得する。claim の削除は UAT や受入の完了を意味しない。

---

<a id="development-tasks"></a>

## 開発タスク

**2026-09-11 / `f0e238f10` で下記3件をREADYと判断した。** 各行は、対象・実装方針・最初の変更が確定し、STG証跡・新機能の発生・codegenを待たず着手できる開発単位である。実装完了やリリース可能を意味しない。実装時はclaimと既存WIPを再確認し、`main`上では1件ずつ進める。

依頼者: 曽我 稔。技術判断: Codex。旧10項目すべての扱いと根拠は [裁定記録](docs/work/development-task-decisions.md) に保存した。検証タスク・実行コマンドは [統合検証TODOの開発検証](todo-verification.md#development-verification)、運用作業は [todo-operations.md](todo-operations.md) で管理する。Linearのライブ状態は今回 `USER_NOT_LOGGED_IN` のためUNKNOWN。

| 順 | ID | 状態 | 着手する開発単位 |
|---|---|---|---|
| 1 | **TASK-444** | READY | ペット送信型のキーを明示し、モデル追加列が送信可能型へ自動混入しないようにする |
| 2 | **BE-RC-009** | READY | LIFF空き枠取得の2つの依存を `FindAll` だけの読取interfaceにする |
| 3 | **BE-RC-017** | READY | 飼主repositoryの公開更新mapを型付きcommandに置き換える |

### TASK-444 — ペット送信型を許可キーで固定

- **対象**: `frontend/src/types/pet.ts`、`frontend/src/lib/transforms/pet.ts` と隣接テスト。`generated/models.ts`、Go DTO、codegen設定は変更対象外。
- **最初の変更**: `PetWritable = Omit<BackendPet, ServerFields>` を、現行フォームが送るキーを列挙した `Pick<BackendPet, ...>` に置き換える。入力型へ `version` / `deceased_at` / `deceased_reason` を取り込まない。未対応の送信キーを新たに増やさない。
- **確定方針**: 作成の必須3キー、`name_kana`、既存enum、更新時のstatus除外、`danger_reason`の省略/null/値を維持する。既存の生成型を参照し、全モデルの移動・新規生成を前提にしない。
- **完了時の状態**: モデルへサーバー専用列が増えても入力型は広がらず、既存の作成・更新payloadと死亡専用APIの役割が維持される。公開response型の全面移行完了とは扱わない。

### BE-RC-009 — LIFF空き枠の依存を読取専用に限定

- **対象**: `backend/internal/reservation/liff_service.go`、`liff_service_mock_test.go`、`liff_service_test.go`。空き枠ロジックとrepository実装は変更対象外。
- **最初の変更**: `liffUnavailableTimeReader` と `liffAvailableSlotReader` を利用側へ定義する。各interfaceは既存と同じ `FindAll(ctx, clinicID, reservationTypeID)` 1メソッドと既存の戻り型だけを持つ。
- **確定方針**: `liffService` の2フィールドと `NewLiffServiceWithType` の対応引数を上記interfaceに変更する。既存repositoryは構造的に代入でき、管理画面のCreate/Delete能力をLIFFの依存へ要求しなくなる。nil時の既存動作を維持する。
- **完了時の状態**: 読取1メソッドだけの実装でLIFFを構成でき、医院scope、空き枠判定、DB query、予約書込の動作が変わらない。

### BE-RC-017 — 飼主更新を型付きcommandへ限定

- **対象**: `backend/internal/owner/repository.go`、`service_core.go`、`service_delivery.go`、`service_line.go`、型付きcommand用の同package新規ファイル、関連ownerテスト。
- **最初の変更**: 現在の更新builderが扱うフィールドだけを具体的な型で持つ `UpdateCommand` を定義する。`UpdateAndFind` のmap引数と `OwnerUpdateApplier` のmap戻り値をcommandへ変更し、DB用mapの組立を非公開処理へ閉じ込める。
- **確定方針**: プロフィール、配信除外・注意、転院、LINE確認の既存呼出元を同時に移行する。割引権限の再判定は `FOR UPDATE` 後のcallback内に残す。列名や値を任意指定する公開factoryは作らない。
- **完了時の状態**: 公開更新APIへ任意mapを渡せず、許可された更新のみを表現できる。clinic条件、nil/false/空文字の意味、割引TOCTOU防御、更新と再読込の同一transaction、失敗時rollbackを維持する。

---

<a id="product-bugs"></a>

## 4. 確認済み製品 FAIL（旧 bug.md）

記録対象は確認済み製品 FAIL のみ（[TEST_ARCHITECTURE.md](docs/ops/testing/TEST_ARCHITECTURE.md) §6）。環境・seed・権限・fixture 不足による BLOCKED / PARTIAL や受入未実施を混ぜない。証跡に credential・token・cookie・idToken・個人情報（PHI）を含めない。起票後は Linear で追跡し、見出し ID は本節内で重複させない。新規項目は `### BUG-XXX` で本節に追加する。

| ID | status | area | severity | scenario | 層 |
|:---|:---|:---|:---|:---|:---|
| （現在の確認済み未対応項目なし） | — | — | — | — | — |

これは現在の製品全体に不具合がないという判定ではない。認証の現在状態は [todo-fix-auth.md](todo-fix-auth.md)、検証は [統合検証TODO](todo-verification.md#認証認可の外部境界) を参照する。対応済み項目は本節に残さず、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照する。

<a id="human-lane"></a>

## 5. PO / 人間レーン（旧 todo-po.md）

実行 SoT は Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) · Project ノア動物病院電子カルテ。人間ゲートの検証は統合検証TODOを正本とし、別の Open 行台帳を再構築しない。

会社側索引: CorpVault `50_Projects/ノア動物病院電子カルテ/05_Linearマップ.md`。旧詳細本文は Git 履歴。

<a id="refactor-constraints"></a>

## 6. FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割、50 行までの機械分割、200–399 行ファイルの薄型化だけを目的とした切断は行わない。
- `utils/` を再作成しない。generated/models の一括移行は行わず、必要性が出た場合だけ [開発タスクの裁定記録](docs/work/development-task-decisions.md#task-444) の境界で分割追従する。
- `app/pages` の合成と owners `loaders.ts` の例外を維持する。
- 権限 ref、死亡 sentinel、`useActionState`、queryKey タプルの契約を維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）は維持する。
- 当時のトリミングフォームの権限・死亡ガード欠落は対象外だった。本履歴から現在の未修正・修正済みを判断しない。

---

## 参照

| 文書 | 役割 |
|------|------|
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 対応案。Linear は UNKNOWN |
| [todo-verification.md](todo-verification.md) | 横断・性能・認証の検証TODO |
| [todo-operations.md](todo-operations.md) | STG・本番・納品などの運用・外部実行TODO |
| [製品 FAIL](#product-bugs) | 確認済み製品 FAIL |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | ローカル handoff |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 破壊境界 |

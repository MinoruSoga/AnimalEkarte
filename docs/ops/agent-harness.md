# Agent harness operations

AnimalEkarte 固有の安全境界と完成条件を定義する。汎用ワークフローはユーザースコープの `agent-task-lifecycle` Skill に置き、本書は未導入の環境でも使える作業契約を残す。

## 設定と参照の責務

| スコープ | 維持する内容 |
|---|---|
| ユーザー | モデル、sandbox、承認方式、汎用Hook、外部接続、汎用Skill、機能の実機確認 |
| プロジェクト | clinic/owner/pet/staff隔離、migration禁止、Docker検証、claim、domain Skillと受入仕様 |
| タスク | 受入条件、担当パス、基準diff、検証対象、承認済みの副作用 |

プロジェクトのMCP server一覧は空に保ち、ユーザー設定の接続を利用する。Chrome DevToolsは固定バージョン、isolatedプロファイル、usage statistics/performance CrUX無効を既定とする。既存ブラウザのloopback `http://127.0.0.1:9222` へのattachは対象を明示したユーザー承認がある場合だけ行う。

設定は実効値を確認する。ユーザー設定を優先する設計でも、runtimeがプロジェクト設定を上書き適用する場合がある。権限の重複定義を避け、設定変更後は新規セッションで確認する。

[AGENTS.md](../../AGENTS.md) → [.claude/CLAUDE.md](../../.claude/CLAUDE.md) → 最も近い `CLAUDE.md` と対象Skillの順で必要部分を読む。仕様は [docs索引](../README.md)、API正本は [api.yaml](../../backend/docs/api.yaml)。同名Skillはプロジェクト特有の適応を含むものを優先し、汎用Skillをコピーして増やさない。生成物の正本は `.claude`、生成先は `.agents` / `.codex`。

CLI・ホストアプリ・APIセッションの機能は別々に確認する。2026-09-08 のローカル診断は Codex CLI 0.153.4 の `multi_agent` / Hooks を stable と報告したが、別runtimeでの利用可能性を保証しない。変更時は `codex --version` と機能一覧・現セッションのtool一覧を再確認する。Experimental / Beta / Deprecated / under development は、その実機表示を結果に明記する。機能の固定一覧を各指示ファイルへ複製しない。

## タスクを閉じる手順

1. 要求と受入条件、現在のdiff、担当パスを確認する。Linearが利用できなければ現在状態はUNKNOWNとし、必要な更新はローカル下書きにする。投稿には明示承認が必要。
2. ledger/packet ID がある場合は [claim規約](../../AGENTS.md#packet-claim-protocol-mandatory)を守る。並列編集は別worktree。調査・レビューはread-onlyで並列化できる。
3. 受入条件の失敗ケースを先に確認し、範囲内の実装と関連テストを行う。docs-onlyや低影響の可逆編集に不要な製品テストを増やさない。
4. 変更した成果物を検証する。独立レビューで見つかった範囲内の不具合を修正し、影響する検証を再実行する。上限・環境障害を成功に変換せず、未完了項目として報告する。
5. 完了報告には動作変更、対象worktree/diff、検証コマンド・終了結果・対象件数、未実行理由、外部状態、残作業を書く。claim解除とmain統合はユーザー操作のまま維持する。

## 変更範囲の検証

事前計画は `python3 -B scripts/verify-agent-task.py --base <BASE_REF> --plan`。未コミットも含む対象diffを確認し、明示パスなら `--paths <PATH>...`、stage対象なら `--staged` を使う。

実行時は既にローカルにある固定imageを `--frontend-image <ID>` / `--backend-image <ID>`、依存volumeを `--frontend-dependency-volume <NAME>` / `--backend-dependency-volume <NAME>` で指定する。Git Hookからも使う場合は、無視対象 `.claude/verification.local.json` に `frontend_image` / `backend_image` / `frontend_dependency_volume` / `backend_dependency_volume` の4キーだけを保存する。優先順位はCLI、`AGENT_VERIFY_*` 環境変数、ローカル設定の順。image/volumeはマシン固有なのでコミットしない。

runnerはimageをimmutable IDへ解決し、networkなし・source/依存volume読み取り専用・capabilityなしの一時コンテナで実行する。image pull・依存インストール・Compose起動・migrationは行わない。Frontendのmountpoint用に空の `frontend/node_modules` ディレクトリだけを必要時に作る。Go実行用一時領域とFrontend native config loaderにより、sourceや依存volumeへの書込みを避ける。既存container指定も可能だが同じ隔離条件と対象worktree mountが必要で、通常のComposeは適合しない。

`--evidence <PATH>` には新しい証跡ファイルを指定する。stage検証はサービス内にunstaged/untracked依存がある場合停止する。終了コードだけでなく、実行したテスト件数とPASS / FAIL / SKIP / BLOCKEDを読む。対応不能な変更はBLOCKED、docs-onlyのSKIPはruntime PASSではない。

- Go/FrontendはDocker。ホストnpm/goは使わない。既存Composeがmainをマウントしていればcandidateの検証に転用しない。
- image、mount、対象diff、コマンド、終了コード、テスト件数を証跡に結ぶ。ツールやテストがない場合は成功キャッシュを作らない。
- `--passWithNoTests` で検証済みとしない。修正後や対象diff変更後は古い証跡を再利用しない。
- DB依存テストは承認済みのdisposableな環境で実行し、共有DBをresetしない。環境を作る権限がなければ、その検証をBLOCKEDとして独立作業を続ける。
- [scoped-verification-gates](../../.claude/skills/scoped-verification-gates/SKILL.md)と[coverage policy](coverage-policy.md)を使用する。限定テスト結果から全体coverageやrelease readinessを推定しない。

`make up` は既存サービスを停止し、backend entrypointがmigrationを適用する。安全な検証準備のコマンドとして自動実行しない。`make migrate` はユーザー手動。起動・停止・依存導入・全体検証の禁止コマンドは `.claude/CLAUDE.md` を参照する。

## ブラウザ受入

UI動作、フォーム、ルーティング、アクセシビリティに変更があれば、関連する [受入シナリオ](testing/scenarios/README.md)を実操作で確認する。専用テストプロファイル、対象URL、合成データと許可された操作範囲を確認する。認証済み個人ブラウザ・実患者データを流用しない。Playwrightは再現テスト、DevToolsは調査、Computer Useは必要な操作に使う。現在呼べるtoolを確認する。

ローカル操作成功はSTG/UAT/PRODの証明ではない。スクリーンショットやネットワーク証跡は患者・飼主・資格情報を除外する。未実施をN/AやPASSへ置き換えない。

## 失敗を仕組みに戻す

再発する失敗は最も近い一つの正本に戻す。業務仕様はdocs、操作手順はSkill、機械判定できる禁止はHook、再現可能な不具合はTest。個人memoryはユーザーの明示依頼がある場合だけ更新する。

命令・Skillを変更したら `python3 -B .claude/scripts/test_instruction_safety_contracts.py`、関連するharness regression tests、生成一致と再生成の冪等性を確認する。欠落Hook、誤mount、テスト不在、秘密出力、他者WIP、禁止migration、外部接続不明を回帰ケースにする。製品runtimeの成功とは分けて報告する。

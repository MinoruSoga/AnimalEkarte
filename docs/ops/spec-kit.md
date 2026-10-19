# Spec Kit の利用手順

2026-10-13、既存ローカルmainにCodex integrationを導入。CLIは公式安定版1.1.3、GitHubタグv1.1.3の取得SHAは0282d03eb6fc3091b3deaa3a855d4030dc40cd6c。uvの隔離ツール環境へインストールし、実行ファイルは ~/.local/bin/specify。生成物は空の隔離ディレクトリで確認した後、衝突のない新規ファイルのみ持ち込んだ。既存.codex/config.tomlやAGENTS.md、既存skillsを上書きしていない。

## 公式とローカルの役割

- 公式core: .specify/scripts/bash/ の6スクリプト、.specify/templates/ のcoreテンプレート、speckit workflow、Codex向け10スキルの生成形式。MIT表示は [.specify/LICENSE.spec-kit](../../.specify/LICENSE.spec-kit)。出典・版は [.specify/UPSTREAM.md](../../.specify/UPSTREAM.md)。
- AnimalEkarte固有: 公式preset機構で導入した animalekarte-canonical 0.1.0。specify/clarify/constitutionの3コマンドはローカルreplace、他7コマンドはローカル前置き＋公式core。4テンプレートは変更参照・計画・tasks・品質確認向けに置換する。公式提供presetではない。
- .specify/memory/constitution.mdは既存AGENTS・臨床安全・意思決定・文書規約への参照口。別の規範正本を作らない。
- 業務仕様の正本は引き続き [docs/spec/](../spec/README.md)。spec.mdは変更依頼と正本FR-ID/GWTへの参照索引、plan/tasksは変更単位の計画。正本の本文を複製しない。

## 開始する

1. `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte` を作業ディレクトリとしてCodexを開く。既に別ディレクトリで開始したセッションのスキル一覧へ自動反映されるとは限らない。新しいセッションで読み込まれたスキルを確認する。
2. [文書規約](../spec/authoring-guide.md)を読み、対象のdomain正本と限定した変更依頼を明示する。
3. Codexチャットで `$speckit-specify` を呼ぶ。`SPECIFY_FEATURE_DIRECTORY=docs/work/spec-kit/<意味を表す変更名>` をプロンプトに明示する。ここは変更参照・計画の場所で、業務正本の新しいコピーではない。新規work資料は [work/README.md](../work/README.md)へ索引登録する。
4. `$speckit-clarify` で未決を確認する。未知の臨床・会計・入力・権限・優先度を推測しない。回答と決定根拠はdomain正本へ反映し、feature側には参照を置く。
5. 計画が許可された範囲で `$speckit-plan` → `$speckit-tasks`。その後 `$speckit-analyze`（read-only）と `$speckit-checklist`（記述品質）で対応を確認する。

これらはシェルコマンドではなくCodexスキルの呼出名。CLIの確認は `specify version`、`specify check`、`specify integration status`、`specify preset list`。テンプレート解決は `specify preset resolve spec-template` で確認できる。

## 既存 scoped Plane チケットとの使い分け（2026-10-19）

スコープ・受入条件・正本への参照が揃った既存の scoped Plane チケットは、spec.md の役割（変更依頼・範囲・正本への参照索引）を代替する。このため `$speckit-specify` を経由せず実装へ進めてよい（AGENTS.md と `.claude/CLAUDE.md` の Spec Kit routing 免除条項「existing scoped Plane tickets」がこれに当たる）。チケットの本文・コメントに業務ルールを複製せず、正本への参照で書く。

spec-kit のスキルを使うのは次の場面である。

- チケット化前の整理: 仕様に無い新要件や業務判断を伴う変更の変更依頼作成と `$speckit-clarify`
- 設計残件の解消: 仕様が「設計残件」と明示した設計判断（例: 予約コース×オプションの対応データ、採用時間の保存場所）
- 複数領域に跨る変更で、リポジトリ内に変更参照を残す価値がある場合

実装の実行状態・Done の正本は Plane のみ（docs/README.md の fact owner 規定どおり）。spec.md を作った場合も実行台帳にはしない。

## 導入状態・注意

Git拡張・任意の第三者拡張・新規認証・権限設定変更は導入していない。branch・commit・push・force初期化は行っていない。実予約のspec/plan/tasksは作成していない。導入検証は隔離した架空fixtureで行う。

ローカルpresetの適用により、integration statusは「Modified managed files: 10」とWARNINGを報告する。生成した10スキルが公式coreから意図的に変更されたことを示す。差分はpresetの3replace／7prependで確認でき、Missing／Invalidは0。検査を弱めたりmanifestハッシュを書き換えて警告を隠していない。アップグレード時も `--force` で消さず、既存変更とpreset適用内容をレビューする。

## 公式参照

- [安定版v1.1.3](https://github.com/github/spec-kit/releases/tag/v1.1.3)
- [既存プロジェクト導入](https://github.github.io/spec-kit/guides/existing-projects.html)
- [Codex integration](https://github.github.io/spec-kit/reference/integrations.html)
- [公式カスタマイズ](https://github.github.io/spec-kit/guides/customization.html)

公式既存プロジェクト手順のforce/commit/stash/branch例は、今回のユーザー指定に合わせて使用せず、隔離生成・衝突検査・新規ファイルのみの持込みに置き換えた。

## 技術設計の新しい配置方針（設定未追従）

予約の技術仕様は[docs/design/domains/reservations/](../design/domains/reservations/README.md)へ、具体的作業・依存関係・進捗はPlaneへ置く。tasks.md・docs/planを作らない。上記docs/work/spec-kitの開始例は導入時の設定を保持したもので、新しい配置方針へ未追従。preset原本・生成skills・出力先は今回変更していない。root allowlistのdesign追加だけは、後続の個別承認（Sentinel_34b62bb163b881919e6f2a9fd05edfc3）で対応済み。coreのspec.md／plan.md固定参照と今回のtechnical-spec.mdとの対応を、将来の許可された運用変更で整える必要がある。業務spec本文や第二技術正本を複製せず、全workflowによる実装開始を今回の文書作成許可へ含めない。

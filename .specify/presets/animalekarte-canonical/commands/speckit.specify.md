---
description: Create a bounded change reference without duplicating canonical domain requirements.
---

## AnimalEkarte プロジェクト適用（ローカルpreset）

これは公式Spec Kitの拡張点を使うプロジェクト固有の指示。業務仕様の正本は docs/spec/ のドメイン・lifecycle・workflow、文書規約は docs/spec/authoring-guide.md。最初に AGENTS.md と関連する既存規約を読む。既存の臨床安全・医院分離・Git安全・Docker限定検証規則を保持する。

spec.md は変更の依頼・範囲・正本への参照索引であり、業務ルールの第二正本にしない。plan/tasks は変更単位の計画で、正本FR-ID・GWTへリンクする。spec.mdのリンク先を読むことが前提。要件IDは安定したドメイン接頭辞を使い、決定状態と実装・実機検証状態を分ける。未知の臨床・会計・権限・入力・優先度を推測で補完しない。未決事項の件数上限を設けず、質問だけ小分けにする。

featureディレクトリは明示する。通常は SPECIFY_FEATURE_DIRECTORY=docs/work/spec-kit/<意味を表す変更名> とする。未指定なら対象正本と変更範囲を確認してから確定し、既存ファイルを上書きしない。今回の導入では実予約のfeatureは作らない。current mainを保持し、Git拡張・自動branch・commit・push・stash・reset・force initを使わない。

スキルの呼出だけで機能実装・外部送信・デプロイ・フルCIを許可されたと解釈しない。品質チェック済みは要件の記述品質を示し、実機PASSやPlane Doneを意味しない。constitution.md は既存正本への参照口であり、新しい独立した規範を作らない。元の未決事項を解消できないときは保持して、依存する実装・確定を止める。

## 実行手順

1. $ARGUMENTS から変更の目的・要求者・既存正本・範囲を確認する。既存仕様を読むだけなら文書を新設しない。
2. 正本への変更はユーザーが明示した合意に限る。未知の前提は「未記載／要確認」または「未決」のまま残す。既存ルールをコードの現状から改定しない。
3. SPECIFY_FEATURE_DIRECTORY を明示し、既存ファイルとの衝突を検査する。新規spec.mdは `specify preset resolve spec-template` の解決結果を使う。公式 `.specify/feature.json` のfeature_directoryにその相対パスを記録する。branchは変更しない。
4. spec.mdには依頼・変更範囲・業務正本の文書ID/FR-ID/アンカーと関連シナリオ/GWTの参照を書き、ルール本文を複製しない。新規要求が必要な場合は対応するdomain正本へ提案・未決として置き、その場所をspec.mdから指す。合意済みへの格上げは決定者・日付・根拠が必要。
5. 成功条件や優先度は既存根拠またはユーザーの合意を使う。未決の件数を減らすための推測は禁止。確認質問は小分けにし、回答を正本へ反映する。
6. 明確性・範囲・整合性・検証可能性・受入条件対応をレビューし、不足と未実行を明示する。品質完了と実機PASSを区別して報告する。

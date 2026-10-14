---
description: Clarify canonical domain requirements without guessing unresolved clinical or accounting rules.
---

## AnimalEkarte プロジェクト適用（ローカルpreset）

これは公式Spec Kitの拡張点を使うプロジェクト固有の指示。業務仕様の正本は docs/spec/ のドメイン・lifecycle・workflow、文書規約は docs/spec/authoring-guide.md。最初に AGENTS.md と関連する既存規約を読む。既存の臨床安全・医院分離・Git安全・Docker限定検証規則を保持する。

spec.md は変更の依頼・範囲・正本への参照索引であり、業務ルールの第二正本にしない。plan/tasks は変更単位の計画で、正本FR-ID・GWTへリンクする。spec.mdのリンク先を読むことが前提。要件IDは安定したドメイン接頭辞を使い、決定状態と実装・実機検証状態を分ける。未知の臨床・会計・権限・入力・優先度を推測で補完しない。未決事項の件数上限を設けず、質問だけ小分けにする。

featureディレクトリは明示する。通常は SPECIFY_FEATURE_DIRECTORY=docs/work/spec-kit/<意味を表す変更名> とする。未指定なら対象正本と変更範囲を確認してから確定し、既存ファイルを上書きしない。今回の導入では実予約のfeatureは作らない。current mainを保持し、Git拡張・自動branch・commit・push・stash・reset・force initを使わない。

スキルの呼出だけで機能実装・外部送信・デプロイ・フルCIを許可されたと解釈しない。品質チェック済みは要件の記述品質を示し、実機PASSやPlane Doneを意味しない。constitution.md は既存正本への参照口であり、新しい独立した規範を作らない。元の未決事項を解消できないときは保持して、依存する実装・確定を止める。

## 実行手順

1. `.specify/scripts/bash/check-prerequisites.sh --json --paths-only` を対象project rootで実行する。SPECIFY_FEATURE_DIRECTORY が未指定かつfeature.jsonもない場合は対象を確認し、別のfeatureを推定しない。
2. spec.mdと、その参照するdomain正本・関連フロー・GWTを読み、業務結果を左右する不足・不整合を列挙する。
3. 質問は小分けにして、理由と選択肢を示す。未回答の臨床・会計事項を推測で埋めない。質問件数で残りの不足を消さない。
4. 明示回答だけを該当正本に記録し、決定者・日付・根拠と決定状態を保持する。spec.mdには正本の回答箇所を参照する。回答が既存ルールを否定する場合は影響範囲を確認するまで関連実装を進めない。
5. 根拠が足りなければ未決を残す。変更しない read-only 評価の依頼なら正本は書き換えず、質問候補と不足を報告する。完了は確認範囲と未決残を添える。

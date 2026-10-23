# 見積（D11）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.estimates.index／入口 |
| 決定状態 | 既存記載を継承。E1の目標は2026-10-15依頼者承認（業務正本参照）。既存詳細すべての再承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-23（移行証拠FR/ACの入口追補。従来確認2026-10-15） |

## 読む順と正本

1. [業務契約・FR／受入条件・例外・未決](business-rules.md)
2. [語彙と他領域境界](../../ontology/finance-operations.md)
3. [一覧](../../screens/22-estimate-list.md)・[作成・編集](../../screens/23-estimate-form.md)・[詳細](../../screens/26-estimate-detail.md)
4. [現行実装差・技術契約](../../../design/domains/finance-operations/technical-spec.md)

業務契約と現行画面の説明は別軸。既存画面本文は削除せず、目標との相違を明示して参照する。未承認の提案を合意済みへ格上げしない。仕様整理、実装完了、実機受入を別に判定する。

## 追加承認：価格と提示原本

2026-10-15追加承認：承認済み・期限内で内容と数量が同じなら提示単価を使う。追加処置や内容・数量変更は再説明し、提示済み見積は返答待ちでも原本を残して後継改訂・再提示する。税・保険・割引は既存適用条件を維持し、総額の無条件固定や自動請求確定に拡張しない。 [根拠と受入条件](business-rules.md#estimate-price-presentation-agreement-20261015)。FR5件を維持し、AC16件へ8件を追加した計24件。実装・実機確認は別工程。

## 2026-10-23 移行原本証拠の追加範囲

[旧見積の保全FR/1AC](business-rules.md#requirement-fr-estimates-legacy-evidence-preservation)と[会計側6共通AC](../accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)へ接続する。上記の既存5FR/24AC確認とは別の追加で、原値保全・認可付きread-only参照を定義する。closed schema/DDL/APIは[技術設計候補](../../../design/domains/accounting/legacy-finance-evidence.md)の未選定で、実装/試験NOT_RUN。

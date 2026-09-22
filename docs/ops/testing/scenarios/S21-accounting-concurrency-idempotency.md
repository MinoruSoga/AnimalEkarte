# S21: 会計確定の並行更新 — 冪等リプレイ・409・確定後保護

> **目的**: 会計確定が二重送信・リトライ・別キー競合で重複作成されず、確定済み会計の明細追加・古い合計での更新・検査由来の重複が正しく拒否/保護されることを実 DB・2 セッションで納品前に証明する。
> **所要目安**: 20分 / **深度**: 深い
> **仕様正本**: [screens/11-accounting-detail.md §3.3](../../../spec/screens/11-accounting-detail.md)。検証キュー: `UAT-R2-EXCLUSIVE-LOCK`（[todo-verification.md](../../../../todo-verification.md)）・DB 制約ゲート `billing-schema-readiness`（[todo-operations.md](../../../../todo-operations.md)）。

## 前提条件

- ローカルの使い捨て clinic、または承認済みの専用 UAT tenant。**DB 制約の適用確認（`billing-schema-readiness`）が完了した環境**で実施する。migration 自動適用禁止。
- 2 セッション（2 ブラウザまたは 2 端末）で同じ会計/カルテを開ける actor 2 名分、または同一 actor の 2 セッション。
- 会計確定 API の `CompletionRequestID`/`CompletionRequestHash` を操作できる手段（devtools での同一リクエスト再送・別 payload 送信）。
- 検査由来の明細を含むカルテ fixture。
- 依存シナリオ: S08（訂正・未収残）・S20（マスタ→会計）。会計の初回確定フローは V02 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 会計を確定する（同一 `CompletionRequestID`・同一 payload） | 1 件の billing header が作成され、確定レスポンスが返る |
| 2 | **同じ request ID・同じ payload** を再送する（リトライ/二重クリック相当） | 冪等リプレイとして既存レコードが返り、重複作成されない（`CompletionRequestID`+`CompletionRequestHash` の digest 一致で replay と判定） |
| 3 | **同じ request ID・異なる payload** を送る | 409 またはエラーで拒否され、無関係な UNIQUE 競合と混同されない。既存レコードは壊れない |
| 4 | **別の request ID** で同じカルテの会計を確定しようとする | 「このカルテには既に会計があります」系の 409 で拒否され、同一カルテに二重会計が作られない（`accounting_complete_tx.go` の UNIQUE 競合分岐） |
| 5 | 確定済み会計へ明細を追加しようとする | 確定後の明細追加が拒否されるか、訂正フローへ誘導される。確定済みの金額が変わらない（[11 §2.1](../../../spec/screens/11-accounting-detail.md)） |
| 6 | 2 セッションで同じ会計を開き、片方で確定→もう片方で古い合計のまま更新/確定する | 古い合計・古い版での更新は拒否または再計算される。楽観ロックまたはサーバー再計算で不整合を残さない |
| 7 | 検査由来の項目が会計明細へ重複して入らないことを確認する | 検査連携の項目は 1 回だけ請求され、再取込・再保存で重複行が増えない |

## 確認観点

- 冪等の判定は `backend/internal/billing/accounting_complete_tx.go` が `CompletionRequestID` + `CompletionRequestHash` で replay（digest 一致）と真の UNIQUE 競合（別 request ID や別内容）を区別する。replay は既存を返し、非 replay は「このカルテには既に会計があります」等で拒否。
- 「2 セッション・実 DB」での確認が本シナリオの核心。mock/合成回帰（古い合計・確定後明細・検査重複）は追加済みだが、実並行は `UAT-R2-EXCLUSIVE-LOCK` で UNKNOWN のまま。ここを実測で証明する。
- 確定後の所見以外の stale 更新設計は別 Issue で継続中。本シナリオで判明した未設計の stale 更新は既存 Issue へ接続し、新規 duplicate issue を作らない。
- 会計の真正性（確定後に金額が変わらない・監査が残る）は [11 §2.1/§2.2](../../../spec/screens/11-accounting-detail.md) の要求。変更が起きた場合は訂正フロー経由かを確認する。
- 失敗時のロールバック・入力救済・1 回分の請求/監査が `UAT-R2-EXCLUSIVE-LOCK` の残条件。部分的な成功（ヘッダだけ残る等）があれば FAIL。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 確定 API の途中でネットワークを切り、クライアント側が再送する | サーバー側で確定が完結していればリプレイで既存を返し、未完結なら再送で 1 回だけ確定。部分作成が残らない |
| A2 | 別 actor が先に確定した会計へ、自分が古い画面から確定を送る | 409 で拒否され、先に確定した会計が上書きされない。画面は最新状態へ再読込を促す |
| A3 | 確定済み会計の明細を API 直叩きで追加/変更する | 権限・状態チェックで拒否。UI を迂回しても確定済みは変わらない |
| A4 | 冪等キーだけ同じで明細金額を変えたリクエストを送る | digest 不一致としてリプレイ扱いされず、409 で拒否される。既存レコードが改ざんされない |

## 実装突合

- 変更サマリ:
  - `accounting_complete_tx.go` の `CompletionRequestID`/`CompletionRequestHash` による replay/非 replay 分岐と「このカルテには既に会計があります」409 を現行コードと突合
  - `UAT-R2-EXCLUSIVE-LOCK` の「異なる key の同一カルテ会計・明細重複・古い合計・確定後明細を実 DB/2 セッションで確認」を手順 3–7 に対応づけ
  - `billing-schema-readiness` の DB 制約適用を前提条件のゲートに明記

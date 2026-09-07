# docs 全件・GitHub Issue照合（2026-09-07）

文書の最新化依頼に対する確認範囲、訂正理由、残る外部依存を記録する。製品タスクの状態・担当・DoneはLinearが正本。今回の受入状態は **awaiting-human-verification**。

## 入力と確認範囲

- ローカル: HEAD `267a17e48ed8f0cc1944d25ad11900cdbd433226` と開始時の未コミット差分。GitHub mainも取得時点では同じSHA。既存WIPを今回の実装修復として数えない。
- GitHub: 全144 Issue（OPEN 15）、Issue/PRコメント957件を2026-09-07に読み取り取得。各文書に関連するIssue本文・採択コメントを現在コードと比較した。過去のタイトル/本文より後の訂正を確認し、相互に矛盾する記録は断定しない。
- 文書: 開始時235ファイル = 199 Markdown + GOAL.yaml + 35証跡/テスト/補助ファイル。MarkdownとGOAL.yamlは全件の役割・主張・根拠・Issue・不確実性を確認。35件は再生成せず同一バイト列で保存する。
- 新規本文は本レポート1件。全件の確認記録は [照合カタログ](EVIDENCE/issue-refresh-catalog.json) にまとめる。`keep`は確認した主張に訂正が不要という判定で、全命題・全環境の無欠陥認定ではない。
- 役割別読者: architecture/product/index 24、spec 80、ops 73、work 6、delivery/governance 17。新規の本書1を含む確認記録は201件。

## 訂正内容

| 文書 | 訂正と直接の根拠 |
|---|---|
| [DELIVERY_PACKAGE](../../delivery/DELIVERY_PACKAGE.md)・[OPERATION_MANUAL](../../delivery/OPERATION_MANUAL.md) | 在庫減算のため薬品/物販マスタ画面で紐付ける案内を訂正。薬品APIには `inventory_id` があるが設定UIは未実装。物販モデルにも在庫紐付けがない。[薬品仕様](../../spec/screens/settings/master-medicine.md)・[物販仕様](../../spec/screens/settings/master-merchandise.md)、`MedicineFormData`・`MedicineSidePanelBody`・`MerchandiseItem`を照合 |
| [STAFF_ACCOUNT_PROVISIONING](../../ops/deploy/STAFF_ACCOUNT_PROVISIONING.md)・[GOLIVE_RUNBOOK](../../delivery/GOLIVE_RUNBOOK.md) | 「受領記録なし/先方提供待ち」の断定を修正。[#255本文訂正](https://github.com/MinoruSoga/AnimalEkarte/issues/255#issuecomment-5141215224)に受領履歴がある一方、[最新整理](https://github.com/MinoruSoga/AnimalEkarte/issues/255#issuecomment-5352950935)は現在版・方針・認可applyを残件とする。名簿や個人情報は転記しない |
| [ADR-006](../../architecture/adr/006-backend-domain-package-boundaries.md) | 2026-09-06の35 package測定を履歴として保持し、9月7日の36 top-level / 14 domainを追補。`package_boundary_gate_test.go`の集合と件数を根拠とし、既存の見出しリンクを維持 |
| [README](README.md)・[INVENTORY](INVENTORY.md)・[GOAL](GOAL.yaml)・[LEDGER](LEDGER.md)・[PROGRESS](PROGRESS.md) | RQ-002未修復の古い現在説明を訂正。開始時点のmanual修復差分、先行completed記録、今回の人間受入待ちを分ける。過去の検証ログや採択理由を新しい結果に置き換えない |

既存のカルテ確定/会計作成案内も再確認した。カルテ更新は会計を自動作成せず、退院時は `CreateAccounting` を選んだ場合のみ会計作成へ進む（`medical_record_crud.go`・`hospitalization_discharge_tx.go`）。このマニュアル修復差分は開始時点から存在し、今回のコード変更ではない。

## 仕様と外部残件

| Issue | 文書に残す境界 / 次の担当 |
|---|---|
| [#249](https://github.com/MinoruSoga/AnimalEkarte/issues/249)・[#261](https://github.com/MinoruSoga/AnimalEkarte/issues/261) | 検査・薬量・健診の実装と臨床/PO承認は別。仕様・臨床承認の既存入口を保持し、PO/臨床担当が受入を確定 |
| [#250](https://github.com/MinoruSoga/AnimalEkarte/issues/250)・[#252](https://github.com/MinoruSoga/AnimalEkarte/issues/252) | 移行リハーサル/当日投入と全院締め値の運用証跡を、実装・seed既定値だけで代用しない。認可済み運用担当のreceipt待ち |
| [#253](https://github.com/MinoruSoga/AnimalEkarte/issues/253)・[#254](https://github.com/MinoruSoga/AnimalEkarte/issues/254) | 本番CI/backup/restore/rollback、デモ確認・現場UAT・実LINE・個別sign-offは実行時証跡で判定。過去のgreenを現在のrelease-readyにしない |
| [#255](https://github.com/MinoruSoga/AnimalEkarte/issues/255) | 過去の受領履歴と、現在版名簿・email/院/役割方針・認可applyを分離。PO/USERが現在版と適用証跡を確定 |
| [#256](https://github.com/MinoruSoga/AnimalEkarte/issues/256)・[#257](https://github.com/MinoruSoga/AnimalEkarte/issues/257) | U13研修は未完の記録。納品後研修とtodo上の切替前提の順序差は既存HOLDを維持しUSERが確定。Issueタイトルの旧日付を新しい切替windowへ流用しない |
| [#258](https://github.com/MinoruSoga/AnimalEkarte/issues/258) | U1–U12の契約・窓口・本番実測は外部入力。U13はマニュアル側の別受入。文書同期だけでは入力欄やIssueを完了にしない |
| [#259](https://github.com/MinoruSoga/AnimalEkarte/issues/259) | cron/Write APIの現在コードと、先方API有効化・承認後の再開を区別。納品後対応のgateを保持 |
| [#235](https://github.com/MinoruSoga/AnimalEkarte/issues/235)・[#284](https://github.com/MinoruSoga/AnimalEkarte/issues/284) | 添付フロー/LIFF実機フォントのIssue受入を、静的な画面仕様照合だけで完了にしない。既存QA担当の実機/実行確認へ |
| [#89](https://github.com/MinoruSoga/AnimalEkarte/issues/89)・[#97](https://github.com/MinoruSoga/AnimalEkarte/issues/97) | 取得時点でOPEN。credential失効/rotationの完了は本調査では検証・実行しない。担当者が既存セキュリティ手順で確認。秘密値や生コメントを文書へ複製しない |

Linearは利用可能な取得経路で現在値を取得できず **UNKNOWN**。GitHubのlinkbackや古いReady/Needs Humanを現在値として転記しない。UAT原証跡、STG/PROD、外部契約や資格情報の状態も今回の静的照合では判定しない。ファイル別の残件と次の担当は照合カタログに残す。

## 検証と受入

`ecc:loop-design-check` の反復需要ゲートでは週次実績を確認できなかったため、常設ループ・cronを設けず一回の手動処理とした。独立担当が編集前に受入仕様/検証器を固定し、本文修復と独立judgeを分離。同一失敗は最大3回までとし、基準を下げて合格にしない。

確認する項目は、全件の確認記録とIssue引用、直接根拠の行一致、全差分の開始時hashとの対応、35補助ファイルとdocs外WIPの保持、ローカルリンク、`git diff --check -- docs`、`bash scripts/check-docs-symbol-drift.sh`。独立レビューは全新規差分と6カテゴリの根拠を確認する。機械検査の成功は本文の全命題やruntime成功を保証しない。

docs-only変更のためruntimeテストは不要。承認を要する外部書込み・本番操作・migration・commit/push/mergeは実行しない。人間の受入は **awaiting-human-verification** のままとする。claim `claim/DOCS-REFRESH-ISSUE-RECONCILIATION` はUSERが統合または明示放棄後に解放する。

実行した静的コマンドの出力は [検証ログ](EVIDENCE/issue-refresh-checks.txt)。最終レビューでは、このログの存在だけで合格にせず、固定した受入検証器と独立した本文照合を使用する。

### 終了前の再取得（2026-09-07 15:02 JST）

共有checkoutのHEADは `4285b06d758746ef63a9ce44c7b84ba3bcc9cab3`。開始時WIPのbackend修正と3件のarchitecture文書は `df936c68f` で別セッションによりコミットされ、その後に依存更新等が加わった。照合対象docs235件のバイト列は開始時から変わっていない。Issue144件のstate/updatedAtも再取得時点で差分はない。

開始時から追加されたdocs外の内容差分は `.github/dependabot.yml`・`frontend/package.json`・`frontend/pnpm-lock.yaml`。依存はZod・Axios・Playwrightの更新を含む。今回の照合対象文書に旧版番号を固定した記載はなく、依存更新のruntime互換性は本調査の検証対象に含めない。これらは別セッションの変更として保存し、隔離候補の受入証拠は開始時snapshotに固定する。反映直前にはHEADだけで同一性を推定せず、対象docsの開始時hashとdocs外差分を改めて照合する。

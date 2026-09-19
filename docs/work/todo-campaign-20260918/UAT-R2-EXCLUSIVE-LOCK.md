# UAT-R2-EXCLUSIVE-LOCK: 同時操作の判断票

状態: **競合ケース設計・不足箇所の調査 READY／実行未実施**。2026-09-19依頼者回答の「両方」により、目的はカルテの上書きと同じ会計の二重確定の防止に確定した。旧システムの画面占有を複製する要件ではない（[要件・状態の正本](../../../todo-issue.md#uat-r2-exclusive-lock)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)）。

## 既存の防御と限界

| 合成データで確認する場面 | 現在の防御 | 防御の範囲 |
| --- | --- | --- |
| 2 セッションが同じカルテ所見・診断を読み、先に片方が保存した後、古い内容をもう片方が保存する | [保存処理](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts)は読み込んだ `existingClinicalPlanVersion` を保存要求へ渡す。[clinical_plan 更新](../../../backend/internal/medicalrecord/clinical_plan_repository.go)は `expectedVersion` を更新条件に含め、更新行がない場合に競合を返す。 | 古い保存の拒否。別端末が同じ画面を開くことは防がない。 |
| 2 セッションが同じ請求の明細を追加・変更する | [明細サービス](../../../backend/internal/billing/billing_item_service_create.go)は請求親を `LockAndFindByID` する。[会計リポジトリ](../../../backend/internal/billing/accounting_repository.go)の同メソッドは ambient transaction 内で請求行に `FOR UPDATE` 相当のロックを取る。[明細リポジトリ](../../../backend/internal/billing/billing_item_repository.go)の作成時参照検証も取引内で請求親行をロックする。 | この明細操作の書き込みを取引内で直列化する。すべての会計更新に同じロック経路があるという意味ではなく、医院の二重会計事故が解消するかは未検証。画面の占有もしない。 |
| 同じタブでカルテを編集し、保存せず画面を離れる | [ReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx)が dirty 状態を管理して [NavigationBlocker](../../../frontend/src/components/shared/NavigationBlocker/NavigationBlocker.tsx)へ渡す。[未保存変更フック](../../../frontend/src/hooks/use-unsaved-changes.ts)はタブ終了・ブラウザ遷移時の `beforeunload` 警告を登録する。 | 同じタブでの未保存離脱の警告。別端末・別タブの入室制御ではない。 |

いずれも**別端末からの画面表示を禁止する占有ロックではない**。既存の競合拒否、取引内の行ロック、離脱警告は目的と適用場面が異なる。旧システムの全面ロックをそのまま複製する判断はしない（[設計思想](../../product-philosophy.md)）。

## 実装担当が検証する事故と期待値

2つの独立したログインセッションA/B、同一医院の合成患者/カルテ/請求を用意する。各行を既存テストへ対応づけ、不足ケースをREDにしてから修正する。検証前に対象revision・専用環境/fixture・後処理を固定する。下表は未実行で、既存防御の存在だけでPASSにしない。

| ケース | 操作 | 受入条件 |
| --- | --- | --- |
| 古いカルテ保存 | A/Bが同版を読む→A保存→Bの古い版を保存。同時送信も確認 | 古い内容で上書きしない。競合表示、再読込/再入力手順があり、保存失敗を成功と見せない。Bの未保存入力を無警告で消さない |
| カルテ内の更新範囲 | 所見/診断だけでなく問診、治療、検査/接種等の更新経路を一覧化。同一行の編集/削除と確定が競合 | 行/資源ごとに守る境界を特定し、確定済みデータの上書き・消失を拒否。clinical_planのversionだけで全タブ保護済みとしない |
| 同一キーの会計再送 | 確定ボタン連打、応答喪失後に同一payload/keyを再送 | 同じ処理結果へ戻り、請求/支払/業務監査が増えない。異なるpayloadで同じkeyは拒否 |
| 異なるキーの二重会計 | A/Bが同じ既存会計を確定、または同一未請求治療/検査/接種を新規会計で同時確定。別々のkeyを使う | 同一業務対象の確定は1回のみ。敗者は既存結果か競合へ。別々の請求/支払として二重作成しない |
| 明細編集と確定 | Aが明細変更、Bが古い画面から確定 | 古い合計で確定せず、最新状態の再検証か競合拒否。部分的な明細/支払保存を残さない |
| 取消/通信断/再接続 | 取消・切断後に状態を再取得し再試行 | 不明な結果を新しいkeyで盲目的に再発行しない。取消後の状態を保持し、失敗は全rollback。永久ロックしない |
| 分離・独立操作 | 他医院で同一ID参照、同一医院の別カルテ/別会計を並行操作 | 他医院を読み書きしない。同一患者でも別の正当な会計まで一律禁止せず、衝突対象を定義する |

[会計complete](../../../backend/internal/billing/accounting_complete.go) は同一 `Idempotency-Key` のreplayと異なるpayloadの拒否を持ち、[transaction](../../../backend/internal/billing/accounting_complete_tx.go) が明細/支払等をまとめる。[frontend](../../../frontend/src/features/accounting/hooks/use-accounting-completion-action.ts) のkeyもmutation単位。この存在だけでは **別端末・別key** の二重会計の防止証拠にならない。

既存入口は [楽観ロックテスト](../../../backend/internal/medicalrecord/clinical_plan_repository_optimistic_lock_test.go)、[確定との競合](../../../backend/internal/medicalrecord/clinical_plan_finalize_concurrency_test.go)、[会計completeテスト](../../../backend/internal/billing/accounting_complete_test.go)。同一キーのmockテストと実DB並行トランザクションの検証を区別する。

## 2 セッションの合成再現と採否ゲート

1. 開発担当が上表と各更新API/既存テストの対応表を作り、衝突キー（clinic＋対象カルテ/明細/請求/未請求項目）と不足防御を特定する。防止目的を医院へ再質問しない。要件責任者の個人名と仕様変更の受入参照は変更前に実行票へ記録する。
2. scopedテストでRED→最小修正→GREENを実施し、候補mount済みの専用Docker/DBで実トランザクション競合と2ブラウザセッションの挙動を検証する。既存version/状態検証/一意性/行ロック/冪等性を優先し、確認ダイアログだけで安全性を成立させない。
3. ケース別にHTTP/競合表示、勝者の保存値、請求/支払/監査件数、rollback、再読込/再試行、医院分離、cleanupを記録する。外部通知がある場合はlocal stubを使い、実送信しない。すべての必須ケースの証拠が揃って完了。環境不足や未実行は残件にする。

**採否:** 全面的な画面占有ロックは未採用。既存防御で満たすケースは再実装しない。それでも占有が必要な不足ケースだけ、開始/期限/解除/切断復旧/権限/監査を別途裁定する。患者情報・秘密を本票に残さない。

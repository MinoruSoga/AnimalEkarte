# 会計・未納

会計一覧・精算・売掛管理を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[会計一覧](pages/accounting-list.md) → [精算詳細](pages/accounting-settlement.md)（§2 業務・安全機能、§3.3 再試行）→ [未納一覧](pages/unpaid-list.md)。来院完了の transaction は [横断フロー §5](../../workflows/reservation-to-record-flow.md)、締めの業務は [cash-register.md](../../cash-register.md)を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S08 会計訂正](../../../ops/testing/scenarios/S08-accounting-corrections.md)・[S21 会計並行操作と冪等性](../../../ops/testing/scenarios/S21-accounting-concurrency-idempotency.md)・[S33 領収書](../../../ops/testing/scenarios/S33-receipt-print-pdf.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

精算詳細のレセプト送信（#240 の着手しない判断）と外部機器連携の未実装を保持します。レジ締め全体の改訂は今回の整理に含みません。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

## 共通事項

関連マスタは [支払方法](../../screens/settings/payment-methods.md)・[保険](../../screens/settings/master-insurance.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

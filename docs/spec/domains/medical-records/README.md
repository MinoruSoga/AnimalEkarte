# カルテ

カルテ一覧・入力画面・保存と確定の契約を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[横断フロー](../../workflows/reservation-to-record-flow.md) §5 → [保存・確定契約](lifecycle.md#hybrid-sync)（特に §2）→ [一覧](pages/medical-record-list.md) → [詳細・入力](pages/medical-record-form.md)を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S06 記録ロックと監査](../../../ops/testing/scenarios/S06-record-lock-audit-trail.md)・[S17 処置数量の保存](../../../ops/testing/scenarios/S17-treatment-quantity-commit.md)・[V01 臨床フォーム](../../../ops/testing/scenarios/V01-clinical-forms.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

入力画面 §2.3 の臨床承認（#261）と横断フロー §7 の既知ギャップを保持します。コードのガードや Issue Closed と臨床承認を同一視しません。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

## 共通事項

関連マスタは [問診](../../screens/settings/master-interview.md)・[処置](../../screens/settings/master-treatment.md)・[薬剤](../../screens/settings/master-medicine.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

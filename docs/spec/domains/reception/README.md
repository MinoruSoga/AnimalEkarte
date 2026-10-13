# 受付

当日の患者進行・日付切替・予約から記録への遷移を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[横断フロー](../../workflows/reservation-to-record-flow.md) §2.3・§2.4・§5 → [当日の受付](pages/reception-board.md)の「主要機能」「操作詳細」を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S40 受付の戻り先](../../../ops/testing/scenarios/S40-reception-return-navigation.md)・[S01 死亡ペットのガード](../../../ops/testing/scenarios/S01-deceased-pet-guard.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

横断フロー §7 の未決・既知ギャップを確認します。通常診療とトリミングの既存遷移ルールは画面本文と横断フローを参照します。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

## 共通事項

関連マスタは [予約種別](../../screens/settings/master-reservation-type.md)・[スタッフ](../../screens/settings/master-staff.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

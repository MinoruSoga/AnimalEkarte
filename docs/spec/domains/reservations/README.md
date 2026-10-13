# 予約

予約作成・空き枠・次回予定・予約状態を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[予約管理](pages/reservation-management.md)の「予約登録フロー」「ステータス管理と連携」→ [横断フロー](../../workflows/reservation-to-record-flow.md) §2.2・§2.3・§3・§5を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S22 予約競合とスタッフ](../../../ops/testing/scenarios/S22-reservation-conflict-and-staff.md)・[S04 LINE予約](../../../ops/testing/scenarios/S04-liff-reservation-journey.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

画面本文の Windows スクロール・実DB並行要求の別途受入、および横断フロー §7 の既知ギャップを保持します。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

## 共通事項

関連マスタは [予約種別](../../screens/settings/master-reservation-type.md)・[スタッフ](../../screens/settings/master-staff.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

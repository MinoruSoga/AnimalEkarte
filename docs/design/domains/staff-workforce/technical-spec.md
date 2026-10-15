# D14 技術接続候補

業務要件の正本：[職員・シフト](../../../spec/domains/staff-workforce/README.md)。基準SHA `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。実装変更なし、実機未検証。

## 対応と所有者

- [docs/spec/screens/settings/master-staff.md](../../../spec/screens/settings/master-staff.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/spec/screens/settings/master-occupation.md](../../../spec/screens/settings/master-occupation.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/spec/screens/settings/master-shift-template.md](../../../spec/screens/settings/master-shift-template.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/spec/screens/24-shift-calendar.md](../../../spec/screens/24-shift-calendar.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。

## 受入trace

| 条件 | 技術確認の観点 |
|:---|:---|
| AC-staff-workforce-01 | B院勤務を追加時、既存人物へ所属を追加し別ログインを増やさないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-02 | 統合しようとする時、名称一致だけでは統合しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-03 | 休日または有休へ変更する時、終日非勤務となり時刻・休憩を勤務扱いしないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-04 | 後日テンプレートを改訂する時、既存勤務を無断変更しない。テンプレートは入力値のコピー元であり、登録済み勤務の一括改訂命令ではないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-05 | 過去勤務を訂正する時、診療・会計・予約の履歴を自動変更しない。各ownerの明示訂正契約を維持することを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-06 | A院での勤務終了を確定時、A院予約の解消を要求しB院勤務・予約・過去履歴は保持することを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-07 | 担当者を退職無効化する時、該当予約の解消まで拒否し自動終了・取消をしないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-08 | 職員を無効化時、過去完了履歴だけを理由に永久に拒否せず履歴を保持することを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-staff-workforce-09 | 指名候補・勤務を参照時、未登録を休日と推測せず他院勤務を自院勤務として扱わない。詳細判定は予約正本へ従うことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |

## 設計・検証残件

業務上の個別質問は追加しない。テンプレート改訂は入力コピー契約、過去勤務訂正の副作用禁止は既存履歴保持とowner分離の境界整理。各画面の実装方式・既存予約不整合の移行救済・医院別能力保存・日跨ぎ表現・並行保存は設計／移行残件。

- 保存前後で対象医院・対象人物・実効権限を拘束する。権限や状態が競合したら最新状態を確認して再操作し、別医院への自動再送を行わない。
- 必須の履歴／監査があるwriteは対応する既存ownerの原子的保存契約へ接続する。全CUDに同一audit要件があるとは仮定しない。
- UI表示／API拒否／DB保存／外部送信の成功は別々に確認し、ソース静的照合だけで実機PASSにしない。
- staff/account付与・医院新規・group新規等の送信後結果不明は、成功と失敗を推測せず対象ID／操作識別子・現在状態を照会する。receipt対応がなく照合不能なら停止し、同じcreateを無条件に再送しない。C05の保存・監査正本に接続し、既存APIに未対応のidempotency機能があるとは主張しない。

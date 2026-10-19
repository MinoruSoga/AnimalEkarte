# C04 技術接続候補

業務要件の正本：[認証・権限・医院scope](../../../spec/shared/access-control.md)。基準SHA `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。実装変更なし、実機未検証。

## 対応と所有者

- [docs/spec/screens/21-login.md](../../../spec/screens/21-login.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/spec/screens/settings/master-permission-group.md](../../../spec/screens/settings/master-permission-group.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/architecture/auth.md](../../../architecture/auth.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/architecture/adr/002-multitenancy-clinic-id-isolation.md](../../../architecture/adr/002-multitenancy-clinic-id-isolation.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [SECURITY.md](../../../../SECURITY.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。

## 受入trace

| 条件 | 技術確認の観点 |
|:---|:---|
| AC-access-control-01 | 再設定を申請時、画面応答で登録有無を露出しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-02 | 同じ操作を試す時、他の医院scope等を満たすならOR許可ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-03 | 直接APIで操作時、拒否されデータが不変ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-04 | A院でwriteし医院復旧する時、A院writeを拒否しB院へ自動再送しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-05 | 保護対象を取得・更新時、古いtokenの医院一覧を根拠に許可しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-06 | その権限を削除時、変更全体を拒否し既存管理権限を保持ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-07 | 資格情報を変更時、資格情報と成功監査が片方だけ確定しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-08 | ヘッダーなしmeで復旧時、有効医院なしと示し管理者確認・ログアウトへ案内。通常403や通信障害と混同しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-09 | 遅延した復旧応答が届く時、医院保存・reloadでログアウトを取り消さないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-10 | 新パスワードを確定時、拒否し成功したと表示しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-access-control-11 | 自己groupを変更時、OR後の権限を評価し自分のgroupという理由だけでは拒否しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |

## 設計・検証残件

業務上の新規grant・権限配布の承認は行わない。各D固有の取消／訂正／確定条件はそのowner正本へ委ねる。全routeの実装適合・権限実設定・cache有効設定・並行性・実機検証は未確認。実装事実のtoken寿命やDB構造はarchitecture正本を参照し複製しない。今回合意したD22公開版の認証職員READ・D23送信前確認を例外一覧へ同期する。

- 保存前後で対象医院・対象人物・実効権限を拘束する。権限や状態が競合したら最新状態を確認して再操作し、別医院への自動再送を行わない。
- 必須の履歴／監査があるwriteは対応する既存ownerの原子的保存契約へ接続する。全CUDに同一audit要件があるとは仮定しない。
- UI表示／API拒否／DB保存／外部送信の成功は別々に確認し、ソース静的照合だけで実機PASSにしない。
- staff/account付与・医院新規・group新規等の送信後結果不明は、成功と失敗を推測せず対象ID／操作識別子・現在状態を照会する。receipt対応がなく照合不能なら停止し、同じcreateを無条件に再送しない。C05の保存・監査正本に接続し、既存APIに未対応のidempotency機能があるとは主張しない。

# D17 技術接続候補

業務要件の正本：[医院・設定入口](../../../spec/domains/clinic-settings/README.md)。基準SHA `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。実装変更なし、実機未検証。

## 対応と所有者

- [docs/spec/screens/19-clinic-settings.md](../../../spec/screens/19-clinic-settings.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/spec/screens/20-master-settings.md](../../../spec/screens/20-master-settings.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。
- [docs/architecture/auth.md](../../../architecture/auth.md)：現行の仕様／実装／認可根拠。API・schemaはここに複製しない。

## 受入trace

| 条件 | 技術確認の観点 |
|:---|:---|
| AC-clinic-settings-01 | scope=allで一覧を見る時、割当外B院を取得できないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-02 | 対象設定を直接更新時、更新拒否、画面表示だけで認可を代替しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-03 | 医院を物理削除時、削除を拒否して既存記録を保持ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-04 | 旧医院を選択した状態で操作時、現在の医院authorityを再評価して拒否・復旧し、過去履歴を一括削除しないことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-05 | 保存前に影響確認時、全医院共通設定であることが分かる（表示方法はUI設計）ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-06 | 旧入力を保存しようとする時、対象医院を再確認しB院へA院の入力を自動転用しない（C04接続）ことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |
| AC-clinic-settings-07 | 一方を編集時、別エンティティを意図せず上書きせず、帳票の採用値はD10正本に従うことを該当ownerの保存前後・拒否結果・画面状態で確認。実行未実施 |

## 設計・検証残件

新しい個別質問を増やさず既存の現在医院認可・履歴保持へ整合する。医院全体無効化前に全未終了業務の解消を強制する新しい規則は今回追加しない。法人情報の発行済帳票への扱いはD10帳票snapshot正本へ委ねる。scope=allの旧記述は既存auth正本との整合修正。

- 保存前後で対象医院・対象人物・実効権限を拘束する。権限や状態が競合したら最新状態を確認して再操作し、別医院への自動再送を行わない。
- 必須の履歴／監査があるwriteは対応する既存ownerの原子的保存契約へ接続する。全CUDに同一audit要件があるとは仮定しない。
- UI表示／API拒否／DB保存／外部送信の成功は別々に確認し、ソース静的照合だけで実機PASSにしない。
- staff/account付与・医院新規・group新規等の送信後結果不明は、成功と失敗を推測せず対象ID／操作識別子・現在状態を照会する。receipt対応がなく照合不能なら停止し、同じcreateを無条件に再送しない。C05の保存・監査正本に接続し、既存APIに未対応のidempotency機能があるとは主張しない。
- 19-clinic-settingsのscope=allの旧説明はarchitecture/auth §4.10に反する。非system_adminは所属院のみ、system_adminは全院という最新正本へ説明を修正する。新たな実権限付与は不要。

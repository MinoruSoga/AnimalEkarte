# spec/ — 仕様系（業務・画面・UI）

> **目的**: システムが「何をするか」の仕様系ドキュメントの索引を提供する。
> **読者**: 全開発者・AI エージェント。
> **タイミング**: 機能実装・仕様確認・画面変更の前。

技術設計（どう作られているか）は [../architecture/](../architecture/README.md)、運用は [../ops/](../ops/README.md) を参照。
新機能・仕様変更の着手前には [../product-philosophy.md](../product-philosophy.md) の実践ゲートを必ず通すこと。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [specification.md](specification.md) | システム全体の機能要件と主要ビジネスフロー（§2.1 臨床安全原則を含む） | 全体像の把握・臨床安全の判断時 |
| [screens/](screens/README.md) | 全画面インデックス（画面別の詳細操作仕様。settings/ にマスタ設定画面群） | 画面単位の仕様確認・UI 変更前 |
| [cash-register.md](cash-register.md) | レジ締め・日次/月次売上集計の業務仕様（AM/PM/EMG 境界） | 会計・締め処理の実装前 |
| [customer-aggregation.md](customer-aggregation.md) | 顧客分析ダッシュボード（累計売上・来院頻度・CPM/LTV） | 集計・分析機能の実装前 |
| [reservation-to-record-flow.md](reservation-to-record-flow.md) | 予約からカルテ作成までの統合フロー（appointments = 単一 source of truth） | 予約・受付・カルテ連携の実装前 |
| [design-system.md](design-system.md) | Notion ライクなデザイン規約とデザイントークン | フロントエンド UI 実装前 |
| [ui-design-compliance.md](ui-design-compliance.md) | UI デザイン規約の準拠状況と機械検査（design-audit）の対応表 | 新規ページ追加時（§2 の表を同コミットで更新） |
| [line/](line/README.md) | LINE / Lステップ連携（CPM・配信トリガー・LIFF 予約・コスト分析） | LINE/Lステップ関連の実装前 |

## AI エージェント向け注記

- screens/ 配下と design-system.md が言及するコンポーネント名・フック名・画面数はローカル `make ci` の docs-symbol-drift ゲートで実在チェックされる。実在しないシンボル名を書くとローカル検査が失敗する。GitHub Actions の必須 gate ではない。
- 「臨床の安全」（specification.md §2.1）は product-philosophy.md より優先される。

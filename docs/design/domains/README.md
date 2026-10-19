# domains/ — 領域別の技術設計

業務領域ごとの実現方式と現行コードとの差分をまとめる。[予約](reservations/README.md)と[受付](reception/README.md)と[カルテの部分設計](medical-records/README.md)と[会計初回〜第十群の文書設計](accounting/README.md)が現在の入口。新しい領域を追加する際も、領域READMEに対象・読む順・正本を記す。

ここにはデータ・API・画面と処理・transaction・移行・検証の設計を置く。業務本文は[spec/domains](../../spec/domains/README.md)、全体アーキテクチャは[architecture](../../architecture/README.md)、実施記録は[work](../../work/README.md)を参照し、本文を二重管理しない。具体的タスク・依存関係・進捗はPlaneへ置く。

読む順：対象領域の業務正本 → 領域README → 技術仕様と参照コード。各記述の「現在実装／提案設計／業務判断待ち／未検証」を別軸で読む。本索引への掲載は実装・公開の許可ではない。


## 全業務の技術契約への接続（2026-10-15）

現行実装観察は特記がなければPR #670の開始head41fcf3eb（起点main7309b3a）の固定読取。後続mainの修正有無を断言せず、目標契約と実装適合を分ける。物理schema／wireの実装・実grant・移行・実機は後続工程だが、意味ID／版・保存単位・競合／結果不明・旧値停止の仕様未記載をここへ追い出さない。

| 管理ID | 技術の正本 | 対応業務 |
|:---|:---|:---|
| D01 | [予約](reservations/technical-spec.md) | [業務・受入](../../spec/domains/reservations/README.md) |
| D02 | [受付・来院](reception/technical-spec.md) | [業務・受入](../../spec/domains/reception/README.md) |
| D03 | [通常カルテ](medical-records/technical-spec.md) | [業務・受入](../../spec/domains/medical-records/README.md) |
| D04 | [飼主・ペット台帳](owners-pets/technical-spec.md) | [業務・受入](../../spec/domains/owners-pets/lifecycle.md) |
| D05 | [入院・ホテル](hospitalization/technical-spec.md) | [業務・受入](../../spec/domains/hospitalization/lifecycle.md) |
| D06 | [トリミング](trimming/technical-spec.md) | [業務・受入](../../spec/domains/trimming/lifecycle.md) |
| D07 | [検査・機器受信](examinations/technical-spec.md) | [業務・受入](../../spec/domains/examinations/lifecycle.md) |
| D08 | [予防接種](vaccinations/technical-spec.md) | [業務・受入](../../spec/domains/vaccinations/lifecycle.md) |
| D09 | [健診](checkups/technical-spec.md) | [業務・受入](../../spec/domains/checkups/lifecycle.md) |
| D10 | [会計・未納](accounting/technical-spec.md) | [業務・受入](../../spec/domains/accounting/lifecycle.md) |
| D11 | [見積](finance-operations/technical-spec.md) | [業務・受入](../../spec/domains/estimates/business-rules.md) |
| D12 | [レジ締め](finance-operations/technical-spec.md) | [業務・受入](../../spec/domains/cash-register/business-rules.md) |
| D13 | [在庫](finance-operations/technical-spec.md) | [業務・受入](../../spec/domains/inventory/business-rules.md) |
| D14 | [職員・シフト](staff-workforce/technical-spec.md) | [業務・受入](../../spec/domains/staff-workforce/README.md) |
| D15 | [CRM・Lステップ](../../spec/line/lstep-integration.md) | [業務・受入](../../spec/line/lstep-integration.md) |
| D16 | [経営・顧客集計](finance-operations/technical-spec.md) | [業務・受入](../../spec/domains/analytics/business-rules.md) |
| D17 | [医院・設定入口](clinic-settings/technical-spec.md) | [業務・受入](../../spec/domains/clinic-settings/README.md) |
| D18 | [LINE予約チャネル](../../spec/screens/28-line-reservation.md) | [業務・受入](../../spec/screens/28-line-reservation.md) |
| D19 | [同一飼主・ペット連携](../../spec/screens/40-identity-links.md) | [業務・受入](../../spec/screens/40-identity-links.md) |
| D20 | [LIFF診察券・健康情報](../../spec/screens/38-liff-pet-health.md) | [業務・受入](../../spec/screens/38-liff-pet-health.md) |
| D21 | [飼主カルテレポート](../../spec/screens/39-owner-report.md) | [業務・受入](../../spec/screens/39-owner-report.md) |
| D22 | [院内マニュアル](internal-manual/technical-spec.md) | [業務・受入](../../spec/domains/internal-manual/README.md) |
| D23 | [サポート](support/technical-spec.md) | [業務・受入](../../spec/domains/support/README.md) |

共有：[認証・権限](access-control/technical-spec.md)・[接続基盤](../shared/integration-contracts.md)。各表への掲載は実装済み・実機合格を意味しない。

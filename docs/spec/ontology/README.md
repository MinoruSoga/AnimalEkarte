# オントロジー

ここは用語・別名・定義・概念関係と多重度・制約の正本入口。全業務仕様が確定したという意味ではない。

- [予約ドメイン](reservations.md)：今回合意の概念と接続境界。個別業務FR・受入条件は元の正本へ参照。
- [カルテ初回合意](medical-records.md)：編集者・確認医師、確定／未完了警告、訂正・過去入力・評価不能の定義。
- [会計初回〜第十群](accounting.md)：請求と入金、未請求と未納、複合決済／部分入金、飼主単位預り金・予約任意／未充当残／利用可能額／充当／実返金、再配分・訂正後表示、割引scope／配賦・税抜計算と税込表示・税率別税額、複数保険会社・資格／残限度のスタッフ確認と保険負担額記録・窓口限定／ペットごと1利用契約、品目の全体割引除外、飼主宛名・ペット別明細／元取引と今回差分、マスタ変更権限、帳票と日付の区別、減額時の充当解除／残高復元・自動失効なし／同医院同飼主限定、再請求と再発行の区別、同時回収の元請求別配分、共通料金、再請求の元snapshot継承・旧入金の確認充当・保険再確認・返金元資金内訳と実返金方法の区別、空明細／0円・値引き前行額・master明示的再適用。
- 更新時は[文書整備手順](../authoring-guide.md#ontology-maintenance)に従い、元文書へ定義を複製しない。

## このフォルダの役割と使い方

用語・別名・定義と概念関係・多重度の正本入口。

概念の意味と関係、制約を持つ正本への参照を置く。個別FRの業務本文、技術モデル定義、履歴記録を複製しない。

子フォルダ：現在、子フォルダはありません。

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：[文書整備手順](../authoring-guide.md#ontology-maintenance) → [予約の定義・関係](reservations.md) → 参照先の各FR・受入条件。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。


## 全領域の意味境界と正本（2026-10-15）

[共通の用語・文書契約](../shared/document-contracts.md)で概念／画面／モデル、状態／IDと保存版の読み分けを整理する。語彙本文を索引へ複製せず、関係・多重度・任意性・制約は次の正本へ。専用ontologyを持たない領域も業務・技術の定義を照合対象とする。

| 領域 | 語彙・関係の正本 |
|:---|:---|
| D01／D02 | [予約・来院](reservations.md)・[受付](../domains/reception/README.md) |
| D03 | [通常カルテ](medical-records.md) |
| D04 | [飼主・ペット](owners-pets.md) |
| D05 | [入院・ホテル](hospitalization.md) |
| D06 | [トリミング](trimming.md) |
| D07／08／09 | [検査](../domains/examinations/lifecycle.md)・[接種](../domains/vaccinations/lifecycle.md)・[健診](../domains/checkups/lifecycle.md) |
| D10 | [会計・金銭](accounting.md) |
| D11／12／13／16 | [見積・締め・在庫・集計](finance-operations.md) |
| D14／D17 | [職員・勤務](../domains/staff-workforce/README.md)・[医院](../domains/clinic-settings/README.md) |
| D15／D18 | [CRM・配信](../line/lstep-integration.md)・[LINE予約](../screens/28-line-reservation.md) |
| D19／D20／D21 | [identity group](../screens/40-identity-links.md)・[本人健康表示](../screens/38-liff-pet-health.md)・[スタッフ読取](../screens/39-owner-report.md) |
| D22／D23 | [マニュアル版](../domains/internal-manual/README.md)・[報告／共有ROOM／個人履歴](../domains/support/README.md) |

実施／記録保存／確定、請求／入金／充当、取消／返金／原本削除、公開版／本人公開／スタッフ読取を同義にしない。現行コードのenumと選定した業務状態の差は各技術契約へ残す。

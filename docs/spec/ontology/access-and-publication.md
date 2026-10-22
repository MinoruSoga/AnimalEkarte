# 認可・医院・公開と参照の境界

<a id="authority-and-clinic-concepts"></a>
## 主体とscopeの定義

2026-10-18の概念整理。既存[認可正本](../shared/access-control.md)、[保持正本](../shared/retention-audit.md)、[全owner対応](../shared/owner-contract-coverage.md)を再利用する。ここでは業務概念と許可の関係を整理し、role/permissionの新設、DB/API/認証方式の選定、医院間公開の拡大を行わない。

| 用語 | 定義・識別・関係 | 制約の正本／追跡 |
|:---|:---|:---|
| 職員／認証主体 | 職員は勤務・役割等の人、認証主体は実際にログインし操作するaccount。職員が必ずログインaccountを持つとは限らない | [D14](../domains/staff-workforce/README.md)、[スタッフ設定](../screens/settings/master-staff.md)。職種・役職・雇用を認可roleへ自動連動しない |
| 操作actor／担当者／確認者 | 実際の操作主体、業務を担当する人、目的別確認を行う人という別役割 | [D03](medical-records.md)、[滞在](hospitalization.md)、[施術](trimming.md)。入力者を実施者/確認医師へ推定しない |
| 所属医院／選択医院／勤務医院 | 所属可能範囲、画面の現在context、その時刻の実勤務という別関係 | [予約語彙](reservations.md)。複数所属でも同時刻他院勤務を指名候補へ含めない |
| 認可された操作 | 現在の対象・医院・操作別の許可を満たすこと。閲覧と更新、master変更と個別会計変更は別 | [FR-access-control-independent-authority](../shared/access-control.md#requirement-fr-access-control-independent-authority)、AC-access-control-02／03 |
| 編集権の現在保持 | 同一カルテの同時編集を止めるための現在の権保持 | [D03編集権](medical-records.md#record-edit-right-concept)。編集操作の認可を付与せず、認可済でも別端末編集中なら同時編集不可 |
| 共通基本情報の利用 | 認可内で同じ飼主/個体の最新基本情報を自院業務へ利用する関係 | [D04](owners-pets.md#common-patient-concept)、AC-owners-pets-other-clinic-use／AC-owners-pets-common-denied。院別診療記録/資金の無制限公開や移管ではない |
| identity group／本人紐付け | D19の手動link/unlink、D20の本人確認を伴う紐付けは各既存目的の関係 | [D19](../screens/40-identity-links.md)・[D20](../screens/38-liff-pet-health.md)。共通基本情報の内部共有方式をこの関係で代用決定しない |
| projection／読取表示 | 元ownerの保存情報を目的・認可に従い参照する表示。別の独立記録本文とは限らない | [領域境界](cross-domain-contracts.md)。表示先の権限が元ownerのwrite許可を生まない |

医院・ログイン・権限失効と保存失敗を区別する。対象医院の所属が失効したwriteを拒否し、別院へ自動再送しない。復旧中のログアウトを遅延応答で取り消さず、有効医院なしと通信障害も分ける。[FR-access-control-revocation-recovery](../shared/access-control.md#requirement-fr-access-control-revocation-recovery)、AC-access-control-04／05／08／09。

<a id="read-and-publication-boundaries"></a>
## 公開・読取・変更の関係

| 対象／参照方向 | 許される意味・対象 | 許可へ読み替えないもの／正本 |
|:---|:---|:---|
| スタッフ患者トップ←D07結果 | 認可された結果版/原値/日付/出典/訂正状態の投影 | D07結果の台帳属性PATCH・無認可他院閲覧ではない。AC-owners-pets-infection-write-boundary／AC-owners-pets-infection-privacy |
| 受付カード←トリミング要望 | 同予約・ペット・対象日の仕上がり要望を表示 | 予約notes上書きや施術記録作成/画像印刷必須ではない。AC-reception-trimming-style-target／AC-reception-trimming-style-empty-failure |
| D20飼主健康表示←許可された健康情報 | 本人確認/紐付けと公開対象に従うread | スタッフ認可・内部特記・原記録writeを与えない。[D20](../screens/38-liff-pet-health.md) |
| D21スタッフ複合読取←各記録owner | 認可された保存済み範囲の読取、section単位の失敗/縮退 | 一般履歴全体の印刷共有は追加採用しない。予防薬処方箋/年間投薬表の専用2帳票だけ既承認例外。[D21](../screens/39-owner-report.md) |
| D15配信→LINE | DM希望/目的別停止/宛先/結果不明保留等を満たす既存送信 | DM必要や台帳変更だけで送信/再送/停止解除しない。[Lステップ](../line/lstep-integration.md) |
| D16分析←請求/実移動/来院/配信ログ | 元指標と母集団・時点に基づくread計算 | 結果表示による配信/売上/金銭移動の生成、未提供画面の新採用ではない。[金融周辺](finance-operations.md) |
| D17設定入口→各master | 各masterのscope・権限と有効状態を個別に扱う | 設定ページ閲覧だけで全master/global動物種の編集を許可しない。[医院設定](../domains/clinic-settings/README.md) |
| D23サポート報告←スクリーンショット添付の外部送信 | Discord・Planeへ渡す内容と送信先を送信前に表示して確認する | スクリーンショットは個人情報を含んでよく、強制マスキングを送信の必須条件にしない（確認は維持）。完全匿名化の保証や加工UI・原画像破棄は採用しない。[2026-10-20本人回答](../domains/support/README.md#support-screenshot-personal-information-20261020) |
| D22公開manual記事←編集/公開版 | 製品内manualの草稿・公開版・履歴を区別 | 開発仕様書や未公開草稿を自動公開しない。[D22](../domains/internal-manual/README.md) |
| D23共有報告一覧／ROOM←各報告／chat元メッセージ | 報告の共有readと報告元医院のmutation、個人chat履歴、同一clinic×staffのchat元メッセージを30分超の空白で区切るROOM表示を区別 | ROOMはバグ報告由来の記録や新しい永続session IDではない。報告元医院のmutation条件は報告に対応し、共有閲覧から他院報告の変更/外部送信許可を生まない。[D23](../domains/support/README.md)・[設定入口](../screens/settings/README.md) |

<a id="retention-and-visibility-concepts"></a>
## 履歴・非表示・物理保存

保存済み原本/必要監査/業務履歴、通常画面の表示、公開対象、配信用一時コピーは別概念である。非表示や論理削除は元履歴の物理廃棄・認可外公開の根拠にならない。[FR-common-retention-interim](../shared/retention-audit.md#requirement-fr-common-retention-interim)、FR-common-retention-visibility、FR-common-retention-trace、AC-common-retention-age／AC-common-retention-hidden／AC-common-retention-money／AC-common-retention-record-revertを追跡する。

保存済み原画像・添付・参照済key・分類不明の業務原本は通常clear/置換/取消で物理削除しない。未保存かつ未送信の放棄作業copyとは区別し、分類不明は保護する。配信一時copyの既存限定cleanup条件を臨床原本へ一般化せず、新保存年限・新物理削除手続きを作らない。

未保存入力のreload回復を可能にする承認と、その保存場所/保持期間・端末共有方式の未選定は両立する。共通の原本保持方針から下書きの無期限保存や新たな他端末公開を導かない。用語・公開境界の整理は法令適合、全table監査、全原媒体網羅、実機PASSの宣言ではない。

# 個別画面詳細仕様書 一覧

本ディレクトリには、システムを構成する各画面の具体的な仕様、レイアウト、データ項目、およびAPI連携の詳細を定義したドキュメントが格納されています。

概要レベルの仕様は **[SPECIFICATION.md](../SPECIFICATION.md)** を参照してください。

---

## 画面一覧

| No | 画面名 | ファイル | 概要 |
|:---|:---|:---|:---|
| 00 | 共通ペット選択 | [00-pet-selection.md](./00-pet-selection.md) | 各機能の新規作成時に使用される共通中間画面 |
| 01 | ダッシュボード | [01-dashboard.md](./01-dashboard.md) | 当日の受付状況管理（カンバンボード） |
| 02 | 予約管理 | [02-reservations.md](./02-reservations.md) | 診療・トリミング等のカレンダー管理 |
| 03 | 飼主・ペット一覧 | [03-owners-list.md](./03-owners-list.md) | 顧客情報の検索・一覧表示 |
| 04 | 飼主・ペット登録/編集 | [04-owners-form.md](./04-owners-form.md) | 顧客情報の入力・更新、ペット管理 |
| 05 | カルテ一覧 | [05-medical-records-list.md](./05-medical-records-list.md) | 過去の診療記録の検索・参照 |
| 06 | カルテ入力/編集 | [06-medical-records-form.md](./06-medical-records-form.md) | SOAPS形式・9タブ構成の診療記録入力 |
| 07 | 入院管理一覧 | [07-hospitalization-list.md](./07-hospitalization-list.md) | 入院・ホテル患者のボード/リスト管理 |
| 08 | 入院詳細・デイリーカルテ | [08-hospitalization-detail.md](./08-hospitalization-detail.md) | 入院中のケア計画と日次実施記録 |
| 09 | 入院登録/編集 | [09-hospitalization-form.md](./09-hospitalization-form.md) | 入院予約・初期プラン設定 |
| 10 | 会計一覧 | [10-accounting-list.md](./10-accounting-list.md) | 会計データの検索・算定状況確認 |
| 11 | 会計精算 | [11-accounting-detail.md](./11-accounting-detail.md) | 保険適用、支払処理、帳票発行 |
| 12 | 検査一覧 | [12-examinations-list.md](./12-examinations-list.md) | 検査オーダーと結果概要の参照 |
| 13 | 検査入力/結果登録 | [13-examinations-form.md](./13-examinations-form.md) | 検査詳細データの入力（カルテ連携） |
| 14 | 予防接種一覧 | [14-vaccinations-list.md](./14-vaccinations-list.md) | 接種履歴の検索・証明書発行導線 |
| 15 | 予防接種フォーム | [15-vaccinations-form.md](./15-vaccinations-form.md) | ワクチン接種記録の入力（カルテ連携） |
| 16 | トリミング一覧 | [16-trimming-list.md](./16-trimming-list.md) | トリミング予約と施術状況の一覧 |
| 17 | トリミング登録/編集 | [17-trimming-form.md](./17-trimming-form.md) | 施術内容・スタイル希望・画像の記録 |
| 18 | 在庫管理 | [18-inventory-list.md](./18-inventory-list.md) | 薬品・消耗品の在庫数とアラート管理 |
| 27 | 在庫登録/編集 | [27-inventory-form.md](./27-inventory-form.md) | 在庫品目の入力・更新 |
| 19 | 病院情報設定 | [19-clinic-settings.md](./19-clinic-settings.md) | 医院名、住所、インボイス番号等の設定 |
| 20 | マスタ設定トップ | [20-master-settings.md](./20-master-settings.md) | システム全体の各種マスタ管理 |
| 21 | ログイン | [21-login.md](./21-login.md) | ユーザー認証と初期遷移 |
| 22 | 見積書一覧 | [22-estimate-list.md](./22-estimate-list.md) | 作成済み見積書の管理 |
| 23 | 見積書作成・編集 | [23-estimate-form.md](./23-estimate-form.md) | 概算見積の算出と発行 |
| 24 | シフト管理カレンダー | [24-shift-calendar.md](./24-shift-calendar.md) | スタッフ勤務シフトの可視化・管理 |
| 25 | 定期健診一覧 | [25-checkups-list.md](./25-checkups-list.md) | 全ペットの定期健診記録の参照 |
| 26 | 見積書詳細 | [26-estimate-detail.md](./26-estimate-detail.md) | 見積内容の確認と管理 |

## マスタ設定詳細

個別マスタの定義は [settings/README.md](./settings/README.md) を参照してください。

---

## 共通要素

- [共通ダイアログ・共有コンポーネント](./common-dialogs.md)

# 画面詳細仕様書 インデックス (Screen Specifications Index)

> **目的**: 全41画面の仕様書インデックスを提供する。
> **読者**: 新規参加エンジニア・PdM。
> **タイミング**: 画面インデックス参照・全体像把握時。

本ディレクトリ直下の番号付き仕様（`[0-9]*.md`）は **41 ファイル**（フロー文書 `99-medical-record-flow` を含む）。製品ルートの product leaf 数（`route-inventory` の 84）とは数え方が異なる。各ファイルに画面の機能、レイアウト、API 連携の詳細を定義する。

---

## 🩺 臨床・診療コア (Clinical Core)

| No | 画面名 | 仕様書 | 概要 |
|:---|:---|:---|:---|
| 00 | **ペット選択** | [00-pet-selection.md](./00-pet-selection.md) | 新規データ作成時の共通検索・特定フロー。 |
| 01 | **当日の受付** | [01-reception.md](./01-reception.md) | カンバン形式による院内稼働状況の管理。 |
| 02 | **予約管理** | [02-reservations.md](./02-reservations.md) | 月/週カレンダーによる予約枠とシフトの可視化。 |
| 05 | **カルテ一覧** | [05-medical-records-list.md](./05-medical-records-list.md) | 全診療記録の時系列検索。 |
| 06 | **カルテ詳細・入力** | [06-medical-records-form.md](./06-medical-records-form.md) | SOAPS 形式の診療録作成（9 タブ構成）。 |
| 12 | **検査一覧** | [12-examinations-list.md](./12-examinations-list.md) | 検査オーダー状況と結果の進捗管理。 |
| 13 | **検査登録・結果** | [13-examinations-form.md](./13-examinations-form.md) | 数値検査の入力と基準値判定。 |
| 14 | **予防接種一覧** | [14-vaccinations-list.md](./14-vaccinations-list.md) | 接種実績と次回予定の時系列リスト。 |
| 15 | **予防接種登録** | [15-vaccinations-form.md](./15-vaccinations-form.md) | ワクチン履歴記録と次回予定自動計算。 |
| 16 | **トリミング一覧** | [16-trimming-list.md](./16-trimming-list.md) | 施術予約と完了ステータスの管理。 |
| 17 | **トリミング登録** | [17-trimming-form.md](./17-trimming-form.md) | 施術内容と仕上がり画像の記録。 |
| 25 | **定期健診一覧** | [25-checkups-list.md](./25-checkups-list.md) | 健診履歴の参照と当日カルテ自動生成。 |
| 39 | **飼主カルテレポート** | [39-owner-report.md](./39-owner-report.md) | 飼主単位の全ペット診療サマリーを別ウィンドウで俯瞰。 |

---

## 🛏️ 入院・ホテル管理 (Inpatient Care)

| No | 画面名 | 仕様書 | 概要 |
|:---|:---|:---|:---|
| 07 | **入院管理一覧** | [07-hospitalization-list.md](./07-hospitalization-list.md) | ボードビューによるケージ稼働状況の監視. |
| 08 | **入院詳細・記録** | [08-hospitalization-detail.md](./08-hospitalization-detail.md) | デイリーケア計画と時系列バイタル記録。 |
| 09 | **入院登録・編集** | [09-hospitalization-form.md](./09-hospitalization-form.md) | ケアプランと値引き設定の定義。 |

---

## 💰 会計・経営管理 (Finance & Admin)

| No | 画面名 | 仕様書 | 概要 |
|:---|:---|:---|:---|
| 03 | **飼主・ペット一覧** | [03-owners-list.md](./03-owners-list.md) | 顧客データベースの検索と参照。 |
| 04 | **飼主・ペット登録** | [04-owners-form.md](./04-owners-form.md) | 顧客情報の編集と Lステップ個別送信。 |
| 10 | **会計一覧** | [10-accounting-list.md](./10-accounting-list.md) | 請求履歴の検索とステータス監視。 |
| 11 | **会計精算** | [11-accounting-detail.md](./11-accounting-detail.md) | 保険窓口精算、決済、インボイス発行。 |
| 22 | **見積書一覧** | [22-estimate-list.md](./22-estimate-list.md) | 発行済み見積の有効期限と承認管理。 |
| 23 | **見積書作成** | [23-estimate-form.md](./23-estimate-form.md) | 概算費用の算出と飼主向け説明文編集。 |
| 26 | **見積書詳細** | [26-estimate-detail.md](./26-estimate-detail.md) | 見積内容の最終確認とステータス制御。 |
| 29 | **レジ締め・履歴** | [29-closing-aggregation.md](./29-closing-aggregation.md) | 日次売上の確定と現金の実査・履歴。 |
| 30 | **未納者一覧** | [30-unpaid-list.md](./30-unpaid-list.md) | 売掛金の把握と督促業務支援。 |
| 32 | **月次集計レポート** | [32-accounting-reports.md](./32-accounting-reports.md) | 経営分析用データの抽出と CSV 出力。 |
| 36 | **顧客集計ダッシュボード** | [36-aggregation-dashboard.md](./36-aggregation-dashboard.md) | 売上・来院頻度による顧客分析。 |

---

## 📦 設定・外部連携・監視 (Infrastructure & CRM)

| No | 画面名 | 仕様書 | 概要 |
|:---|:---|:---|:---|
| 18 | **在庫管理一覧** | [18-inventory-list.md](./18-inventory-list.md) | 品目別の在庫監視と発注点管理。 |
| 27 | **在庫登録・編集** | [27-inventory-form.md](./27-inventory-form.md) | 薬品・備品の基本情報管理。 |
| 24 | **シフト管理** | [24-shift-calendar.md](./24-shift-calendar.md) | スタッフ勤務と LINE 予約の連動. |
| 28 | **LINE 予約設定** | [28-line-reservation.md](./28-line-reservation.md) | 予約システム稼働ルールと文言編集・予約枠カレンダー. |
| 37 | **LINE 予約（飼主側）** | [37-line-reserve-owner-flow.md](./37-line-reserve-owner-flow.md) | 飼主が LINE から予約を作成・確認・キャンセルする 13 ステップのフロー。 |
| 38 | **LIFF 診察券・健康情報** | [38-liff-pet-health.md](./38-liff-pet-health.md) | 飼い主向け LINE ミニアプリ。健康手帳（ワクチン記録）表示と LINE アカウント紐付け。 |
| 31 | **Lステップ連携設定** | [31-lstep-integration.md](./31-lstep-integration.md) | CPM 判定、配信自動化、各種プレフィックス設定。 |
| - | ├ **Lステップタグ管理** | [31-lstep-integration.md](./31-lstep-integration.md) | 連携タグのマスタ管理・コードマッピング。 |
| - | ├ **健診タグ一括同期** | [31-lstep-integration.md](./31-lstep-integration.md) | 健診対象者の抽出と一括タグ連携（プレビュー付）。 |
| - | └ **Lステップ顧客分析** | [31-lstep-integration.md](./31-lstep-integration.md) | CPM 分析と友だち属性推移ダッシュボード。 |
| 34 | **Lステップ配信監視** | [34-lstep-delivery-monitor.md](./34-lstep-delivery-monitor.md) | 自動配信トリガーの実行ログと失敗検知。 |
| 35 | **取扱説明書** | [35-internal-manual.md](./35-internal-manual.md) | システム内マニュアルの閲覧と編集。 |
| 40 | **同一飼主・ペット連携** | [40-identity-links.md](./40-identity-links.md) | 所属医院内の飼主・ペット手動 identity link と最小連携履歴。 |
| 21 | **ログイン** | [21-login.md](./21-login.md) | 認証プロセスとパスワード再設定。 |
| 19 | **医院マスタ設定** | [19-clinic-settings.md](./19-clinic-settings.md) | 拠点基本情報と税務設定。 |

---

## 🛠️ 技術・共通基盤

- **[共通ダイアログ](./common-dialogs.md)**: 全画面で共有される検索、入力部品。
- **[カルテ保存フロー](./99-medical-record-flow.md)**: 診療記録の複雑なライフサイクル。
- **[マスタ設定ポータル](./20-master-settings.md)**: 各種定義データ管理の入り口。
- **[マスタ設定画面群 (settings/)](./settings/README.md)**: 個別マスタ設定画面の詳細仕様インデックス。

---

**最新更新**: 2026-08-14 | **ステータス**: Static/Code Sync (123 Tables / 37 Resources; Fresh DB Apply Pending — `q&a.html` OPS-13)

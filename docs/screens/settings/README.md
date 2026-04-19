# 個別マスタ設定 仕様書 一覧

本ディレクトリには、システム全体で共有されるマスタデータの管理画面（個別カテゴリ）の仕様が定義されています。

全体の一覧とパターン分類は **[master-pages.md](./master-pages.md)** を参照してください。

---

## 主要マスタ仕様

| カテゴリ | ファイル | 概要 |
|:---|:---|:---|
| スタッフ | [master-staff.md](./master-staff.md) | 獣医師・スタッフおよびログインアカウント管理 |
| 権限グループ | [master-permission-group.md](./master-permission-group.md) | ロールベースのアクセス制御（RBAC）設定 |
| 診療項目 | [master-treatment.md](./master-treatment.md) | 診察・検査・処置・予防・健診の共通ツリー管理 |
| 薬剤 | [master-medicine.md](./master-medicine.md) | カテゴリ別薬剤単価と剤形管理 |
| 診断 | [master-diagnosis.md](./master-diagnosis.md) | 診断カテゴリと具体的な病名の紐付け |
| トリミング | [master-trimming.md](./master-trimming.md) | コースとオプションの組合せ・単価管理 |
| 物販 | [master-merchandise.md](./master-merchandise.md) | フードやグッズ等の名称・単価・税率管理 |
| ケージ | [master-cage.md](./master-cage.md) | 入院用ケージのエリア・サイズ・料金管理 |
| 問診・テンプレート | [master-interview.md](./master-interview.md) | 主訴カテゴリと問診票テンプレート定型文 |
| 保険 | [master-insurance.md](./master-insurance.md) | ペット保険会社の補償率・連絡先管理 |
| シフトテンプレート | [master-shift-template.md](./master-shift-template.md) | シフト入力パターンの管理 |
| 基礎マスタ | [master-basics.md](./master-basics.md) | 動物種、職能、予約区分、入院プラン等の共通設定 |
| **集計・締め時間** | `docs/tasks/pending/accounting/FEAT-368_closing-aggregation.md` Section 5 | AM/PM境界・診療終了時刻・特別期間の設定（管理者のみ） |

---

## 共通パターン

個別マスタの多くは以下の共通コンポーネントパターンを使用しています。
- **一覧**: `[C] MasterCRUDPage` (DataTable + 検索フィルタ)
- **編集**: `[C] MasterSidePanel` (Notionスタイルのスライドインパネル)

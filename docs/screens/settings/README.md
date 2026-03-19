# 個別マスタ設定 仕様書 一覧

本ディレクトリには、システム全体で共有されるマスタデータの管理画面（個別カテゴリ）の仕様が定義されています。

全体の一覧とパターン分類は **[master-pages.md](./master-pages.md)** を参照してください。

---

## 主要マスタ仕様

| カテゴリ | ファイル | 概要 |
|:---|:---|:---|
| スタッフ | [master-staff.md](./master-staff.md) | 獣医師・スタッフおよびログインアカウント管理 |
| 診療項目 | [master-treatment.md](./master-treatment.md) | 診察・検査・処置・予防・健診の共通ツリー管理 |
| 薬剤 | [master-medicine.md](./master-medicine.md) | カテゴリ別薬剤単価と剤形管理 |
| 診断 | [master-diagnosis.md](./master-diagnosis.md) | 診断カテゴリと具体的な病名の紐付け |
| トリミング | [master-trimming.md](./master-trimming.md) | コースとオプションの組合せ・単価管理 |

---

## 共通パターン

個別マスタの多くは以下の共通コンポーネントパターンを使用しています。
- **一覧**: `[C] MasterCRUDPage` (DataTable + 検索フィルタ)
- **編集**: `[C] MasterSidePanel` (Notionスタイルのスライドインパネル)

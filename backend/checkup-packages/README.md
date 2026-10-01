# 健診パッケージ投入マニフェスト（EMR-169）

バージョン付き健診パッケージの JSON マニフェスト群。`checkup_types` /
`checkup_type_fields` を医院スコープで作成するための定義ソース。
直下 `backend/migrations/*.sql` は DDL 専用、CSV seed は医院骨格のみのため本方式を使う。

> **EMR-249**: `POST /api/v1/checkup-package-imports[/preview]` import API は
> 削除された。投入・修正は健診タイプ/項目のマスタ CRUD
> （`/api/v1/masters/checkup-types*`、EMR-225）で行う。
> 既存医院の投入済みデータと `checkup_package_import_receipts` の
> 適用履歴はこの変更の影響を受けない。

## ファイル一覧と内容

| ファイル | 健診タイプ | 項目 |
|---|---|---|
| `annual-checkup.json` | 年4健診 | なし（尿検査項目は要PO確認） |
| `valentine-checkup.json` | バレンタイン健診 | なし（尿検査項目は要PO確認） |
| `may-checkup.json` | 5月健診 | なし（尿検査項目は要PO確認） |
| `adpuritto-checkup.json` | アドプリット検診 | アドプリットレベル（単一選択 0〜5） |
| `skin-checkup.json` | 皮膚検診 | 異常の有無・治療の要否（boolean） |
| `ear-checkup.json` | 耳検診 | 異常の有無・治療の要否（boolean） |
| `eye-checkup.json` | 眼科検診 | 眼圧 左右・涙量 左右（number）＋傷の有無（boolean） |

## スコープ外・未確定（詳細は EMR-169 レポート）

- 尿検査の項目名・基準値、皮膚・耳の状態チェック項目の内訳、眼圧・涙量の基準値は
  共有原文に具体的な記述がないため未設定（PO 確認後にマスタ CRUD で追加する）。
- イラスト記入・写真添付は未実装機能のため対象外。代替としてカルテ添付機能を利用する。

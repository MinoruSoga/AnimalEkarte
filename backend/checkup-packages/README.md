# 健診パッケージ投入マニフェスト（EMR-169）

バージョン付き健診パッケージ import API（TASK-374、
`backend/internal/medicalrecord/checkup_package_*.go`）へ投入する JSON マニフェスト群。
`checkup_types` / `checkup_type_fields` を医院スコープで作成する正式経路。
直下 `backend/migrations/*.sql` は DDL 専用、CSV seed は医院骨格のみのため本方式を使う。

## 適用（ユーザー手動・医院ごと）

1. 実行 staff に `checkup-package-import:create` 権限を付与する
   （現行の権限表では全ロール DENY のため、設定画面で対象ロールへ付与してから実行）。
2. preview（dry-run・DB 非書込）:
   `POST /api/v1/checkup-package-imports/preview` に対象 manifest JSON を POST する。
3. apply:
   `POST /api/v1/checkup-package-imports` に同じ JSON を POST する。
   - 同一内容の再投入は `noop`（冪等）
   - 同 `namespace`/`version` で内容が異なる再投入は `conflict` で拒否
   - 同名 `checkup_type` が既に存在する医院では conflict。
     医院側で名称を解消してから再投入する
4. 各 manifest は医院非依存。必要な医院それぞれに同じ JSON を apply する。

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
  共有原文に具体的な記述がないため未設定（PO 確認後に別 manifest で追加する。
  現行の import API は既投入タイプへの項目追加をサポートしない点に注意）。
- イラスト記入・写真添付は未実装機能のため対象外。代替としてカルテ添付機能を利用する。

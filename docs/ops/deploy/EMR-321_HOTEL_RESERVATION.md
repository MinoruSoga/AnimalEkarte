# EMR-321: ホテル予約区分とケージ選択の手動確認

本票は標準予約区分の整備とケージ選択不可の調査を扱う。ケージの任意化は EMR-316、一時連結は EMR-306 の担当範囲。既存医院へのデータ投入、migration、CSV 生成は実行者が手動で行う。

## ケージ選択不可の切り分け

対象医院で、入院・ホテル登録とケージマスタの閲覧権限を持つスタッフが確認する。

1. 対象医院を確認し、設定の [ケージマスタ](../../spec/screens/settings/master-cage.md)（`/settings/cage`）を開く。医院を切り替えた後は再読み込みする。
2. ブラウザの network で `GET /api/v1/masters/cages` の HTTP status、返却件数、選択中の医院を記録する。氏名・患者情報・認証ヘッダーは証跡に含めない。
3. `200` で空配列なら、その医院の live ケージマスタがない。必要なケージ名・エリア・サイズを医院責任者に確認し、マスタの「新規登録」で有効なケージを登録する。別医院のケージを流用しない。
4. `401` / `403` ならログイン状態と `master-cage` 閲覧権限を確認する。通信失敗や `5xx` を「マスタ未登録」と判定しない。
5. 行がある場合は、削除状態・有効/無効・連結状態を確認する。連結メンバーは入院フォームの候補から除かれ、代表ケージが候補になる。連結を解除する操作は EMR-306 の手順と医院責任者の判断に従う。
6. 入院・ホテル登録を開き直し、候補の名称と代表ケージを確認する。ケージを使わない預かりは「選択なし」で登録する（EMR-316）。ボードの「ケージ未割当」で確認する。

`useGetMasterItems` は現在エラーを呼出側へ返さず、取得できない場合も空配列として表示され得る。そのため、候補が空という見た目だけでは未登録と通信・権限エラーを区別できない。医院別の実データと STG 応答はこの PR では未確認。

医院別の件数を確認する場合、実行者は承認済みの接続方法で次の読み取り専用 SQL を実行する。医院 ID と件数だけを記録する。

```sql
BEGIN READ ONLY;
SELECT c.id AS clinic_id,
       count(g.id) FILTER (WHERE g.deleted_at IS NULL) AS live_cages,
       count(g.id) FILTER (WHERE g.deleted_at IS NULL AND g.is_active) AS active_cages,
       count(g.id) FILTER (WHERE g.deleted_at IS NULL AND g.is_active
                           AND g.merged_into_cage_id IS NULL) AS active_representatives
FROM clinics c
LEFT JOIN cages g ON g.clinic_id = c.id
GROUP BY c.id ORDER BY c.id;
ROLLBACK;
```

## ホテル区分の決定と適用前確認

所要時間はチェックイン時の来院枠であり、預かり日数ではない。`duration_minutes` は院内と LINE で共通（5〜480 分）。カテゴリは現行 enum の `general` / `trimming` の2値で、ホテル登録への導線は区分名の「入院」または「ホテル」で判定する。

標準区分の所要時間・色・順序・LIFF 公開方針は EMR-321 の決定コメントを正本とし、未決定の値で投入しない。

2026-10-08 のユーザー回答で、名称「ホテル」、`category=general`、`duration_minutes=15`、色 `#06B6D4`、`sort_order=5`（健診の後）、`reservation_visible=false`、`is_internal=false` と決定。既存同名 live 有効行の属性にはこの標準値を強制しない。

実行前に、対象環境・接続先・全医院への適用範囲を実行者が確認する。バックアップと既存行の属性・ID を安全な運用記録に保持する。以下は読み取り専用の事前確認であり、無効行を再活性化する操作ではない。

```sql
BEGIN READ ONLY;
SELECT c.id AS clinic_id, r.id, r.name, r.is_active, r.deleted_at,
       r.duration_minutes, r.color, r.sort_order, r.category,
       r.reservation_visible, r.is_internal
FROM clinics c
LEFT JOIN reservation_types r ON r.clinic_id = c.id AND r.name = 'ホテル'
ORDER BY c.id, r.id;
ROLLBACK;
```

- 同名の live 有効行がある医院: 既存行の属性を維持する。標準値で上書きしない。
- 同名の live 無効行がある医院: 停止する。医院責任者が既存行の扱いを判断する。投入 SQL は再活性化せず、全体を rollback する。
- 同名の削除済み行だけがある医院: 停止して、復元と新規作成のどちらにするか医院責任者へ確認する。削除理由を無視して自動復元しない。新規作成を承認した場合は、その決定を記録してから手動投入する。
- 行がない医院: 決定済みの標準値で新規投入できる。

## 手動投入と事後確認

仕様決定と事前確認を終えてから実行する。`PGSERVICE` / 保護された接続設定を使用し、接続文字列・パスワードをコマンドや証跡に貼らない。

```sh
psql -X -v ON_ERROR_STOP=1 \
  -f backend/migrations/seeds/live_insert_standard_reservation_types.sql
```

この SQL は `cmd/migrate` の自動適用対象ではない。既存同名 live 行は上書きせず、医院×標準区分の全組合せが有効でなければ transaction 全体を rollback する。正常終了だけで UI 受入済みとはしない。

```sql
BEGIN READ ONLY;
SELECT c.id AS clinic_id,
       count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) AS active_hotels
FROM clinics c
LEFT JOIN reservation_types r ON r.clinic_id = c.id AND r.name = 'ホテル'
GROUP BY c.id
HAVING count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) <> 1;
ROLLBACK;
```

結果が0行なら全医院にホテルの live 有効行が1件ずつ存在する。同名既存行の ID・属性が事前記録と一致することを確認する。再実行後も行数・ID・属性が変わらないことを確認する。

院内予約で「ホテル」を選択して予約を保存し、予約詳細から「入院・ホテル登録」へ進む。LIFF 非公開の決定なら飼主向けの選択肢に現れないことも確認する。別医院の行が見えないことを確認する。

## fresh DB と失敗時の扱い

この手動 SQL の適用だけでは `002_master/reservation_types.csv` は更新されない。fresh DB への収録は [Seed / Migration 運用](SEED_MIGRATION_OPERATIONS.md) §3 の正規経路（承認済みの使い捨て DB に実適用し、`cmd/seed-export` で再生成）を別途ユーザーが実行する。CSV や manifest を手編集しない。既存・共有 DB を reset しない。

SQL 失敗時は transaction 全体の rollback を確認し、原因を解消して再実行する。commit 後に問題が分かった場合は、事前記録で今回新規作成した ID を特定し、医院責任者に報告する。予約への参照を確認してから既存マスタ UI の無効化等を判断する。既存行や他医院の行を一括削除しない。

EMR-321 のコメントへ対象環境、実行者、仕様決定、適用日時、医院別件数、再実行結果、UI 受入結果を記録する。実データや認証情報は記載しない。

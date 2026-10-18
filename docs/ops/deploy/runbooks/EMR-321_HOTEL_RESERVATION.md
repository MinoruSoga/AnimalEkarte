# EMR-321: ホテル予約区分とケージ選択の手動確認

本票は標準予約区分の整備とケージ選択不可の調査を扱う。ケージの任意化は EMR-316、一時連結は EMR-306 の担当範囲。既存医院へのデータ投入、migration、CSV 生成は実行者が手動で行う。

## ケージ選択不可の切り分け

対象医院で、入院・ホテル登録とケージマスタの閲覧権限を持つスタッフが確認する。

1. 対象医院を確認し、設定の [ケージマスタ](../../../spec/screens/settings/master-cage.md)（`/settings/cage`）を開く。医院を切り替えた後は再読み込みする。
2. ブラウザの network で `GET /api/v1/masters/cages` の HTTP status、返却件数、選択中の医院を記録する。氏名・患者情報・認証ヘッダーは証跡に含めない。
3. `200` で空配列なら、その医院の live ケージマスタがない。必要なケージ名・エリア・サイズを医院責任者に確認し、マスタの「新規登録」で有効なケージを登録する。別医院のケージを流用しない。
4. `401` / `403` ならログイン状態と `master-hospitalization:view` 閲覧権限（[ケージマスタ正本](../../../spec/screens/settings/master-cage.md)の `ResourceMasterHospitalization`）を確認する。通信失敗や `5xx` を「マスタ未登録」と判定しない。
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

## 標準5区分の決定と適用前確認

**基準と現在の適用境界**：以下はEMR-321／2026-10-08当時のenum・名称判定と標準値投入の手順記録であり、最新の[明示業務分類による接続](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-routing)へ適合済みという意味ではない。現在の診療・トリミング・ホテル・入院の接続を区分名だけで推測せず、[予約技術契約](../../../design/domains/reservations/technical-spec.md)の分類対応と照合する。SQL・当時の決定値は履歴として維持するが、現行仕様向け投入として利用する前に分類・対象版・互換性と実施承認を確認する。本書更新はSQL実行許可ではない。

所要時間はチェックイン時の来院枠であり、預かり日数ではない。`duration_minutes` は院内と LINE で共通（5〜480 分）。カテゴリは現行 enum の `general` / `trimming` の2値で、ホテル登録への導線は区分名の「入院」または「ホテル」で判定する。

標準区分（診察・お手入れ・ワクチン・健診・ホテルの5種）の所要時間・色・順序・LIFF 公開方針は EMR-321 の決定コメントを正本とし、未決定の値で投入しない。

2026-10-08 のユーザー回答で、名称「ホテル」、`category=general`、`duration_minutes=15`、色 `#06B6D4`、`sort_order=5`（健診の後）、`reservation_visible=false`、`is_internal=false` と決定。既存同名 live 有効行の属性にはこの標準値を強制しない。

実行前に実行者が、対象環境・接続先（承認済みの接続方法と接続情報の照合）・実施承認・実行者権限・バックアップ取得と、適用範囲が `clinics` 全医院であることを確認する。医院の絞り込み条件は設けず、画面で選択中の医院だけを対象にしない。既存行の属性・ID は安全な運用記録に保持する。以下は読み取り専用の事前確認であり、無効行を再活性化する操作ではない。投入スクリプトは `cmd/migrate` の入力ではなく単一 transaction で実行され、既存の live 行の再活性化・上書きをしない。

migrate 済み・002_master 投入済みの環境では `clinics` の医院 inventory（件数と ID 集合）が非ゼロであることを期待値として事前に記録する。医院が1件も存在しない、または事後確認で事前記録と inventory が乖離する場合は停止する — postcheck が0行でも成功としない。

```sql
BEGIN READ ONLY;
SELECT count(*) AS clinics FROM clinics;
SELECT c.id AS clinic_id, d.name AS desired_name, r.id, r.is_active, r.deleted_at,
       r.duration_minutes, r.color, r.sort_order, r.category,
       r.reservation_visible, r.is_internal
FROM clinics c
CROSS JOIN (VALUES ('診察'), ('お手入れ'), ('ワクチン'), ('健診'), ('ホテル')) AS d(name)
LEFT JOIN reservation_types r ON r.clinic_id = c.id AND r.name = d.name
ORDER BY c.id, d.name, r.id;
ROLLBACK;
```

実行者は医院×名称の各組合せを、行なし（`r.id` が NULL）・live 有効・live 無効・削除済みのみに区別する。

- 同名の live 有効行がある組合せ: 既存行の属性を維持する。標準値で上書きしない。
- 同名の live 無効行がある組合せ: 停止する。医院責任者が既存行の扱いを判断する。投入 SQL は再活性化せず、全体を rollback する。
- 同名の削除済み行だけがある組合せ: 停止して、復元と新規作成のどちらにするか医院責任者へ確認する。削除理由を無視して自動復元しない。新規作成を承認した場合は、その決定を記録してから手動投入する。
- 行がない組合せ: 決定済みの標準値で新規投入できる。

## 手動投入と事後確認

仕様決定と事前確認を終えてから実行する。`PGSERVICE` / 保護された接続設定を使用し、接続文字列・パスワードをコマンドや証跡に貼らない。

```sh
psql -X -v ON_ERROR_STOP=1 \
  -f backend/migrations/seeds/live_insert_standard_reservation_types.sql
```

この SQL は `cmd/migrate` の自動適用対象ではない。既存同名 live 行は上書きせず、医院×標準区分の全組合せが有効でなければ transaction 全体を rollback する。正常終了だけで UI 受入済みとはしない。

```sql
BEGIN READ ONLY;
SELECT c.id AS clinic_id, d.name,
       count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) AS live_active
FROM clinics c
CROSS JOIN (VALUES ('診察'), ('お手入れ'), ('ワクチン'), ('健診'), ('ホテル')) AS d(name)
LEFT JOIN reservation_types r ON r.clinic_id = c.id AND r.name = d.name
GROUP BY c.id, d.name
HAVING count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) <> 1
ORDER BY c.id, d.name;
ROLLBACK;
```

上の `HAVING` 付き異常検出クエリは、live 有効行が1件でない医院×名称だけを返す。このクエリの結果が0行であることを確認する。

正常な COMMIT 後にも、事前確認と同じ読み取り専用ブロック（`clinics` の件数と全医院×5名称の ID・属性 SELECT）を再実行して事後 inventory を取得する。inventory は正常でも行を返すため、0行を合格条件にしない。同名の削除済み履歴も返るため、行数を一律に「5 × 医院数」と比較しない。医院 inventory（件数と ID 集合）が非ゼロで事前/事後に一致し、同名既存行の ID・属性が事前記録と一致して上書きされていないことを確認する。異常検出が0行であることと、この inventory 照合の両方を満たした場合のみ合格とする。医院が存在しない、または inventory に乖離がある場合は停止する。冪等性の立証だけを目的に書き込み SQL を再実行しない。

基準トリミング行の重複・欠落も読み取り専用で確認する。

```sql
BEGIN READ ONLY;
SELECT c.id AS clinic_id,
       count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) AS active_baseline_trimming
FROM clinics c
LEFT JOIN reservation_types r ON r.clinic_id = c.id AND r.name = 'トリミング'
GROUP BY c.id
HAVING count(r.id) FILTER (WHERE r.deleted_at IS NULL AND r.is_active) <> 1
ORDER BY c.id;
ROLLBACK;
```

`002_master` が各医院へ入れる基準行は名称 `トリミング` の1件であり、確認対象はその名称だけである — trimming カテゴリ全体の件数は数えない。医院が独自名で追加した区分は重複ではなく、そのまま維持する。基準 `トリミング` 行が複数（重複）または0件（欠落）の医院が返った場合は医院責任者への確認対象とし、本手順では修復しない。対象行の ID・属性が事前記録と一致することも確認する。

管理画面で各医院に5区分が表示され、トリミングが重複表示されず、別医院の行が見えないことを確認する。院内予約で「ホテル」を選択して予約を保存し、予約詳細から「入院・ホテル登録」へ進む。今回新規投入した区分は飼主向け LIFF の選択肢に現れないことも確認する。既存同名行の公開設定は維持するため、既存行が公開中なら医院責任者にその状態を報告し、無断で変更しない。

## fresh DB と失敗時の扱い

この手動 SQL の適用だけでは `002_master/reservation_types.csv` は更新されない。fresh DB への収録は [Seed / Migration 運用](../seeds/SEED_MIGRATION_OPERATIONS.md) §3 の正規経路（承認済みの使い捨て DB に実適用し、`cmd/seed-export` で再生成）を別途ユーザーが実行する。CSV や manifest を手編集しない。既存・共有 DB を reset しない。

SQL 失敗時は transaction 全体の rollback を確認し、原因を解消して再実行する。通信切断や COMMIT 応答が不明な場合は盲目で再実行せず、まず読み取り専用の事後確認クエリで実際の状態を人間が確認する。判別できない場合は未確認として停止し、エスカレーションする。commit 済みと判明した後に rollback を強制したり既存行を削除しない。commit 後に問題が分かった場合は、事前確認と同じ全医院×5名称の読み取り専用 SELECT を再実行して行の ID・属性 inventory を取得し、事前 inventory と照合する。事前に欠落し事後に出現した行は、今回投入した行の候補に留める。事前確認後に通常 UI で同名行を登録すると、投入 SQL の `NOT EXISTS` はその組合せを skip するが、snapshot 差分には出現する。SQL の advisory lock は通常登録経路と共有されず、`INSERT RETURNING` による挿入 ID の証跡もないため、snapshot 差分だけでは作成元を断定できない。今回操作と行 ID を対応付ける独立した挿入証跡がない行は無効化対象にしない。由来が不明なら停止してエスカレーションする。postcheck が0行でもそれだけでは成功とは判断しない。独立証跡で今回挿入したと確定できた ID は医院責任者に報告し、予約への参照を確認してから既存マスタ UI の無効化等を判断する。既存行や他医院の行を一括削除しない。

非本番環境での実施記録（対象環境、実行者、仕様決定、適用日時、医院別件数、UI 受入結果）は EMR-321 のコメントへ記録する。本番環境での実施は EMR-348 の作業であり、実行記録は EMR-348 へ残し、本書は正規手順としてリンクする。承認された復旧フローで実際に再実行した場合に限り、その再実行結果も対応する記録先へ記載する。実データや認証情報は記載しない。

# ステージング CRUD スモークテスト

> **目的**: デプロイ直後に医院・権限グループ・スタッフの認可とresource stateを確認する。
> **境界**: 共有STGへの実行は承認済みoperatorのみ。資格情報、cookie値、PHIを記録しない。

自動 `stg-smoke.yml` はhealthのみである。この手順は手動で実施し、作成したIDと元値をrun sheetに記録する。

## 1. 準備

```bash
export API_V1=https://api.stg.noah-karte.com/api/v1
umask 077
export COOKIE_JAR=/secure/path/stg-smoke.cookies
RESPONSE_FILE="$(mktemp)"
chmod 600 "$COOKIE_JAR" "$RESPONSE_FILE"
trap 'rm -f "$RESPONSE_FILE"' EXIT
```

- `hospital-settings:view` を持つ承認済みアカウントと、持たないアカウントを用意する。
- POST/PATCH/DELETE は `X-Requested-With: XMLHttpRequest` を必須とする（`RequireXRequestedWith`）。対象医院は認可済み session の `X-Clinic-ID` 契約に従い、他院へ流用しない。
- browser/curlのHttpOnly cookie認証を使う。cookie/token/password値を文書、shell history、artifactへ残さない。
- 既存医院を編集する場合は、変更前の値を安全なrun sheetへ記録し、同じAPIで必ず復元する。

## 2. 医院

### A-1 全医院scope

`hospital-settings:view` を持つsessionで実行する。

```bash
curl -sS -o ${RESPONSE_FILE} -w '%{http_code}
' \
  "${API_V1}/clinics?scope=all" -b "$COOKIE_JAR" -H 'Accept: application/json'
```

**期待:** `200`。`system_admin`という名前ではなく、必要permissionで判定する。

### A-2 権限拒否

`hospital-settings:view` を持たない別sessionで同じ `?scope=all` を実行する。

**期待:** `403`。scopeを省略した一覧はassigned clinicsを返す別contractであり、この拒否testの代用にしない。

### A-3 既存医院の編集と復元

承認済みtest clinicだけをPATCHし、`200`とGETしたresource stateを確認する。直後に保存済みの元値へPATCHして復元し、再GETする。別clinic IDへのaccessも認可どおり拒否されることを確認する。医院をこのrunで新規作成しない。

## 3. 権限グループ

### B-1 作成

```bash
curl -sS -X POST "${API_V1}/masters/permission-groups" \
  -b "$COOKIE_JAR" -H 'X-Requested-With: XMLHttpRequest' \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"TEST-GROUP-<run-id>",
    "description":"Smoke test temporary group",
    "color":"#64748B",
    "is_active":true,
    "sort_order":9999,
    "rules":[{
      "resource":"medical-records",
      "can_view":true,
      "can_create":false,
      "can_edit":false,
      "can_delete":false
    }]
  }'
```

**期待:** `201`。返されたIDを `TEST_GROUP_ID` として記録する。`permissions/action`形式は使わない。

### B-2 in-use保護

同じclinic内でactive staff assignmentが存在する、事前に確認済みのgroupを選ぶ。そのIDへのDELETEが `409` になることを確認する。ID `1` や `system_admin`という名前を保護contractとしてhard-codeしない。

### B-3 cleanup

このrunがB-1で作成し、staff assignmentが無い `TEST_GROUP_ID` だけをDELETEする。

**期待:** `204`。list/detailで非activeまたはnot foundとなる実装contractを確認する。

## 4. スタッフ

### C-1 CRUD-only staff作成

```bash
curl -sS -X POST "${API_V1}/masters/staffs" \
  -b "$COOKIE_JAR" -H 'X-Requested-With: XMLHttpRequest' \
  -H 'Content-Type: application/json' \
  -d '{"name":"TEST-STAFF-<run-id>","sort_order":9999}'
```

**期待:** `201`。返されたIDを `TEST_STAFF_ID` として記録する。このCRUD payloadに存在しない `role` を送らず、email/password無しのstaffでlogin testをしない。

### C-2 account loginが必要な場合

CRUD作成とは分離する。[スタッフアカウント払い出し](./STAFF_ACCOUNT_PROVISIONING.md) の承認済み経路で、有効な `staff_type` と生成済み8〜72文字passwordを持つaccountを作る。login routeは `POST /api/v1/login`。credential値は記録しない。remote provisioning mechanismが未承認なら、このcaseは **BLOCKED** と記録し、CRUD staffへemailだけを追加して代替しない。

### C-3 FK保護とcleanup

- active child recordを持つ同一clinic staffへのDELETEが `409` になることを確認する。
- このrunでC-1に作成しchild recordが無い `TEST_STAFF_ID` だけをDELETEし、`204`と後続GET/list stateを確認する。

## 5. 期待status

| case | 期待 |
|---|---|
| authorized GET/PATCH | `200` |
| valid POST | `201` |
| unassigned test resource DELETE | `204` |
| active dependencyありDELETE | `409` |
| `scope=all` permissionなし | `403` |
| invalid payload | `400` |

想定外の4xx/5xx、tenant越境、復元失敗はrelease stopとする。

## 6. 監査contract

| route family | この手順での合格証跡 | `audit_logs` |
|---|---|---|
| clinics | HTTP status + GETしたresource state + 元値復元 | uniformな明示contractなし。存在を必須にしない |
| permission groups | HTTP status + resource state | mutationに明示的監査contractあり。actor/action/resource/clinicを確認 |
| staffs | HTTP status + resource state | uniformな明示contractなし。存在を必須にしない |

Workers Logsはruntime failureの調査用で、application auditの代替ではない。未実装のclinic/staff auditを手動INSERTして成功扱いにしない。

## 7. cleanupと直接DB操作の境界

1. このrunが作成したIDだけを対象にする。
2. dependencyの子から親へAPIで削除する。
3. GET/listでresource stateを確認する。
4. permission-groupだけは明示contractに従いauditも確認する。

直接DB削除は通常cleanupではない。API不能時の承認済みincident procedureに限り、target、clinic scope、対象ID、backup/restore、transaction、rollback条件、post-check、実施者を事前に固定する。「大量」「APIでは面倒」だけを理由にしない。application由来に見せるmanual audit rowを作らない。

## 8. 結果記録

commit/deploy run、実施時刻、caseごとのstatus、作成/cleanupした非PHI ID、復元結果、blockerを記録する。cookie、password、responseの個人情報は残さない。

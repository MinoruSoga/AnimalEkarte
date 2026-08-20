# Staff Account Batch Provisioning（#255）

> **目的**: 受領予定スタッフ一覧を **preflight → 全件 atomic apply** で account / staff / permission / clinic へ一括発行する。  
> **読者**: 認可済みオペレータ（system admin または scope 全クリニックの `master-staff` create 権限保持者）。  
> **境界**: 本ドキュメントは **repo 側の準備とコマンド契約** のみ。実スタッフ一覧・初期パスワード・PROD 適用は USER 作業。

更新日: 2026-07-30

## 何を解決するか

| 以前 | 本コマンド |
|------|-----------|
| account / staff / permission / clinic を分割 request → 部分成功しうる | 1 transaction で全件 create + 同 tx audit |
| 再実行の根拠がない | clinic-scoped な PII-free receipt（batch_id / digest / count） |
| 入力検証が apply と混在 | `preflight` は **write 0**、receipt 照合は apply の advisory lock 内 |

## 安全境界（必須）

1. **実スタッフ data / 初期パスワードを git に置かない。** manifest / secrets は repo 外の absolute realpath、**mode 0600 のみ**。
2. **log / receipt / 標準出力に name・email・password・file body を出さない。** digests / counts / batch_id のみ。
3. **role → permission 推論はしない。** `permission_group_ids` は manifest の明示値のみ（main clinic 所属の group）。
4. **PROD / 共有 STG への apply は USER 承認後。** 開発者 laptop から非 local `DB_HOST` への apply は `STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND` が無いと拒否される。対象環境のコンテナ内実行を正とする。
5. **本 repo の synthetic fixture は架空 ID / `@example.test` のみ。** 外部入力を fixture として commit しない。

## 入力契約

### Manifest（JSON, strict decode / unknown field 拒否）

| フィールド | 制約 |
|-----------|------|
| `schema_version` | 固定 `staff-provision-v1` |
| `clinic_scope` | 昇順・重複なし・非空。**全 staff の main/assignment clinic の集合と完全一致** |
| `batch_id` | `staff-provision:` + `SHA-256(hex)` of `clinic_scope` joined by `,`（例: `1,2,10`） |
| `actor_account_id` | 有効 account。system admin **または** scope 全クリニックで `master-staff` create |
| `staff[]` | 非空。各行に以下を明示 |

各 `staff` 行:

- `external_staff_id`（batch 内一意）
- `name`, `email`（batch 内 email 一意）
- `main_clinic_id`, `clinic_ids`（main を含む、scope の部分集合、重複なし）
- `permission_group_ids`（main clinic 所属 group のみ。空配列可）
- `occupation_id`（null 可。指定時は main clinic の active occupation）
- `staff_type`（`doctor` / `nurse` / `trimmer` / `resource`）
- `is_active`, `reservation_visible`
- `secret_ref`（batch 内一意。secrets ファイルと 1:1）

### Secrets（JSON, strict decode）

```json
{
  "secrets": [
    { "secret_ref": "sec-001", "password": "……" }
  ]
}
```

- `secret_ref` は manifest と **完全一致の 1:1**（余剰・欠落・重複は preflight 失敗）
- パスワード方針は通常の staff 作成と同一（8+ 文字、英字+数字、72 バイト以下）

### batch_id の計算例（オペレータ側）

```bash
# clinic_scope が [1,2,10] のとき
printf '%s' '1,2,10' | shasum -a 256
# => <hex>
# batch_id = staff-provision:<hex>
```

## コマンド

バイナリ: `backend/cmd/staff-provision`

```bash
# 作業用: 入力は /secure/... など repo 外・0600
chmod 600 /secure/staff-batch/manifest.json /secure/staff-batch/secrets.json

# 1) preflight（write 0）
docker compose exec backend go run ./cmd/staff-provision preflight \
  --manifest=/secure/staff-batch/manifest.json \
  --secrets=/secure/staff-batch/secrets.json

# 2) apply（USER 承認後・対象環境の DB）
docker compose exec backend go run ./cmd/staff-provision apply \
  --manifest=/secure/staff-batch/manifest.json \
  --secrets=/secure/staff-batch/secrets.json
```

オプション:

| flag / env | 意味 |
|------------|------|
| `--repo-root` | 入力禁止ルートを追加（absolute） |
| `STAFF_PROVISION_REPO_ROOT` | 同上 |
| `STAFF_PROVISION_ALLOW_REMOTE=YES_I_UNDERSTAND` | 非 local `DB_HOST` への apply を明示許可（通常は環境内実行） |

成功時 stdout は PII-free JSON（`status` / `batch_id` / `digest` / `staff_count` / `clinic_scope`）。

## 処理順序

### preflight

1. 両ファイル: absolute / realpath / **mode 0600** / **repo 外** / regular file / symlink 拒否  
2. strict JSON decode（unknown field 拒否）  
3. 件数・重複・scope 集合・batch_id namespace・password 方針  
4. DB 参照のみ: clinic / occupation / permission_group / email 未使用 / actor 認可  
5. **write 0**。receipt は見ない（認可前の存在漏洩を避ける）

### apply

1. preflight と同一の入力・認可検証  
2. transaction 開始 → `pg_advisory_xact_lock(hashtextextended('staff-provision:'||batch_id, 0))`  
3. **認可済み clinic_scope に限定**して receipt 照合  
   - 全 clinic が同一 digest → **noop**（再実行安全）  
   - 一部欠落 / digest 不一致 / 同一 batch 異 digest → **conflict**（並行含む）  
   - scope 外 receipt は検索しない。error に存在有無を出さない  
4. 全 staff を 1 tx で create（account + staff + assignments + permission groups）  
5. staff 単位 audit（`staff.provision.create`）と、影響 clinic ごとの receipt audit（`staff.provision.receipt`）を **同 tx**  
6. commit

## 監査・receipt

| action | resource | payload（PII-free） |
|--------|----------|---------------------|
| `staff.provision.create` | `staff` | `batch_id`, `digest`, `external_staff_id`, `staff_id` |
| `staff.provision.receipt` | `staff_provision_batch` | `batch_id`, `digest`, `count` |

`digest` は manifest 内容の正規化ハッシュ（name/email は fingerprint のみ。password 非含有）。

## 残る USER 作業（repo 外）

#255 の実適用には次が必要（本 packet では UNREPORTED）:

1. email 方針の確定  
2. clinic 対応表  
3. 休職 / 退職 / 委託者の発行可否  
4. role → **明示** permission_group_ids マッピング  
5. 実一覧・secret 配布・対象環境・authorized actor での apply

### 不足入力チェックリスト（2026-08-20）

値・氏名・email・パスワード・架空スタッフは書かない。未供給は **未記入**。本番 apply はしない。

| ID | 不足入力 | なぜ必要か | 供給者 | 状態 |
|---|---|---|---|---|
| I-ROSTER | 実スタッフ一覧（氏名・所属院・役割） | #255 本文のブロッカー。manifest `staff[]` の源泉 | 先方 | **未記入**（受領記録なし） |
| I-EMAIL | email 方針（個人必須 / 共有禁止は Q&A No.30） | manifest `email` を埋められるか | PO | **未記入** |
| I-CLINIC | 院 → `clinic_id` 対応表 | `clinic_scope` / `main_clinic_id` | 運用 | **未記入** |
| I-ROLE | 役割 → **明示** `permission_group_ids`（推論禁止） | 権限グループ割当 | PO | **未記入** |
| I-LEAVE | 休職 / 退職 / 委託者を発行するか | `is_active` 方針 | PO | **未記入** |
| I-ACTOR | 認可済み `actor_account_id` | preflight 認可 | USER | **未記入** |
| I-ENV | 適用先（local / STG / PROD）と承認 | apply は USER。PROD は #253/#254 gate 後 | USER | **未記入** |
| I-RECEIPT | 認可済み適用証跡（PII-free `batch_id` / digest / count） | #255 AC（発行・権限・audit） | USER apply 後 | **未記入** |

**やらない:** 架空スタッフの invent、repo への実 roster コミット、本番 apply、PII を Issue / Linear に書く。

## 検証（開発）

worktree を backend にマウントしたうえで:

```bash
docker compose run --rm --no-deps --entrypoint '' -T \
  -v "$WT/backend:/app" backend \
  go test -p 1 ./internal/staff -run 'TestStaffProvisioning' -count=1

docker compose run --rm --no-deps --entrypoint '' -T \
  -v "$WT/backend:/app" backend \
  go test -p 1 ./cmd/staff-provision -count=1
```

## 関連コード

- `backend/internal/staff/staff_provisioning.go`
- `backend/internal/staff/staff_provisioning_repository.go`
- `backend/cmd/staff-provision/`
- `backend/internal/model/audit_log.go`（action/resource 定数）

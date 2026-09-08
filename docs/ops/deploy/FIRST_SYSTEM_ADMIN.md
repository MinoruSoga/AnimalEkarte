# 本番初回システム管理者の作成手順

> **目的**: 公開の管理者登録画面を作らず、指定運用担当者が本番の最初のシステム管理者を 1 回だけ用意する。  
> **読者**: 認可済みの運用担当者。  
> **境界**: 人間が実行する初回専用トランザクション。実アドレス・パスワード・担当者名は git に書かない。本番への実付与は別承認。

更新日: 2026-09-09

## この手順が必要な理由

- ログイン画面とスタッフ新規作成画面から、最初のシステム管理者を作ることはできない。
- 既存スタッフへのアカウント追加 `POST /api/v1/masters/staffs/{id}/account` は **既存のシステム管理者** が前提である。
- 一括発行 `staff-provision` も **既存の認可済み actor** が前提である。ゼロから作る経路ではない。詳細は [STAFF_ACCOUNT_PROVISIONING.md](./STAFF_ACCOUNT_PROVISIONING.md)。

## やらないこと

- 一括発行 CSV / LoginForm / デモ catalog に本番管理者を載せない。
- 誰でも使える管理者登録画面を作らない。
- 既存のシステム管理者がいる環境で、この手順を再実行して追加のシステム管理者を作らない。
- 共有パスワードや非本番デモ資格情報で代替しない。
- 本書へ実メール・氏名・パスワードを記入しない。

## 環境の切り分け

| 環境 | 経路 |
| --- | --- |
| local / development / staging | migrate フェーズ3の `SEEDLOGIN_OPERATOR_*`。未設定なら upsert しない。値は repo 外。`APP_ENV=production` では `seedlogin.ShouldApply` が拒否する。 |
| production | この文書の初回専用手順。migrate の operator upsert は使わない。 |

## 本番初回の条件

1. 対象環境の `is_system_admin=true` が **0 件** であること。無効・論理削除済みも含めて拒否する。既存管理者の復旧は別の承認済み手順で扱う。
2. 指定された運用担当者だけが、安全な経路（秘密管理・対面または承認済みチャネル）で資格情報を扱う。
3. 作成する account は `is_system_admin=true`・`is_active=true`。ログイン用 staff を重複作成しない。既存の診療担当 staff がある場合は、別 staff を作らずその staff へ account を付ける。
4. 作成結果（account ID / staff ID / clinic 所属の有無）と監査を残す。パスワード・hash・token はログに出さない。
5. 作成後、担当者は通常ログインできることを確認し、初期秘密値を破棄する。以降のスタッフ追加は画面のアカウント追加、または既存の staff-provision を使う。

## 重複・競合

- email 衝突、既存 account の再作成、既存システム管理者の追加作成は拒否する。
- 通信失敗後の再実行は、同じ email の既存管理者を上書きしない条件を先に確認する。
- 失敗したら業務 write を部分成功のまま残さない。

## 人間による実行

本節は `001_init.sql` の `accounts` / `staffs` / `clinics` / `staff_clinic_assignments` / `audit_logs` に対応する。対象の適用済みスキーマが一致することを事前確認する。対象 staff と有効な主所属 clinic / assignment が既にある場合に限り使う。staff や所属が未整備なら停止し、承認済みのマスタ整備を先に行う。本手順は既存 staff を変更せずに複製する代替経路を持たない。

### 1. 対象と入力の準備

認可済みオペレータが対象環境内の承認済み管理用コンテナで実行する。DB 接続先・TLS 検証・実行ロールを、対象環境の秘密管理で設定した libpq service `first-system-admin` に固定する。接続文字列やパスワードをコマンド引数に渡さない。PostgreSQL 18 の専用運用ロールに次の権限が準備済みであることを確認する。権限不足なら停止し、この作業中に権限を拡張しない。

| 対象 | 必要な権限 |
| --- | --- |
| `accounts` | `SELECT`, `INSERT`, `MAINTAIN` |
| `staffs` | `SELECT`, `UPDATE (account_id, updated_at)`, `MAINTAIN` |
| `clinics`, `staff_clinic_assignments` | `SELECT`, `MAINTAIN` |
| `audit_logs` | `SELECT`, `INSERT` |
| account / audit の関連sequence、database、public schema | sequence `USAGE`、database `CONNECT` / `TEMPORARY`、schema `USAGE` |

`SHARE ROW EXCLUSIVE` は `SELECT` / `INSERT` だけでは取得できないため、ロックする4テーブルには `MAINTAIN` を要求する。独立した `LOCK` 権限という名称のGRANTはない。専用運用ロールは通常アプリ接続に使わない。[PostgreSQL 18 LOCK の権限要件](https://www.postgresql.org/docs/18/sql-lock.html#SQL-LOCK-NOTES)を参照。

service 定義は repo 外の `/secure/first-system-admin/pg_service.conf`（所有者は実行者、mode `0600`）を read-only mount し、`host`、`port`、`dbname`、`user`、`sslmode=verify-full` と信頼する CA を明示する。入力ファイルを読み込む前に、同じ定義で次を実行し、接続先を承認記録の provider project/branch・hostname/port・database・role と照合する。database 名と role が一致するだけでは別サーバを除外できない。確認結果には資格情報を含めず、公開チケットには接続先詳細を転載しない。

```sh
PGSERVICEFILE=/secure/first-system-admin/pg_service.conf PSQL_HISTORY=/dev/null \
  psql -X -w --dbname='service=first-system-admin' --command='\conninfo'
```

以下の 8 列・**ヘッダなし、データ 1 行だけ**の UTF-8 CSV を、秘密管理から `/secure/first-system-admin/input.csv` に供給する。repo 外の通常ファイル、所有者は実行者、mode `0600`、親ディレクトリ `0700`、symlink 不可とし、コンテナへ read-only mount する。実値をシェル履歴・Issue・画面共有・git に載せない。

| 列順 | 内容 |
| --- | --- |
| 1 | 対象 `current_database()` と一致する database 名 |
| 2 | 対象 `session_user` と一致する DB ロール名 |
| 3 | 既存 `staff_id`（正の整数、18 桁以内） |
| 4 | staff の既存主所属 `clinic_id`（正の整数、18 桁以内） |
| 5 | 承認済み個人メールアドレス（前後空白なし） |
| 6 | 本人だけに配布するパスワードの bcrypt hash（`$2a$12$` または `$2b$12$`、60 文字） |
| 7 | 承認記録の参照 ID（英数字・`._:-`、1〜100 文字） |
| 8 | 実行者の運用記録参照 ID（同じ文字制約。氏名ではなく、承認記録で実行者を特定できる ID） |

パスワードは通常アカウントと同じ方針（8 文字以上、英字と数字、72 バイト以内）で、秘密をログ出力しない承認済みのオフライン bcrypt ツールで cost 12 の hash を作り、元パスワードとの照合成功も確認する。SQL は hash の形式のみ検証し、元パスワードの強度・一致を検証できない。`pgcrypto` 拡張や DB 側の平文ハッシュ化は使わない。

入力ファイルの CSV 構造・列数を秘密を表示しない方法で事前検証する。`COPY` の失敗詳細には入力行が含まれ得るため、管理端末・DB/proxy のログ設定が COPY データとエラー中の個人情報・hash を記録しないことを運用管理者が確認する。確認できなければ実行しない。本手順を理由に既存監査やアクセス制御を無効化しない。

### 2. 1 トランザクションで作成

下記 SQL ブロックを**実値を埋め込まず**管理用コンテナの `/secure/first-system-admin/create.sql` に保存する。SQL 内の `\copy` が入力ファイルを読み込むため、メールや hash が SQL リテラル・コマンド引数にならない。psql の起動設定・履歴・SQL エコーを避けて実行する（人間のみ）。

```sh
PGSERVICEFILE=/secure/first-system-admin/pg_service.conf PSQL_HISTORY=/dev/null \
  psql -X -w --dbname='service=first-system-admin' \
  --file=/secure/first-system-admin/create.sql
```

初回は短い保守枠で行う。テーブルロックは通常の account 発行・管理者変更・所属更新の INSERT/UPDATE/DELETE とも競合するため、事前に進行中の管理操作が終わるのを待つ。待機中の既存操作があれば先に完了し、その後の状態を READ COMMITTED で検証する。同じ bootstrap 同士だけに効く advisory lock には依存しない。ロックタイムアウト・deadlock は全体失敗とし、自動再試行しない。

```sql
\set ON_ERROR_STOP on
\set ECHO none
\set VERBOSITY terse
BEGIN ISOLATION LEVEL READ COMMITTED;
SET LOCAL search_path = public, pg_temp;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
SET LOCAL idle_in_transaction_session_timeout = '60s';

CREATE TEMP TABLE bootstrap_input (
    database_name text, database_role text, staff_id text, clinic_id text,
    email text, password_hash text, approval_ref text, operator_ref text
) ON COMMIT DROP;
\copy pg_temp.bootstrap_input FROM '/secure/first-system-admin/input.csv' WITH (FORMAT csv)

-- Normal account writers take ROW EXCLUSIVE, which conflicts with this lock.
-- Lock memberships as well, including INSERTs that do not exist yet.
LOCK TABLE public.accounts, public.clinics, public.staffs,
    public.staff_clinic_assignments IN SHARE ROW EXCLUSIVE MODE;

DO $bootstrap$
DECLARE
    i record;
    new_account_id bigint;
BEGIN
    IF (SELECT count(*) FROM pg_temp.bootstrap_input) <> 1 THEN
        RAISE EXCEPTION 'bootstrap requires exactly one input row';
    END IF;
    SELECT * INTO STRICT i FROM pg_temp.bootstrap_input;
    IF i.database_name IS DISTINCT FROM current_database()
       OR i.database_role IS DISTINCT FROM session_user THEN
        RAISE EXCEPTION 'bootstrap target mismatch';
    END IF;
    IF NOT coalesce(i.staff_id ~ '^[1-9][0-9]{0,17}$', false)
       OR NOT coalesce(i.clinic_id ~ '^[1-9][0-9]{0,17}$', false)
       OR NOT coalesce(i.email ~ '^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$', false)
       OR NOT coalesce(length(i.email) <= 254, false)
       OR NOT coalesce(i.password_hash ~ '^\$2[ab]\$12\$[./A-Za-z0-9]{53}$', false)
       OR NOT coalesce(i.approval_ref ~ '^[A-Za-z0-9._:-]{1,100}$', false)
       OR NOT coalesce(i.operator_ref ~ '^[A-Za-z0-9._:-]{1,100}$', false) THEN
        RAISE EXCEPTION 'bootstrap input invalid';
    END IF;
    -- Intentionally includes inactive and soft-deleted administrators/accounts.
    IF EXISTS (SELECT 1 FROM public.accounts WHERE is_system_admin)
       OR EXISTS (SELECT 1 FROM public.accounts WHERE lower(email) = lower(i.email)) THEN
        RAISE EXCEPTION 'bootstrap account conflict; do not retry';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM public.staffs s
        JOIN public.clinics c ON c.id = s.clinic_id AND c.is_active
        JOIN public.staff_clinic_assignments a
          ON a.staff_id = s.id AND a.clinic_id = c.id
         AND a.is_main AND a.deleted_at IS NULL
        WHERE s.id = i.staff_id::bigint AND s.clinic_id = i.clinic_id::bigint
          AND s.is_active AND s.deleted_at IS NULL AND s.account_id IS NULL
    ) THEN
        RAISE EXCEPTION 'bootstrap staff or active main membership unavailable';
    END IF;

    INSERT INTO public.accounts (email, password_hash, is_active, is_system_admin)
    VALUES (i.email, i.password_hash, true, true)
    RETURNING id INTO new_account_id;
    UPDATE public.staffs SET account_id = new_account_id, updated_at = now()
    WHERE id = i.staff_id::bigint AND account_id IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'bootstrap staff attachment failed';
    END IF;
    INSERT INTO public.audit_logs (
        clinic_id, actor_id, actor_type, action, resource, resource_id,
        old_value, new_value, metadata
    ) VALUES (
        i.clinic_id::bigint, NULL, 'system', 'account.bootstrap.create',
        'account', new_account_id, NULL,
        jsonb_build_object('staff_id', i.staff_id::bigint,
                           'is_active', true, 'is_system_admin', true),
        jsonb_build_object('approval_ref', i.approval_ref,
                           'operator_ref', i.operator_ref,
                           'procedure', 'first-system-admin-v1')
    );
END;
$bootstrap$;
COMMIT;

-- Only a committed account with its committed audit is a success receipt.
SELECT a.id AS account_id, s.id AS staff_id, s.clinic_id, l.id AS audit_id,
       l.metadata ->> 'approval_ref' AS approval_ref,
       l.metadata ->> 'operator_ref' AS operator_ref
FROM public.accounts a
JOIN public.staffs s ON s.account_id = a.id
JOIN public.audit_logs l ON l.resource = 'account' AND l.resource_id = a.id
WHERE a.is_system_admin AND a.is_active AND a.deleted_at IS NULL
  AND l.action = 'account.bootstrap.create';
```

`actor_type=system` / `actor_id=NULL` は、まだ認証済み staff actor がいない初回操作を表す。実行者は `metadata.operator_ref` と承認記録で追跡する。監査 INSERT が失敗すれば account 作成・staff への付与も rollback する。既存 staff ID、主所属、他の所属、権限グループは維持する。パスワード・hash・email は監査へ保存しない。

### 3. 結果確認と失敗時の扱い

- psql が終了コード 0、`COMMIT` 成功、receipt 1 行であることを確認し、account/staff/clinic/audit ID を承認済み運用記録に残す。SQL 実行のみではログイン確認は完了しない。
- 途中エラーは `ON_ERROR_STOP` で接続を終了し、未 commit の transaction を rollback する。sequence の欠番は許容する。通信切断や COMMIT 応答不明の場合は**再実行せず**、同じ service への読み取り専用接続で末尾の receipt SELECT と承認参照を照合する。成功済みなら重複発行しない。判別できなければ運用上の未確認として停止する。
- receipt のない失敗でも再実行は原因・残存 account・監査を確認してから人間が判断する。成功後の取り消しに account/staff/audit の DELETE を使わない。必要な無効化・復旧は別の承認済み操作とする。
- 本人が通常ログインし、対象 clinic を選択できることを確認する。安全な配布経路で本人が秘密を保持した後、作業用の平文・hash・CSV は秘密管理の廃棄手順で処分する。監査記録を削除しない。

## 実行状態

手順はリポジトリの schema と認証の所属参照経路に静的照合した。本番への実付与・実環境での存在確認・SQL の実 DB 実行・本人ログイン確認は **未実施**。対象環境・担当者・資格情報の受領と実行は別承認が必要である。

# 権限まわり現状ダンプ（ログイン / 権限グループ / 所属医院 / 医院横断）

> 作業用の **リバースエンジニアリング結果**。正規設計書は [`docs/architecture/auth.md`](docs/architecture/auth.md)。
> 対象は **2026-09-08 時点の作業ツリー**。資格情報の値は複製しない。
>
> **この文書の境界:** 認可の機械（誰が入れるか、grant の計算、医院スコープ、スタッフ/権限グループ/医院マスタの横断ルール）を閉じる。
> 37 リソースすべての臨床 API 1本ずつの対応表は、アプリ全体の再列挙になるため **ここには置かない**（§13）。

---

## 0. 読み方

| 層 | 何が効くか |
|:---|:---|
| **システム管理者** | `accounts.is_system_admin = TRUE`。RBAC をバイパスし、全リソース全アクションを許可したものとして扱う。選択できる医院は **現存かつ `is_active=true` の医院だけ**。所属が空でも、active clinic が1件以上あればログインできる。 |
| **医院スコープのスタッフ** | リクエストごとに account / staff / 所属を再解決し、**選択中医院**の権限グループ grant で判定する。 |

「院長」「クリニック管理者」は独立フラグではない。院内で執行グループなどを付けたスタッフである。

**`is_system_admin` を API で立てる経路はない。** スタッフ作成は常に `FALSE`。スタッフ PATCH の `IsSystemAdmin` は **呼び出し元のバイパス用**であり、対象を昇格するフィールドではない。立てられるのは `seedlogin` のオペレータ upsert（環境変数）だけ。

**STG と作業ツリーの差**

- スタッフ **プロフィール PATCH** の医院横断ルールは、この作業ツリーで修正済み（§8.1）。
- **STG に未デプロイ**のあいだ、複数医院所属スタッフのプロフィール更新は、非システム管理者だと 403（トースト「更新の権限がありません。」）のまま。
- 所属医院 PUT はもともと「対象の全所属医院で `master-staff:edit`」。STG でもその契約。
- 複数医院所属スタッフの **削除は所属数 > 1 で Conflict**。システム管理者でもスタッフ DELETE では消せない（§7）。

---

## 1. データモデル

```
accounts
  email UNIQUE
  password_hash
  is_system_admin
  is_active
  deleted_at

staffs
  clinic_id          … ホーム医院（主所属）。所属 PUT で更新される
  account_id         … ログインがある場合。INDEX のみで UNIQUE ではない
  occupation_id     … occupations（医院スコープ）。seedlogin は未設定
  staff_type         … doctor / nurse
  is_active
  deleted_at

staff_clinic_assignments
  staff_id + clinic_id UNIQUE
  is_main            … ログイン直後の既定医院
  deleted_at

permission_groups          … clinic_id 付き。医院内 name unique
  name, description, color, is_active, sort_order

permission_group_rules
  group_id × resource × can_view / can_create / can_edit / can_delete

staff_permission_groups
  staff_id + group_id（clinic_id 列なし）
  医院分離は JOIN 先の permission_groups.clinic_id
```

アプリは `FindByAccountID` で **最初の1スタッフ**を取る。DB は同一 `account_id` の複数スタッフを拒まない。運用上は 1:1 前提。

実効権限は `FindAllEffectivePermissionsByStaffID(staffID, clinicID)`。

- `permission_groups.is_active = true` かつ `deleted_at IS NULL`
- `pg.clinic_id = 選択医院`
- ルールを resource ごとに `bool_or`（OR）
- グループ 0 件、またはルール無しリソース → **そのアクションは deny**
- 無効化したグループ（`is_active=false`）は実効権限に入らない。割当 GET は `deleted_at` だけ見るので、画面上は残って見えることがある

職種名だけでは権限は付かない。

---

## 2. ログイン解決

### 2.1 認証が通る条件

`AuthenticateUser`:

1. email で account がある
2. account が `is_active` かつ未削除
3. パスワードが bcrypt 一致、**または** 非本番かつカタログ email に対する共通デモパスワード（`AcceptSharedPassword`。定数時間比較。オペレータ・カタログ外・production/空/未知 env は不可）
4. その account に紐づく staff がある
5. staff が `is_active` かつ未削除

失敗は一律「メールアドレスまたはパスワードが正しくありません」。アカウント無しスタッフはログインできない。

ログイン JSON は password `min=8,max=72`。CSRF 用 `X-Requested-With` は login/refresh/logout と保護 route の非 GET に必要。無い 403 を権限不足と混同しない。`/auth/forgot-password` と `/auth/reset-password` は CSRF 対象外（専用 rate limit）。

### 2.2 医院スコープ

`ResolveClinicInfo`: 所属の `is_main` を既定医院にする。無ければ所属配列の先頭。
所属 0 件（非 admin）→ 403 `no clinic access is available`。

システム管理者は所属が空でも、active clinic の先頭（または残っている preferred）を使う。

以降のリクエストは JWT の `clinic_ids` スナップショットを最終権限に使わない。`current_access_service` が毎回再解決する。無効化・所属解除後の stale token は fail-closed。lookup 障害は 503（JWT を継続権限にしない）。

一般スタッフの `X-Clinic-ID` は、**いま有効な所属医院のいずれか**。システム管理者も任意 ID は選べず、active clinic に限定。

### 2.3 `GET /api/v1/me`（ログイン応答の `user` と同型）

| フィールド | 内容 |
|:---|:---|
| `id` | staff id |
| `email` / `display_name` | account / staff |
| `is_system_admin` | account フラグ |
| `occupation` | `staff.Occupation.Name`。`occupation_id` 未設定なら omit |
| `main_clinic_id` | 主医院 |
| `clinic` | **選択中医院**の詳細（帳票設定など） |
| `clinics[]` | 切替候補 `{clinic_id, clinic_name, is_main}` |
| `permissions` | 選択中医院の実効マップ。admin は全リソース V/C/E/D すべて true。clinic 解決失敗時は **空 map = deny** |

フロントのボタンは `/me.permissions`（選択医院のみ）。staleTime 5 分。他院の執行はここには出ない。

### 2.4 医院切替（フロント）

`AuthProvider.switchClinic`:

1. `user.clinics` に無い ID は無視
2. localStorage に clinic id を書く（axios が `X-Clinic-ID` に載せる）
3. React Query を `clear`
4. フルリロード → `/me` を新医院で取り直す

ストレージ書込失敗時はリロードしない。

### 2.5 自分のパスワード

`PUT /api/v1/users/me/password`。スタッフマスタ権限は不要。本人認証済みなら可。

---

## 3. 医院（002_master / seedlogin カタログ）

| ID | 名称 |
|:--:|:---|
| 1 | 八王子病院 |
| 2 | 城東センター病院 |
| 3 | ノア動物病院　敷島病院 |
| 4 | ノア動物病院　Hako bu neco |

### 3.1 医院 API と権限

`/api/v1/clinics` はすべて `hospital-settings` の view/create/edit/delete。

| 操作 | RBAC | 追加ゲート |
|:---|:---|:---|
| GET 一覧（デフォルト） | view | 自分の所属医院だけ |
| GET 一覧 `?scope=all` | view | **全医院**（所属外含む）。医院マスタ画面が使う。system_admin 不要。執行も一般も seed では view あり |
| GET `:id` | view | 非 admin は **選択中医院と ID が一致**するときだけ。他院は 403 `cannot access other clinics` |
| PATCH `:id` | edit | 同上。執行は切替後に自選択医院の設定を編集できる |
| POST 作成 | create | **さらに `is_system_admin`**。執行の seed は create 無し |
| DELETE | delete | **さらに `is_system_admin`**。執行の seed は delete 無し |

新規医院作成時、同一トランザクションで **執行 + 一般** と `defaultPermissionRuleTable` のルールを作る。閲覧専用は作らない。

動物種マスタの mutation は `is_system_admin` のみ（執行の `master-animal-species` は閲覧だけ、かつ handler が admin を要求）。

---

## 4. 権限グループ

### 4.1 Seed（`002_master/accounts/`）

| ID | clinic | 名前 | color | sort |
|:--:|:------:|:---|:---|:--:|
| 1 | 1 | 執行 | #6366F1 | 1 |
| 2 | 1 | 一般 | #10B981 | 2 |
| 9 | 1 | 閲覧専用（**医院1のみ**） | #9CA3AF | 3 |
| 3 | 2 | 執行 | #6366F1 | 1 |
| 4 | 2 | 一般 | #10B981 | 2 |
| 5 | 3 | 執行 | #6366F1 | 1 |
| 6 | 3 | 一般 | #10B981 | 2 |
| 7 | 4 | 執行 | #6366F1 | 1 |
| 8 | 4 | 一般 | #10B981 | 2 |

執行 1/3/5/7 のルールは同一。一般 2/4/6/8 も同一。医院内名称は unique（衝突は 409 `permission_group_name_conflict`）。

### 4.2 グループマスタ CRUD

画面 `/settings/permission-groups`。API はすべて `master-permission`。一般は seed で全拒否なので一覧も不可。

| 操作 | 権限 | 追加ルール |
|:---|:---|:---|
| 一覧・詳細 | view | **選択医院のグループだけ** |
| 作成 | create | |
| メタデータ / ルール PUT / 並び替え | edit | ルールはグループ単位で全置換 |
| 削除 | delete | **割当スタッフが1人でもいれば Conflict**「スタッフに割り当てられているため削除できません」 |

他院のグループ ID を今の医院の PUT に混ぜると、clinic-scoped 検証で失敗する。

### 4.3 リソース（37）と seed の穴

`model.AllResources` は 37 キー。seed CSV は 1 グループあたり 36 行。
**`examination-unconfirm` は CSV に行が無い** → deny。新規医院 default では明示 4 false。

執行でも既定で全拒否: `examination-unconfirm` / `identity-links` / `checkup-package-import`。

`master-animal-species` は執行・一般とも閲覧のみ。mutation は system admin。

### 4.4 マトリックス（V/C/E/D）

凡例: `○` = 許可、`—` = 拒否。

| リソース | 執行 | 一般 | 閲覧専用 |
|:---|:---|:---|:---|
| `reception` | ○○○○ | ○——— | ○——— |
| `owners` | ○○○○ | ○○○— | ○——— |
| `reservations` | ○○○○ | ○○○— | ○——— |
| `medical-records` | ○○○○ | ○○○— | ○——— |
| `hospitalization` | ○○○○ | ○○○— | ○——— |
| `trimming` | ○○○○ | ○○○— | ○——— |
| `examinations` | ○○○○ | ○○○— | ○——— |
| `examination-unconfirm` | ————（CSV 行なし） | 同左 | 同左 |
| `accounting` | ○○○○ | ○——— | ○——— |
| `vaccinations` | ○○○○ | ○○○— | ○——— |
| `checkups` | ○○○○ | ○——— | ○——— |
| `inventory` | ○○○○ | ○——— | ○——— |
| `estimates` | ○○○○ | ○——— | ○——— |
| `shifts` | ○○○○ | ○○○— | ○——— |
| `hospital-settings` | ○—○— | ○——— | ○——— |
| `master-animal-species` | ○——— | ○——— | ○——— |
| `master-medical` | ○○○○ | ○——— | ○——— |
| `master-reservation-type` | ○○○○ | ○——— | ○——— |
| `master-hospitalization` | ○○○○ | ○——— | ○——— |
| `master-trimming` | ○○○○ | ○——— | ○——— |
| `master-permission` | ○○○○ | ———— | ———— |
| `master-staff` | ○○○○ | ○——— | ○——— |
| `master-insurance` | ○○○○ | ○——— | ○——— |
| `master-merchandise` | ○○○○ | ○——— | ○——— |
| `discount` | ○○○○ | ———— | ○——— |
| `accounting-cancel` | ○—○— | ○——— | ○——— |
| `accounting-post-close-edit` | ○—○— | ○——— | ○——— |
| `closing-settings` | ○○○○ | ○——— | ○——— |
| `cash-register-close` | ○—○— | ○——— | ○——— |
| `accounting-reports` | ○—○— | ○——— | ○——— |
| `master-payment-method` | ○—○— | ○——— | ○——— |
| `lstep-csv-import` | ○—○— | ○——— | ○——— |
| `lstep-analytics` | ○—○— | ○——— | ○——— |
| `manual-edit` | ○—○— | ○——— | ○——— |
| `lab-import` | ○—○— | ○——— | ○——— |
| `identity-links` | ———— | ———— | ———— |
| `checkup-package-import` | ———— | ———— | ———— |

スタッフ運用で効く差:

- **執行**: `master-staff` 全操作、`master-permission` 全操作。
- **一般**: `master-staff` は閲覧のみ。`master-permission` 全拒否。
- **閲覧専用**: ほぼ閲覧。`discount` だけ一般より広く閲覧がある。

新規医院 default は執行/一般と同じ（`examination-unconfirm` は明示 false）。

### 4.5 HTTP での grant の効き方

`RequirePermission`:

| メソッド | 判定 |
|:---|:---|
| 書き込み | **選択中医院**にその resource/action |
| GET/HEAD | 選択中医院に無くても、**所属する別医院**に grant があれば一旦通す。一覧・詳細は `FilterClinicIDsForPermission` 等で絞る |

フロントは `/me` の現医院だけなので、BE の GET 緩和より狭い。他院に執行があっても、今の医院が一般ならスタッフ編集 UI は出ない。

### 4.6 医院 ID 集合の二つの関数

| 関数 | 意味 |
|:---|:---|
| `AuthorizeClinicIDs` | 要求 ID が trusted 所属（admin は active clinic）の **部分集合**。1件でも外なら 403。部分成功なし |
| `AuthorizeClinicIDsForPermission` | 上に加え、**要求した全医院**で resource/action。admin は所属チェックのあと grant をスキップ |
| `FilterClinicIDsForPermission` | 所属集合から grant がある医院だけ残す。checker 欠落・結果 0 件は 403。admin は grant 検査せず compact 集合を返す（空は 403） |

飼主作成の登録先医院など、**宛先医院**の write は `AuthorizeClinicIDsForPermission`（その医院の create が必要）。スタッフ横断更新は Filter（認可集合を作り、対象所属が包含されるか見る）。

---

## 5. ログインアカウント

### 5.1 いつ作られるか

`internal/seedlogin`。CSV ではない。migrate 時 upsert。
`APP_ENV` ∈ `development` / `local` / `dev` / `test` / `staging`。
**production / 空 / 未知値はスキップ。**

カタログ email の `ON CONFLICT` は password_hash / is_active / deleted_at を直す。**`is_system_admin` は更新しない**（INSERT 時 FALSE）。

### 5.2 ID / email

- スタッフ ID = `clinicID * 10_000_000 + suffix`
- email = `stg-staff-{staffID}@example.test`
- `occupation_id` なし。職種ラベル → `staff_type`（獣医師=doctor、他=nurse）
- 林の `is_main` はホーム医院 1 のみ。他 3 医院は `is_main=false`
- 旧医院別林クローン（`20000021` 等）は account を inactive+deleted。残る執行ログインは 1 人

seedlogin は `staff_permission_groups` を **staff 単位で全削除**してから、所属医院ごとに名前一致グループを付け直す。手で付けた閲覧専用も migrate で消える。

### 5.3 林 文明（唯一の複数医院執行）

| 項目 | 値 |
|:---|:---|
| staff id | `10000021` |
| email | `stg-staff-10000021@example.test` |
| ホーム / is_main | 1 八王子 |
| 所属 | 1, 2, 3, 4 |
| 権限 | 各院の執行（group 1, 3, 5, 7） |
| `is_system_admin` | FALSE |

### 5.4 医院ローカル（一般・単一所属）

各医院に同じ 9 人。所属はその医院だけ。権限はホーム医院の一般のみ。

| suffix | 氏名 | 職種ラベル |
|:------:|:---|:---|
| 3 | 高橋 純子 | 獣医師 |
| 7 | 鈴木 諒平 | 獣医師 |
| 8 | 加藤 茉里 | 獣医師 |
| 25 | チャン ハン | 看護師 |
| 31 | 近喰 千瞳 | 動物看護師 |
| 34 | 川野 称希 | 動物看護師 |
| 5 | 冨田 美佳 | VT |
| 6 | 井冨 和美 | VT |
| 9 | 原 梨吏華 | スタッフ |

| 医院 | 高橋 | 鈴木 | 加藤 | チャン | 近喰 | 川野 | 冨田 | 井冨 | 原 |
|:---|:---|:---|:---|:---|:---|:---|:---|:---|:---|
| 1 | 10000003 | 10000007 | 10000008 | 10000025 | 10000031 | 10000034 | 10000005 | 10000006 | 10000009 |
| 2 | 20000003 | 20000007 | 20000008 | 20000025 | 20000031 | 20000034 | 20000005 | 20000006 | 20000009 |
| 3 | 30000003 | 30000007 | 30000008 | 30000025 | 30000031 | 30000034 | 30000005 | 30000006 | 30000009 |
| 4 | 40000003 | 40000007 | 40000008 | 40000025 | 40000031 | 40000034 | 40000005 | 40000006 | 40000009 |

email は `stg-staff-{id}@example.test`。合計 **37**。同名の高橋は医院ごとに別スタッフ。

### 5.5 オペレータ（任意・git 外）

`SEEDLOGIN_OPERATOR_EMAIL` / `NAME` / `PASSWORD` が揃ったときだけ。

- `is_system_admin = TRUE`
- カタログ 4 医院 + 各院執行
- カタログ email と衝突したら失敗
- パスワード 8 文字以上・英字と数字（値は git に置かない）
- 共通デモパスワードでは入れない

### 5.6 ログイン画面デモ一覧

`SHOW_DEMO` = Vite DEV または Vercel preview。STG が `VERCEL_ENV=production` ならピッカー無し。アカウントは手入力可。
ピッカーはカタログ 37 件（4 医院すべて）。オペレータは出ない。

### 5.7 Cutover / 旧DB スタッフ

旧DB handoff の `staffs.csv` はログイン列を持たない。**account 無し = ログイン不可**。所属と権限グループは付けられる。seedlogin はカタログだけ upsert し、旧スタッフを消さない。

UI からスタッフ+アカウントを新規作成すると、所属は選択医院 1 件、**権限グループは空 = deny-all**。ログインできても何もできない。グループは別 PUT で付ける。

---

## 6. 所属医院

### 6.1 何が所属か

選択できる医院 = 有効な `staff_clinic_assignments`。
複数所属だけ医院切替が出る。カタログ一般 36 人は自院固定。

### 6.2 スタッフ一覧に誰が出るか

`GET /masters/staffs` は `staff_clinic_assignments` の **INNER JOIN**（`staffs.clinic_id` ホームではない）。
林は医院 2 を選択すると **医院2の一覧に出る**。
`deleted_at` のみ除外。`is_active=false` も一覧に残る。

詳細 `GET /staffs/:id`: 非 admin は `GetByIDInClinic`（その医院の所属が無ければ 404）。admin は `GetByID`。

### 6.3 所属の見え方

`GET /staffs/:id/clinics`

- ミドルウェア: `master-staff:view`（GET なので他院 view でも一旦通過しうる）
- 返却は対象の所属 ∩ 呼び出し元が `master-staff:view` を持つ医院
- admin は全所属

執行 1 医院だけの人が林を見ると、自分が view を持つ医院の所属だけ見える。

### 6.4 所属の書き換え

`PUT /staffs/:id/clinics`

1. 選択医院で `master-staff:edit`
2. 呼び出し元の `master-staff:edit` 医院へ Filter
3. 追加先が認可外 → 403 `cannot assign staff outside authorized clinics`
4. 既存所属のいずれかが認可外 → 403 `cannot replace staff assignments outside authorized clinics`  
   **他院の所属を残した部分更新は非管理者にはできない**
5. 予約・シフト等が残っている医院の所属削除は Conflict
6. 成功時、主医院 `staffs.clinic_id` も更新
7. システム管理者は 1–4 をバイパス（新規追加の active lock は残る）

フロント: 現医院の `master-staff:edit` が無いと「所属医院を変更する権限がありません」。

---

## 7. 権限グループ割当とスタッフ API

### 7.1 割当（医院ごと独立）

`GET/PUT /staffs/:id/permission-groups`

- **`X-Clinic-ID` のグループだけ**。他院リンクは消さない
- PUT 空配列 = その医院のグループ全解除（その医院は deny-all）。他院は残る
- 対象がその医院に所属していなければ失敗（assignment を FOR UPDATE）
- アカウントの有無は見ない（cutover スタッフにも付けられる）
- PUT 必須: 選択医院の `master-staff:edit` **かつ** `master-permission:edit`

林を城東で開いてグループを変えるには、**城東**で両方の edit。八王子の執行だけでは足りない。

### 7.2 スタッフ API 権限表

| 操作 | 権限 |
|:---|:---|
| 一覧・詳細・グループ GET・所属 GET・職種一覧 | `master-staff:view` |
| 作成・職種作成 | `master-staff:create` |
| プロフィール更新・並び替え・所属 PUT・予約タイプ制限 PUT・職種更新 | `master-staff:edit` |
| グループ PUT | `master-staff:edit` かつ `master-permission:edit` |
| 削除・職種削除 | `master-staff:delete` |

追加ガード:

- 自己削除 400「自分自身を削除することはできません」
- 自己無効化 403「自分自身を無効化することはできません」
- 最後のシステム管理者の削除/無効化は拒否
- **所属が 2 医院以上のスタッフは削除できない**（Conflict「複数のクリニックに所属しているスタッフは削除できません」）。**system admin でも同じ**。林はスタッフ DELETE では消せない
- パスワード置き換えには選択医院の `master-permission:edit` が追加で必要
- email 重複は AlreadyExists
- 職種は医院スコープ（他院の `occupation_id` は付けられない）
- 並び替え・予約タイプ制限は **選択医院** のマスタ

職種マスタも `master-staff`。**002_master に occupations CSV は無い。** STG の職種画面が空なのはデータ未投入。

### 7.3 フロント（スタッフ設定）

`StaffSettings` は `/me` の現医院。

| UI | 現医院 grant |
|:---|:---|
| 保存 | create / edit |
| 削除 | delete |
| グループ保存 | staff edit かつ permission edit |
| 所属保存 | staff edit |

設定ルートは `RequirePermission`（既定 action=view）。Sidebar も view。`isSystemAdmin` は FE でも常時 true。
最終判定は API。英語 403 は「{操作}の権限がありません。」。日本語 403（所属するすべての医院で…）はそのまま出す。

---

## 8. 医院横断：別の所属医院に対する権限

### 8.1 スタッフプロフィール PATCH

1 スタッフ行を全所属が共有する。

作業ツリー（`authorizeGlobalStaffUpdate` + Filter `master-staff:edit`）:

1. 選択医院で `master-staff:edit`
2. 呼び出し元所属のうち edit がある医院が認可集合
3. 対象の **有効所属すべて**が認可集合に入らなければ 403「所属するすべての医院でスタッフ編集権限が必要です」
4. 対象が選択医院に所属していない（非 admin）→ 404
5. システム管理者は 2–4 をバイパス

複数医院所属だから admin 専用、という分岐は **この作業ツリーでは削除済み**。

| 操作者 | 対象 | 作業ツリー |
|:---|:---|:---|
| 林（4医院執行） | 林自身 | 可 |
| 林 | 八王子の高橋 | 可 |
| 八王子だけ執行 | 林 | 不可 |
| 一般 | 誰でも | ミドルウェア 403 |
| オペレータ | 誰でも | 可 |

**STG 未デプロイ:** 所属 ≥2 なら非 admin は常に 403。林は自分を更新できない。

### 8.2 所属 PUT

対象が今所属している医院すべてで `master-staff:edit` が要る。欠けていると置き換え不可。

### 8.3 グループ PUT

医院ローカル。選択医院の edit 2種。林は切替すれば各院独立に変えられる。

### 8.4 選択医院 vs 他院 grant

| やりたいこと | 他院の執行で足りるか |
|:---|:---|
| 城東選択のままスタッフ更新 | 足りない。城東で staff:edit |
| 城東選択のままスタッフ一覧 GET | 八王子に view があれば GET は通ることがある。返却は医院2所属者 |
| 林として切替 → 各院で編集 | 各院執行なので可 |
| 高橋（八王子一般）が城東を選ぶ | 所属無し。`X-Clinic-ID` 拒否 |
| 飼主を他所属医院へ作成 | 宛先医院で `owners:create` が必要（林は可、一般の単一所属は自院のみ） |
| 医院設定を他院へ PATCH | 非 admin は選択医院と ID 一致が必要。切替してから |

### 8.5 臨床データの `clinic_ids`

query/body の医院集合は trusted 所属の部分集合。admin でも inactive / 集合外は拒否。
`identity-links` は既定付与なし。

---

## 9. 運用で起きること

### 9.1 林 文明

- 4 医院を切替できる。各院執行なのでスタッフ編集ボタンは出る
- 医院2のスタッフ一覧に自分が出る
- **STG:** 自分のプロフィール保存 → 403。所属 PUT は全医院 edit があるので通る想定
- **この作業ツリーをデプロイ後:** プロフィール保存も通る
- 自分をスタッフ DELETE できない（複数所属 Conflict）
- 職種マスタは空。ログイン画面の「獣医師」はハードコード

### 9.2 高橋 純子（各医院の一般）

- 自院のみ
- スタッフ一覧は見られる。編集・作成・削除・グループ・所属変更は不可
- 医院マスタ `scope=all` は `hospital-settings:view` があるので、その画面を開けば他院名も見える
- 他院の林・他院の高橋は、スタッフ一覧 JOIN では出ない

### 9.3 執行を 1 医院だけ付けた実スタッフ

- 自院の単一所属スタッフは編集可
- 林のプロフィール/所属は不可
- 切替先が一般なら、その医院ではスタッフ編集 UI も出ない
- 新規スタッフを作るとグループ空。別途執行がグループを付けないとログインしても deny-all

### 9.4 オペレータ

- RBAC 無視。active clinic ならスタッフ更新・所属置換・医院作成/削除が可能
- 林の削除（複数所属）はスタッフ API では不可のまま
- デモ共通パスワードでは入れない

---

## 10. 認可以外で 403 になるもの（混同防止）

- `X-Requested-With` 欠落（CSRF）
- `X-Clinic-ID` が所属外 / inactive
- Filter/Authorize の checker 欠落（実装バグ時 fail-closed）
- ログイン後に所属 0 件

---

## 11. 関連ファイル

| 内容 | 場所 |
|:---|:---|
| 認可設計 | `docs/architecture/auth.md` |
| 権限グループ画面 | `docs/spec/screens/settings/master-permission-group.md` |
| seed | `backend/migrations/seeds/002_master/accounts/permission_groups.csv` / `permission_group_rules.csv` |
| 新規医院デフォルト | `backend/internal/clinic/clinic_service.go` |
| デモログイン | `backend/internal/seedlogin/` |
| 実効権限 SQL | `backend/internal/auth/permission_group_repository.go` |
| HTTP RBAC | `backend/internal/auth/http_permission.go` |
| 医院集合 | `backend/internal/httpapi/clinic_permission.go` `context.go` |
| リクエスト権限再解決 | `backend/internal/auth/current_access_service.go` |
| スタッフ横断更新 | `backend/internal/staff/staff_service_core.go` |
| 所属置換 | `backend/internal/staff/staff_clinic_assignment_service.go` |
| 医院作成の admin ゲート | `backend/internal/clinic/clinic_handler.go` |
| FE 切替 / hasPermission | `frontend/src/features/auth/components/AuthProvider.tsx` |
| 設定ルートガード | `frontend/src/app/routes/settings-routes.tsx` |
| スタッフ画面 | `frontend/src/features/master/routes/StaffSettings.tsx` |

---

## 12. 未デプロイ・未決定

- [ ] スタッフ複数所属のプロフィール PATCH 緩和を STG に載せる
- [ ] 職種マスタを seed するか手入力のままにするか
- [ ] 閲覧専用を医院 2–4 にも seed するか
- [ ] `examination-unconfirm` / `identity-links` / `checkup-package-import` を執行へ付けるか
- [ ] 複数所属スタッフを削除する運用経路（所属を1医院にしてから DELETE、など）

---

## 13. このダンプに入れてないもの

意図的に省略した範囲。権限「機械」ではなくアプリ機能カタログになるため。

- 受付・カルテ・会計など、37 リソースに紐づく **全エンドポイント一覧**
- dual-token の寿命・refresh 再利用検知の詳細（`auth.md` §4.1）
- 資格情報監査のトランザクション契約（`auth.md` §4.3）
- LIFF / 飼主向け公開 API（staff RBAC の外）
- STG 実 DB の実測（seed と手編集の差分は未知）

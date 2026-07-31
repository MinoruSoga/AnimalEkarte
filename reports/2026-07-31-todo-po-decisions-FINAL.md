# PO 最終決裁 — TASK-002 / 003 / 012 / 013 / 014 / 021

> **決裁日**: 2026-07-31
> **Unit**: `TODO-MD-PO-DECISIONS-GPT56-20260731`
> **決裁者**: PO (gpt-5.6-sol session 2026-07-31)
> **権限の前提**: 人間の個人名は入力されていないため、ユーザーから本セッションへ委任された PO 権限を上記名義で記録する。
> **Runtime note**: 親セッションの実モデル・reasoning effort は CLI から確認できず unverified。read-only probe / independent review は `gpt-5.6-sol`・`max` を明示指定した。決裁者名は runtime attestation ではなく、本 unit の委任 PO 名義である。
> **効力**: 本書の六件は推奨ではなく binding decision。再度の PO 確認なしに、下記 follow-up の実装計画へ進める。
> **非実施**: 本 unit では製品コード、migration、seed、API、FE を変更しない。

## 決裁サマリー

| TASK | Chosen | Binding disposition | Follow-up |
|:---|:---|:---|:---|
| TASK-002 | **B** | 編集フォームでの治療プラン unlock は WONTFIX。登録時の見積明細 snapshot として固定する | UI/spec honesty と親削除導線の follow-up |
| TASK-003 | **B** | 入院フォーム一括割引は永続化しない | 無効な入力欄を削除し、概算表示へ整理する FE follow-up |
| TASK-012 | **B** | `approved` / `rejected` は不可逆。unlock しない | High: 後継見積の安全な新規作成経路 + S07 docs |
| TASK-013 | **B** | レジ締めは append-only。reverse は製品に持たない | High: persistence hardening + 誤った manual の修正 |
| TASK-014 | **A** | 予約済み `system_key` は immutable・編集 UI 非公開 | system row の update/deactivate/delete guard + V04 docs |
| TASK-021 | **staged B → A** | capabilities を唯一の SoT とし、exclusion は期限付き互換 facade の後に撤去 | 段階実装。`available-staffs` は WONTFILE |

## 共通判断原則

- 責任者と業務目的を明示し、画面上の要望だけを要件にしない（`docs/product-philosophy.md:32-34`）。
- 同じ business fact の二重入力・二重管理を作らない（同 `:56-70`, `:155-162`）。
- 臨床・会計の確定状態は confirmation dialog ではなく、不可逆 lock・append-only・監査で守る。
- 現在の実装や運用経路が証明されていない場合、「利用可能」とは決裁しない。follow-up 完了までは fail-closed とする。

## Binding decisions

### TASK-002 — Decision

- **Chosen**: **B — 入院編集フォームの治療プランは恒久 read-only**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 臨床運用 PO
- **Purpose**: 入院登録時の治療プランを、登録時点の費用・予定明細 snapshot として保持し、進行中の投薬・給餌・処置を入院詳細の `care-plan-items` と二重管理しない。
- **Binding policy**:
  - 新規登録時に作成した treatment-plan 行は、本フォームの編集では変更・追加・削除しない。
  - 継続診療の write owner は入院詳細の care plan とする。care plan と treatment plan を同一データであるかのように案内しない。
  - treatment plan が存在する入院を、行削除によって親ごと消せるようにはしない。確定済み臨床・費用 snapshot の保持を優先する。
- **Risks accepted**:
  - 登録後に treatment-plan 行の誤りを本フォームから訂正できない。
  - 現在は treatment plan がある親入院の削除を BE が拒否する一方、FE に親削除ボタンがあり、操作時に Conflict となり得る。
  - BE に PATCH/DELETE route が存在しても、退院・請求との lock 境界が未定義であるため、route の存在を unlock 要件とは扱わない。
- **Follow-up**:
  - **Unlock 自体は WONTFIX close**。
  - 実装 TASK: treatment plan と care plan の名称・説明を明確に分離し、子明細がある場合の親削除を非表示または物理ブロックとして事前表示する。
  - hospitalization 配下の treatment-plan PATCH/DELETE は consumer inventory 後に route を撤去する。撤去までの間も service が mutation を拒否し、API 直叩きで snapshot を変更できないことを回帰 test で固定する。medical-record 配下の別 contract は本決裁の対象外。
  - 受け入れ: 編集時 read-only、care plan との用語分離、親削除が事後 Conflict だけに依存しないこと、hospitalization treatment-plan の PATCH/DELETE が UI/API の双方から不可能であること。
- **Rationale (vs product philosophy)**: 同じ診療予定を二つの write surface で持つ案 A/C を追加せず、登録 snapshot と継続 care の責務を分離する。
- **Overrides packet recommendation?**: **No, B を採用**。ただし packet の「detail/care 画面で同じ treatment plan を更新できる」と読める表現は採用しない。両者は別の business fact である。
- **Evidence**: `docs/spec/screens/09-hospitalization-form.md:27-36`; `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:252-280`; `backend/internal/medicalrecord/routes.go:343-346`。

### TASK-003 — Decision

- **Chosen**: **B — 一括割引は永続化しない**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 会計運用 PO
- **Purpose**: 割引の会計 fact を見積・請求の write owner に集約し、入院フォームに第二の割引 SoT を作らない。
- **Binding policy**:
  - hospitalization 親に一括割引フィールドを追加せず、create/update payload にも追加しない。
  - 行単位 discount と会計側 discount を正本とし、入院フォームの一括値を会計 fact として扱わない。
  - 現在の `%` / `円` 欄は常時 disabled・初期値 0 であり calculator ではないため、follow-up で削除する。小計・税・合計の read-only 概算と「保存されない」honesty は残す。
- **Risks accepted**:
  - 入院フォームだけでは一括割引を保存できない。
  - follow-up まで disabled 欄が保存可能な値に見える UX risk が残る。
- **Follow-up**:
  - **永続化は WONTFIX close**。
  - FE 実装 TASK: disabled の一括割引入力二つを削除し、read-only 概算サマリーと honesty copy に整理する。既存の行割引・見積・会計 contract は変更しない。
  - 受け入れ: 一括割引の request/schema 追加なし、保存されるように誤認する入力欄なし、概算の出所が明示されること。
- **Rationale (vs product philosophy)**: 不使用入力を削除し、割引の二重管理を禁止する。機能追加ではなく削除・単純化を選ぶ。
- **Overrides packet recommendation?**: **No, B を採用**。ただし「calculator UX」という説明は現行コードと一致しないため、dead control 削除へ具体化した。
- **Evidence**: `docs/spec/screens/09-hospitalization-form.md:34-36`; `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:270-280`; `frontend/src/features/hospitalization/components/HospitalizationCostSummary.tsx:49-77`。

### TASK-012 — Decision

- **Chosen**: **B — `approved` / `rejected` から draft へ戻さない**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 臨床・会計 PO
- **Purpose**: 飼主へ提示・判断済みの見積を改変せず、提示時点の臨床・金額・判断履歴を保持する。
- **Binding policy**:
  - `approved` と `rejected` は不可逆の terminal status。既存見積の edit/delete/unlock を禁止する。
  - 訂正時は原本を残し、別 ID・別番号の draft 後継見積を作る。原本への上書きや status 巻き戻しはしない。
  - 現在の製品には安全な後継見積経路が証明されていないため、follow-up 完了までは「新規作成で訂正できる」と運用案内しない。
- **Risks accepted**:
  - 原本は訂正不能で、後継作成には転記・確認コストがある。
  - 現状は `estimate_no` の atomic 採番が source 上で確認できず、clinic 単位 UNIQUE と衝突し得る。
  - カルテ内 UI は既存一件を PATCH するため、locked 見積から後継を作成する経路がない。
- **Follow-up**:
  - **High 実装 TASK — 後継見積による訂正**:
    1. clinic-scoped の `estimate_no` を並行作成でも重複しない方法で採番する。
    2. locked 原本から、理由必須・権限必須で別 draft を作成し、原本との supersedes 関係を一方向に保持する。
    3. 原本を変更せず、原本 ID・後継 ID・理由・実行者を audit に残す。後継作成・supersedes 関係・audit write は同一 transaction で行い、監査依存欠落または監査失敗時は全体を rollback する。
    4. カルテ確定状態を巻き戻さず、見積訂正 action だけを明示的に許可する。
    5. 同一 clinic の連続/並行二件、他 clinic 分離、locked 原本不変、カルテ内導線を統合 test で証明する。
  - **Docs TASK**: `docs/ops/testing/scenarios/S07-estimate-status-control.md:40` の「下書きへ戻す」を削除し、26§2.1 と terminal-status rule に統一する。follow-up 完了までは後継作成を未実装と明記する。
- **Rationale (vs product philosophy)**: unlock による履歴改変を避ける。後継は二重管理ではなく、原本と訂正版の関係・監査を明示する append-only versioning として実装する。
- **Overrides packet recommendation?**: **No, B を採用**。ただし packet の「新規見積で訂正可能」という未証明の運用前提は採用せず、High follow-up に格上げした。
- **Evidence**: `docs/spec/screens/26-estimate-detail.md:29-32`; `docs/ops/testing/scenarios/S07-estimate-status-control.md:29-40`; `backend/internal/billing/estimate_service.go:87-89`; `backend/internal/billing/estimate_repository.go:100-143`; `backend/internal/model/estimate.go:21-28`; `backend/migrations/001_init.sql:2168-2169`。

### TASK-013 — Decision

- **Chosen**: **B — cash-register close は append-only、reverse なし**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 会計・現金管理 PO
- **Purpose**: AM/PM/EMG の現金引継ぎと締め時点 snapshot を後から消さず、差異・訂正を監査可能にする。
- **Binding policy**:
  - close row の update/delete/soft-delete/reopen/reverse API を提供しない。
  - 誤った締めでも元 close row を保持する。基礎会計の訂正は理由必須の post-close audited edit で行い、close snapshot 自体を書き換えない。
  - 訂正差異は元 `close_id`、対象 billing ID、会計差額、現金移動額、理由、actor、実行日時を持つ専用の append-only adjustment record に一件ずつ記録する。memo だけの運用や close row 上書きは禁止する。
  - 月次の訂正後表示は「元 close snapshot + 元期間に帰属する adjustment」を分離表示する。後日実際に現金が動いた場合だけ、その現金移動額を実行日の次回 close に一度だけ反映し、会計差額だけの訂正を次期間の現金移動として二重計上しない。
  - 同一 `(clinic_id, close_date, period)` の close を隠して再作成する挙動を許可しない。
- **Risks accepted**:
  - 誤締め時に reverse でやり直せず、差異の説明・次回照合の運用コストが残る。
  - 現状の public route は append-only だが、DDL は `deleted_at` と soft-delete 後を除外する UNIQUE を持ち、DB 不変条件としては未完成。
  - 複数の manual に存在しない reverse 手順が残っている。
- **Follow-up**:
  - **High hardening TASK**: 既存データを監査したうえで、production 経路から close row の update/delete/soft-delete を不可能にし、行を隠して同期間を再 close できない DB/application invariant を実装する。同一 transaction で post-close edit と append-only adjustment を記録し、記録失敗時は訂正を rollback する。
  - **Docs TASK**: reverse を案内する stale manual を canonical な「reverse なし・元締め保持」の手順へ統一する。
  - 受け入れ: route/service/repository/DB の全境界で close 不変、同期間再作成不可、元期間の会計差額と実行日の現金移動が分離され各一度だけ集計されること、誤締め時の監査付き訂正手順が一箇所から辿れること。
- **Rationale (vs product philosophy)**: 会計 snapshot の取消・再作成という二重管理を許さず、確定 lock と監査で守る。
- **Overrides packet recommendation?**: **No, B を採用**。ただし append-only を API 表面だけで満たしたとは判定せず、persistence hardening を必須 follow-up とした。
- **Evidence**: `backend/internal/billing/routes.go:138-143`; `backend/internal/billing/cash_register_service.go:176-184`; `backend/migrations/001_init.sql:2126-2144`; `docs/ops/testing/scenarios/S09-closing-time-boundaries.md:30-34`。

### TASK-014 — Decision

- **Chosen**: **A — reserved `system_key` は immutable、編集 UI 非公開**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 会計マスタ PO
- **Purpose**: 表示名を医院ごとに変更できても、会計・締め・import が同じ支払カテゴリを安定識別できるようにする。
- **Binding policy**:
  - 予約値は `cash`, `credit_card`, `electronic_money`, `bank_transfer` の四つ。
  - reserved `system_key` は作成後に変更・付替え・再利用しない。master UI/API request に編集項目として公開しない。
  - `name` は表示専用ラベルとして変更可。ただし identity 判定に使用しない。
  - reserved key を持つ system row は deactivate/delete 不可。安全な選択肢制御が別 contract として設計されるまでは active のまま保持する。
  - cash-register snapshot 内の raw key は internal wire identifier であり、master UI の編集可能項目ではない。
- **Risks accepted**:
  - 医院が標準四分類の system row を削除・無効化できない。
  - 現状は UI 非露出である一方、DB immutability と system-row deactivate/delete guard は未実装。
- **Follow-up**:
  - 実装 TASK: request/service/repository/DB のいずれからも reserved key の変更を拒否し、system row の deactivate/delete を fail-closed にする。custom row の通常 CRUD は維持する。
  - Docs/test TASK: V04 の「要実測」を本決裁へ置換し、system row の rename 可、key 非公開、deactivate/delete 不可を FE/BE test で固定する。
- **Rationale (vs product philosophy)**: stable identity と表示ラベルを役割分離し、名称と会計分類の二重解釈を除く。power-user remap は追加しない。
- **Overrides packet recommendation?**: **No, A を採用**。未決だった active/delete を fail-closed 方針まで具体化した。
- **Evidence**: `backend/internal/model/payment_method_master.go:9-19`; `backend/internal/billing/payment_method_master_request.go:3-26`; `backend/internal/billing/payment_method_master_service.go:102-147`; `backend/migrations/001_init.sql:1922-1939,4353-4395`。

### TASK-021 — Decision

- **Chosen**: **staged B → A — capabilities-only SoT へ収束**
- **Owner**: PO (gpt-5.6-sol session 2026-07-31) / 業務責任ロール: 予約・スタッフ運用 PO
- **Purpose**: 予約区分ごとに対応可能なスタッフだけを一貫して提示・検証し、肯定形 capability と否定形 exclusion の二重管理を除去する。
- **Binding policy**:
  - 唯一の write/read decision SoT は `staff_reservation_capabilities`。
  - 現行 exclusion endpoint/field を直ちに破壊削除せず、Stage B で capabilities から導出する期限付き compatibility facade にする。
  - `staff_reservation_exclusions` への production write を Stage B 完了時にゼロにする。別 junction を維持する facade は B と認めない。
  - 院内候補は `active ∩ on-duty ∩ capable`。capability metadata の欠落・未取得時は候補可と推測せず fail-closed にする。
  - `GET /v1/reservations/available-staffs` は作らない（WONTFILE）。
- **Risks accepted**:
  - Stage B 中は deprecated shape を維持する互換コストがある。
  - 現状は exclusion/capability が別 table に書かれ、院内 FE も `excluded_courses` を読むため、本決裁は未実装。
  - 互換 PUT の inverse mapping は clinic-scoped active/non-deleted reservation types を universe として一意に定義する必要がある。
- **Follow-up**:
  - **Stage B**:
    1. 既存の overlap/矛盾を report し、capabilities を authoritative として再構築する。
    2. exclusion GET/response は capabilities から導出し、exclusion PUT/`excluded_type_ids` write は一つの atomic capability replacement へ変換する。
    3. `/clinics/:clinic_id/reservation-staffs` の既存 response に positive capability IDs を加える。新 endpoint は作らない。
    4. FE 候補を肯定形 capability filter へ変更し、missing/pending metadata を fail-closed にする。capability 更新後に候補 query を invalidate する。
    5. UI 名称を「対応可能」に統一し、clinic isolation、legacy contradiction、inverse mapping、UI/POST agreement を regression test で固定する。
    6. `available-staffs` route/client の追加を禁止する static gate を置く。
  - **Stage A**: consumer inventory と別途の破壊変更承認後、deprecated exclusion route/payload/model/table/seed/OpenAPI を migration 付きで削除する。既存 `reservation-staffs`, `on-duty-staffs`, `available-times` は維持する。
- **Rationale (vs product philosophy)**: 即時破壊による診療中の予約障害を避けつつ、二重管理を恒久許容しない。B は移行手段であり最終状態ではない。
- **Overrides packet recommendation?**: **No, staged B → A を採用**。現状を B 済みとは扱わず、dual persistence を Stage B の修正対象として明記した。
- **Evidence**: `docs/spec/reservation-to-record-flow.md:308-310,343-347,495-514`; `backend/internal/staff/staff_handler.go:313-383`; `backend/internal/staff/staff_service_permissions.go:76-128`; `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:95-114`。

## 決裁後の実装優先度

1. **High**: TASK-012 後継見積経路（採番・supersedes・audit）と S07 honesty。
2. **High**: TASK-013 close persistence hardening と reverse manual の撤去。
3. **Medium**: TASK-021 Stage B capabilities-only facade。Stage A は consumer inventory 後。
4. **Medium**: TASK-014 system-row guard と V04 固定。
5. **Medium**: TASK-002 / TASK-003 の UI/spec truthfulness cleanup。

六件とも PO 再決裁は不要。follow-up が実装済みであるとの主張は、それぞれの deterministic test / docs check が PASS してから行う。

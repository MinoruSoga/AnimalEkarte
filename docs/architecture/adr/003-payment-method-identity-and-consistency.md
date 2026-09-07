# ADR-003: 支払方法の安定識別と payment_methods 整合性の設計判断

**Status**: Decided — system_key / TRIGGER（method⇔system_key match）/ payments.clinic_id・複合 FK 実装済（#197・TASK-445・TASK-ADR003）。論点3（representativeMethod）#198 済。論点4（ENUM DROP）WONTFIX 維持。残差: レガシー `payment_method_id IS NULL` 許容・TRIGGER は soft-deleted master を拒否しない・確定後訂正経路は保存済 method/payment_method_id 組合せを再検証しない。
**Date**: 2026-06-25 / Updated: 2026-07-31
**Deciders**: Engineering（論点2/3 決裁・論点1 実装反映）。
**Drafted by**: Engineering（#185 follow-up）

## Context

#128（`d9bcd387` マージ済）で `payment_method` ENUM ⇔ `payment_method_id` 二重管理の中核バグ
（アプリ書込みの非現金会計が `payment_method_id=NULL` で保存され、レジ締め・月次集計で全て「現金」に倒れる）を解消した。
**Option C**（ENUM を真実の源泉として維持 + 書込み時に当該 clinic の master id を同期）を採用済み。

残る4論点は #128 クローズ時に PO 決裁へ預けられ、#185 を単一追跡先として検討された。以下の案・トレードオフ・DoD は **決定前の historical proposal** である。現行 outcome は Status、各「決定」節、末尾の実装メモを正とする。

### マイグレーション安全性（全論点共通の前提）

適用済み migration の **in-place 編集は checksum mismatch を起こす**。2026-07-27に当時のincremental 002〜009は `001_init.sql` 末尾へ統合され、現行DDLは001の単一ファイルになった。この統合前の001が記録済みのDBは `DB_RESET=true` 相当の手動再構築が必要であり、共有環境の復旧・再構築には明示承認が必要、現行workflowに自動reset経路はない。
一方、かつて `007_add_bank_transfer_payment_method.sql`（#127）が実証した通り（同ファイルは 2026-06-26 の統合で `001_init.sql` に取り込み済み）、
**新規ファイルで additive に変更すれば適用済みファイルの checksum は破壊されず、通常デプロイで適用される（db_reset 不要）**。
本 ADR 作成時点では全案を新規ファイル（当時の `004+`）で実現可能と評価していた。現行ベースラインは統合済み `001` のみであり、今後の追加方式はmigration運用規約とchecksum影響を再評価する。
よって各論点を退ける/保留する根拠は「技術的危険」ではなく「設計の意味論を PO が確定すべき」というガバナンス事由である。

---

## Decision Point 1: DB レベルの整合制約（CHECK / TRIGGER）

### 現状（2026-07-31 live）
- 新規の `method` ⇔ `payment_method_id` 矛盾はアプリ層 `resolvePaymentMethodMasterID`
  （`backend/internal/billing/accounting_service_builders.go`）で防止。
  他院 id 混入・method 不一致を拒否、master 欠落は明示エラーで会計確定を止める（NULL→現金フォールバックなし）。
- DB 側では `app_private.enforce_payment_method_system_key_match` を `payments` / `payment_splits` の BEFORE INSERT/UPDATE に接続し、`method` ⇔ `payment_methods.system_key` 一致を強制する（旧006相当・`001_init.sql` 末尾アーカイブ）。
- `payments.clinic_id` と clinic 軸複合 FK（TASK-445 / 旧005相当）により payment 行のテナント境界を DB でも harden 済み。
- 残差: レガシー `payment_method_id IS NULL` 行は許容し得る。TRIGGER は soft-deleted master を拒否しない。確定後訂正経路は保存済 method/payment_method_id 組合せを再検証しない。

### Historical proposal（決定前）
`payment_method_id` は per-clinic master を参照するため、**単純 CHECK では表現不可**（クロステーブル参照が必要）。
- **案 1A**: `payments` / `payment_splits` の `BEFORE INSERT/UPDATE` トリガーで `NEW.payment_method_id` の行を引き、
  `clinic_id` 一致 + `name = expected(NEW.method)` を検証。レガシー NULL 行は後方互換で許容。
- **案 1B（推奨・論点2 と結合）**: system_key 導入後は `name` 比較でなく `system_key ⇔ method` の固定マップで検証。
  改名耐性があり、トリガーが簡潔になる。

### トレードオフ
- 書込み毎に追加クエリ1回（軽微）。アプリ層検証との二重防御で堅牢性は上がるが保守対象が増える。
- レガシー `payment_method_id=NULL` 行（旧 seed・hotfix 前データ）の扱いを決める必要がある。

### Historical DoD（satisfied）
「DB制約を追加する／しない」を決定。追加する場合は 1A/1B のいずれか・NULL 行の扱い・検証範囲を実装 Issue で定義。

### 推奨
論点2（system_key）採用を前提に **1B**。単独優先度は中（アプリ層で既に新規矛盾は防止済みのため）。

---

## Decision Point 2: 安定識別子 `system_key` / `code` 列の導入

### 決定前の現状（historical, before #197）
当時は `payment_methods` が per-clinic id と自由文字列 `name` だけを持ち、書込み・現金集計が name 一致に依存していた。この rename 脆弱性は #197 の `system_key` 導入と key-based lookup への切替で解消済み。現行定義は `001_init.sql` の `CREATE TABLE payment_methods`、`create_default_payment_methods`、および billing の `paymentMethodSystemKeys` / `findCashMethodID` を参照する。

### Historical proposal（決定前）
`payment_methods` に `system_key varchar(50) NULL` を追加。予約体系 = `cash` / `credit_card` / `electronic_money` / `bank_transfer`。
新規の additive マイグレーション（当時の増分ファイルとして提案、現在は `001_init.sql` に統合済み）で：
1. `ALTER TABLE payment_methods ADD COLUMN system_key varchar(50);`
2. 既存行を name 一致で backfill（`'現金'→'cash'` 等。非標準のクリニック独自 method は NULL のまま）。
3. `create_default_payment_methods()` を `CREATE OR REPLACE`（新規ファイル内）で `system_key` 込みに更新。
4. （任意）部分 UNIQUE INDEX `(clinic_id, system_key) WHERE system_key IS NOT NULL AND deleted_at IS NULL`。
   `NOT NULL` 化は全 backfill 完了後の別段階。

コード切替: `paymentMethodMasterNames`（name マップ）を system_key マップへ移行。
`findCashMethodID` / `resolvePaymentMethodMasterID` を system_key 基準に。`name` は表示専用に降格。

### トレードオフ
- 追加列 + backfill。クリニック独自追加の非標準 method は `system_key=NULL`（標準4種のみ予約 key）。
- 技術リスクは低い（additive・db_reset 不要）。本質コストは「予約 code 体系・列名の確定」という設計判断。

### Historical DoD（satisfied）
「列追加 する／しない」+ 列名（`system_key` vs `code`）+ 予約体系を決定。

### 推奨
**採用**。rename 脆弱性の恒久解消はこれ以外にない。技術リスク低・db_reset 不要・論点1/3 の前提にもなる。

### 決定（#197 — 2026-06-26）

**採用**。`payment_methods.system_key varchar(50)` 列を additive migration として追加（当時の増分マイグレーションとして実装し、2026-06-26 の統合により `001_init.sql` に取り込み済み）。

実装内容:
- `ALTER TABLE payment_methods ADD COLUMN IF NOT EXISTS system_key varchar(50)` (db_reset 不要)
- 既存行 backfill（標準4種: cash/credit_card/electronic_money/bank_transfer）
- `create_default_payment_methods()` を `system_key` 込みに `CREATE OR REPLACE`
- 部分 UNIQUE INDEX `(clinic_id, system_key) WHERE system_key IS NOT NULL AND deleted_at IS NULL`

コード切替:
- `paymentMethodMasterNames`（name マップ）→ `paymentMethodSystemKeys`（system_key マップ）
- `loadPaymentMethodNameToID` → `loadPaymentMethodSystemKeyToID`（system_key→id マップ構築、NULL スキップ）
- `findCashMethodID` を `SystemKey='cash'` ベースへ変更（name 比較廃止）
- rename 耐性テスト追加（`renamedPayMethodMock` + `TestFindCashMethodID/rename耐性`）

---

## Decision Point 3: 代表支払方法 `representativeMethod` の優先順位

### 現状（実装前）
- `accounting_service_builders.go` の `representativeMethod` は、当時 `cash > credit_card > else(electronic_money)` だった（PO判断B 2026-05-25 確定）。
- **既知バグ**: `bank_transfer`（#127 で ENUM 追加済）は最終 `else` に落ち、
  **bank_transfer 単独 split でも `electronic_money` と表示される**。`accounting_service_test.go:1155` が旧仕様を固定していた。

### Historical proposal（決定前）
- **案 3A**: 優先チェーンに bank_transfer を追加（`cash > credit_card > bank_transfer > electronic_money`、最終 fallback を見直し）。最小変更。
- **案 3B（推奨）**: 「代表 = 金額最大の split の method」へ意味論を変更。混在会計で業務的に妥当。同額時の tiebreak を固定優先順位で解決。

### トレードオフ
- 3A は最小差分だが「代表」の意味が曖昧なまま。
- 3B は意味が明確だが既存表示・test 更新が必要、`payments.method` の歴史的値とずれる可能性。

### Historical DoD（satisfied）
bank_transfer を含む優先順位ルール、または金額最大方式への変更を確定。

### 推奨
**3B**（業務的妥当性）。ただし「代表」表示の業務意図に依存するため PO 確定が前提。

### 決定（#198 — 2026-06-26）

**案 3A を採用**。`#127` 対応漏れバグとして最小差分で解消（実装 #198）。

優先順位: `cash > credit_card > bank_transfer > electronic_money`

`representativeMethod` は priority slice によるループに書き換え、すべての ENUM 値を明示カバーする。
`accounting_service_test.go` のピンテストを正常系に更新し、`bank_transfer + electronic_money` 混在ケースを追加。

3B（金額最大方式）は将来の独立 Issue で別途判断する。

---

## Decision Point 4: ENUM DROP（Option B）の可否

### 現状
#128 で **Option C**（ENUM 維持 + 書込み時 master id 同期）を採用・マージ済。
`payments.method` / `payment_splits.method` ENUM は維持。

### 判定
**本 follow-up では Option B（ENUM 廃止・master 主軸化）を採用しない。**
返金 API 契約（`BillingRefund.payment_method` ENUM、#60 方法別返金上限）・FE ラベル・seed・集計の全面改修を伴い、不可逆かつ広範。

将来 Option B を採るなら、段階移行（後方互換期間・データ移行・契約バージョニング）を**独立した大型 Issue** で策定する。

### 推奨
**Option C 継続**。Option B は現時点 WONTFIX（将来再評価）。

---

## Consequences（live outcome）

**実装済み:**
- `system_key` と部分 unique indexにより、標準支払方法の識別を表示名から分離した。
- application lookup と DB trigger は `method` と `system_key` の不一致を拒否する。
- `payments.clinic_id` と clinic 軸複合 FK が支払 graph の tenant 整合性を補強する。
- `representativeMethod` は `cash > credit_card > bank_transfer > electronic_money` を明示的に扱う。
- Option C（ENUM 維持）を継続し、ENUM DROP は WONTFIX とした。

**残差:**
- legacy の `payment_method_id IS NULL` は許容される。
- trigger は soft-deleted master を拒否しない。
- 確定後訂正経路は保存済み `method` / `payment_method_id` の組合せを再検証しない。

## Historical implementation order（完了済み）

決定前は `system_key` →整合 trigger、並行して `representativeMethod`、ENUM DROP は WONTFIX の順序を想定した。#197、#198、TASK-445、TASK-ADR003 で選択項目は実装済みであり、これは active work list ではない。

## References

- #185（本 follow-up・残論点の単一追跡先）/ #128（中核バグ解消済・CLOSED）/ #127（bank_transfer ENUM 追加）
- hotfix `d9bcd387` / `007_add_bank_transfer_payment_method.sql`（additive migration の実証・2026-06-26 統合により `001_init.sql` に取り込み済み）
- `backend/internal/billing/accounting_service_builders.go`（`resolvePaymentMethodMasterID` / `representativeMethod`；BE9 以前は `internal/service`）
- `backend/internal/billing/cash_register_service.go`（`findCashMethodID` / `calcTheoreticalCash`；BE9 以前は `internal/service`）
- `backend/migrations/001_init.sql` の `CREATE TABLE payment_methods` / `create_default_payment_methods`
- [ADR-002: マルチテナント設計](002-multitenancy-clinic-id-isolation.md)

## 実装メモ（2026-07-25・TASK-ADR003）

Decision Point 1 のうち、`payment_splits.payment_method_id` の clinic 一致は、ADR 作成後に導入された既存の複合 FK パターンを使って宣言的に実装した。旧 `backend/migrations/006_payment_splits_payment_method_clinic_fk.sql`（`c434c4e66`、2026-07-27に001へ統合。現行所在は `001_init.sql` 末尾の旧006アーカイブブロック）は `payment_methods` に述語なしの `UNIQUE (id, clinic_id)` を追加し、`payment_splits (payment_method_id, clinic_id)` から `payment_methods (id, clinic_id)` への複合 FK を追加する。既定の `MATCH SIMPLE` により legacy の `payment_method_id IS NULL` 行は許容し、削除動作は `ON DELETE RESTRICT` とした。soft-delete 済み master への既存参照を許す挙動は変えない。

続報（2026-07-29 統合 / live）: `method` ⇔ `system_key` 一致は `app_private.enforce_payment_method_system_key_match` と `payments`/`payment_splits` トリガーで実装済み。`payments.clinic_id` と複合 FK（billing / payment_methods との clinic 軸）も TASK-445 相当として `001_init.sql` に統合済み。通常の会計作成・更新経路は引き続き `backend/internal/billing/accounting_service_builders.go` の `resolvePaymentMethodMasterID` が不一致を拒否する。確定後訂正経路は `method` / `payment_method_id` 自体を変更しないが、保存済みの組合せは再検証しない。レガシー `payment_method_id IS NULL` 行と soft-deleted master 参照の扱いは Status の残差を参照。

## GitHub 決定記録との照合（2026-09-06）

[GitHub #185](https://github.com/MinoruSoga/AnimalEkarte/issues/185) の CLOSED は 2026-06-26 の設計判断の完了を示し、同本文には当時の TRIGGER 不採用が残る。その後の `001_init.sql` と本 ADR の TASK-ADR003 実装メモでは整合 trigger が存在する。Issue の当初判断を現在の「trigger 未実装」の根拠にしない。[#197](https://github.com/MinoruSoga/AnimalEkarte/issues/197) / [#198](https://github.com/MinoruSoga/AnimalEkarte/issues/198) の system_key / bank_transfer 対応も実装済みだが、Issue にある旧 `internal/service` や削除済み incremental migration は historical path として読む。

通常 Create/Update から `completed` へ遷移する旧会計経路は現在拒否される。新規確定は `accounting_complete.go` の `Complete` に集約され、現行支払検証はこの entry と `accounting_service_builders.go` を合わせて確認する。ここで過去の意思決定や実 DB 適用状態を変更・認定していない。

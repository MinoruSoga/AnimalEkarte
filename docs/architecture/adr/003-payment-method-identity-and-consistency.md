# ADR-003: 支払方法の安定識別と payment_methods 整合性の設計判断

**Status**: Decided — 論点2（system_key）#197 済・論点3（representativeMethod）#198 済・論点1（TRIGGER）保留（独立 Issue で再評価）・論点4（ENUM DROP）WONTFIX 維持。
**Date**: 2026-06-25 / Updated: 2026-06-26
**Deciders**: Engineering（論点2/3）。論点1 は PO 判断待ち。
**Drafted by**: Engineering（#185 follow-up）

## Context

#128（`d9bcd387` マージ済）で `payment_method` ENUM ⇔ `payment_method_id` 二重管理の中核バグ
（アプリ書込みの非現金会計が `payment_method_id=NULL` で保存され、レジ締め・月次集計で全て「現金」に倒れる）を解消した。
**Option C**（ENUM を真実の源泉として維持 + 書込み時に当該 clinic の master id を同期）を採用済み。

残る4論点は恒久的アーキテクチャ判断であり、**#128 クローズ時に PO 決裁へ預けられた**。#185 がその単一追跡先である。
本 ADR は各論点に対し **技術案・トレードオフ・推奨** を提示し、PO 決裁の材料とする。
**実装は本 ADR では行わない。** 決裁確定後、論点ごとに個別の実装 Issue を起票して着手する。

### マイグレーション安全性（全論点共通の前提）

適用済み migration（`001`〜`004`）の **in-place 編集は checksum mismatch を起こし STG db_reset 必須**になる。
一方、かつて `007_add_bank_transfer_payment_method.sql`（#127）が実証した通り（同ファイルは 2026-06-26 の統合で `001_init.sql` に取り込み済み）、
**新規ファイルで additive に変更すれば適用済みファイルの checksum は破壊されず、通常デプロイで適用される（db_reset 不要）**。
本 ADR の全案は新規ファイル（`005+`、現行ベースラインは `001`〜`004`）で実現可能であり、**技術リスクは低い**。
よって各論点を退ける/保留する根拠は「技術的危険」ではなく「設計の意味論を PO が確定すべき」というガバナンス事由である。

---

## Decision Point 1: DB レベルの整合制約（CHECK / TRIGGER）

### 現状
- 新規の `method` ⇔ `payment_method_id` 矛盾はアプリ層 `resolvePaymentMethodMasterID`
  （`backend/internal/service/accounting_service_builders.go:28-44`）で防止。
  他院 id 混入・method 不一致を拒否、master 欠落は明示エラーで会計確定を止める（NULL→現金フォールバックなし）。
- ただし**直接 SQL 書込み（手動・外部ツール・将来の雑な migration）は防げない**。
  payment 系トリガーは `trg_create_default_payment_methods`（`001_init.sql:2735`）のみで整合制約は存在しない。

### 設計案
`payment_method_id` は per-clinic master を参照するため、**単純 CHECK では表現不可**（クロステーブル参照が必要）。
- **案 1A**: `payments` / `payment_splits` の `BEFORE INSERT/UPDATE` トリガーで `NEW.payment_method_id` の行を引き、
  `clinic_id` 一致 + `name = expected(NEW.method)` を検証。レガシー NULL 行は後方互換で許容。
- **案 1B（推奨・論点2 と結合）**: system_key 導入後は `name` 比較でなく `system_key ⇔ method` の固定マップで検証。
  改名耐性があり、トリガーが簡潔になる。

### トレードオフ
- 書込み毎に追加クエリ1回（軽微）。アプリ層検証との二重防御で堅牢性は上がるが保守対象が増える。
- レガシー `payment_method_id=NULL` 行（旧 seed・hotfix 前データ）の扱いを決める必要がある。

### DoD（PO 判断条件）
「DB制約を追加する／しない」を決定。追加する場合は 1A/1B のいずれか・NULL 行の扱い・検証範囲を実装 Issue で定義。

### 推奨
論点2（system_key）採用を前提に **1B**。単独優先度は中（アプリ層で既に新規矛盾は防止済みのため）。

---

## Decision Point 2: 安定識別子 `system_key` / `code` 列の導入

### 現状
- `payment_methods` は **per-clinic id + 自由文字列 `name`** のみ
  （`001_init.sql:1885-1894`、`UNIQUE(clinic_id, name)`）。`PaymentMethodMaster` に安定列なし。
- 現金判定はマスタ `name="現金"` の文字列一致
  （`accounting_service_builders.go:15-20` `paymentMethodMasterNames` / `cash_register_service.go:414` `findCashMethodID`）。
- **rename 脆弱性**: クリニックが既定 name を改名すると、
  - 書込み側 `resolvePaymentMethodMasterID` は **明示エラーで安全に倒れる**（データ破損なし、ただし会計確定不能）。
  - 集計側 `findCashMethodID` は既存現金 master を**見失い、現金を過少計上**しうる（`name` ベースのため）。

### 設計案
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

### DoD（PO 判断条件）
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
- `accounting_service_builders.go:60-72`。優先順位 `cash > credit_card > else(electronic_money)`（PO判断B 2026-05-25 確定）。
- **既知バグ**: `bank_transfer`（#127 で ENUM 追加済）は最終 `else` に落ち、
  **bank_transfer 単独 split でも `electronic_money` と表示される**。`accounting_service_test.go:1155` が旧仕様を固定していた。

### 設計案
- **案 3A**: 優先チェーンに bank_transfer を追加（`cash > credit_card > bank_transfer > electronic_money`、最終 fallback を見直し）。最小変更。
- **案 3B（推奨）**: 「代表 = 金額最大の split の method」へ意味論を変更。混在会計で業務的に妥当。同額時の tiebreak を固定優先順位で解決。

### トレードオフ
- 3A は最小差分だが「代表」の意味が曖昧なまま。
- 3B は意味が明確だが既存表示・test 更新が必要、`payments.method` の歴史的値とずれる可能性。

### DoD（PO 判断条件）
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

## Consequences

**ポジティブ:**
- 残4論点が技術案・トレードオフ・db_reset 影響付きで可視化され、PO が判断材料込みで決裁できる。
- 実装は決裁後に論点ごとの最小差分 Issue へ分離でき、schema-wide rewrite を回避できる。
- 論点2→1/3 の依存順序が明確（system_key を先に入れると後続が簡潔化）。

**ネガティブ:**
- 決裁が滞ると #185 は OPEN 維持となり、rename 脆弱性（集計側の現金過少計上）が残存する。
  ただし書込み側は安全 fail のため**データ破損は発生しない**（運用上の脆さに留まる）。

## 実装順序（PO 決裁後）

```
論点2 (system_key 導入) ──▶ 論点1 (整合 TRIGGER, 1B で簡潔化)
                        └─▶ 論点3 (representativeMethod 再設計) ※独立だが system_key と同時が効率的
論点4 (Option B) ── WONTFIX / 将来の独立大型 Issue
```

## References

- #185（本 follow-up・残論点の単一追跡先）/ #128（中核バグ解消済・CLOSED）/ #127（bank_transfer ENUM 追加）
- hotfix `d9bcd387` / `007_add_bank_transfer_payment_method.sql`（additive migration の実証・2026-06-26 統合により `001_init.sql` に取り込み済み）
- `backend/internal/service/accounting_service_builders.go`（`resolvePaymentMethodMasterID` / `representativeMethod`）
- `backend/internal/service/cash_register_service.go`（`findCashMethodID` / `calcTheoreticalCash`）
- `backend/migrations/001_init.sql:1885`（payment_methods スキーマ）/ `:2722`（create_default_payment_methods）
- [ADR-002: マルチテナント設計](002-multitenancy-clinic-id-isolation.md)

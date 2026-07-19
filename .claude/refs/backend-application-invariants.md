---
description: AnimalEkarte backendのセキュリティ・データ分離に関するアプリケーション固有不変条件
---

# Backend Application Invariants

この文書は Go/Gin 公式ベストプラクティスではない。AnimalEkarte が医療関連データを安全に扱うための、architecture の形に依存しない不変条件を定義する。

## Tenant and ownership isolation

- clinic-scoped data のすべての read/write/delete は、認証済み `clinic_id` で制約する。
- client が送信した `clinic_id` を認可根拠にしない。認証済み identity から決定する。
- owner、pet、staff、medical record、billing 等の関連 ID は、同じ clinic に属することを server-side で確認する。
- not-found と cross-tenant resource の差で、他 clinic の存在を推測できる応答を返さない。
- bulk query、join、preload、count、export、background job にも同じ scope を適用する。
- transaction の内外で scope を失わない。raw SQL も例外にしない。

具体的な isolation 方式と監査証跡は [ADR-002](../../docs/architecture/adr/002-multitenancy-clinic-id-isolation.md) を正本とする。

## Authorization and sensitive data

- authentication、role/permission authorization、resource ownership をそれぞれ検証する。
- patient/owner/staff 情報、診療内容、credential、token、secret を error や log に出さない。
- response、export、event、audit log は最小限の field だけを含める。
- security boundary の変更は、unauthorized/cross-tenant test を必ず含める。

## Data integrity

- schema migration は project の migration 規約と ADR に従い、application 起動時の暗黙 `AutoMigrate` に依存しない。
- foreign key、unique constraint、transaction を application check の代替ではなく、追加の防御として使う。
- destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。

## Separation from official guidance

これらの不変条件を特定の `handler`、`service`、`repository` package の存在に結び付けない。package 構成が変わっても、すべての入口・use case・persistence path で検証できる形を維持する。

# adr/ — アーキテクチャ意思決定記録 (Architecture Decision Records)

> **目的**: 重要な設計判断の「なぜ」を記録し、後から経緯を追跡可能にする。
> **読者**: 全開発者・AI エージェント。
> **タイミング**: 既存設計の変更を検討する前（先に該当 ADR の判断理由を確認する）。

## 索引

| ADR | 内容 |
|:---|:---|
| [001-system-architecture.md](001-system-architecture.md) | システムアーキテクチャ基本設計（backend構成判断はADR-005でsuperseded） |
| [002-multitenancy-clinic-id-isolation.md](002-multitenancy-clinic-id-isolation.md) | マルチテナント設計 — clinic_id 完全隔離 |
| [003-payment-method-identity-and-consistency.md](003-payment-method-identity-and-consistency.md) | 支払方法の安定識別と payment_methods 整合性 |
| [004-checkup-canonical-system.md](004-checkup-canonical-system.md) | 健診機能の正系統 — Checkup パッケージ系に一本化 |
| [005-go-gin-backend-guidelines.md](005-go-gin-backend-guidelines.md) | Go/Gin公式ベースライン採用と固定3層構成の廃止 |
| [006-backend-domain-package-boundaries.md](006-backend-domain-package-boundaries.md) | backend domain package境界・許可依存グラフ（Status: Accepted/Implemented、2026-07-24 amended） |

## 運用ルール

- ADR は追記型。決定を覆す場合は既存 ADR を書き換えず、新番号の ADR で「supersedes ADR-XXX」と明記する。
- 新規 ADR の起票対象は「複数の実装に波及する不可逆性の高い判断」のみ。個別機能の仕様は [../../spec/](../../spec/README.md) に書く。

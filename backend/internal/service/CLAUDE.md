# Application Logic Packages

このdirectoryはBE9移行完了後（2026-07-24）のtest-only surfaceであり、残存する14 fileは全て`_test.go`、production実装は0件で、新規production実装の追加先ではない。新規実装は[ADR-006](../../../docs/architecture/adr/006-backend-domain-package-boundaries.md)のtarget domain packageへ置く。既存testは移行後の互換性と回帰検出のために維持する。`service` layer自体はGo/Gin公式要件ではなく、package境界の正本はADR-006と[boundary map](../../../docs/architecture/be9-2a-boundary-map.md)とする。

## Review points

- request-scoped 処理は `context.Context` を第1引数で受け、DB/外部 API へ伝播する。
- Context を struct に保存しない。timeout/cancel の cleanup を保証する。
- interface は利用側が必要とする最小メソッドだけを定義する。
- input は HTTP/Gin に依存させる必要がない限り、通常の Go type として表現する。
- business invariant、authorization、ownership を persistence error に依存せず明示する。
- error chain を `%w` で保持し、同じ error を重複ログしない。
- transaction/concurrency の境界と failure 時の atomicity を test する。
- goroutine の ownership、終了条件、cancel、error/panic 経路を明確にする。

constructor 名、input type 名、定義順、validator の配置、logging する package は Go/Gin公式未規定である。

clinic/owner/pet/staff の分離と確定済み診療データの完全性は [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md) と関連 ADR/test を維持する。

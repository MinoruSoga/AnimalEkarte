# Application Logic Packages

このdirectoryは未移行application logicと期限付きcompatibility facade/adapterのmigration surfaceであり、新規production実装の追加先ではない。新規実装は[ADR-006](../../../docs/architecture/adr/006-backend-domain-package-boundaries.md)のtarget domain packageへ置く。ここを変更するのは、未移行実装の保守・安全修正、domain移動、または実consumerと削除phaseを持つcompatibility変更に限る。`service` layer自体はGo/Gin公式要件ではなく、詳細は[`BE-refactor.md`](../../../BE-refactor.md)に従う。

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

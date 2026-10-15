# 飼主・ペットの技術設計

[技術契約](technical-spec.md)は既存writer／API／DB／transactionと合意業務の差を記述する入口。業務要件と受入条件は[業務正本](../../../spec/domains/owners-pets/lifecycle.md)に置き、API型・DB定義を重複定義しない。

保存単位、型付き参照、競合・結果不明／再送、必要監査、外部副作用、旧値／未実装差、他ownerへの引継ぎを確認する。文書整理は実装済み・運用可能の宣言ではない。2026-10-15静的照合、実機未実施。

2026-10-15追加合意4・5の[連絡先重複](technical-spec.md#owner-contact-duplicate-target)と[主飼主影響確認](technical-spec.md#owner-transfer-target)を、従来の409拒否・限定移管実装とは分けて記載する。候補とID、影響確認と同一txの保存、台帳commitと通知到達を区別し、過去請求・預り金の元ownerを保持する。

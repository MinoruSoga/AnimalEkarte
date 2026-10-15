# 入院・ホテルの技術設計

[技術契約](technical-spec.md)は既存writer／API／DB／transactionと合意業務の差を記述する入口。業務要件と受入条件は[業務正本](../../../spec/domains/hospitalization/lifecycle.md)に置き、API型・DB定義を重複定義しない。

保存単位、型付き参照、競合・結果不明／再送、必要監査、外部副作用、旧値／未実装差、他ownerへの引継ぎを確認する。文書整理は実装済み・運用可能の宣言ではない。2026-10-15静的照合、実機未実施。


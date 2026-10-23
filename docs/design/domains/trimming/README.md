# トリミングの技術設計

[技術契約](technical-spec.md)は既存writer／API／DB／transactionと合意業務の差を記述する入口。業務要件と受入条件は[業務正本](../../../spec/domains/trimming/lifecycle.md)に置き、API型・DB定義を重複定義しない。

保存単位、型付き参照、競合・結果不明／再送、必要監査、外部副作用、旧値／未実装差、他ownerへの引継ぎを確認する。文書整理は実装済み・運用可能の宣言ではない。2026-10-15静的照合、実機未実施。


<a id="trimming-start-end-technical-selection"></a>
## 2026-10-23開始・終了の技術接続

[開始・終了の限定契約](trimming-start-end-contract.md#trimming-start-end-closed-wire)で468単一writer、T4 schema2の実担当者（actor/予約指名と別）、閉じたstart/end fingerprintと最小準備readproofを選定。準備はmetadataのみで記録作成/本文save/状態変更なし。現在認可/版検査を維持し、同key成功はproof再検証前に扱う。途中保存/開始に実担当者・料金の新gateを加えず、終了の既決9ACを保持。共有供給/実装/試験/全9AC受入/activationは未完（NOT_RUN）。

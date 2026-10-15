# カルテの技術設計

**最新の読む順**：[2026-10-15会議合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)を先に確認する。自動保存維持・診療終了後の通常編集・履歴付きタブ開始時復元／複製直前取消・終了時警告／終了履歴を追加し、旧医師確定必須と本文ロックの目標を置換。以下の完了記録は当時の範囲で、今回の統合レビューは重大指摘なしで完了。保存済み版・会計確定歴・検査自体の確定等の独立保護は保持する。全過去の承認証跡網羅や実装完了は意味しない。

通常カルテの初回5項目と入力4項目について[部分技術仕様](technical-spec.md)を整理する。読む順は[業務入口](../../../spec/domains/medical-records/README.md) → [保存・初回合意](../../../spec/domains/medical-records/lifecycle.md) → 技術仕様 → 既存コード参照。予約／受付・専用記録・会計の正本を複製しない。

初期の部分合意反映と今回の最終照合を区別する。現在は16FR／71ACと通常カルテ／D03master／受け手のdocs契約の明示範囲を文書完了と判定。全機能・実装・他owner内部・全体横断最終確認の完了ではない。API YAML・DB・コード・実機受入は未変更／未検証。全体範囲と状態は[仕様整理進捗](../../../spec/domains/SPECIFICATION_PROGRESS.md)で管理し、実装タスク管理はPlaneへ保持する。

## 訂正・例外操作の追加範囲

以下2節は部分合意時点の進行中／文書残件の履歴。現在の判定は末尾の最終照合と完成条件を参照。

[追加合意4FR／14AC](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-correction-four) → [接続設計](technical-spec.md#訂正例外4項目の接続設計)。初回5・入力4と合わせた部分合意13FR／45AC。全タブ・API／DB・全owner・実機は未完了で、D03は進行中。

## 追加3項目と全タブの契約

[業務3FR／12AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three) → [共通14AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-shared-coverage-ac) → [追加技術契約](technical-spec.md#medical-record-integration-contracts) → [全9タブ＋バイタルの型・操作](technical-spec.md#medical-record-full-coverage-contract)。合意計16FR／71ACを文書整理。通常業務の追加質問は不要、実wire／全DTO・旧値／grant／全writerの最終照合と他owner契約は文書残件。D03は進行中、薬量具体値の別承認・実機未実施を維持する。

## 最終照合と完成判定

[API／DTO・互換](technical-spec.md#medical-record-reconciled-contract) → [writer／監査／gate](technical-spec.md#medical-record-writer-matrix) → [旧値／専用intent](technical-spec.md#medical-record-legacy-and-special-contract) → [合意16FR／71AC trace](technical-spec.md#medical-record-final-trace) → [文書判定範囲](technical-spec.md#medical-record-document-completion)。初期の部分設計時点の残件はこの最終contractへ接続する。独立レビュー3名・限定文書検査を通過し、明示範囲を文書完成と判定。実権限・実旧値・薬量数値承認・全writer適合／利用開始・実機性能・他owner内部／全体統合の未確認を保持する。次の推奨はD10会計。

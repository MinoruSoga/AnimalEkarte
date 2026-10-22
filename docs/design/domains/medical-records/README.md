# カルテの技術設計

**診断・バイタルの追加差分（2026-10-15）**：[各区分で複数診断・最新値の測定時刻／同日履歴の業務2FR／12AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-agreement-diagnosis-vitals-20261015) → [論理契約・旧wireとの差分](technical-spec.md#medical-record-diagnosis-vitals-contract) → [追加trace](technical-spec.md#medical-record-diagnosis-vitals-trace)。合意追跡は24FR／101AC（旧継承ID・C05追加1ACは別計数）。以下の過去完了記録は当時の範囲。第1・第2の区分・各複数件、時刻付き表示と以前の測定履歴を文書化し、自動保存・終了後通常編集・復元保護を保持。実装／API・DB・実機検証は未実施。

**最新の読む順**：[2026-10-15会議合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)を先に確認する。自動保存維持・診療終了後の通常編集・履歴付きタブ開始時復元／複製直前取消・終了時警告／終了履歴を追加し、旧医師確定必須と本文ロックの目標を置換。以下の完了記録は当時の範囲で、今回の統合レビューは重大指摘なしで完了。保存済み版・会計確定歴・検査自体の確定等の独立保護は保持する。全過去の承認証跡網羅や実装完了は意味しない。

通常カルテの初回5項目と入力4項目について[部分技術仕様](technical-spec.md)を整理する。読む順は[業務入口](../../../spec/domains/medical-records/README.md) → [保存・初回合意](../../../spec/domains/medical-records/lifecycle.md) → 技術仕様 → 既存コード参照。予約／受付・専用記録・会計の正本を複製しない。

初期の部分合意反映と今回の最終照合を区別する。現在は16FR／71ACと通常カルテ／D03master／受け手のdocs契約の明示範囲を文書完了と判定。全機能・実装・他owner内部・全体横断最終確認の完了ではない。API YAML・DB・コード・実機受入は未変更／未検証。全体範囲と状態は[仕様整理進捗](../../../spec/domains/SPECIFICATION_PROGRESS.md)で管理し、実装タスク管理はPlaneへ保持する。

## 訂正・例外操作の追加範囲

以下2節は部分合意時点の進行中／文書残件の履歴。現在の判定は末尾の最終照合と完成条件を参照。

[追加合意4FR／14AC](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-correction-four) → [接続設計](technical-spec.md#訂正例外4項目の接続設計)。初回5・入力4と合わせた部分合意13FR／45AC。全タブ・API／DB・全owner・実機は未完了で、D03は進行中。

## 追加3項目と全タブの契約

[業務3FR／12AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three) → [共通14AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-shared-coverage-ac) → [追加技術契約](technical-spec.md#medical-record-integration-contracts) → [統合前の全9タブ＋バイタルの型・操作](technical-spec.md#medical-record-full-coverage-contract)。合意計16FR／71ACを文書整理。通常業務の追加質問は不要、実wire／全DTO・旧値／grant／全writerの最終照合と他owner契約は文書残件。D03は進行中、薬量具体値の別承認・実機未実施を維持する。

## 最終照合と完成判定

[API／DTO・互換](technical-spec.md#medical-record-reconciled-contract) → [writer／監査／gate](technical-spec.md#medical-record-writer-matrix) → [旧値／専用intent](technical-spec.md#medical-record-legacy-and-special-contract) → [合意16FR／71AC trace](technical-spec.md#medical-record-final-trace) → [文書判定範囲](technical-spec.md#medical-record-document-completion)。初期の部分設計時点の残件はこの最終contractへ接続する。独立レビュー3名・限定文書検査を通過し、明示範囲を文書完成と判定。実権限・実旧値・薬量数値承認・全writer適合／利用開始・実機性能・他owner内部／全体統合の未確認を保持する。次の推奨はD10会計。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/medical-records/lifecycle.md#integrated-plan-20261017)。問診統合・右側履歴移設、飼主原文保護、非表示の来院別要約保存、1回量／用法／処方総量表示、価格入り4種明細の実施前後単価変更を適用する。旧9タブ／問診単独送信の記述は統合前の保存先・実装参照で、画面を別タブへ戻す根拠としない。owner別保存で全体一括txにしない（選定済み）。同一カルテの同時編集禁止・強制取得可は[2026-10-18追加合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-edit-exclusivity-20261018)に従い、閲覧のみでは取得せず「編集を開始」で取得し、同一カルテのタブ移動・単なる無操作では維持する。別画面離脱は保存結果確認後に解放し、切断時は入力保持と編集・送信停止、再接続時は編集権と保存結果を確認する。再読み込み後も未保存入力を復元可能とし、再取得後は最新内容と比較して明示選択適用する（2026-10-18 04:38:21 UTC本人承認）。API・編集session・保存／鍵方式は[EMR-494採用契約](../../../work/emr/emr-494-edit-draft-contract-20261020.md)で選定済み。保持期限・自動破棄・端末間共有は新設しない（同契約:14）。比較UIの未選定詳細と実装残は維持。通常の単価割引理由は任意、その他行・取消・提示済み見積・確定会計の必要理由と既決薬量安全条件は維持する。

## EMR-518 診療項目パッケージの採用方向

[package・共通予定の技術方向](treatment-packages.md)と[固定残件/編集owner/PR単位](../../../work/emr/emr-518-treatment-packages-20261020.md)を参照。共通予定schemaはEMR-516窓口へ集約し、実績/入院表の意味を変更しない。論理要件の正本は[4FR/12AC](../../../spec/domains/medical-records/treatment-packages.md)。表名は採用設計名で全列/FK/DTO/APIの固定は未完、runtime/試験はNOT_RUN。

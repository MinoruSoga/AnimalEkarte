# BE-refactor.md 統合計画（並行編集リスク明示・直接編集は保留）

> **結論（先出し）**: BE-refactor.md本体への直接編集は**今回もBLOCKED**。理由は並行セッションが
> 同ファイルを未コミットで編集中であり、衝突リスクが解消されたと確認できないため。
> Scopeの制約「衝突リスクがないことを確認し、無理なら統合計画のみ作る」に従い、本ファイルは
> 統合計画のみを提供する。

## 衝突リスクの確認結果

`git diff --stat -- BE-refactor.md` = `1 file changed, 3 insertions(+), 1 deletion(-)`。
working treeに**未コミットの並行編集が現在も存在する**（前回検証時と差分なし＝進行中か放置中かは
不明だが、いずれにせよ「衝突リスクがないこと」を積極的に確認できる材料がない）。

→ **安全と確認できないため、直接編集は行わない**（Scope指示通り）。

## 現在の未コミット差分の内容（統合対象の実体）

並行セッションが `BE-refactor.md` に加えた3箇所の変更は、`docs/be-refactor-followup-status.md`
の「BE-refactor.md 陳腐化訂正」節（Phase A-5 = 旧 #225a 相当）の内容と**整合している**
（同じ結論：R3-3 openapi drift対応完了・R3-4 checkup_field_results複合FK完了・exam_results BLOCKED）。
つまり並行セッションの編集は本タスクの結論と**矛盾しておらず、むしろ先行して同じ訂正を
BE-refactor.md側に反映しようとしている**と解釈できる。

差分の3箇所（`git diff -- BE-refactor.md` より）:

1. **R3-3節（L132付近）**: 「状態（2026-07-02 検証・完了）」として、`openapi_date_format_drift_test.go`
   実装済み・22箇所drift検出・CI job `openapi-date-format-drift` 配線済みの旨を追記。
2. **R3-4節（L165付近）**: 「状態（2026-07-02 検証）」として、checkup_field_results側の複合FK完了、
   exam_results側は構造的BLOCKED（clinic_id列不在）の旨を追記。
3. **Definition of Doneセクション（L211付近）**: 「exam_results / checkup_field_results に...」の行を
   「checkup_field_results に...完了。exam_results は...正当BLOCKED」に書き換え。

## マージ条件（いつ直接編集して良いか）

以下のいずれかが満たされた時点で、`BE-refactor.md`への直接統合を再検討可能:

1. **並行セッションが自身の変更をcommitする** → 本タスクの結論との整合を再確認した上で
   fast-forwardベースでの追加編集が可能（衝突が起きない）。
2. **並行セッションが一定期間（例: 24時間）応答なく、working treeの変更が意図的な放置と
   判断できる** → ユーザー/POに確認の上、変更を引き継ぐか破棄するかの判断を仰ぐ。
3. **ユーザーから明示的に「並行編集は解消済み、直接編集して良い」との承認がある**。

いずれの場合も、直接編集前に必ず `git diff -- BE-refactor.md` を再実行し、想定外の変更が
追加されていないことを確認すること。

## 統合時に反映すべき内容（Phase A分・並行セッション差分とは別に必要な追記）

並行セッションの3箇所の差分に加えて、本タスク（Phase A）が新たに確定した以下の事実は、
**まだBE-refactor.md本体のどこにも反映されていない**。将来の統合作業でこれらも追記対象とすること:

| BE-refactor.md 想定反映箇所 | 追記内容 | 出典 |
|---|---|---|
| RLS関連セクション | RLSは実測でdormant確認済み（`ekarte_user`がrolsuper=true）。full実効化は別Issue化(Draft 1参照)して着手 | Phase A-2a |
| tx開始点関連セクション（もしあれば新設） | tx開始点は合計27箇所（内訳: ambient1/dbOrTx参加可能4/ignore-ambient21/service1）。旧「22箇所」記述は誤り | Phase A-2b |
| RLS関連セクション | 非ownerロールでのRLS実効性はローカル実証済み（`rls_effectiveness_test.go`）。**ただし当該テストファイルは未コミット**であり、統合時に忘れずコミット対象へ含めること | Phase A-3 |
| セキュリティ/権限セクション | `system_admin`は審査済みapp層バイパス（既知・AUDIT-H2）。DB接続ロールレベルのbypassは存在しない | Phase A-4 |
| DoD/未完了リスト | date-only wire統一はFE/PO確認待ちでBLOCKED継続。RLS full実効化もarchitect/PO判断待ちでBLOCKED継続（両方とも「未完了」であることを明記し、完了と誤読されないようにする） | Phase A-1, A-2 |

## 未実施事項の明示

本ファイル作成にあたり `BE-refactor.md` への直接編集（Write/Edit）は**一度も実行していない**。
本ファイルは統合計画（将来の作業指示書）としてのみ機能する。

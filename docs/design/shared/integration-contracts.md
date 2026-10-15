# 技術接続基盤と文書検査（C06）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.shared.integration-contracts／共通契約の整理候補 |
| 決定状態 | 既存記載の継承。新業務承認を含まない。根拠不明は未確認のまま。未回答案は別節 |
| 実装状態 | 既存資料の記載を継承（コード全体の再監査なし）。選定仕様と既存挙動の差を明記 |
| 実機検証状態 | 未実施 |
| 文書確認日・基準 | 2026-10-15／PR #670 commit 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8 |
| 対象 | 全owner接続の技術文書契約・最終文書照合。実装変更と全writerの実装適合検証は別工程 |

本書は既存正本への対応と受入観点を束ねる。数値・語彙・業務条件の第二正本は作らず、リンク先の有効な個別契約を優先する。FR/ACの識別子付与は追加承認や実装済みを表さない。未確認の不足を対象外へ移して完了としない。

## 正本・読む順

[Architecture](../../architecture/README.md) → [overview](../../architecture/overview.md) → [FE／BE map](../../architecture/fe-feature-be-domain-map.md) → [write-owner](../../architecture/model-write-owner-catalog.md) → [横断処理catalog](../../architecture/cross-domain-orchestration-catalog.md) → 各domain技術仕様。wireは[api.yaml](../../../backend/docs/api.yaml)、DBはmigration（[ERD入口](../../architecture/erd.md)）。本書へ全DTO・全table定義をコピーしない。

C01の部品技術、C04の認証／医院scopeと個別Dの操作認可・ID・状態は別の責務。D01–D23の文書管理境界はBE packageを再編しない。API仕様・実装事実だけで業務判断を採択しない。

<a id="requirement-fr-common-tech-owner"></a>
### FR-common-tech-owner — 一事実一writer

一つのbusiness factは一つのwrite-ownerへ。appointmentsはreservation、billing graphはbilling、staff勤務はstaff等の既存ownerを維持。shared model配置は共有writerの許可ではない。consumer-sideの最小interfaceまたはtyped intentを通し、他ownerへ任意field更新・独立write pathを増やさない。共通ID／enumを複製しない。FEのshared昇格は実consumer2以上・理由・domain固有API/RBAC知識を持たない条件を満たすものに限る。

<a id="requirement-fr-common-tech-boundary"></a>
### FR-common-tech-boundary — tx・失敗・回復

複数owner writeはambient transactionへの参加か、commit順・部分成功・回復・観測を明記したorchestrationへ。必要なtx／依存／検証／監査が欠けるintegrity経路はfail-closed。best-effortは明記した経路だけに許容し、全処理を勝手に同一txへ格上げしない。逆に既存実装のbest-effortを、最新合意の原子性を弱める根拠にしない。

既存catalogにある各候補単位same-txと、全医院batch loopのmulti-txを区別する。timeout／応答喪失／失敗時は各ownerのID・receipt・再読取・再送契約に従い、単にPOSTを再実行して冪等と呼ばない。未定の一律Version欄・ETag・HTTP code・全業務共通retry回数を新設しない。

<a id="requirement-fr-common-tech-contract"></a>
### FR-common-tech-contract — 各接続に必要な確認項目

| 項目 | 各Dで具体化する内容 | 未記載時の扱い |
|:---|:---|:---|
| 対象とwriter | initiator、owner、操作intent、読取と書込 | package名だけで適合扱いしない |
| identity／scope | clinic、owner、pet、record／appointment／billing ID、認可と対象照合 | client指定IDを認可根拠にしない |
| wireと版 | request/response実キー・型、未送信／null／clear、状態enum、依存版／旧値 | api.yaml／DTOの差は未確認として残す |
| 保存単位 | draft/local／即時保存／確定、tx owner、lock順、必須監査sink | 同一画面を一括txと推測しない |
| 競合・失敗 | 入力拒否、権限失効、not-found、競合、通信失敗、commit後応答喪失 | 各状態で入力保持／再取得／訂正の期待結果が必要 |
| 再送・取消 | 同一効果識別、payload相違、receipt、対象取消／補償 | 一律Undo・再POST・自動再試行を追加しない |
| 読取consumer | snapshot／原事実／派生表示、遅延応答、患者context、section scope | 最新masterで過去事実を置換しない |
| 外部通知 | commit前後、部分成功、停止、観測、再実行の主体 | 内部成功と外部成功を混同しない |
| 保存・監査 | actor、時刻、理由／元先、必要tx、保持／照会・非再利用 | C05当面方針へ接続し法令適合とはしない |
| 互換・検証 | 新旧client、旧値停止／移行前提、FR/AC→技術→証跡 | 実装未完了と仕様未定を別記 |

<a id="selected-vs-current"></a>
## 選定仕様と現行実装資料の差

| 接続／資料 | 既存資料の記載 | 現在の選定仕様／扱い | 仕様残件・確認先 |
|:---|:---|:---|:---|
| PATH-RES-MR-AUTOCREATE | 予約confirmed/checked_in保存後にseparate-tx best-effortで通常カルテ自動作成 | [D03保存正本](../../spec/domains/medical-records/lifecycle.md)は通常予約時作成なし、受付時同時成功。古い経路を新承認にしない | D01/D02/D03の選定設計と旧caller停止／移行照合 |
| PATH-RES-MR-CANCEL-CLEANUP | 予約取消／削除後にdraft cleanup、separate-tx best-effort | 同正本は通常予約取消からの自動削除禁止。draftというだけで空判定しない | 専用の空・関連なし削除と旧cleanupを分離 |
| PATH-BILL-LSTEP-CPM-SYNC | 会計成功後Created=trueのみ同期、失敗は会計を戻さず、同会計replayでは再同期しない | intentional best-effortを保持。後続同期・停止条件とD15/D10を照合 | 「再送すれば回復済み」としない。実稼働未確認 |
| PATH-BATCH-TAG-HEALTH | implemented capabilityだがproduction scheduler/cron未接続、戻り値error | durable scheduled batchと呼ばない | D15の導入／配信再開を別扱い |
| support write-owner行 | 全医院報告を全staffが更新・削除・Plane起票可という記載 | FE/BE mapは共有READ、報告医院mutation、他医院404と記載 | [service](../../../backend/internal/support/service.go)／[repository](../../../backend/internal/support/repository.go)のFindByIDForClinic／clinic_id条件を読取確認。報告医院mutationが一致し、ownerカタログは旧誤記として修正対象。自動外部通知の新方針とは別 |
| RequestID | Gin context／応答header／完了logに保持 | request Contextへ自動伝播しない既存制約を保持 | 全downstream logへ存在すると主張しない |

当表は既存カタログ本文の全面書換えを承認するものではない。統合時に古い現行実装記述へ「選定仕様との差」を明記し、原子性・監査・回復の重要例外を消さない。

<a id="requirement-fr-common-tech-evidence"></a>
### FR-common-tech-evidence — 文書検査と証拠の境界

[testing](../../ops/testing/README.md)のL0–L5入口と実行記録を使う。文書検査は相対リンク／アンカー、FR/AC一意性とtrace、本文／既存ID保持、表構造、明示scope、source時点を扱う。コードsymbol drift・88route実行検査・runtime・性能・全make ciは実施していなければ未実施とする。過去83route実測・85route E2E inventoryや個別PASSを現在88route全件PASSへ拡張しない。

## 受入条件

| 条件ID | Given／When | Then／独立確認・成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-common-tech-owner | 他領域の事実を変更 | owner intent／interface・tx主体が特定できる | owner catalogと各D設計 |
| AC-common-tech-same-tx | 必要監査／依存が失敗 | integrity経路は対象writeと一緒に失敗し、成功を偽装しない | 各pathのfail-closed AC |
| AC-common-tech-partial | documented best-effortの副作用失敗 | 部分成功・回復主体・監査／観測が区別できる | catalogと最新選定仕様 |
| AC-common-tech-wire | 未送信／null／旧値が届く | ownerごとの保持／clear／拒否条件とwire根拠がある | API/DTO/技術trace静的照合 |
| AC-common-tech-replay | commit後応答喪失・再送・payload相違 | 同一効果と新規操作を判別し、競合復旧の期待結果が記載される | 各D receipt／lock契約 |
| AC-common-tech-selected | 旧auto-create／cleanupを見る | 現行実装資料と受付時原子性／自動削除禁止の選定仕様を区別 | 本表＋D01/D03正本 |
| AC-common-tech-exception | support／global master／jobを扱う | 通常clinic-scopeを乱暴に一律適用せず、正しい例外根拠を確認 | C04/D23／ADR／owner |
| AC-common-tech-doc-test | docs-only検査を報告 | 実施内容・基準commit・未実行チェックを明示 | 検証manifest／独立レビュー |
| AC-common-tech-owner-coverage | 全D候補を統合 | 上記接続項目の具体化とFR/AC traceが全ownerで確認済み | 全ownerの意味ID／版・保存／失敗・旧値・受入traceを独立照合済み |

## 未確認・完了境界

本書の技術契約と全Dの意味型・ID／版・tx／lock・通知・失敗回復・互換をowner対応表とX接続で最終照合した。D04／D06、D11〜D13、D22／D23の操作別保存・版・receipt／結果不明などの不足は本文とACへ具体化し、独立レビューで未解消の重大指摘なし。物理名や実装方式の確定だけを後続へ分離し、呼出側／受け手が意味契約を推測する穴を実装対象外として隠さない。実装・API/DB・設定変更、稼働再開、実機／性能／法的適合の実行は別工程。

## 全ownerへの適用照合

[全23領域の具体ID・保存単位・再送／旧値・安全と保持](../../spec/shared/owner-contract-coverage.md)と[横断8接続](../../spec/workflows/reservation-to-record-flow.md#cross-domain-review-20261015)を参照する。全writerの実装適合とは別に、呼出側／受け手が意味契約を推測しない文書境界を照合する。

<a id="posthumous-operation-contract"></a>
## 追加承認の死亡後記録と専用操作の接続

[C03患者状態と死亡後記録](../../spec/shared/clinical-safety.md#requirement-fr-common-safety-patient)のAC-common-safety-posthumous-historyは、D03〜D10の当該操作の期待版・現在認可・対象元事実・理由／監査・結果照会・原子receiptへ接続する。死亡状態を生存へ書き換えて通常新規ゲートを迂回せず、元患者・生前実施事実・原記録を保った専用の後日入力／誤記訂正／退院／精算として識別する。必須履歴を保存できない操作を成功にせず、結果不明時は同じ操作を照会し、新しい操作として二重実行しない。死亡後の新しい臨床実施や未来予約、金銭・在庫・通知の自動作用を追加しない。物理schema／wireの実装状態とは分離する。

[追加9点の進捗](../../spec/domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-nine)の業務承認と文書再整合を別に追跡する。過去の文書完成・静的検査合格から、既存記載全ての業務承認を導かない。

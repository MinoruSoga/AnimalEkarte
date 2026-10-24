# EMR-506 受領・配賦と金銭誤記訂正：設計候補・未採択

<a id="compatibility-scope-20261024"></a>
## 2026-10-24 現在の適用範囲（506）

旧Complete初回201/再送200・Billing整数wireの維持とその互換fixtureを新仕様の必須条件から外す。確定APIと必要callerの同時置換へ接続し、現在認可・元target anchor・生前精算・金銭/監査/原子receiptを維持する。本書の一般506 wireは引き続き設計候補であり、本追補でHTTP/DTOを新たに選定しない。 詳細根拠は[全票監査の採用是正方針](backward-compatibility-audit-20261024.md#accepted-corrections)。以下の旧日付の記録はこの限定追補と区別する。製品置換・実行受入済みではない。

2026-10-22。照合main=`a5b4a9553b3f003e90ea47660a109c5d526be59c`。本稿のroute/新DTO/金額・ID wire/CAS tokenは**設計候補・未採択**。既決業務と既存保存型を静的照合した残技術論点の一覧で、closed wire完成・実在API・506全体完了とは扱わない。実装/API YAML/DDL/model/生成/設定/Planeは変更しない。tests/build/codegen/DB/Docker/browserはNOT_RUN、merge・レビュー依頼なし。

## 1. 正本・目的と既決境界

業務判断窓口はMinoru Soga。[会計業務正本](../../spec/domains/accounting/lifecycle.md)の実受領/部分入金/過受領/元飼主帰属/金銭誤記訂正、[会計技術正本](../../design/domains/accounting/technical-spec.md)、[死亡後精算の限定参加選定](emr-540-308-462-506-selected-handoff-20261022.md#posthumous-selected-participation)と[C06](../../design/shared/operation-participation.md)を維持する。既決を未決へ戻さず、現旧Payment/Split writerを最終fund仕様へ昇格しない。

目的は実受領・請求配賦・実返金・入力訂正を二重計上せず、一度の明示操作の保存結果を再取得できるようにすること。二重受領/転記/架空の差額調整を除く。業務メトリクスの実績測定は未実施。今回の計画指標は下記2familyと未選定論点の保存・認可・CAS・復旧・受領先対応があること。未決方式を実装/自動化しない。

- 受領と複数請求配賦を同じ操作へ含める候補と、方法/日時の金銭誤記訂正候補を分ける。実授受なしの訂正を実返金/再受領として記録しない。
- 非現金5,000受領・3,000充当・返金必要2,000は既決の有効例。配賦合計=受領額という制限を追加しない。未配賦差額を自動預り金/返金済みにせず、実返金/了承付き預りは各既決の別操作へ。
- 元owner/元請求/元資金を保持し、現在主飼主へ資金を付け替えない。停止方法は新受領465 guard、既資金利用/誤記訂正へ新受領active要求を機械転用しない。
- lookupは認証医院・同一actor・保存元対象anchorへの現在`accounting:view`。mutationは既存金融write grant、勤務consumerの元write grantをlookupへ追加しない。失権/元対象解決失敗は非開示、incoming targetで補修しない。死亡後専用namespaceを作らず通常506 operationの原子的receiptへ参加する。
- 既存462 `POST /api/v1/accountings/complete`の整数金額wire、初回201/再送200・Billing responseを維持する。新候補の小数字句/ID型を既存Completeへ一括適用しない。

### 1.1 2026-10-23の限定業務承認（技術候補の全面採択ではない）

[業務正本の追加承認](../../spec/domains/accounting/lifecycle.md#accounting-cash-method-correction-unknown-pair-20261023)により、実受領額が既知でカード等から現金へ方法誤記訂正するとき、当時の預かり額/釣銭額は両不明nullのまま理由付きで訂正可。理由必須、原記録/訂正履歴保全、不明額0補完なし、新金銭移動なし、実受領額変更なし。方法訂正だけで実授受日時を変更しない既決条件を維持する。新受領の現金入力要件を緩和しない。

この限定業務判断は本人確認済みで再質問しない。§3–4の詳細wire/reader/CAS/死亡後adapter・残る方法組合せ等の技術候補は引き続き未採択。DB CHECKの両NULL許容から業務許可を推測したものではなく、上記本人承認が根拠。今回はその同期のみ、実装/API YAML/DDL/生成/試験なし。

## 2. 現mainで存在するもの・未接続

| 根拠path | 静的観測 | それだけでは証明しないもの |
|---|---|---|
| [storage_finance.go PaymentFund](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/model/storage_finance.go#L567) | received_amountはint64、方法ID/日時/actor/operation、cash2列、original_source_referenceを保持。fund version列はない | 新fund専有route/全writer・read DTO・CAS tokenの完成 |
| [FundRecordCorrection](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/model/storage_finance.go#L499) | revisionと訂正方法/日時/金額/現金対、before_values/理由/operationがある | fund全体の版や全配賦writerのCAS参加。方法/日時訂正の全wire採択 |
| [001のfund_correction_source_guard](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/migrations/001_init.sql#L11112) | fund rowをlockしてmax correction revision+1等を検査する | correction revisionだけで全配賦変更/返金/残高の期待状態を表せること |
| [001のcash入力CHECK](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/migrations/001_init.sql#L10915) | 両NULL、又は両非NULLでnet=tendered-changeの形を許す | 新受領全般での両NULL許可。方法誤記訂正の限定許可だけは§1.1で本人承認済み、DDL許容とは根拠を分ける |
| [billing routes](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/billing/routes.go#L133) | completeとcredit-correctionが存在。後者はカード金額訂正の限定経路 | 一般fund/方法/日時誤記訂正・複数請求配賦の完成 |
| [C06 DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/operation/dto.go#L30) | result_refs/committed_versionsはmap[string]string | arrayをそのまま成功refsへ返せること、金融固有refs/key集合の選定 |

masterの`system_key`は現在master識別として参照できるが、fundに保存時snapshotがあるとは確認できない。現master値をoriginalの保存時system_keyと表示しない。方法変更前の事実は保存ID・immutable原資金/訂正履歴から読み、欠落を名称/現在activeで推定しない。

## 3. 2familyの候補と採択が必要な点

以下の`/payment-funds`、`/{fund_id}/record-corrections`、`/accounting-operations`は**候補名**。既存prefixへの収録・method・path parameter型・operationId・intentは未採択で、呼出可能なAPIとして文書/FEへ接続しない。

| family | 候補の最小操作単位 | 未採択の具体契約 |
|---|---|---|
| 受領＋複数請求配賦 | 一回の実受領fund、指定した請求ごとの配賦、残額/返金必要額のreader、必要監査・receiptを同最外TXへ | 候補POST `/payment-funds`のexact request/response、fund/請求/配賦の期待版、配賦順の意味、read DTO、成功refs/versions、現在grant/元anchor |
| 方法/日時誤記訂正 | 元fundを保持し、固定revisionの訂正と履歴reader。実追加授受なしなら移動0 | 候補POST `/payment-funds/{fund_id}/record-corrections`のexact差分/全置換、現金対の扱い、許される日時精度、期待token/全writer参加、成功refs/versions |
| 限定結果照会 | 上記familyの元scope/actor/anchorから現在認可し固定receiptを返す | 候補GET `/accounting-operations/...`のpath/query/intent allowlist・元対象shape・reader。死亡専用lookupを併設しない |

金額/新IDをcanonical decimal stringにする案は、binary floatを通さず厳密境界を作るための候補。storage int64で許される値域・整数円と新wire文法の整合、負/小数/桁超過/先頭0の拒否、既存caller adapterは選定が必要。decimal stringというだけで小数額の新業務を許可しない。旧Complete整数ID/金額の破壊変更を行わない。

C06のrefs/versionsはstring mapのまま。複数配賦をarrayで直置きする案は不適合で、aggregate参照/固定reader又は定義済み文字列符号化等の選択とexact key集合が必要。どれかを本稿で採択したとはしない。

## 4. 残技術論点と受領owner

| 論点 | 固定してはならない推測・次の照合 | 受領先 |
|---|---|---|
| fund CAS | fund.versionは存在しない。correction revisionと配賦/返金/残高の依存状態をどう束縛するかを全writer参加表から選ぶ。独自token/第2counterを未採択のまま実装しない | 506/464/462、必要保存差分だけ446/466 |
| 方法変更時の現金対 | 元/訂正方法の組合せとtendered/changeの保持/再入力/不適用を照合。非現金への変更で旧現金値を黙って消さない。実受領既知・現金への方法訂正時の両不明nullは§1.1で業務承認済み。残る組合せ/技術表現・新受領条件まで一括採択しない | 506、465参照保護との必要接点 |
| 日時 | 受理offset/RFC3339と保存DB精度のcanonicalization、表示timezone、同key fingerprint等値を定める。server時刻補完や丸めで別内容を同一としない | 506/API単一owner |
| read DTO | original/effective/correction revision、方法ID/表示根拠、配賦/返金必要/実返金を区別。現在master system_keyを保存時snapshotと偽らない | 506 reader・専有FE |
| 死亡後typed source | 生前事実・元owner/source/版/理由のtyped reader、現在認可と最外TX再検査。会計日/client flagで生前根拠を捏造しない。通常死亡gate全面解除なし | 506/462、元source ownerの必要reader |
| 配賦writer閉包 | receive/allocation/release/reallocate/refund/訂正/預り参加を一覧化し、fund row・請求/口座のlock/CAS順と期待状態を揃える。helper単体/特定経路だけで完成にしない | 506/464/462、465 method参照lock |
| C06参加・error | Q/VerifiedCommand、保存元anchor、lookup allowlist、Content/Preconditions、strict decode、成功refs/版を単一ownerで固定。記録済みunknownと現在認可後のresult-store障害を区別 | 489/506/API単一owner |

新506のprocessing202/unknown503/validation400/conflict409は先の限定引渡しで挙げた**整合案**であり、一般金融の既採択HTTP表ではない。C06既存意味を維持してAPI ownerがroute別参加表へ固定する。既存Complete 201/200や信用訂正routeを共通化理由で一括変更しない。

## 5. 最初の実装単位とfixture準備（採択前は着手許可ではない）

506-Aは一回の実受領＋指定配賦＋未払再読込、506-Bは方法/日時の誤記訂正一種類＋履歴再読込へ分ける候補。必要共有API・保存・Q/anchorを固定SHAで受領し、全506/全金融Done待ちにはしない。ただし未採択wire/CASを独自stubで先行実装せず、現実装sessionのscopeに自動追加しない。

fixture候補（全NOT_RUN）：非現金5,000/充当3,000/返金必要2,000、部分入金、複数請求配賦、差額を無断預り/返金済みにしない、実授受なし訂正移動0、方法変更各現金対（承認済みの現金訂正/両不明null・理由必須・実受領額/日時保持・0補完/新移動なしを含む）、日時offset/精度/同key、元owner保持、source不明/死亡後、全配賦writer競合、read失敗/失権lookup、同keyreplay/異内容/unknown、監査/receipt失敗全rollback、既存Complete整数wire・201/200保持。parser単体/mock/実DB/実入出金を別証拠とし、未実行をPASSにしない。

未採択事項は本稿§3–4。既決業務に新しい本人回答を要求するものではなく、技術ownerが既存契約と保存経路を照合して閉じる設計残。実医院値・旧source照合や実DB適用は後続の限定受領条件として別管理する。

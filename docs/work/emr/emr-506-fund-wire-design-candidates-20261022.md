# EMR-506 受領・配賦と金銭誤記訂正：候補履歴と採択範囲

<a id="compatibility-scope-20261024"></a>
## 2026-10-24 現在の適用範囲（506）

旧Complete初回201/再送200・Billing整数wireの維持とその互換fixtureを新仕様の必須条件から外す。確定APIと必要callerの同時置換へ接続し、現在認可・元target anchor・生前精算・金銭/監査/原子receiptを維持する。この互換撤回自体はwire選定ではないが、その後の[初回A/B・lookup採択契約](../../design/domains/accounting/fund-receive-method-contract.md)を現在の適用契約とする。 詳細根拠は[全票監査の採用是正方針](backward-compatibility-audit-20261024.md#accepted-corrections)。以下の旧日付の記録はこの限定追補と区別する。製品置換・実行受入済みではない。

2026-10-22の候補整理を基礎とし、現在はA初回受領＋明示配賦、B方法のみ訂正、lookupを技術選定済み。詳細は[採択先](../../design/domains/accounting/fund-receive-method-contract.md)。実装/API YAML/DDL/model/生成/activation/試験の完成ではない。日時・金額訂正/refund等は後続familyとして分離する。

## 1. 正本・目的と既決境界

業務判断窓口はMinoru Soga。[会計業務正本](../../spec/domains/accounting/lifecycle.md)の実受領/部分入金/過受領/元飼主帰属/金銭誤記訂正、[会計技術正本](../../design/domains/accounting/technical-spec.md)、[死亡後精算の限定参加選定](emr-540-308-462-506-selected-handoff-20261022.md#posthumous-selected-participation)と[C06](../../design/shared/operation-participation.md)を維持する。既決を未決へ戻さず、現旧Payment/Split writerを最終fund仕様へ昇格しない。

目的は実受領・請求配賦・実返金・入力訂正を二重計上せず、一度の明示操作の保存結果を再取得できるようにすること。二重受領/転記/架空の差額調整を除く。業務メトリクスの実績測定は未実施。現在の成功単位はA/Bの保存・認可・CAS・復旧・受領先を閉じること。後続familyの未採択方式は実装/自動化しない。

- 受領と複数請求配賦を同じ操作へ含める候補と、方法/日時の金銭誤記訂正候補を分ける。実授受なしの訂正を実返金/再受領として記録しない。
- 非現金5,000受領・3,000充当・返金必要2,000は既決の有効例。配賦合計=受領額という制限を追加しない。未配賦差額を自動預り金/返金済みにせず、実返金/了承付き預りは各既決の別操作へ。
- 元owner/元請求/元資金を保持し、現在主飼主へ資金を付け替えない。停止方法は新受領465 guard、既資金利用/誤記訂正へ新受領active要求を機械転用しない。
- lookupは認証医院・同一actor・保存元対象anchorへの現在`accounting:view`。mutationは既存金融write grant、勤務consumerの元write grantをlookupへ追加しない。失権/元対象解決失敗は非開示、incoming targetで補修しない。死亡後専用namespaceを作らず通常506 operationの原子的receiptへ参加する。
- 旧Complete整数wire/201・200/Billing互換維持は必須ではない。今回A/B選定をComplete全体の採択や実装に換算せず、必要callerを完結単位で置換する。

### 1.1 2026-10-23の限定業務承認（技術候補の全面採択ではない）

[業務正本の追加承認](../../spec/domains/accounting/lifecycle.md#accounting-cash-method-correction-unknown-pair-20261023)により、実受領額が既知でカード等から現金へ方法誤記訂正するとき、当時の預かり額/釣銭額は両不明nullのまま理由付きで訂正可。理由必須、原記録/訂正履歴保全、不明額0補完なし、新金銭移動なし、実受領額変更なし。方法訂正だけで実授受日時を変更しない既決条件を維持する。新受領の現金入力要件を緩和しない。

この限定業務判断は本人確認済みで再質問しない。A/Bの詳細wire/reader/CAS/死亡後adapterは下記採択先に従い、日時・金額変更/refund等の後続familyと分ける。DB CHECKの両NULL許容から業務許可を推測したものではなく、上記本人承認が根拠。今回はその同期のみ、実装/API YAML/DDL/生成/試験なし。

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

## 3. 現在の採択範囲と後続

| family/論点 | 現在の技術選定・受領先 |
|:---|:---|
| A 実受領＋明示配賦 | POST /api/v1/payment-funds、billing.fund.receive。closed Q/厳密Preconditions、fund/請求settlement CAS、元owner anchor、manifest/audit/receipt同TX |
| B 方法のみ訂正 | POST /api/v1/payment-funds/{fund_id}/record-corrections、billing.fund.correct_method。最新effective金額/日時保持、承認済両null、配賦event追加なし |
| lookup/current read | 元actor/clinic/保存anchor・現在accounting:view。成功再送view、未成功Apply write。fund tokenはresult_refs、数値版だけcommitted_versions |
| fund/請求CAS | fund state/settlement stateのexact JCS preimageを採択先で固定。fund.version新設やbilling no-op/timestamp更新による増分なし |
| close/life/source | graph上位gateとowner別下位順、実source区間のtyped resolver。日lockを日全体拒否にせず、Completeのpre-billing close段で根拠受領 |
| 後続 | received_at訂正は別intent、金額変更/release/reallocate/refund/505 cancel-rebill/464 depositは各既存票で別成功単位。同CAS/anchor/TXへ参加後接続 |

具体的な型・制約・API/HTTP・manifest・CAS preimage・lock参加順は[506初回A/B・lookup技術正本](../../design/domains/accounting/fund-receive-method-contract.md)に一本化する。旧候補の互換義務や単一新rankを併存させない。fund/event/correction保存モデルと業務金銭方針は維持。505/481/462/508の既採択必要subsetを受領し、全親Done待ちにしない。全実装・試験・DB操作はNOT_RUN。

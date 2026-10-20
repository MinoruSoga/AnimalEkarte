# EMR-499 → EMR-461 有効状態 subset の実装と未確定case

本表はPR734の実装境界と担当間の技術契約の証跡。新しい業務仕様やDDLの採用ではない。PR740のschema/DTO/routerは別ownerであり、この文書で変更しない。

| 入力履歴 | 今回の判定 | 根拠・残差 |
|---|---|---|
| 実admit（actual_atあり）から終了eventなし | admitted | [開始の明示確定](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-entry-and-start)。予定終了超過は実退院ではない |
| admit → 有効discharge、実退院時刻が入院以上 | released | [明示退院](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-discharge-and-billing)。通常の再check-inは追加しない |
| admit前のcancel/no_show | released | [計数正本](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-hotel-stay-capacity)の取消・未来院除外 |
| no movement（予定entry/exitだけあり） | unknown | **未確定Q1**: stay行の作成だけで予約成立を表すか、予約受入/成立の別参照が必要か。予定日時の存在からreservedを捏造しない。499の明示契約が必要 |
| correction・無効化・訂正参照あり | 当該stayのみunknown | **未確定Q2**: 置換後effectの指定、元effect継承か、無効化、actual_at NULLの意味。corrected_event_id/reasonのDDLだけから推測しない |
| discharge単独、admit後cancel/no_show、退院後再admit、実時刻逆転、scope/ID/revision不正 | 当該stayのみunknown | **未確定Q3**: 専用訂正を通した有効履歴の解決規則。最大revisionのevent_kindでは決めない |

`StayCageEffectiveStateReader` は医院付きStayCaseIDについて状態と`EvidenceVersion`を返す。今回の根拠版は検査したmovement履歴の最新revisionで、CAS/fenceやstay行versionの代用ではない。全履歴をrevision順に検査し、同revision・同ID・scope不整合をunknownにする。予定期間は占有解放に使わない。

## readerとlockの成立範囲

readerはambient transaction内のstay行をID昇順`FOR SHARE`で保護し、同じtransactionでmovementを読む。候補読取にambient transactionがなければ短いtransactionを作り、stay読取とevent読取の間のINSERT競合を避ける。

採用済み001の [`stay_movement_guard`](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L10388) はmovementのBEFORE INSERTで同stay行を`FOR UPDATE`する。このINSERTとの直列化は静的な対応関係であり、実DBのlock/2接続試験PASSではない。

採用済み001には [`stay_movement_events_immutable`](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L10529) のBEFORE UPDATE OR DELETE→`contract_history_immutable`も存在する。DDL上の原event不変条件は確認済みで、実DB適用/受入は未検証。

**未確定Q4**: 499全writerの取得順と、461のhospitalization/cage-group→stay行の順との整合。現reader注入だけで訂正writer全体の安全性を認定しない。

## 現在の未完範囲

通常caseのreaderをcompositionへ注入し、nil依存だけで全stayを常時503にする構成は除去する。一方、Q1–Q3の対象stayはunknownであり、対象ケージの空き確認を止める。現在の一覧処理は対象のstayが1件でもunknownなら、一覧全体を503で返す。既知の他stayだけから部分一覧を返す実装ではない。unknownをstay単位で判定することと、HTTP一覧が部分成功することは異なる。

hospitalization create/update/swap等の旧status/type/start/end利用の499 writer接続、必要認可・監査・receipt、PR740のknown/unknown・current_assignment等へのruntime追随、実DB/API/UI受入は別途必要。DDL変更の必要性はQ1/Q2の意味・表現が確定してから446/466へ具体差分を渡す。今回はDDLを追加しない。


## 担当間で閉じる技術契約と最小の次工程

Q1–Q4は承認済み業務方針を実装へ落とすための技術契約であり、ユーザーへの業務再承認依頼ではない。以下の型名・取得順は検討用の具体案で、現行の採用契約・DDLではない。461は確定した499 reader契約を利用し、各consumerに独自の予約成立／訂正解釈を複製しない。

### Q1: 予約成立を確認する入力（499、予約writer担当、446）

[001のstay接続](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L9247)には`appointments.stay_case_id`と医院付きFKがある。一方、[reservation_contract_states](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L8915)の`enabled`は医院・契約種別のcutover状態で、個別予約の受入成功を示す行ではない。stay作成・予定日時・cutoverだけで予約済みにしない。

次工程は予約writerの採用実装で、同医院のappointment→stay接続が受入成功時にどの状態／履歴／成功receiptと同一TXで確定するかを追跡する。499 readerへの入力案は`AcceptedStayReservationProof{ClinicID, StayCaseID, AppointmentID, AppointmentVersion, AcceptanceEvidenceRef}`。`AcceptanceEvidenceRef`は既存writerが成功確定に使用する履歴・receiptの参照を渡す想定で、存在未確認の新しい受入eventを作る提案ではない。予約writer担当と446が、現在有効な受入と取消・未来院の判別元、参照の版、同stayの重複予約の扱いを既存保存契約から特定し、499が予約成立proofの有無を解決する。証明できた予約のみreservedへ投影するfixtureを追加する。参照先の採用契約が確定するまではno eventをunknownのままとする。

### Q2: 訂正後のeffectを明示する型（499、446、必要時466）

[現movement行](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L9278)は`corrected_event_id`、`actual_at`、`reason`を保持するが、correction行が元effectの時刻だけを置換するのか、kindまで置換するのか、無効化するのかを区別する明示フィールドはない。NULL時刻だけで無効化や継承を判定しない。[専用訂正の正本](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-special-correction)は原記録・変更内容・理由・actor・時刻・関連の保持と頭数再照合を要求する。

499の内部解決結果の型案は`ResolvedStayMovementEffect{SourceEventID, ResolutionRevision, Disposition, ReplacementEffect}`。`Disposition`を`retain / replace / void`の識別子にし、`replace`だけ`ReplacementEffect{Kind, ActualAt, HospitalizationID, AdmissionRecordRef, DoctorID}`を持つ判別共用体とする。これはkindや時刻をNULLから推測しないための表現案で、void操作の新規業務許可を意味しない。499/446が許可済み専用訂正の保存内容と照合して、実際に必要な分岐・訂正対象の重複／連鎖解決・版を確定する。既存保存先で表せない場合に限り、466へ必要列・制約・移行方針を具体化して渡す。461は解決済みeffect列だけを通常遷移へ入力し、raw correctionを解釈しない。

### Q3: 通常遷移の拒否と訂正解決の分離（499）

通常writerは`未開始→admit→discharge`と`未開始→cancel/no_show`を検証し、discharge単独・admit後cancel/no_show・同stay退院後再admit・実時刻逆転を通常書込として拒否する。[退院の正本](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-discharge-and-billing)どおり通常の再check-inは追加しない。既存の異常履歴を新しい通常操作で救済せず、専用訂正の対象として扱う。

499の訂正resolverがQ2の確定契約に従って原履歴を保持した有効effect列へ解決した後、同じ通常遷移検証へ渡す。解決不能・なお不正な場合はunknownを返す。fixtureは「通常writerで拒否」「原履歴を保持した解決済み訂正」「解決不能unknown」を分け、最新kind採用や自動再入院を追加しない。

### Q4: 全writerのlock順と実DB検証（499、461、予約writer担当、466）

現在の461はhospitalization/cage-groupを扱う経路からstayをロックし、movement INSERT triggerはstayを`FOR UPDATE`する。trigger単体の存在だけでは、全writerがその前に取るlockとの循環を否定できない。499と461で全経路（admit/discharge/cancel/no_show/correction、ケージcreate/edit/swap、予約接続）のlock一覧を作り、複数対象のID昇順を含む共通順を1つ決める。候補案は`cage-group ID昇順 → hospitalization ID昇順 → stay ID昇順 → event/receipt書込`で、予約writerなど先にstayを取る経路と突き合わせるまでは採用しない。

466の実001 fixtureが利用可能になった後、同医院2接続で候補読取対movement INSERT、ケージ保存再検査対訂正、逆順のswap、予約接続対開始を実行し、ロック待ち後に再検査されること、deadlock／stale stateで保存成功しないこと、rollback後に変更とreceiptが残らないことを確認する。immutable triggerとRLSの実適用も確認する。現在のcontext fixtureやfile-list PASSをこの実DB検証の代用にしない。

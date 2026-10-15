# D07 検査・機器受信の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D07-examinations-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。新規の未決事項は未決のまま |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 対象と証拠
業務正本は[検査ライフサイクル](../../../spec/domains/examinations/lifecycle.md)。2026-10-15、固定commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`の文書／source静的照合。API／DB／実装／実grantは変更せず、実機・並行性・性能未検証。既存実装とD03継承目標を分ける。

<a id="examination-state-contract"></a>
## 状態・版と保存単位
[lock](../../../../backend/internal/medicalrecord/examination_lock.go)、[revision](../../../../backend/internal/medicalrecord/examination_revision_service.go)、[更新](../../../../backend/internal/medicalrecord/examination_service_update.go)を参照。
- confirmedは全更新ロック。completedはcurrent_revision_versionなしなら結果更新・削除不可、ありなら作業版である可能性を保持する。
- 初回confirmはofficial revisionをappendしCASでpointer更新。unconfirmは理由必須（trim後非空、500文字以内）、認証actor、専用examination-unconfirm:edit、official版あり、親draftを要求しworking版をappendする。
- revisionあり再confirmは変更を先に保存。変更field/itemsと再confirm同送は拒否。修正結果と公式版を別にする。
- revision存在後はpatient／recordの関連変更を拒否。legacyでAPIが関連PATCHを受理できることはD03の患者取り違え訂正の業務許可ではない。
- result／revision／必要監査の同一tx成功、競合CAS失敗時rollbackを目標とし、現revision実装だけでD03全writerが適合済みとはしない。working updateの理由は現実装では内部固定文言を使う部分があり、業務上の訂正理由全体の充足とは別。

<a id="examination-wire-contract"></a>
## 型・未送信／null／旧互換
| 対象 | 現契約 | 目標・停止条件 |
|:---|:---|:---|
| exam create/PATCH | [request](../../../../backend/internal/medicalrecord/examination_request.go)：数値ID、date日時、status5値、result_summary／machine、items pointer | 医院＋IDを各owner refとして扱い、JS安全整数外・不正IDを0にしない。担当医と認証actorは別 |
| PATCH items | 未送信／nullは不変更、[]は空置換 | 意図を保持。空payloadを全削除扱いしない |
| 旧PUT /items | 未送信／nullも空配列化する | D03継承目標は未送信／nullで消去しない・明示[]だけ。旧client切替前に適合済みとしない |
| 項目 | field ID、name、inspection_value、normal_value、unit、reference_value、sort_order | 原値を文字列保持。status/is_abnormalはrequestで信頼せずサーバ導出 |
| 基準値 | [items](../../../../backend/internal/medicalrecord/examination_items.go)で医院／種／fieldを解決しrefMin/Max・定性範囲をsnapshot | 取得失敗と範囲なしを分離。保存snapshotを当時値として使用。医師の具体値承認とは別 |
| 結果ref | clinic＋exam ID、current_revision_version、revision item | examは結果全体、版は固定、itemは版内。跨版IDの同一性を仮定しない |
| 旧版 | legacy confirmedはofficial projectionなし | 最新値からofficialを架空生成せず、旧資料照合まで専用解除・追記adapterを停止 |

[現items service](../../../../backend/internal/medicalrecord/examination_items.go)は型・所属照合→範囲解決→置換・削除監査を同一txに置く。純挿入の監査やD03が求める全変更ログの充足は個別確認が必要。item取得や詳細取得はlab usage receiptを記録し得るため、完全無副作用readだと仮定しない。

<a id="examination-device-contract"></a>
## 機器・ジョブ・互換経路
現行運用は[接続正本](../../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)と[ADR008 current daily path](../../../architecture/adr/008-local-lab-device-agent.md)。日常はbrowser direct Web Serial、NX600/AU10Vのみ・ページ表示中・origin-wide排他。legacy loopback agentは任意手動経路であり、旧raw token取得説明を現行API仕様にしない。現agent-consumerは短命signed capability、署名鍵を返さない。

| 論理操作 | 保存／境界 | 失敗・再送 |
|:---|:---|:---|
| receive | 1frame=1job、医院/source/fingerprint一意、decoded items保存、生bytes破棄 | 重複は既存job。invalid_payloadは保存成功カードにしない |
| wait | 医院1有効待機、期限／変更／解除、現在生存pet | 期限切れ到着は未紐付け、前患者を再使用しない |
| persist | 保存済みjob/items＋医院内mapping、N exam_typeならN exams | 未対応行をjobへ残す。receivedとpersistedの成功を分離 |
| attach | jobをlock、対象petと既存関連を再照合しpersist | 他患者既存関連は409、結果不明はjob照会 |
| detach | 使用履歴／確定／親保護を照合、job由来全検査の補償 | 安全ならreceived／pet NULL、revertedへ終端化しない |
| fixture commit/revert | fixtureのみcommit、完全一致skip、revertは終端補償 | deviceをfixture allowlistへ追加しない、auto_commit既定false |
| mapping master | clinic/source/device_item_code→field→exam_type | unknownはneeds_review。名称一致やLLM候補のみで確定しない |

[受信service](../../../../backend/internal/medicalrecord/lab_device_receive_service.go)のreceive後persist失敗にはjob pet解除の別処理があり、受信・保存全体atomicと記載しない。GET board／stationはensureStationによる初期化や待機期限処理を伴い得るため、writer棚卸しに含める。通常view専用印刷projectionにこれらを流用しない。

## 権限・全writer
| 経路 | 既存grant／owner | 継承する照合 |
|:---|:---|:---|
| exam CRUD／items／print | examinations view/create/edit/delete、medicalrecord write owner | clinic/pet/record関係、検査ロック、親finalized、必要ログ |
| unconfirm | examination-unconfirm:edit | actor・理由・official版・親draft |
| frames／wait／board | lab-import:create | trusted clinic、alive pet、受信consumer排他 |
| unlinked／station／masters read | lab-import:view | 医院scope、閲覧可能範囲 |
| attach/detach/revert／機器・item master更新 | lab-import:edit | job source・使用receipt・全由来検査・親依存 |
| 検査type／field／基準範囲 | master-medical操作別grant | 同医院／種／field、旧snapshot保護 |
| D03専用訂正／追記adapter | D03 intentとD07内部を接続 | 両方の現在grant・固定版・理由・結果ref、通常PATCH代用なし |

現在のstaff scopeはJWT snapshotだけで決めず、current access再解決とX-Clinic-ID検証のtrusted scopeを使う。医院／患者／認証世代が変わった遅い応答を別コンテキストへ反映しない。ロック順はD03の全writer規約と整合するよう後続実装で統一確認する。検査内部のexam-first経路と他owner parent-firstが存在するため、静的表だけで全体deadlock-freeを宣言しない。

<a id="examination-owner-contract"></a>
## D03／読取ownerへの返却
検査結果の受渡し先はD03とD21の認可済みスタッフ読取であり、今回D20へ検査結果を公開しない。D20の限定健康表示とD21のスタッフ専用Owner Report閲覧認可を区別する。D21への返却に飼主公開や新しい公開承認操作を含めない。
返却は医院、患者／カルテ対応、exam ID、固定revision、版内項目ref、状態、原値・単位・保存範囲、変更前後／理由／actor／時刻、取得状態。取得不能・非開示・旧版不明は区別する。D03の[最終writer／専用intent](../medical-records/technical-spec.md#medical-record-writer-matrix)へ接続し、draftは許可された関連保存、finalizedは原記載と確定時参照版を保持する専用結果追記へ渡す。現カルテExamGroupがrevisionを保持しない差、addendumが構造化結果refを持たない差を残す。lab usage receiptは転記／復元／公開のapproval receiptではない。

## AC追跡
以下は業務正本の全36ACを技術契約へ対応する索引。前提／期待結果の正本は業務側のGWT。
| 技術部分 | AC suffix |
|:---|:---|
| 入口・型・scope | entry, scope, required, value, unassessed |
| 状態・版・必要ログ | confirm-lock, legacy-seal, unconfirm, unconfirm-denied, reconfirm, finalized-parent, cross-version, write-failure |
| 日常受信・待機 | browser, lock-denied, unsupported, wait, wait-expired, specimen |
| mapping・原値 | clock, unknown-code, multiple-types |
| 冪等・補償 | receive-duplicate, fixture-difference, attach, detach, detach-used |
| 失敗・再開 | frame-loss, result-unknown, autocommit |
| master・医師確認境界 | master-grant, master-history, clinical-approval |
| read・projection | list-scope, print, owner-handoff |

## 利用開始前に残す項目
基準値具体値の医師確認、医院実grant、臨床旧版・item対応、receipt／全監査／競合adapter、物理配線・UAT・PIMS運用承認・医院rollout、旧PUT互換切替は未確認。文書の整合をこれらのPASSにしない。追加の製品要件・自動受信対象拡大は別の判断を要する。


D21 Owner Reportはスタッフの閲覧専用であり、同画面からの印刷・共有を設けない。D03や検査自ドメインに既存の印刷経路があることを、D21経由の印刷許可へ拡張しない。

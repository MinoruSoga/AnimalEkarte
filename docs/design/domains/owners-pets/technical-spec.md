# 飼主・ペットの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D04-TECH / 技術設計・現行差分 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存契約の継承と[追加合意4・5](../../../spec/domains/owners-pets/lifecycle.md#agreement-owner-contacts-transfer-20261015)の実現目標。[台帳・会員区分・LINE希望の共同確認](../../../spec/domains/owners-pets/lifecycle.md#agreement-owner-membership-dm-20261015)を追加。具体的wire／DB変更は未実施 |
| 実装状態 | 下記限定ソースを2026-10-15に静的照合。全writer・HTTP型・DB・実権限の再監査は未実施 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-17追加差分（従来確認2026-10-15） |

業務FRとACは[台帳lifecycle](../../../spec/domains/owners-pets/lifecycle.md)が正本。画面詳細の重複、API YAML・DB定義の複製をしない。wireは[API](../../../../backend/docs/api.yaml)、DBはmigrations、writer境界は[カタログ](../../../architecture/model-write-owner-catalog.md)。

## ID・write owner・保存単位

| 対象 | 現行owner／入口 | 保存単位・注意 |
|:---|:---|:---|
| 飼主 | owner / owners CRUD | clinic+owner ID。owner_numberはID表示別名 |
| 患者台帳 | pet / pets CRUD | clinic+pet ID。pet_numberは別管理番号。名前で連結しない |
| 飼主と複数ペットの新規登録 | owner.CreateWithPets→pet.CreateForOwnerRegistration | ambient tx必須、全write＋response reload成功でcommit |
| 主飼主変更 | pet.Update→reservation.TransferUpcomingPetReservations | 現行はpet、対象予約、移管監査がsame-tx。目標は影響一覧確認を同じ対象の保存へ結び付ける。下節で現行と目標を分離 |
| 副飼主 | pet.PetOwnerService | version付き全置換、リンク差分の必要監査same-tx |
| 全医院動物種 | pet.AnimalSpecies | system-admin限定、CRUD/reorder＋監査same-tx |
| 死亡／復活 | 専用HandlePetDeath/HandlePetRevival | generic UpdatePetInputにstatusなし。D15の外部連携と区別 |
| 配信除外／注意／LINE ID／送信／共有 | owner／LINE専用経路 | 通常飼主form全体のrollback対象ではない。外部送信はcommitと到達を別結果 |

死亡状態を保つ生前診療の後日入力・誤記訂正・退院処理・精算は、追加承認された[各領域の専用経路](../../../spec/shared/clinical-safety.md)で理由と履歴を残す。これらを死亡／復活writerへ変換したり、generic status PATCHを解禁したりしない。台帳deceasedを維持し、未来の新規予約・実施の死亡ゲートと区別する。

[PATH-OWNER-PET-REGISTER](../../../architecture/cross-domain-orchestration-catalog.md)はsame-txだが必須audit writeを持たない。全台帳更新に監査があると一般化しない。新たに必要監査を増やす際は保存方式・移行・権限とともに設計し、現状の証明に使わない。

<a id="owner-transfer-current"></a>
## 主飼主移管の現在の限定契約

静的参照：[pet service](../../../../backend/internal/pet/service.go)・[reservation移管](../../../../backend/internal/reservation/pet_owner_transfer.go)。後者だけ現head da38895d21d60085b3190c785660a2867b81be26（blob 568e0581bc96e5670e432cc50054df697eed040f）で読取再確認した。以下のpet側・監査・commit後処理は従来照合を継承し、全writer再監査ではない。

1. 同医院の変更先ownerを検証し、現在の副飼主でないことを確認。pet行lock後も主副重複を再確認。
2. 対象予約は同医院・同pet・JST当日00:00以降・pending/confirmedのみ。現在の実装では最大1000件、超過は全体拒否。
3. 予約側で新ownerをSHARE NOWAIT、対象予約をID順UPDATE NOWAITで取得。対象ownerが旧ownerと違う／NULL、関連カルテがある場合はConflict。
4. pet変更と対象予約owner更新をsame-tx。元owner、新owner、移管appointment IDs、基準日を必要監査へ記録し、監査失敗はrollback。
5. 元予約がLINE customer付きの場合だけ、新ownerに確認済LINE IDと同医院で一意に対応するcustomerがあればそのIDへ。なければNULL。LINE予約のcustomer_fieldsは空にする既存処理。manual予約（元line_customer_idなし）はNULLと既存customer_fieldsを保持する。氏名類似・未確認IDを使わず、元から未リンクの予約を自動新規リンクしない。
6. 過去予約、checked_in以降、cancelled/no_show/tentative、カルテ、過去請求、預り金、identity groupによる連携はこの移管の対象外。予約source・staff・schedule/status・provenanceを保全し、現行writerはowner_idとupdated_atを更新する。選定仕様の予約版保護・receiptへの適合は未確認。
7. 次回来院等のtag syncはcommit後best-effort。tag失敗を台帳rollbackと誤認しない。D15の停止・再開許可を本整理から変更しない。

これは既存実装の照合結果で、今回この全挙動を新規承認したものではない。D01/D18/D20の合意正本と差があれば現行差として明示し、移管対象を拡張しない。副飼主解除と主飼主変更の2回送信は一体transactionではなく、片方だけ成功した場合は同じpetで現在状態を照合する。

<a id="owner-contact-duplicate-target"></a>
## 電話・メール重複の目標と現行差

追加合意4の目標は、同医院の連絡先重複を警告・候補提示に変え、別ownerとしての登録を可能にすること。従来のBE409重複拒否は目標と異なる現行説明として残す。画面の409握り潰しや別の連絡先への書換えで成功を装わず、owner登録writer・既存制約・応答を整合させる必要がある。現在のDB制約／全caller適合・実装変更は未検証。

- 候補は現在認可された医院・閲覧scopeのowner IDに基づく参照結果。候補提示と既存ownerの採用・別owner登録を混同しない。自動名寄せ・LINE本人確認・identity group操作へ接続しない。
- 重複検出の新しい正規化方式・閾値・類似度・追加必須確認の選定は行わない。一覧検索のカナ／空白／部分一致規則を連絡先重複判定へ流用したと断定しない。
- 登録scope・他項目の形式検証・owner＋全petのsame-txを保全。結果不明では元操作を照合し、同じ電話／メールの候補があるだけで保存済み又は未保存と判定しない。別ownerでの無条件再createは不可。

<a id="owner-transfer-target"></a>
## 主飼主変更の影響確認と保存目標

追加合意5の確認は、petを保存してから影響を知らせる方式ではなく、変更前の影響一覧を確認して同じ変更を確定する方式とする。現在の会員区分／値引率だけの警告は十分な実装証拠ではない。新しいAPIパス・物理schema・確認token形式は未選定だが、以下の安全条件は保持する。

1. 同医院・pet ID・旧owner ID・明示選択した新owner IDを固定し、reservation ownerから現行範囲の対象を取得する。予約ID／日時／状態、変更前後のowner、LINE付替え／解除／変更なしの影響をスタッフに判別可能にする。確認済みの宛先がない状態、対象0件、取得失敗を分ける。生のLINE認証情報を表示・監査へ追加しない。
2. 一覧取得でwrite・外部送信を行わない。確認された対象集合・帰属・宛先解決・必要な版を保存操作と結び付け、保存時に現在認可、医院／pet／owner、予約状態・関連カルテ、LINE確認／一意対応を再検証する。影響が変われば原記録を変更せず再取得・差分確認し、期待版だけ更新した無警告再送をしない。
3. 予約技術正本の[共通gate・取得順と参加契約](../reservations/technical-spec.md)へ従い、pet行・owner保護・対象予約NOWAIT、切替検査、必要監査same-txを保全し、選定済みの予約版保護を満たす。現行移管writerでの明示的な予約version増分は確認できず、updated_at更新と区別する。pet→対象予約owner／LINE参照→必要監査の一部だけを確定しない。監査は既存の元／新owner ID・appointment IDs・基準日の範囲を維持し、氏名・電話・LINE identityを追加しない。
4. 新飼主に確認済み・同医院で一意のLINE customer対応がない場合、既存LINE予約は現行の解除結果を影響一覧へ出す。未確認ID・氏名・共用電話／メールからcustomerを推定せず、未リンクmanual予約へ新規リンクを作らない。LINE本人紐付け・D15の停止／除外／結果不明保留をこの確認で上書きしない。
5. 過去請求・預り金は元clinic／owner IDのまま、残高・金銭履歴を変更しない。過去予約・カルテ・tentative等の対象外履歴もそのまま保持する。新規リンク発行・過去通知の送信先変更・結果不明通知の再送はこの操作に含めない。
6. commitと外部通知到達は別結果。既存tag syncはcommit後best-effortで、通知失敗による台帳再移管・再createをしない。保存結果不明は同一操作識別／内容で照合する。永続receiptの実装適合は未確認であり、電話・メール候補や通知送信結果を台帳receiptの代替にしない。

[FR-owners-pets-primary-transfer](../../../spec/domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-primary-transfer)の実装／実機受入は未実施。max1000件等の現行制約を新しい業務上限と承認したり、影響一覧の導入から移管対象や通知の送信権限を拡大したりしない。

<a id="owner-membership-dm-contract"></a>
## 会員区分・DMの保存とD15への受渡し

D04がclinic／owner IDに属する会員区分とDM希望の正本を持つ。表示ラベル、ペットの生死、個別値引率、独立した停止／配信除外、LINE本人紐付けを別に扱う。副飼主リンクの追加もこれらのwriterや通知先変更を呼ばない。会員区分変更だけで価格再計算・予約可否変更・LINE配信可否変更を発生させない。退亡者と死亡を相互変換せず、他診/準から未知の動作を導かない。

DM未設定を必要へ自動移行・補完しない。未設定から必要への保存は希望確認後の明示変更として扱い、取得失敗を未設定・必要のどちらにもすり替えない。保存は既存owner編集認可と医院scope、元表示版／現在版、操作結果照合へ参加し、競合時は原値を保持して再取得・差分確認する。省略した別項目や他者変更を上書きしない。保存成功と外部配信は別で、値変更だけから送信・過去結果不明の再送・独立停止解除を起動しない。

D15は送信対象の現在clinic／owner・現在DM値と[D15目的別判定](../../../spec/line/lstep-integration.md#crm-dm-purpose-contract)を送信前に照合する。旧LstepOptOut／delivery_excludedとの対応を今回新設・同一視しない。wire field・移行の具体形は実装設計で照合し、現行ゲートがこの目的別契約へ適合済みとは報告しない。主飼主移管の明示影響確認やD10過去資金帰属は保持する。

## 更新・null・版・参照保護

| 項目／経路 | 現行契約 | 誤用の防止 |
|:---|:---|:---|
| 名前由来／出逢い | 省略=保持、null=clear、空白はNULL。500／2000文字 | 空GET結果をclear payloadへしない |
| 飼主生年月日 | 消去はJSON null、サイグラムは派生表示 | 生年月日からの計算正本を保持。結果DB保存を採用しDOB正本・変更時再計算・ルール版保存との整合性を維持 |
| 副飼主PUT | FEは取得version付き全置換、409は再読込 | service内部Versionはoptionalな互換型。HTTPで全caller必須を保証済みとはしない |
| owner削除 | 主飼主ペットcount＋副飼主count保護。productionはguard付き構成 | fallback構成の副飼主保護欠落を「全経路保護済」にしない |
| pet削除 | 関連カルテcount確認後delete、commit後tag sync | 全専用記録／FKを網羅した削除保証は未確認。臨床記録削除を案内しない |
| owner登録／基本更新 | 現行説明では重複メール／電話の検査と原子的更新経路 | 登録時の重複拒否は追加合意4の警告・候補提示・別登録と差がある。UI noValidateは他のBE検査免除ではない |
| 台帳一般PATCH | 現在全項目に共通revision／receiptがあるとは確認できない | 編集者間上書き防止の新方式を既実装扱いしない |

## 取得・通信・再送の設計境界

- 所属医院・owner/pet ID・現在認証・画面generationを取得／保存contextへ固定。医院切替後の遅延応答は表示stateへ反映しない。
- 一覧はURL queryをSoT、300ms debounce、サーバ20件ページ。検索結果の0件、読込中、取得失敗を分ける。検索条件変更でpageの妥当性を再確認し、別検索の応答で上書きしない。
- 新規の送信前／送信中／成功／明確拒否／結果不明を分け、成功対象だけdirty解除。結果不明は同じ操作・対象を照合する。メール／電話一致を保存receiptの代替にはしない。
- 現行新規APIに永続operation receipt／一意な再送キーの完備は未確認。安全な再送にはサーバ側照会可能な操作識別とpayload対応を設計・実装検証する必要がある。現状のGET0件だけでは遅延commitを否定できず、自動再createを止める。
- 期待版のない既存PATCHを「楽観lock済」としない。副飼主と主飼主移管の具体的競合保護を他の一般属性へ拡張して報告しない。
- UI離脱取消は未送信draftだけに効く。送信済mutation・専用LINE送信を取り消したことにはならない。

## 受け渡しと検証対応

| FR ID | AC ID（全件・一意） | 技術対応 |
|:---|:---|:---|
| FR-owners-pets-register-scope | AC-owners-pets-register-atomic / AC-owners-pets-scope / AC-owners-pets-optional-clear / AC-owners-pets-unknown-result / AC-owners-pets-stale-update | owner→pet typed draft／同一tx・所属検証 |
| FR-owners-pets-contact-duplicate-warning | AC-owners-pets-contact-duplicate-warning / AC-owners-pets-contact-duplicate-separate | 警告候補と別登録の分離・医院／認可scope・独立owner ID・receipt非代替 |
| FR-owners-pets-identity-and-search | AC-owners-pets-search / AC-owners-pets-checkup-boundary / AC-owners-pets-read-only / AC-owners-pets-cross-clinic-list | 検索query／ページ・権限・遅延応答分離 |
| FR-owners-pets-primary-secondary | AC-owners-pets-secondary-replace / AC-owners-pets-secondary-conflict | 主副重複／全置換version・監査 |
| FR-owners-pets-primary-transfer | AC-owners-pets-owner-transfer-confirm / AC-owners-pets-owner-transfer-atomic / AC-owners-pets-owner-transfer-impact / AC-owners-pets-owner-transfer-impact-stale / AC-owners-pets-owner-transfer-line-boundary / AC-owners-pets-owner-transfer-history | 確認済み影響と保存対象の一致・現在認可／版・pet→reservation typed intent・JST対象／同一tx監査・LINE条件付き移管／元資金保持 |
| FR-owners-pets-safety-and-private-notes | AC-owners-pets-danger-privacy | 死亡専用writer／staff・公開DTO境界 |
| FR-owners-pets-delete-dependencies | AC-owners-pets-delete-guard | 参照依存／不明停止・原記録保持 |
| FR-owners-pets-species-master | AC-owners-pets-species-authority / AC-owners-pets-species-reference-audit | global species／system-admin・参照拒否・監査 |
| FR-owners-pets-membership-dm | AC-owners-pets-membership-labels / AC-owners-pets-membership-independent / AC-owners-pets-dm-line-preference / AC-owners-pets-dm-unset-confirmation | [会員・DM保存と受渡し](#owner-membership-dm-contract)：独立値／現在版・認可／希望確認／送信目的の再照合 |

参照ソース：[owner service](../../../../backend/internal/owner/service_core.go)、[副飼主service](../../../../backend/internal/pet/pet_owner_service.go)、[同時登録カタログ](../../../architecture/cross-domain-orchestration-catalog.md)。確認refは41fcf3eb0a8befb7c0b5f613d1df749a60e30be8。動作テスト・実DBの既存不整合・全writer依存・移管時の外部通知の実稼働は未確認。

## 文書整理の完了境界

先行整理は既存3画面・台帳FR9件／AC27件・語彙・保存単位・移管／副飼主／死亡writer・主要異常系・他ownerへの引継ぎを記述した範囲。追加合意4・5の目標と現行差を反映し、既存IDは保持する。独立レビューとリンク確認を経て判定する。実装／receipt導入・全参照ガード・認可設定・保存期間・全体横断X04/X06の完了を本書から宣言しない。


## 一般更新の同時編集契約（選定仕様）

元表示版と保存時の現在版を比較し、不一致なら現在の原記録を変更せず競合を返す。入力を保持し、再取得した現値と差分を表示してスタッフの再確認を受ける。期待版だけを新版へ置換して無警告再送しない。明確拒否と結果不明を区別し、結果不明は元医院・対象ID・同一操作識別／内容で照合する。別操作IDや別createで救済しない。既存一般PATCH/UPDATEの期待版・永続receipt対応は実装差であり、この選定動作の実装済みを意味しない。

<a id="owner-infection-summary-read-contract"></a>
## 患者トップの感染症結果：D07読取adapter

[追加FR](../../../spec/domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-infection-summary)はD04の新しいスタッフ向けread consumerで、結果のwrite ownerはD07のまま。D04取得contextのtrusted clinic・pet・species・現在認証世代を[D07 projection](../examinations/technical-spec.md#examination-infection-summary-contract)へ渡す。owners:viewに加えD07結果閲覧認可を検証し、台帳DTOを持つだけで検査結果を取得可能にしない。

D07が返す原値・各検査日／精度・院内／他院・固定結果／item ref・版／訂正状態・曖昧さ／取得結果をそのまま表示する。台帳の備考・danger_level、保存／更新日時へ置換しない。未取得／error／権限不足を未検査・陰性へ変換せず、医院／患者／種／認証世代をまたぐ遅延応答は破棄する。手入力入口はD07のcreate／edit・期待版・履歴／結果不明照合へ渡し、owners:editをD07 write権限として代用しない。

主飼主移管はpet IDの結果履歴を複製／移管しない。過去結果の原医院／患者／版帰属を保持する。D20、LIFF／line-reserve公開DTOのallowlistは拡張せず、D21の検査履歴sectionにも印刷／共有を加えない。別途採用の予防薬専用2帳票とは区別する。UI表示以外の自動通知・予約・課金・在庫副作用はない。現行owner/pet DTO・query・D07 endpointでの実装適合は未検証。

| 追加FR | 追加AC（全4件） | 技術対応 |
|:---|:---|:---|
| FR-owners-pets-infection-summary | AC-owners-pets-infection-top-reference / AC-owners-pets-infection-ambiguity / AC-owners-pets-infection-write-boundary / AC-owners-pets-infection-privacy | 本節のD07元ref・出典／日付精度・取得状態・両owner認可・台帳write分離・公開DTO境界 |

本節追加時点では先行9FR／27ACに1FR／4ACを追加した10FR／31AC。2026-10-17の共通利用追加時点は11FR／36ACであり、その後の自動採番1ACとサイグラム1FR／9ACを含む現在の契約は[業務索引](../../../spec/domains/owners-pets/README.md)の12FR／46ACへ従う。先行ID・既存の業務採否／実装検証状態を保持し、追加分の実機受入は未実施。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/owners-pets/lifecycle.md#common-owner-pet-20261017)。氏名または動物種＋ペット名の候補をスタッフ確認し、同一飼主は警告して新規停止、別人は可、新ペットは既存飼主へ追加する。他院既存患者の自院利用と共通最新基本情報を採用し、院別診療・金銭記録と認可は維持する。旧他院行／ID一律拒否を自院利用へ機械適用しない。共有ID・既存重複移行・link/unlink方式は未選定で、重複作成を自院利用条件にしない。


### 共通利用の追加受入trace

| 正本条件 | 技術接続・検証観点 |
|:---|:---|
| AC-owners-pets-duplicate-confirm | 権限内の候補参照とスタッフ同一確認結果を区別。同一確認なら新規停止、別人なら既存登録検証へ。名前一致だけの自動統合・新規停止なし |
| AC-owners-pets-new-pet-existing-owner | 確認済み既存飼主と新ペットをIDで対応し、飼主重複createを要求しない。既存同時登録の原子性と認可を保護 |
| AC-owners-pets-other-clinic-use | 確認済み共通患者の自院利用と、元院の診療・金銭資源の操作認可を分離。院別利用と内部IDの対応方式は未選定、履歴移管を起動しない |
| AC-owners-pets-common-latest | 共通基本情報の現在値と、院別診療・金銭原本／発行版を区別して参照。基本情報更新で過去snapshotを再構成しない |
| AC-owners-pets-common-denied | 自院利用の導線でも現在の医院・対象患者・操作別認可とLINE公開境界を再検査。認可外データを返さず、既存の手動group認可を緩めない |

上記は[業務正本](../../../spec/domains/owners-pets/lifecycle.md#common-owner-pet-20261017)の意味対応であり、共有ID・移行・link/unlink再設計を選定せず、実装・実機検証は未実施。

## 本人指定サイグラム計算基準との接続

[計算・保存の業務正本・固定表・9AC](../../../spec/domains/owners-pets/lifecycle.md#saigram-calculation-reference-20261018)を参照する。2026-07-16保存Drive版が本人の現在の指定基準であり、旧Slack添付の同一性や旧ソフト一致を実証済みとはしない。UTC暦日差と非負mod・固定表を仕様どおり扱う方式は実装設計へ渡すが、今回コード適合は未確認。

2026-10-18 02:59:05 UTCの本人回答（Sentinel_01ed294ac8988191b77e6c33e4a92bbf）により、サイグラム結果をDB保存する方針へ変更する。生年月日を正本とし、保存する派生結果と計算ルールのバージョンを対応付ける。生年月日の有効な変更を保存するときは同じ計算基準で結果を再計算し、結果だけを手入力・独立修正しない。未保存の入力中表示を保存済み結果へ混同しない。生年月日を明示消去して保存した場合は派生結果も消去し、分類なしとする。未入力・不正・削除済みの生年月日に古い結果を有効分類として残さない。不正入力の保存は既存検証で拒否し、その場合に保存済みの有効な生年月日と対応結果を片方だけ変更しない。既存データの不正・欠落が判明した場合もその古い結果を表示根拠にしない。DOB未送信は既存の保持契約を維持する。将来の絞り込み・集計を見据えた保存であり、その検索・集計UI／機能は今回は追加しない。実際のDB・schema・コード変更、移行・再計算処理の実行は今回行わない。 具体列・API契約・整合性を満たすトランザクション方式は未選定。既存飼主についても導入時に有効DOBから算出し結果とルール版を保存、未登録・不正DOBは未設定とする業務方針が03:05:29 UTC本人承認で確定した。導入処理の具体手順・検証は未選定であり、実データ処理は今回実行しない。将来の計算ルール改定は具体要求がないため改定時の技術・運用設計事項であり、現行リリースの必須業務判断待ちとしない。現在の画面縦積み、DOBの保存/null/省略保持と既存認可を維持する。

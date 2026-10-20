# 飼主・ペットの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D04-TECH / 技術設計・現行差分 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存契約の継承と[追加合意4・5](../../../spec/domains/owners-pets/lifecycle.md#agreement-owner-contacts-transfer-20261015)の実現目標。[台帳・会員区分・LINE希望の共同確認](../../../spec/domains/owners-pets/lifecycle.md#agreement-owner-membership-dm-20261015)を追加。具体的wire／DB変更は未実施 |
| 実装状態 | 下記限定ソースを2026-10-15に静的照合。全writer・HTTP型・DB・実権限の再監査は未実施 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-20 EMR-463先行wire契約・死亡取消追従（従来確認2026-10-17） |

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

<a id="pet-revival-target-gap"></a>
### 誤登録取消の目標と現行差（2026-10-20追従）

[本人承認済みの死亡誤登録取消](../../../spec/domains/owners-pets/lifecycle.md#pet-death-misregistration-undo-20261020)では取消理由を必須とし、医院・患者・操作権限と現在状態を検証した専用writerで、取消による状態変更と変更前後・actor・理由を含む取消履歴を同じtransactionで保存する。理由未入力・履歴保存失敗では状態変更を確定しない。取消だけで予約・通知を自動再開せず、生前分の後日入力のために取消を要求しない。

現行の[HandlePetRevival](../../../../backend/internal/lstep/lstep_lifecycle_service.go)は理由引数を持たず、状態変更と一次監査 `pet_revival` はsame-txだが、その監査に取消理由・変更前後を渡していない。commit後には既存のtag再同期がある。これは現行観察であり、上記目標への適合や予約／通知の再開許可を意味しない。理由入力・受渡しと履歴内容の充足は未実装差分として扱い、API／DB／実装の変更と実機受入は今回未実施。新grant・職種条件は追加しない。

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
- 一覧は適用済み条件のURL queryをSoT、サーバ20件ページ。[10/19明示検索](../../../spec/shared/ui-contracts.md#owner-pet-explicit-search-20261019)に従い入力中値は別に保持し、検索押下で適用する。300ms debounceは旧文書/旧実装観察であり現在全callerの差と断定せず、未適用文字列をURLへ流す目標ではない。検索結果の0件、読込中、取得失敗を分ける。検索条件変更でpageの妥当性を再確認し、別検索の応答で上書きしない。
- 新規の送信前／送信中／成功／明確拒否／結果不明を分け、成功対象だけdirty解除。結果不明は同じ操作・対象を照合する。メール／電話一致を保存receiptの代替にはしない。
- 現行新規APIに永続operation receipt／一意な再送キーの完備は未確認。安全な再送にはサーバ側照会可能な操作識別とpayload対応を設計・実装検証する必要がある。現状のGET0件だけでは遅延commitを否定できず、自動再createを止める。
- 期待版のない既存PATCHを「楽観lock済」としない。副飼主と主飼主移管の具体的競合保護を他の一般属性へ拡張して報告しない。
- UI離脱取消は未送信draftだけに効く。送信済mutation・専用LINE送信を取り消したことにはならない。

<a id="owners-list-scope-contract-emr463"></a>
## EMR-463：台帳一覧の現在許可医院と選択保持（採用先行契約）

2026-10-20のDots採用設計。業務根拠は[医院横断一覧](../../../spec/domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-identity-and-search)と[台帳一覧](../../../spec/screens/03-owners-list.md)。wireの正本は[api.yamlのGET /me・GetMeResponse](../../../../backend/docs/api.yaml)。ここでは処理順・既存との差・実装引渡しを定め、DTO定義を複製しない。今回の463-Cは文書のみで、認可拡大・DB追加・runtime実装・動作検証を含まない。

### BE：GET /meだけに許可集合を付ける

- `GetMeResponse`は既存`MeResponse`を同じJSON階層に包むGET専用DTOとする。`owners_view_clinic_ids`はGET /meの200だけに必須追加。共通`MeResponse`、login/refreshの応答、既存builderのsignatureは変えない。`id`は現行BEの`staff.ID`を文字列化した値でありaccount IDへ置換しない。
- [GetMe](../../../../backend/internal/auth/http_session_me.go)の`g.Wait()`成功後、既存のaccount取得・所属補完を終えてprofileを組み立てる。`ExtractClinicIDs(c)`が返す現在検証済み医院集合と、`profile.Clinics`に含まれるactive医院集合を交差する。profileの名称や`main_clinic_id`から所属を追加しない。system_adminもこの交差内に限定する。
- 交差した各IDへ既存の[checkPermissionInClinic](../../../../backend/internal/auth/http_permission.go)で`owners/view`を**順次**評価する。trueだけ採用、falseは除外、errorは5xxで要求全体を失敗させる。bool互換の`HasPermissionInClinic`でerrorをfalseへ潰さない。Gin context内のrequest memoへ並行書込する方式は採らない。
- [ExtractClinicIDs](../../../../backend/internal/httpapi/context.go)が失敗応答を書いた場合は直ちにreturnし、200や二重応答を返さない。認可評価errorに部分的な許可集合を付けて成功させない。許可ゼロは非NULLの`[]`。重複を含めず、各IDを正の正規形10進文字列で返す。返却順は認可の意味を持たない。
- [ToMeResponse / meClinicMemberships](../../../../backend/internal/auth/http_response.go)は既にactive医院を選別する。これを現在contextと交差させる設計であり、全医院readや新grantを追加する設計ではない。この集合は一覧選択の材料であり、後続`GET /pets`の最終認可を省略するtokenや永続した許可証ではない。

### FE：owners専有adapterから一覧を開始する

1. `ownersLoader`の各実行・再検証で、owners専有adapterが`GET /me`を取得する。fieldは必須parseし、欠落・NULL・非配列・非文字列ID・正規形でないID・重複を契約未提供／不正な取得結果として扱う。`[]`へ補完しない。既存の全体`ME_QUERY_KEY`のcached profileやlogin/refresh結果を現在許可集合の根拠として流用しない。
2. scope取得成功後、URLの適用済み選択／保存済み選択を現在許可集合と交差する。未選択の初回は全許可、保存選択の全件が失効した場合は現在の全許可を選ぶ。URLは適用済み選択のSoTであり、保存値は権限ではない。許可集合そのものが`[]`なら一覧APIを呼ばない。空選択を`clinic_ids`省略へ変換すると現在医院へ戻るため、空集合をそのように送らない。
3. 非空の有効選択を確定してから`GET /pets`を呼ぶ。IDをJS Numberへ変換せず、文字列のまま`clinic_ids`へ渡す。scope取得前の先行list requestや、取得不能時の所属全件・現在医院・旧scopeへのfallbackは行わない。
4. `GET /pets`の既存最終認可を維持する。403は現在scopeを再取得して選択を照合し直す。失敗したscopeを使ってlistを再送しない。通信障害／5xx／契約未提供は失権の証明ではなく、保存選択を削除しない。
5. 既存`clinic_ids`はカンマ区切り最大50 IDの制約を維持する。許可集合の応答自体を50件へ制限せず、選択が上限超過なら超過として一覧取得を停止する。先頭50件への切捨てや一部医院だけの成功表示、上限回避の分割全件取得は採用しない。

localStorage keyは`owners-list.clinics.v1.<user.id>`とし、保存内容は選択した医院IDの配列だけ。認可集合、患者／飼主情報は保存しない。keyへ医院IDを追加して横断選択を分断しない。別ユーザーの保存値は読まない。ログアウト時の物理削除方式など、未採用の保存ライフサイクルはここで追加選定しない。

login/logout、user/clinic切替、`ME_QUERY_KEY`更新では、進行中のscope/list両応答を無効化する。request abortと要求世代の照合等により旧応答を表示・選択・保存値へ反映しない。これは既存全体/me cacheを毎回改変する要求ではなく、owners専有取得とその要求contextの責務である。scope取得失敗、許可ゼロ、選択に対する一覧0件を区別する。

### 基準mainとの限定差分と実装順

静的照合基準は`da27989fa812c1319b62a1c3b4684c419b110fe1`。現行`GetMe`は共通builderの`MeResponse`を直接返し、新fieldと上記交差・順次評価はまだない。現行[ownersLoader](../../../../frontend/src/features/owners/loaders.ts)はURLの`clinics`をそのまま/petsへ送り、[useClinicScope](../../../../frontend/src/hooks/use-clinic-scope.ts)は未指定時に現在医院へfallbackする。[get-me](../../../../frontend/src/features/auth/api/get-me.ts)は全体profile cacheへ接続している。これらを採用目標へ適合済みとはしない。共有hookや全体AuthUserへの一括変更を今回の契約から要求しない。

既存`api.yaml/MeResponse.permissions`には医院ID配下のネストという旧説明が残る一方、現行`http_response.go`はresource-keyedの`EffectivePermissions`を返している。これは既存の文書差であり、今回共通MeResponseを全面改訂したり、owners集合をその旧fieldからFEで合成したりしない。login schemaの旧token説明等も今回の追加契約から正当化しない。

| 引渡し | 完了条件 | 今回の状態 |
|:---|:---|:---|
| 463-C 契約公開 | api.yamlとこの処理境界の採用SHAを公開 | 文書PRでレビュー対象。runtime未変更 |
| 463-BE 最小DTO | GetMe専用wrapper・交差・三値順次評価。login/refreshとbuilder互換を保持 | 未実装／NOT_RUN |
| 463-FE 専有adapter | 必須parse・scope先行loader・選択保持・要求無効化 | 採用SHA後の契約mock PRで着手可。未実装／NOT_RUN |
| 実API結合 | BE/FEが同じ採用SHAを参照し、現在認可・失敗・切替を実APIで確認 | 契約mock合格と別。未実施／NOT_RUN |

### 後続用fixture・受入ケース定義（全件NOT_RUN）

値は合成IDによるfield断片であり、実患者や完全な/me応答の記録ではない。下表は後続で作るmock／BE／結合検証の定義で、今回実行した証跡ではない。

| ケース | 期待 |
|:---|:---|
| 現在context `[1,2,3]`、active profile `[1,2,4]`、1=true・2=false | `owners_view_clinic_ids=["1"]`。3/4を評価・採用しない |
| system_admin、現在context `[1]`、active profile `[1,2]` | 1だけを採用。adminでも2を追加しない |
| 全候補false／交差0件 | 200+`[]`。FEは/petsを呼ばず、取得失敗と区別 |
| 1=trueの後に2が評価error | 5xx。`["1"]`や`[]`の200を返さない |
| context欠落／不正型 | ExtractClinicIDsの既存401／400で終了、二重応答なし |
| field欠落／NULL／数値ID／`"0"`／`"01"`／重複 | owners adapter取得エラー。所属や旧scopeへfallbackせず、保存値を消さない |
| `"9007199254740993"`が現在許可される | scope→選択保存→clinic_idsまで文字列を保ち精度損失なし |
| 初回、許可 `["1","2"]` | 全許可を選択し、scope完了後にだけ/petsを呼ぶ |
| 保存 `["2","3"]`、現在許可 `["1","2"]` | 交差`["2"]`だけを有効選択にする |
| 保存全失効、現在許可 `["1","2"]` | 現在の全許可へ戻す。未検証の3や現在医院を補わない |
| URLに許可外IDが混入 | 現在許可との交差外を送らず、空選択をquery省略で送らない |
| 現在選択50件／51件 | 50件は既存上限内、51件は超過として停止。切捨て・分割取得なし |
| /petsが403 | scope再取得後に再照合。失敗scopeで再送しない |
| scope通信障害／5xx | 一覧取得エラー。許可ゼロへ縮退せず、保存選択を失権扱いで削除しない |
| login/logout・user/clinic切替・ME_QUERY_KEY更新後に旧応答到着 | 旧scope/listを表示・保存へ反映しない。別userのkeyを読まない |
| login／refresh／既存builder | 新fieldを必須にしない。GET /me専用wrapper以外へ追加しない |

## 受け渡しと検証対応

| FR ID | 先行台帳FRのAC ID（追加分は後続節の対応表） | 技術対応 |
|:---|:---|:---|
| FR-owners-pets-register-scope | AC-owners-pets-owner-number-auto / AC-owners-pets-register-atomic / AC-owners-pets-scope / AC-owners-pets-optional-clear / AC-owners-pets-unknown-result / AC-owners-pets-stale-update | owner→pet typed draft／同一tx・所属検証。新規飼主番号の自動割当・保存後読戻しを追跡し、既存番号再採番や結果不明時の別createをしない。採番の物理方式は未選定 |
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

<a id="common-subject-link-proposal"></a>
## 共通患者・移管・LINE解除の詳細設計案

[C06提案状態・再送契約](../../shared/integration-contracts.md#detail-design-proposals-20261018)を適用する未実装の内部方式案。[共通最新基本情報・自院利用](../../../spec/domains/owners-pets/lifecycle.md#common-owner-pet-20261017)の既決と、本人の物理ID統合承認を混同しない。既存D19 groupとLINE本人対応を別に維持する。

| 論理資源／writer | ID・版・多重度の提案 | 不変条件 |
|:---|:---|:---|
| 共通飼主／共通ペット：D04 | 種類別opaque UUID＋基本情報revision | スタッフ同一確認又は明示新規で対応。氏名・電話・LINE ID一致から自動確定しない |
| 医院利用binding：D04 | clinic＋common subjectにactive利用は最大1。元clinic/local IDとのmapping版を保持 | 自院利用で診療・請求の元patient ID／医院／履歴を一括更新しない。利用解除・患者分割をこのintentへ追加しない |
| 共通基本情報：D04 | subject＋expected revision＋項目別changes | 現在の自院利用と既存owner/pet編集grantを再検査。過去診療・請求宛名・発行証明書snapshotは更新しない |
| identity group：D19 | 既存clinic/local member、生成院anchor・group版 | 共通subjectの物理PKへ流用せず、group unlinkで元台帳・履歴を削除しない |
| LINE binding：D04/D20 | clinic/local owner＋binding版、本人確認が対応する版とactor/時刻 | 共通患者・主飼主・LINE customer・外部本人IDを区別。旧bindingの確認を別LINE IDへ継承しない |

### 自院利用と移行

候補readは権限内の共通refと必要最小項目だけ返す。保存時は現在認可・候補版・スタッフ同一確認・既存利用を再照合し、利用binding／必要local adapter／監査／成功receiptを同時commit。同じclinic/subjectの別key同時利用も一意制約で同じ利用へ収束し、患者を二重createしない。mapping不明／複数は停止し、他院の診療内容を同一確認の材料として無認可公開しない。

全consumerはcommon ref→自院local refと、原資源のorigin clinic/local refを分離する。臨床・金銭read/writeは元資源の医院scope/grantで再認可。旧APIの他院ID拒否を単純解除せず、専用自院利用intentへ限定する。移行案は既存ID対応のread-only一覧→矛盾検出→スタッフ同一確認→mapping dry-run→対象限定切替。最新updated_atだけで基本情報の採用値を選ばず、過去請求・預り金・診療を移管しない。旧writer停止・全reader対応・mappingの一意性／件数確認が前提。実DB／移行実行は未実施。

### 主飼主移管のpreviewとphantom保護

既存[影響確認](#owner-transfer-target)の対象条件、LINE条件付き付替え／解除を維持する。previewはpet/旧新owner・予約ID集合と各版・LINE binding版・選定基準JST日・確認digestを返すread-only操作。保存は共通gate→receipt→既存pet/owner保護→対象予約NOWAITへ参加し、対象を**再検索して集合差分も比較**する。preview後の予約新設・受付・取消・LINE変更・日付境界通過は、既存行のversion一致だけで無変更としない。

対象／宛先結果が変われば元台帳を変更せず新影響一覧へ戻す。同じgateへ新予約writerも参加させ、検査後のphantomを防ぐ。既存上限・NOWAIT拒否・カルテ関連拒否・過去資金不変を保ち、pet＋予約帰属／LINE参照＋版＋必須監査＋成功receiptを同時commit。post-commit tag sync失敗で移管を再実行しない。

### LINE本人対応と解除

入力案は現在binding版・検証済み変更先ref又は明示解除・必要な影響確認・key。生token／LINE IDを監査や表示cacheへ追加しない。D20のtoken consume、対応保存、必須監査、成功結果を同じtxへ結ぶ。token期限切れ／再利用拒否だけでは初回保存失敗と判定せず、同認証主体/keyの結果を照会する。変更先IDへ旧確認を自動継承しない。

解除previewはbinding・未来予約参照・保留配送・D19依存をreadし、保存時に版と対象集合を再検査する。D19のowner member unlinkは子pet group依存があれば拒否し、対象医院の認可・同tx監査・元台帳保持を維持。D19 unlink、独立LINE解除、主飼主移管を一つの汎用unlinkへ統合しない。

送信reader D15には永続dispatch claimを提案する。clinic/owner/binding版・送信目的・payload digest・attempt ID・状態（準備／発信中／結果既知／結果不明）を持ち、認証／本人対応／停止policyを検査してbinding rowと同じgate下で発信認可を確定する。解除要求は同じrowで新claimを止め、発信前claimを取消可能かCASで照合する。先行する発信中／結果不明attemptがあれば解除完了を返さず処理中として照合へ。結果既知で発信が終了し、今後発信可能なclaimがないことを確認してからbinding解除と成功receiptをcommitする。claimの解放やworker再起動だけで外部再送しない。

内部tx中に外部HTTPを待たない。外部呼出直前の単純な再読だけでは、再読直後のunlinkを防げないため、上記claim／解除完了の線形化点を実装検証する。解除要求より先に認可・発信開始されたattemptは到達し得るが、既送信を回収可能と保証せず、結果不明を自動で未送信にしない。解除完了後の旧版新発信を止める契約と、解除要求前に始まった発信の回収不能を分ける。D15の既存結果不明保留／非冪等外部write再送禁止は維持する。このclaimは新たな配信や送信許可を採用するものではない。Q-I2は[実際の表示第2項A](../../../spec/line/lstep-integration.md#line-unlink-future-notification-20261018)により確定。予約自体・患者関連・参照履歴は保持し、通知は停止する。旧比較JSONのBに対応するため旧Aの宛先参照解除へ取り違えない。再連携だけでは再開せず、スタッフが再開する予約を確認する。

### 再開対象確認の技術adapter案

独立解除はreservation/customer参照を削除するwriteではなく、bindingの連携状態と通知資格を変更し、停止状態・必要履歴・成功receiptを同時保存する。残存参照は出所として保持するだけで送信許可に使わない。再連携時は新しいbinding版に対して再開未確認の状態を維持し、一括再開しない。

再開previewは明示対象の予約ID集合・各版・患者/owner・binding版・既存停止／結果不明状態を現在認可付きreadで返す。保存入力案は選択した予約ref/期待版集合＋binding版＋確認意思＋operation key。共通gate→receipt→binding/対象予約の既存安定順で選択集合・現在認可を再検査し、版／所属／宛先が変化すれば再previewへ。成功は確認した予約集合とactor/時刻・binding版への再開確認receiptを同時保存する。未選択予約、独立停止、結果不明attemptを変更しない。外部送信は既存D15が後で現在資格を判定し、確認txから起動しない。

再開確認receiptは古いbindingや別予約に転用せず、並行解除／再連携・予約変更・通知claimを同じgateへ接続する。応答不明は同keyで照会し、新keyの一括再開や再送で救済しない。保管／API fieldは技術提案で、本人がwireを承認した意味ではない。

### 限定コード観察・条件付き事項と確定済み回答

基準HEADの`backend/internal/identitylink/service_owner.go`はgroup lock・子依存拒否・論理削除・同tx監査。`backend/internal/owner/service_line.go`は手動LINE更新の監査best-effortで、当該repository更新からbinding版／確認解除までは確認できない。`backend/internal/lstep/line_link_service.go`のtoken linkはconsume・owner更新・監査を同txで行う。前者を後者と同じ原子監査済みとはしない。`reservation/pet_owner_transfer.go`のupdated_at更新は明示予約version増分と別。新案に適合済みの証拠ではない。

| 事項／根拠 | 現在の決定・扱い | 適用範囲 |
|:---|:---|:---|
| 条件付きQ-I1：具体的な矛盾データ移行／患者分割が必要になった場合の採用値・確認責任。[共通最新と内部方式留保](../../../spec/domains/owners-pets/lifecycle.md#common-owner-pet-20261017) | スタッフ同一確認・共通最新・自院利用は既決で聞き直さない。実データの矛盾や新たな分割要求が具体化した場合だけ対象一覧・採用案を提示する。更新日時のみの自動統合／患者分割を今回追加しない | 現在の本人確認待ち・全体blockerから除外。具体的な移行／分割対象が発生した時の条件付き検討に限定 |
| Q-I2：12:53 UTC本人の表示第2項Aで確定。[独立LINE解除と再開](../../../spec/line/lstep-integration.md#line-unlink-future-notification-20261018) | 予約・患者関連・参照履歴保持、LINE通知停止。再連携後に再開予約をstaff確認。元JSONのBと意味対応し、宛先参照も消す元Aは不採用 | 自動再開・即時送信・結果不明再送なし。主飼主移管とD19 unlinkの独自契約は維持 |

DC-I01：同一／別人／新ペット、自院利用の二重要求・mapping不明。DC-I02：2院同時基本情報変更、権限失効、過去記録不変。DC-I03：preview後の予約新設／受付／JST境界／binding変更を両順実行。DC-I04：D19子依存拒否、同key再送・異内容・応答断。DC-I05：付替え後の旧本人確認、token再送、必須監査失敗、旧宛先配送、結果不明再連携。DC-I06：dispatch claimと解除要求を両先行順序で実行し、発信中／結果不明は解除完了にせず、新claimを止め、worker停止・遅延HTTP・結果照会後も旧版を新発信しない。DC-I07：解除／再連携／staff選択／予約変更を両順実行。予約ID・患者関連保持、再連携のみの再開なし、未選択予約と独立停止／不明保持、古い版での確認拒否、確認保存だけの送信なしを検査。実装時の検証案で未実行。

## 本人指定サイグラム計算基準との接続

[計算・保存の業務正本・固定表・9AC](../../../spec/domains/owners-pets/lifecycle.md#saigram-calculation-reference-20261018)を参照する。2026-07-16保存Drive版が本人の現在の指定基準であり、旧Slack添付の同一性や旧ソフト一致を実証済みとはしない。UTC暦日差と非負mod・固定表を仕様どおり扱う方式は実装設計へ渡すが、今回コード適合は未確認。

2026-10-18 02:59:05 UTCの本人回答（Sentinel_01ed294ac8988191b77e6c33e4a92bbf）により、サイグラム結果をDB保存する方針へ変更する。生年月日を正本とし、保存する派生結果と計算ルールのバージョンを対応付ける。生年月日の有効な変更を保存するときは同じ計算基準で結果を再計算し、結果だけを手入力・独立修正しない。未保存の入力中表示を保存済み結果へ混同しない。生年月日を明示消去して保存した場合は派生結果も消去し、分類なしとする。未入力・不正・削除済みの生年月日に古い結果を有効分類として残さない。不正入力の保存は既存検証で拒否し、その場合に保存済みの有効な生年月日と対応結果を片方だけ変更しない。既存データの不正・欠落が判明した場合もその古い結果を表示根拠にしない。DOB未送信は既存の保持契約を維持する。将来の絞り込み・集計を見据えた保存であり、その検索・集計UI／機能は今回は追加しない。実際のDB・schema・コード変更、移行・再計算処理の実行は今回行わない。 具体列・API契約・整合性を満たすトランザクション方式は未選定。既存飼主についても導入時に有効DOBから算出し結果とルール版を保存、未登録・不正DOBは未設定とする業務方針が03:05:29 UTC本人承認で確定した。導入処理の具体手順・検証は未選定であり、実データ処理は今回実行しない。将来の計算ルール改定は具体要求がないため改定時の技術・運用設計事項であり、現行リリースの必須業務判断待ちとしない。現在の画面縦積み、DOBの保存/null/省略保持と既存認可を維持する。


<a id="saigram-acceptance-trace"></a>
### 既存サイグラム9ACの技術対応

上記の[業務正本・固定表・9ACへの委譲](../../../spec/domains/owners-pets/lifecycle.md#saigram-calculation-reference-20261018)は維持する。以下は既存の計算参照規則・保存／初期投入説明への対応を明示するもので、新しい算法・業務採否・実装完了を追加しない。先行の台帳・追加感染症・共通利用の37ACと合わせ、現行46ACの追跡範囲を示す。

| 正本AC（prefix AC-owners-pets-saigram-） | 本節の既存説明への接続・確認点 |
|:---|:---|
| AC-owners-pets-saigram-base-negative-cycle | 本節のUTC暦日差・非負modを正本の基準日前後／±60日の期待値へ照合し、負番号・0へ補正する独自算法を足さない。 |
| AC-owners-pets-saigram-fixed-mapping | 正本の固定表60エントリを参照して結果を解決し、表示名や並び替えから対応を再生成しない。 |
| AC-owners-pets-saigram-code-group-ha | 既存のタイプ→コード／グループ／派の変換を正本へ照合する。画面縦積みの既決を保ち、計算定義を新しい配置義務にしない。 |
| AC-owners-pets-saigram-five-examples | 正本の5例について生年月日から派まで同じ参照規則を通し、例の一致とコード適合・実機合格を分けて記録する。 |
| AC-owners-pets-saigram-utc-leap | 端末timezoneに依存しない暦日差と正本の閏日期待値へ対応付ける。既存日付検証を迂回した不正日の補完をしない。 |
| AC-owners-pets-saigram-persist-consistency | 本節のDOB正本・派生結果・ルール版の保存説明へ接続し、登録／有効変更と再読取で対応を照合する。片側更新や未保存入力を保存済みとしない。 |
| AC-owners-pets-saigram-clear-invalid | 本節の明示消去時の結果消去、不正保存拒否、既存欠落／不正の旧結果不採用へ接続する。DOB未送信は既存値保持を維持する。 |
| AC-owners-pets-saigram-storage-scope | 本節の保存対象はDOB対応結果とルール版であり、新しい絞込み・集計UI、独立結果編集や外部連携を追加しない。 |
| AC-owners-pets-saigram-initial-population | 本節の導入方針に従い、有効DOBは同じ基準で算出して結果／ルール版を保存する対象、未登録／不正は未設定として分ける。具体移行手順は未選定で、今回実データ処理を行わない。 |


<a id="identity-physical-adapter-candidates-20261019"></a>
## 2026-10-19 内部adapter・保存候補と未実行検査

IR-01〜06／19の文書具体化。以下の資源名・列・wireは**内部実装候補**で、本人による物理統合方式の承認・API実装・DDL適用を意味しない。予約共通gate・目的別intent・再送規約は既選定のD01正本へ従い、再選定しない。

| 対象 | 現入口から候補adapter／保存構造への対応 | 失敗・照会・移行検査 |
|:---|:---|:---|
| IR-01 共通最新・自院利用 | 既存owners/petsのBIGINT PKとclinic_idを残す。候補`subject_refs`はopaque参照・種類・生成院anchor・revision、`clinic_subject_bindings`はclinic_id・subject_ref・local_owner/pet_id・mapping_revision。activeの(clinic_id,subject_ref,種類)と(clinic_id,local_id,種類)をそれぞれ一意にする。共通refを既存D19 group IDやLINE identityへ流用しない。院内APIはlocal ref、共通基本情報readerは認可付きbinding解決、診療・金銭・過去帳票はorigin clinic/local refを維持する。全面PK置換案は過去FKの一括移管を必要とするため本候補に採らない | `use-in-clinic`候補intentはsubject_ref・期待mapping/basic版・スタッフ同一確認参照・key。現在認可と一意性再検査、必要local adapter・binding・監査・receiptを同tx。異key競合も同bindingへ収束。未知／複数mappingは拒否。調査対象はowner/pet CRUD、予約、D03、D07、D10/18、D19/20/21。実データ件数・由来・全consumer列挙が揃うまで切替しない。生成院anchorは認可ではなく、認可外基本情報を返さない |
| IR-03 主飼主移管 | 既存pet更新→`TransferUpcomingPetReservations`へ、read-only previewと確認digest adapterを追加する候補。入力はpet/旧新owner ref、pet期待版、preview参照、key。preview保存候補はJST基準日・対象予約ID/各版集合・LINE binding版・宛先解決結果のdigest。個人情報や生LINE IDはdigest入力の監査表示へ複写しない | 同gate下で対象を再検索し、集合追加・消失・日付境界・LINE変化も比較。既存最大1000/NOWAIT・カルテ関連拒否を維持。pet・対象予約版増分・帰属・必要監査・receiptを同tx。応答断は元intent/key照会。tag sync失敗で再移管しない |
| IR-04 一般台帳更新・段階登録 | 既存owner nullableDate/Boolの存在フラグを維持し、候補入力はexpected_revision＋許可changes＋必要な同一確認参照。owner+pet作成の既存same-txを保ち、登録試行→成功owner/pet IDsの最小対応をreceiptに置く。住所・電話等の基本情報は共通owner資源へ、医院別既存会員・DM等は現scopeへ分け、全owner列を共通化しない | 同一飼主確認なら新規停止、別人登録は既存形式検証へ。電話メール一致だけを保存済みreceiptにしない。全caller・現unique制約・旧409処理を調べてから変更案を確定。追加類似閾値・新正規化は選ばない。省略保持／明示null消去と版競合拒否を検証 |
| IR-05 LINE対応・独立解除 | 手動`LinkLineUserID`とtoken `LinkAccount`を同じ内部binding adapterへ寄せる候補。`owner_line_bindings`候補にclinic_id/local_owner_id・binding_revision・連携状態・確認actor/時刻、現秘密情報への参照を持つ。確認はbinding版に帰属し、別IDへ継承しない。`line_dispatch_attempts`候補はbinding版・目的・payload digest・attempt ID・状態。新送信機能や目的の採用ではない | 解除と発信claimは同binding rowで直列化。発信前claimはCASで取消、発信中/不明があれば解除は処理中。結果照会で不明を未送信に落とさず、自動再送なし。完了時binding変更・停止・必要監査・成功receiptを同tx。外部HTTPはtx外。D19 unlink・移管とは別intent。旧予約/患者参照は消さない |
| IR-06 再連携後の選択再開 | 候補`reservation_line_resume_confirmations`はclinic_id・予約ID・binding_revision・予約期待版・actor/時刻・operation結果参照。送信資格は現在bindingと確認集合の一致＋独立停止条件で判定。previewは予約ID集合を返し、保存は選択済みID/版集合だけ受ける | 解除/再連携/予約変更をgate下で再確認。成功receiptを取得してもそのtxから送信しない。未選択予約、不明attempt、独立停止は保持。旧bindingの確認を新bindingへ流用せず、結果不明は同key照会 |
| IR-19 DOB派生保存 | 現owner DOB nullable fieldを唯一の入力とする。候補派生列はsaigram_personality_no(1..60)/type_no・rule_version・source_birth_date・source_revision。結果列はwrite DTOから除外し、固定表からserver計算。全派生列がnullか同じDOB/版へ整合する組だけを保存。表示名は固定ルール版から解決し入力欄を増やさない | 有効DOB変更と派生結果を同tx、明示nullは派生列もnull、省略は保持、不正入力は旧DOB/結果とも不変。導入はread-only分類→dry-run manifest(対象ID/元revision/ルール版/期待値digest)→有効DOBだけ条件付き更新→件数/9AC照合。新規・変更writerも同契約へ接続後に切替。再実行は同DOB/版なら無変更、途中でDOBが変われば再評価し古い計算を書かない。未登録/不正/削除済みを推測補完しない。処理は未実行 |

### 限定静的根拠と残工程

このpassの基準refで`backend/internal/model/owner.go`はclinic-scoped BIGINT ID・DOB・DM nullableと既存LINE属性を保持し、本ファイル内にサイグラム派生列はない。`owner/http_request.go`のnullableDateは未送信/nullを区別する。`owner/service_line.go`の手動変更監査はbest-effortで、`lstep/line_link_service.go`のLinkAccountはtoken consume・owner変更・必須監査を同txにする。両者を同じ原子監査済みとしない。`reservation/pet_owner_transfer.go`は既存対象をNOWAIT取得しowner_id/updated_atを更新するが、この限定読取で明示予約version増分・永続receiptは確認できない。repository全体や他経路の未実装を断定する根拠にはしない。

IR-02はmappingの実在件数・矛盾・由来とdry-runが残る。共通候補の具体的な矛盾採用値は実対象が出た場合だけQ-I1へ限定質問し、現在の同一確認・共通最新・自院利用・サイグラム保存を聞き直さない。IR-08はLIFF DTO allowlist・single-use linkの固定ref照合と別医院/取消/訂正版/応答断の実機、IR-17は現grant/失効/lookup障害の実機、IR-20はS14/S23の期待行・件数・非第一医院帰属・入力者の実機が残る。静的観察からPASSに変更しない。


追加静的照合（IR-02/08/17/20）：`model/identity_link.go`の既存groupは生成院anchorとclinic/local memberを持つ手動連携で、共通subject方式の実装ではない。`reservation/liff_handler.go`はpath clinic＋認証customerをhealth-card serviceへ渡し、`liff_response.go`は飼主名・生存pet基本表示・ワクチン日付・最終来院日へ閉じる。`liff_service_health_card.go`は本人owner未対応で空pet、死亡status/死亡日のOR除外を行う。この限定照合から取消/削除/訂正版filter・別医院token・単回CAS競合・応答断の全成立は宣言しない。`auth/current_access_service.go`はstaff/account/assignment/active clinicを再解決する資源を持ち、`current_access_cache.go`の現在authorityキャッシュと権限キャッシュは別資源である。関数既定TTLを実稼働設定や即時失効保証へ転用しない。`use-clinic-scope.ts`の現fallbackはcurrentClinicIdであり、選択未保存時の権限内全院初期値・次回保持という最新目標への適合証拠ではない。`medical_record_crud.go`の入力者検証呼出しは存在するが、B院実作成と再読取の証跡は未取得。

## 10/19明示検索の状態・取得adapter（内部設計）

[目標・AC](../../../spec/shared/ui-contracts.md#owner-pet-explicit-search-20261019)へ接続する。clientは入力中search/owner_noと、適用済みsearch条件・独立filter・page・医院scopeを別に持つ候補とする。入力effectからloader/queryを起動する旧debounce配線を新目標として継承せず、検索ボタンの明示actionで既存queryを適用する。台帳/共通患者選択では空白だけをtrimして空条件へ変えない。D19では既存trim・最小1文字未満の取得停止・上限付き候補取得を維持し、新たなページングを追加しない。ページングを持つ画面では条件変更時だけ既存pageを先頭へ戻し、ページ変更は適用済み値を使う。

URLを持つ一覧は適用済み条件だけを復帰根拠にし、URLを持たない候補検索も同じ入力/適用分離を保つ。scope/filter更新は既存操作で適用し、未適用検索入力を付け足さない。候補選択用は入力中値と適用済み条件が一致しない間、旧結果からの新たな選択確定を抑止する。既選択IDを入力変更だけで書換えない。query key/要求世代に医院scopeと適用済み条件を含め、応答採用時にも現在条件を照合する。resetは既存解除対象だけを入力/適用の両方へ反映し、0件・未実行・loading・失敗を別状態にする。新API・全件取得・追加認可・入力最小文字数・新しい検索述語を選定しない。旧hook/loaderと各候補検索への適合・実機は別工程。

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。

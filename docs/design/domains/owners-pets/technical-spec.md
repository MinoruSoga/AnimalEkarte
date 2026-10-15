# 飼主・ペットの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D04-TECH / 技術設計・現行差分 |
| 決定状態 | 既存契約の継承＋業務結果を変えない安全な設計整理。新しい業務承認ではない |
| 実装状態 | 下記限定ソースを2026-10-15に静的照合。全writer・HTTP型・DB・実権限の再監査は未実施 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

業務FRとACは[台帳lifecycle](../../../spec/domains/owners-pets/lifecycle.md)が正本。画面詳細の重複、API YAML・DB定義の複製をしない。wireは[API](../../../../backend/docs/api.yaml)、DBはmigrations、writer境界は[カタログ](../../../architecture/model-write-owner-catalog.md)。

## ID・write owner・保存単位

| 対象 | 現行owner／入口 | 保存単位・注意 |
|:---|:---|:---|
| 飼主 | owner / owners CRUD | clinic+owner ID。owner_numberはID表示別名 |
| 患者台帳 | pet / pets CRUD | clinic+pet ID。pet_numberは別管理番号。名前で連結しない |
| 飼主と複数ペットの新規登録 | owner.CreateWithPets→pet.CreateForOwnerRegistration | ambient tx必須、全write＋response reload成功でcommit |
| 主飼主変更 | pet.Update→reservation.TransferUpcomingPetReservations | pet、対象予約、移管監査はsame-tx。下節の限定条件 |
| 副飼主 | pet.PetOwnerService | version付き全置換、リンク差分の必要監査same-tx |
| 全医院動物種 | pet.AnimalSpecies | system-admin限定、CRUD/reorder＋監査same-tx |
| 死亡／復活 | 専用HandlePetDeath/HandlePetRevival | generic UpdatePetInputにstatusなし。D15の外部連携と区別 |
| 配信除外／注意／LINE ID／送信／共有 | owner／LINE専用経路 | 通常飼主form全体のrollback対象ではない。外部送信はcommitと到達を別結果 |

[PATH-OWNER-PET-REGISTER](../../../architecture/cross-domain-orchestration-catalog.md)はsame-txだが必須audit writeを持たない。全台帳更新に監査があると一般化しない。新たに必要監査を増やす際は保存方式・移行・権限とともに設計し、現状の証明に使わない。

<a id="owner-transfer-current"></a>
## 主飼主移管の現在の限定契約

静的参照：[pet service](../../../../backend/internal/pet/service.go)・[reservation移管](../../../../backend/internal/reservation/pet_owner_transfer.go)。

1. 同医院の変更先ownerを検証し、現在の副飼主でないことを確認。pet行lock後も主副重複を再確認。
2. 対象予約は同医院・同pet・JST当日00:00以降・pending/confirmedのみ。現在の実装では最大1000件、超過は全体拒否。
3. 予約側で新ownerをSHARE NOWAIT、対象予約をID順UPDATE NOWAITで取得。対象ownerが旧ownerと違う／NULL、関連カルテがある場合はConflict。
4. pet変更と対象予約owner更新をsame-tx。元owner、新owner、移管appointment IDs、基準日を必要監査へ記録し、監査失敗はrollback。
5. 元予約がLINE customer付きの場合だけ、新ownerに確認済LINE IDと一意に対応するcustomerがあればそのIDへ。なければNULL。customer_fieldsは空にする既存処理。氏名類似・未確認IDを使わない。
6. 過去予約、checked_in以降、カルテ、確定請求、預り金、identity groupによる連携はこの移管の対象外。
7. 次回来院等のtag syncはcommit後best-effort。tag失敗を台帳rollbackと誤認しない。D15の停止・再開許可を本整理から変更しない。

これは既存実装の照合結果で、今回この全挙動を新規承認したものではない。D01/D18/D20の合意正本と差があれば現行差として明示し、移管対象を拡張しない。副飼主解除と主飼主変更の2回送信は一体transactionではなく、片方だけ成功した場合は同じpetで現在状態を照合する。

## 更新・null・版・参照保護

| 項目／経路 | 現行契約 | 誤用の防止 |
|:---|:---|:---|
| 名前由来／出逢い | 省略=保持、null=clear、空白はNULL。500／2000文字 | 空GET結果をclear payloadへしない |
| 飼主生年月日 | 消去はJSON null、サイグラムは派生表示 | 派生分類をDBの独立事実にしない |
| 副飼主PUT | FEは取得version付き全置換、409は再読込 | service内部Versionはoptionalな互換型。HTTPで全caller必須を保証済みとはしない |
| owner削除 | 主飼主ペットcount＋副飼主count保護。productionはguard付き構成 | fallback構成の副飼主保護欠落を「全経路保護済」にしない |
| pet削除 | 関連カルテcount確認後delete、commit後tag sync | 全専用記録／FKを網羅した削除保証は未確認。臨床記録削除を案内しない |
| owner基本更新 | 重複メール／電話の検査と原子的更新経路 | UI noValidateはBE検査免除ではない |
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
| FR-owners-pets-identity-and-search | AC-owners-pets-search / AC-owners-pets-checkup-boundary / AC-owners-pets-read-only / AC-owners-pets-cross-clinic-list | 検索query／ページ・権限・遅延応答分離 |
| FR-owners-pets-primary-secondary | AC-owners-pets-secondary-replace / AC-owners-pets-secondary-conflict | 主副重複／全置換version・監査 |
| FR-owners-pets-primary-transfer | AC-owners-pets-owner-transfer-confirm / AC-owners-pets-owner-transfer-atomic | pet→reservation typed intent・JST対象／同一tx監査 |
| FR-owners-pets-safety-and-private-notes | AC-owners-pets-danger-privacy | 死亡専用writer／staff・公開DTO境界 |
| FR-owners-pets-delete-dependencies | AC-owners-pets-delete-guard | 参照依存／不明停止・原記録保持 |
| FR-owners-pets-species-master | AC-owners-pets-species-authority / AC-owners-pets-species-reference-audit | global species／system-admin・参照拒否・監査 |

参照ソース：[owner service](../../../../backend/internal/owner/service_core.go)、[副飼主service](../../../../backend/internal/pet/pet_owner_service.go)、[同時登録カタログ](../../../architecture/cross-domain-orchestration-catalog.md)。確認refは41fcf3eb0a8befb7c0b5f613d1df749a60e30be8。動作テスト・実DBの既存不整合・全writer依存・移管時の外部通知の実稼働は未確認。

## 文書整理の完了境界

既存3画面・台帳FR7件／AC17件・語彙・保存単位・移管／副飼主／死亡writer・主要異常系・他ownerへの引継ぎを記述した範囲。独立レビューとリンク確認を経て判定する。実装／receipt導入・全参照ガード・認可設定・保存期間・全体横断X04/X06の完了を本書から宣言しない。


## 一般更新の同時編集契約（選定仕様）

元表示版と保存時の現在版を比較し、不一致なら現在の原記録を変更せず競合を返す。入力を保持し、再取得した現値と差分を表示してスタッフの再確認を受ける。期待版だけを新版へ置換して無警告再送しない。明確拒否と結果不明を区別し、結果不明は元医院・対象ID・同一操作識別／内容で照合する。別操作IDや別createで救済しない。既存一般PATCH/UPDATEの期待版・永続receipt対応は実装差であり、この選定動作の実装済みを意味しない。

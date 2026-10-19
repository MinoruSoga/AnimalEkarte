# EMR-445：仕様回答と最終migrationへのモデル整合

基準SHA `7f164f8a71baf5bfb23207ddce5b3049939e81a3`。PR #727はmerge済み（merge SHA `81b9f076f6d592b6003a1b8589bbc06843bde65d`）で、未mergeを作業条件にしていない。専用branch `docs-models/emr445-followup-20261020`、専用worktreeを使用。共有mainと他の既存作業は変更しない。本書はローカル修正・静的確認の記録で、公開状態は[PR #729](https://github.com/MinoruSoga/AnimalEkarte/pull/729)で追跡する。公開を利用開始・実機合格の証拠にはしない。

## レビューする成果物

- [全241表の索引](table-inventory.md)：既存122モデルと新規119モデル。最終DDLで有効な2,691列を全件対応し、対象外として省略したテーブルはない。
- [全列TSV](columns.tsv)／[詳細JSON](model-schema-map.json)：SQL型・Go型・NULL・PK・既定値・最終FK/制約/index・実ソースhashを記録。
- [静的照合結果](static-check.json)：GORMの実解析結果とDDLを比較した30,020項目、不一致0。DBの制約実行・同時保存・認可の試験ではない。
- [未追従consumer一覧](consumer-impact.json)：廃止型参照、型/列変更、構造体literal、旧計算helper呼出しの位置。文字列候補の棚卸しであり全プログラムのコンパイル診断ではない。

## 型・キーの対応方針

CREATEだけでなくALTER/DROP後の有効定義を、既存 `scripts/tools/erd_schema.py` から取得した。未対応構文を黙って無視する方式ではない。最終SQLは正本のまま変更していない。

SQL text/varcharはstring。既存Goの名前付き文字列型をSQL enumへ割り当てる推定は、実際にCREATE TYPEされたenum名へ限定した。元の推定がSQL textにCheckupPackageImportStatusを全体適用する誤りを起こしたため、個別修正ではなく元のモデルsnapshotから全件再生成した。JSON既定値の引用符も生成元でescapeし、reflect/GORM解析で検証した。

NUMERIC/DECIMALは既存依存pgx/v5のpgtype.Numericを包むローカルExactNumericとし、2進floatへ変換しない。GormDataTypeでnumericを明示し、pgtypeのValidからboolを誤推定させない。列ごとの精度/scaleはtypeタグを保持。Scanner/Valuer/JSONの既存インターフェース適合はcompile-timeで確認した。新ライブラリ・go.mod/go.sumの変更、計算式・丸め・NULLの0補完は追加しない。NULL列はポインタ、既存のsoft-delete列はgorm.DeletedAtを保持する。ExactNumericのValid=falseもSQL NULLを意味するため、非NULL列のwriterは値の成立を明示する必要がある。

JSONBはjson.RawMessage、UUIDは既存google/uuid、日付/時刻はtime.Time（SQL timeはstring）、byteaは[]byte、配列は既存lib/pqの対応型（real[]はFloat32Array、smallint[]はInt64Array、text/enum配列はStringArray）とSQL型タグを用いる。これはdriver経由の実DB往復を試験した意味ではない。bigint IDは既存uint64規約、非IDの整数は幅を保持する。

全主キーを明示し、非serialの複合PKへ暗黙auto-incrementを付けない。既存フィールドの明示／暗黙の自動時刻動作を保持し、新規モデル／新規列だけDB既定値と各write ownerの時刻を優先する。独立レビューで見つかった一律抑止の回帰を修正した。既存ORM関連121件は最終FKの列組/参照組へ合わせ、廃止表への6関連を外した。直接FKのない8項目はgorm:"-"の明示projectionにし、暗黙joinや誤ったmutable master読取を生じさせない。対応と理由はJSONのassociationsへ記録。新規モデルは全FK列を持つが、自動ナビゲーション・自動保存graphは追加していない。FK/unique/index等の正本はmigrationであり、これらのモデルからAutoMigrateを行わない。

新規列はjson:"-"とし、無条件の公開JSON項目を増やしていない。一方、既存列の型やNULL表現は変わるため、旧DTO/responseとの互換を保証しない。3件のpackage-private rowは全列とNULLの有無を読取確認し、ローカルworkflow/cache adapterとして保持した（JSONのlocal_row_projections）。canonical counterpartは追加済みで、caller側の統合や計算algorithmは変更しない。

## 廃止対象と他層の残件

廃止9モデルは、Payment/payments、PaymentSplit/payment_splits、StaffReservationExclusion/staff_reservation_exclusions、ReservationTypeOccupation/reservation_type_occupations、LineSendLog/line_send_logs、BillingRefund/billing_refunds、AppointmentTrimmingOption/appointment_trimming_options、CashRegisterCloseAdjustment/cash_register_close_adjustments、BillingConfirmation/billing_confirmations。旧backendのために削除済みschemaや列を復活させない。非永続DTO/enum語彙は他層の互換参照として残るものがあるが、廃止TableNameへ戻していない。

既存75モデルに形状差分（削除98フィールド、型変更139、追加89）がある。旧billing/payment/refund/confirmation、LINE送信履歴、予約属性/個人適格性、共通owner/pet基本情報、旧計算/DTO等に追従が必要。廃止モデル型だけでもproduction31ファイル（test含む106ファイル）に参照が残る。現状の他層は未追従であり、そのままアプリを利用・deployできるとは扱わない。全体buildは今回実施していない。

旧BillingItem/EstimateItem.CalculateTaxAmountはfloat・旧形状へ依存するため外し、代替税式は導入していない。productionの5呼出し位置をconsumer-impact.jsonへ記録。既存calculator/DTOの新しいexact型・NULL・入力税込額契約への適合、旧テストの追従は所有領域の残件。今回API/service/repositoryの業務実装へ範囲を広げない。

## 仕様の反映と旧記述の解消

税込入力を無条件の税抜再換算で減額する読み方を訂正し、税抜入力の既決例と分けた。行別値引率、手動の小額値引き、明細/税集計/請求/受領・釣銭・繰越の段階を明記。既決順序と旧画面例が矛盾しないという照合結果を、唯一式の証明や実際の現行診療システムとの完全一致へ拡張しない。小数数量、内税換算、率と定額の併用、混在税率を追加ACへ残した。資料の私的な原数値・画像・Driveリンクは転記していない。

初回本番切替時の001凍結をREADME/CLAUDE/ERDへ整合し、即時の全編集禁止や2026-10-20の自動凍結としない。保存時氏名を選定済みへ更新し、物理保存先/全writer/APIの設計とは分離した。飼主原文reader477は今回phase、音声/外部320/322全体はPhase2と明示。ケージreadは同院入院作成/編集のID最小範囲だけを承認済みとし、旧route観察をgrant変更済みへ書き換えない。品目保険の新明細引継ぎは454–456へ含め、旧欠落の推測や健診保存→会計生成を追加しない。270/370は本人の解消済み認識に従い、当時の環境不足を現在blockerから外すが、技術解消の検証済みとはしない。

## 回答の出典

Plane本文/最新コメントをread-onlyで確認した。下記は2026-10-19のコメントIDであり、チケットやstateは変更していない。追加会計照合の結論は本セッションの親側から送付された確定範囲を適用した。

| Ticket | 最新回答コメントID | 反映する契約 |
|:---|:---|:---|
| [EMR-445](https://app.plane.so/baritechllc/browse/EMR-445/) | `492f16e6-9533-48f7-b581-b443b945dc30` | 全体境界。270/370の当時の環境不足を現在blockerへ残さない |
| [EMR-270](https://app.plane.so/baritechllc/browse/EMR-270/) | `aa570a04-c680-4f2c-82e1-21fb66e09356` | 本人は解消済みと認識。再現・修正・実機合格とは別 |
| [EMR-370](https://app.plane.so/baritechllc/browse/EMR-370/) | `66db4f47-e8c6-4a41-9967-1eb0e78dd5a5` | 本人は解消済みと認識。Chrome/Edge最新推奨、最低版追加なし |
| [EMR-367](https://app.plane.so/baritechllc/browse/EMR-367/) | `52dc7803-c201-4619-99aa-557f8c59b5ef` | 小額値引きは手動。税込入力維持、条件付きSTG判断と同等性未確認を区別 |
| [EMR-368](https://app.plane.so/baritechllc/browse/EMR-368/) | `3c7a14f2-761e-4430-aba6-6b5df3484364` | 各正常保存時の氏名とIDを保持、改名だけで過去名を変更しない |
| [EMR-452](https://app.plane.so/baritechllc/browse/EMR-452/) | `a9f103ce-4ba3-4932-814e-c948809a3912` | 477 readerは今回phase、476/472依存。320/322全体はPhase2 |
| [EMR-446](https://app.plane.so/baritechllc/browse/EMR-446/) | `aafeb13b-626c-4666-8de1-75f6c8e923bc` | 別途承認された初回本番切替で001凍結、その後増分。日付の自動期限なし |
| [EMR-461](https://app.plane.so/baritechllc/browse/EMR-461/) | `50458a22-7c05-4dc2-8629-b0de8fa7a975` | 同院の入院作成/編集権限で空き確認に必要なケージIDだけread |
| [EMR-448](https://app.plane.so/baritechllc/browse/EMR-448/) | `5b9bfbb1-ce7a-4663-be13-24a42053f078` | 457/458/480/461、時間二重加算なし、正の個人区分＋勤務、241/493へ引継ぎ |
| [EMR-447](https://app.plane.so/baritechllc/browse/EMR-447/) | `a6658664-a7b3-430e-8060-c75541755f5b` | 品目保険初期値→新明細と参照版は454–456、共有owner454 |

## 実施確認と未実施

- 実施：SQL静的解釈、生成元修正、gofmt、modelの正規固定依存でのコンパイル、productionモデルファイル限定go vet、GORM schema.Parse全241型、実解析フィールドと最終DDL/対応表の照合、呼出し候補棚卸し、差分/追加リンク確認。既存詳細ERDのindex/manifestに残っていた旧DDL hashのみ更新し、詳細ERD生成チェックも確認。
- 初回はキャッシュ不足のためx/text v0.41.0への一時置換で補助確認した。その後の明示指示により、既存固定版x/text v0.43.0とpostgres driver v1.6.3を公式proxy.golang.org／sum.golang.org経路で取得できた。最終確認は正規go.mod/go.sum、置換なし、ネットワークなしでモデルbuildとproductionモデル限定vetを実施。正規版の結果は[依存・コマンド結果](dependency-verification.json)へ。通常のpackage単位vetは旧accounting_test.go等が数値literalをExactNumericへ代入するコンパイルエラーでFAILとなる。これはキャッシュ不足ではなく旧テスト未追従。テスト本体は実行していない。依存定義・他層・旧テストは今回変更しない。
- Dockerは既存golang:1.26-alpine、read-only root、隔離tmpを使用。固定依存取得時のみ公式経路へ接続し、最終build/vet/メタデータ解析はnetwork noneで実施。DBを起動/接続するコードを実行せず、解析helperだけを実行した。新規testは作らず、既存testも実行していない。
- [独立レビュー結果](independent-review.md)：Standards1件とSpec2件を修正し、各担当が読み取りで再確認した。生成元と異なる観点でALTER/DROP、型/キー、driver I/Oと文書を照合し、検証数の大きさだけを保証にしていない。
- 未実施：DB作成/適用、migration/制約/RLS/並行性試験、実アプリ起動、API/UI、実ユーザー/実データ、税計算の実行受入、STG/本番、CI手動起動、commit/push/PR/merge。

既存DBのchecksum guardを回避せず、resetや切替の許可は推定しない。現行システムとの計算同等性は未確認なので、条件付きのSTG制約緩和の成立も未確認。本番過去取引・発行物、医院/飼主/pet/staff境界、原文の不変履歴、認可・write owner・監査/operation receiptの契約を維持する。今回の保存型定義から実装や権限成立を推定しない。

共有checkoutは最終読取時点で別作業の `feat/emr-465-inactive-payment-gate` と未commitのbilling変更を持っていた。今回の専用worktreeとは分離されており、これらの差分には触れていない。

## 旧調査案の再掲に対する追加照合

本人回答の撤回ではない旧調査コメントの再掲と照合した結果、カルテ設計とリンク先の財務設計に、廃止billing_confirmationsを残す最小Complete gateと、金額変更時に医師再確認を求める旧文が残っていた。2文書だけを訂正し、最終001のdoctor_review_sets/doctor_review_recordsと現在の対象集合・臨床世代の照合へ接続。金額だけの変更は警告了承で可・医師再確認不要、薬剤処置追加削除/数量変更は再確認まで確定精算停止という本人決定を維持した。API/業務処理/DDLの実装は追加していない。別担当が局所修正と4決定の維持を再読確認した。既存文書の変更対象は、この直接のリンク先1件を加えて18件である。

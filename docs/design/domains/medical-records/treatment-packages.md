# 診療項目パッケージの採用技術方向（EMR-518）

2026-10-20採用方向。[業務4FR/12AC](../../../spec/domains/medical-records/treatment-packages.md)が正本。本書は実装済みschema/APIではない。表名は設計名で、全列DDL・全FK/index・最終DTOは[固定チェック](../../../work/emr/emr-518-treatment-packages-20261020.md#schema-freeze)の未完項目。出典: [EMR-518](https://app.plane.so/baritechllc/browse/EMR-518/)。

<a id="common-plan-identity"></a>
## 共通予定identityと保存責務

通常カルテの単品/package追加は新しい共通 `medical_record_plan_items` を正本とする。関連revisionとowner linkを持つ。現 `treatment_plans` は入院兼用でtyped参照・4状態・版がなく、`treatments` createには在庫減算副作用があるため、意味を変えてpackage予定へ流用しない。

| 採用設計名 | 責務 |
|:---|:---|
| treatment_packages / treatment_package_items | 医院別master・安定component key。構成変更でも親versionを増やす |
| treatment_package_applications | 適用package版/構成snapshot、record、actor、C06 operation対応 |
| medical_record_plan_items | 単品/package共通identity、7typed master FK＋manual_other、clinic/record/pet整合、数量・金額等snapshot、由来・現在表示状態 |
| medical_record_plan_item_revisions | 変更前後/実施状態、理由、actor/時刻、operationの履歴 |
| medical_record_plan_owner_links | 実績/結果/会計ownerへの型付き対応、必要版/固定参照版、関連取消の履歴 |

共通schemaは[EMR-516](https://app.plane.so/baritechllc/browse/EMR-516/)窓口で統合する。[517](https://app.plane.so/baritechllc/browse/EMR-517/)/[500](https://app.plane.so/baritechllc/browse/EMR-500/)/package用に別々の予定表を増やさない。[既存3タブ契約](technical-spec.md#plan-domain-tabs-target-gap)の同一予定・結果追跡に接続する。

## ExactNumericとowner別精度

共通quantityは固定precision/scaleなしnumeric＋ExactNumeric、wireは正確な十進文字列とする。既存 `treatments.quantity` のnumeric(10,2)を全owner共通制約へ拡張しない。billing/estimateのnumeric精度を維持し、各ownerの許容数量・単位をadapterで検証する。保存先で表現できない値は丸めず不整合として拒否する。既存薬量・在庫整数等の条件をpackageが回避しない。

<a id="execution-coordinator"></a>
## 単一の実施状態writer

PlanExecutionCoordinatorを唯一のwriterとし、plan revisionのexecution eventを正本、clinical_stateを同TXで更新する読取投影とする。汎用PATCHからperformedを書けない。専用state-change intentは既存モーダル/確認・理由・版・編集権とtyped owner adapter検証を通す。

同じ明示実施操作にowner実績writeが含まれる場合は、owner write＋状態revision/投影＋必要監査＋C06 receiptを同じ最外TX/CASに参加させる。結果到着・attachだけでperformedにしない。D03の実施済み・結果待ちを保つため、結果レコードの存在もperformedの一律条件にはしない。検査等の明示実施記録をexecution eventに保持し、結果本文を先行生成しない。結果正本は各ownerのまま。EMR-500の親→子→結果の段階receiptを全体atomicへ変更しない。

clinicalの未実施/実施済み/後日予定/中止と、物販等のcommercial fulfillmentを別概念にする。登録料・済票料の新しい共通実施イベントは作らず、既存明示会計候補追加を使う。非臨床料金をclinical completedに偽装しない。

<a id="atomic-apply"></a>
## 原子的applyと履歴

previewはread-only。applyは全構成品の現在認可・record clinic・typed参照/有効性/必要版・価格等・既存安全条件を検査後、一つのTXでapplication＋全plan項目＋revision/由来＋監査＋receiptを保存する。実施・結果・会計確定はこのTXに含めない。

C06のimmutable contentと一時preconditionsを分け、同key異内容を拒否する。結果不明は元key照会へ戻す。同名・同日の別予定を併合しない。master更新は既存snapshotへ遡及させず、論理削除/取消後も対応・履歴を保持する。

旧treatment_plans/実績は再分類しない。旧移行は裏付けのあるtyped mappingのみ別計画にし、不明をcompleted・other・単価0で補完しない。

<a id="billing-certificate-adapter"></a>
## 会計・証明書adapter

同componentから請求ownerを二つ生成しない。実接種はvaccinationの単一sourceとし、登録料等を同vaccination_idへ複数bindingしない。既存 `uq_active_billing_vaccination_source` を緩めない。物販/非臨床料金は既存の明示会計候補追加経路を再利用し、packageから自動課金しない。[D10](../accounting/technical-spec.md)が会計正本。

証明書は[D08料金projection](../vaccinations/technical-spec.md#certificate-fee-projection-20261020)と[edition/renderer契約](../vaccinations/technical-spec.md#certificate-edition-proposal)へ接続する。`certificate_fee_role` は一般packageの必須属性ではなく証明書adapter側の任意mappingとする。発行前編集値はedition snapshotで固定し、元会計へ逆writeしない。preview/issue/reprintを新料金生成入口にしない。EMR-486/487担当の最終契約で具体化し、本票で重複実装しない。

現mainの限定確認範囲と固定SHAのrepo根拠は[作業記録](../../../work/emr/emr-518-treatment-packages-20261020.md#static-evidence)に置く。新技術方向と現実装を混同しない。

# EMR-518 selected contract fixtures

[OpenAPI](../../api.yaml)と[公開対応表](../../../../docs/work/emr/emr-518-package-api-contract-20261020.md)に付属する新規ダミーJSON群。正本は[業務4FR/12AC](../../../../docs/spec/domains/medical-records/treatment-packages.md)、[採用storage契約](../../../../docs/work/emr/emr-518-storage-contract-20261020.md)、[494共通操作](../../../../docs/work/emr/emr-494-edit-draft-contract-20261020.md)。実患者・医院料金・税務分類・医学的用量・実装完了の証拠ではない。

全fixtureは **NOT_RUN**。JSON/YAMLの構文・参照先の静的読取とfixtureの作成だけを行う。OpenAPI instance schema検証、runtime/Go/FE/DB試験、JCS canonicalizer実行、hash検算、Dockerは実行していない。manifestのexpected_schema valid/invalidは期待値であり検証PASSではない。

`manifest.json`にfile、operationId、実在component_ref、request/response区別、応答HTTP status、scenario、期待適否、NOT_RUNを記録する。通常scenario fileはHTTP bodyそのものでmetadata wrapperを付けない。`digest/`の2projectionはhash入力資料、canonicalization-vectorsは文字列境界資料でHTTP bodyではない。この3件はbody_kindで区別する。vectorsにはHTTP component_refを作らずnot_applicableとする。

各scenarioは独立したダミー状態であり、全JSONを一つの実データ列として投入しない。

成功例はrecord91/clinic7/pet801/package201/version3、collection版17→18、application302、plan401〜408。診察・処置・薬剤・検査・接種・健診・物販・正規手入力の8行を保持する。薬剤だけquantity0.25で、接種回数を小数にしない。保存値としての税込110を100へ変換しない。3masterの税null、物販保険null、版なし型native_version nullを保持し、単品で予定入力が許されるunknownはpreviewからwarningで保存する。合計は独自計算せずunresolved/null。明示0は欠落と別にする。

一般successのpreview tokenは`pkg-preview-v1:`＋64個の`a`、source tokenは`pkg-source-v1:`＋01〜08の左zero埋め64hex。いずれもschema形に合わせた**未検算dummy**で、projectionから計算したdigestではない。fixture同士の対応・適用履歴を説明する値であり、hash等価のgoldenには使わない。dummy以外の期待SHAを捏造せず、digest資料のexpected_sha256はnull/NOT_RUN。

実digest定義は[storage §3.4](../../../../docs/work/emr/emr-518-storage-contract-20261020.md#preview-digest-projection)。全key必須・未知key不可のSourceProjection/PreviewProjectionをserver実値から構築する。ComponentProjection.source_digestはprefixなし64hex、transportだけsource/preview prefix付き。生成plan/application ID・actor・時刻・edit-session/epoch・plan collection CAS・表示整形・合計可否をpreview投影へ含めない。保存application snapshotは生成plan IDsも含む別物。

RFC8785 JCSはUTF-16 key順、array順保持、UTF-8、BOM/末尾newlineなし。Unicode非正規化、duplicate JSON key/lone surrogate/invalid UTF-8拒否。quote/backslash/controlをJCS escapeし、slash/日本語/正常emojiは実文字のまま。null/false/文字列"0"を区別する。ID/version/円はcanonical整数文字列、quantity/tax_rateはexact canonical十進文字列でfloat演算なし。DB bytea32にprefixは保存せず、JSONB文字列hashを使わない。vectorsのexpected_canonical_utf8は末尾newlineなしの文字列bytes表現で、JSON内escapeはその文字列を格納するための表記である。

FEは公開された同版OpenAPIとこのfixtureを直接参照/コピーし、参照commitを記録する。手書き独自DTOや架空generic-master APIを作らない。FE契約mockはruntime/SQL完成待ちにしない。実API結合・臨床実施・会計・証明書等の本番有効化は各backend必要subset/owner受入後であり、この公開だけで完成しない。

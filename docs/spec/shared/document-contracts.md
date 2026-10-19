# 用語・文書状態・概念境界（C02）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | shared.document-contracts／共通契約の整理候補 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載の継承。新業務承認を含まない。根拠不明は未確認のまま。未回答案は別節 |
| 実装状態 | 既存資料の記載を継承（コード全体の再監査なし）。選定仕様と既存挙動の差を明記 |
| 実機検証状態 | 未実施 |
| 文書確認日・基準 | 2026-10-15／PR #670 commit 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8 |
| 対象 | 全D01–D23の概念境界確認と既存語彙正本への接続 |

本書は既存正本への対応と受入観点を束ねる。数値・語彙・業務条件の第二正本は作らず、リンク先の有効な個別契約を優先する。FR/ACの識別子付与は追加承認や実装済みを表さない。未確認の不足を対象外へ移して完了としない。

## 正本と責務

[記述規約](../authoring-guide.md#status)と[ontology](../ontology/README.md)が正本。本書は全領域の意味の衝突を点検する入口で、各概念の完全な定義・関係・多重度を複製しない。D番号は仕様整理の管理単位、BE package・table・画面route数とは異なる。

<a id="requirement-fr-common-doc-status"></a>
### FR-common-doc-status — 決定・実装・検証の別軸

未決／提案／合意済みを区別し、合意は決定者・日付・根拠・範囲を持つ。決定根拠不明な既存本文は「既存記載を継承（決定根拠未確認）」とする。実装・実機は別軸で、ADR Accepted・Issue Closed・索引掲載・ID付与を承認／runtime PASSへ読み替えない。過去の「未決」はその時点の履歴で、後続合意で解決した項目を再び未決へ戻さない。

<a id="requirement-fr-common-doc-identity"></a>
### FR-common-doc-identity — 正本とID

安定文書ID・意味アンカー・既存FR/ACを保持する。業務はspec、実現方式はdesign、wire型はapi.yaml、DBはmigration、進捗はPlane、実行証跡はwork/opsへ。旧番号入口を新たな正本にしない。名称変更をコード・ラベルの一括改名に広げない。

<a id="requirement-fr-common-doc-ontology"></a>
### FR-common-doc-ontology — 変更時の概念照合

変更した概念の対象・対象外、推奨語／別名、関係方向・多重度・必須任意・状態・制約をFR/ACと照合する。図だけ／表だけの更新にしない。不足は未記載／要確認、検証不足は未確認、非適用は理由付きとし、空欄で混ぜない。既存の[予約語彙](../ontology/reservations.md)、[カルテ語彙](../ontology/medical-records.md)、[会計語彙](../ontology/accounting.md)を再定義しない。

## 全領域の概念境界チェック

以下は意味を混ぜないための確認対応。専用ontologyがない領域の詳細定義・多重度をこの一覧だけで確定しない。各D候補と最新正本の照合が完了条件。

| 領域 | 区別して保つ概念／他ownerとの境界 | 正本の入口 |
|:---|:---|:---|
| D01 | 予約区分・表示グループ・業務分類、指名／担当、予約と受付／記録 | [予約語彙](../ontology/reservations.md) |
| D02 | 表示分類・カード位置と予約実状態、片付けと業務取消 | [受付](../domains/reception/README.md) |
| D03 | 入力者・確認医師・担当医・指名医、保存／確定／追記／過去入力 | [カルテ語彙](../ontology/medical-records.md) |
| D04 | 飼主・ペット基本台帳と所有関係、identity groupとは別 | [飼主](../screens/04-owners-form.md) |
| D05 | 入院とホテル、予定予約と実滞在、退院と請求 | [入院詳細](../screens/08-hospitalization-detail.md) |
| D06 | 予約intentと施術記録、コース／optionと会計明細 | [施術フォーム](../screens/17-trimming-form.md) |
| D07 | 受信・患者対応・結果保存／確定／訂正、fixtureとdevice経路 | [検査](../screens/13-examinations-form.md) |
| D08 | 接種事実と次回予定、記録担当とD20公開 | [接種](../screens/15-vaccinations-form.md) |
| D09 | checkup正系統と旧検査系、package／field／result | [健診](../screens/25-checkups-list.md) |
| D10 | 請求／実入金／充当／残高復元／実返金、再請求と再発行 | [会計語彙](../ontology/accounting.md) |
| D11 | 見積と確定請求、転記と入金 | [見積](../screens/26-estimate-detail.md) |
| D12 | 実査／締め／売上／実入出金、原請求日と処理日 | [締め](../cash-register.md) |
| D13 | 在庫記録／残数、物販品目マスタと自動引落し | [在庫](../screens/18-inventory-list.md) |
| D14 | 職員／職種／勤務／休憩／予約対応能力 | [職員](../screens/settings/master-staff.md) |
| D15 | tag／配信policy／送信ログ／停止、定義済みと稼働 | [Lステップ](../line/lstep-integration.md) |
| D16 | 集計読取／CPM評価とD15配信、締めとレポート | [集計](../customer-aggregation.md) |
| D17 | 設定入口と個別master write-owner、医院とglobal設定 | [医院設定](../screens/19-clinic-settings.md) |
| D18 | 院内設定と飼主予約app、公開枠と空き判定 | [予約チャネル](../line/reservation-spec.md) |
| D19 | 手動link/unlinkと台帳統合、自動同一判定を追加しない | [identity](../screens/40-identity-links.md) |
| D20 | 自己認証／紐付け／健康公開とstaff認可、原記録write | [健康手帳](../screens/38-liff-pet-health.md) |
| D21 | スタッフ専用の複合読取／section縮退と記録更新、患者context。一般履歴は印刷・共有なし、最新採用の予防薬専用処方箋・年間投薬表だけ印刷を許容。飼主公開はD20、D03等の既存印刷とは別 | [飼主レポート](../screens/39-owner-report.md) |
| D22 | 製品manual記事／版と仕様管理文書 | [manual](../screens/35-internal-manual.md) |
| D23 | 共有ボードREAD／報告医院mutation／個人history | [support本文](../screens/settings/README.md) |

## 受入条件

| 条件ID | Given／When | Then／独立確認・成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-common-doc-status | 同ファイルに合意・旧実装・提案が混在 | 節ごとの状態と根拠が区別され、ファイル全体を合意済みにしない | authoring-guideと候補本文照合 |
| AC-common-doc-identifiers | 配置／見出し更新 | 旧IDと参照先を保持し、正本が重複しない | link/anchor/ID静的照合 |
| AC-common-doc-relations | 概念・多重度変更 | 語彙・関係図・FR/AC・利用側が同じ制約を指す | 変更概念のtraceレビュー |
| AC-common-doc-history | 旧未決を最新合意で解消 | 旧履歴を根拠に解決済みを未決へ戻さない | 決定根拠と現在節の照合 |
| AC-common-doc-domain-map | D01–D23候補を再照合 | 各行の概念境界・定義不足・他owner受渡しを記録 | 全23領域の正本・語彙／IDとowner対応表の独立文書照合済み |
| AC-common-doc-verification | 文書検査のみ実施 | runtime／実装適合／法令適合のPASSを主張しない | 検査証跡に実行範囲を明記 |

## 未確認・完了境界

専用ontology（予約・カルテ・会計・飼主患者・滞在・施術・金融周辺）と、各領域本文内の定義を全23領域のowner対応表で照合した。ホテル任意担当・D20限定公開・D21閲覧専用などの語彙／状態の衝突を修正し、独立レビューで未解消の重大指摘なし。完成判定は今回明示した製品仕様の文書範囲で、未発見のあらゆる概念を網羅する無制限な辞書完成や、既存全ファイルの改名・実装適合を意味しない。新しいconsumer／概念は正本とACを更新してから適用する。

## 全ownerへの適用照合

[全23領域の具体的な意味ID・保存・安全・保持・失敗条件との対応](owner-contract-coverage.md)を参照する。原則の再掲だけで個別操作の受入を満たしたとは扱わず、引用先のFR／AC・技術契約と照合する。

## 2026-10-15の用語・優先関係

通常カルテの「保存」「診療終了」「確定済み請求」「検査確定」は別概念。[D03最新合意](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)では医師確定を通常本文編集やスタッフ診療履歴掲載の必須条件にしない。旧医師確定・ロック表は現行実装／過去合意として保持し、最新目標へ混ぜない。クリアはタブ開始時、履歴複製取消は複製直前という基準を識別する。D06コース種別はコースのカテゴリ、勤務templateは勤務入力案であり、独自割引scopeや無断上書き権限を生まない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../domains/vaccinations/lifecycle.md#certificate-editions-20261017)。通常縦横・狂犬病の3様式は旧見本に従う。訂正後新発行・旧版履歴保持、再印刷は最新版、旧訂正版は有効印刷不可。共通基本情報更新で発行版は自動変更しない。旧印刷時自動登録等は復活させない。専用証明書実装の所在・適合と仕様採用を分ける。

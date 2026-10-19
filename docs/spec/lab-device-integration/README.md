# 検査機器連携（Lab Device Integration）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | reference.lab-device-integration.index／解析資料索引・既存正本への接続 |
| 決定状態 | 旧解析記録を継承。2026-10-19の承認は記述訂正・正本接続・未確認事項の明記のみ。新規連携の採用なし |
| 実装状態 | 記載コードへの参照。全コードの適合・配備・撤去完了は未確認 |
| 実機検証状態 | 今回未実施。過去観測と現在の運用承認を分ける |
| 文書確認日 | 2026-10-19 |


本ディレクトリは城東の旧システム（Drワン）の構造・通信条件を記録した解析資料と、追加解析の手順例です。旧観測、コード参照、確定済みの移行先業務仕様、実機受入・医院切替の状態を分けます。解析結果だけで旧環境の撤去、新機器の運用、外注・画像専用連携やデータ移行を承認しません。

---

## 1. 概要と背景

動物病院では日々の診療において、血液検査や尿検査などの院内機器が稼働しています。
従来（旧城東医院等）は、Windows端末上で動作する専用クライアント「**Drワン**」がRS-232C（シリアル通信）経由で各機器から測定データを受信し、ローカルデータベース（Access MDB）および旧カルテ（NOAH）に展開していました。

新カルテの日常受信は[現行接続正本](../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)と[ADR008 current daily path](../../architecture/adr/008-local-lab-device-agent.md#history-and-current-daily-path-2026-10-14)を優先します。HTTPS ChromeでNX600/AU10V専用ポートを明示選択し、最大2ポート・同一origin排他・ページ表示中のみ受信専用で読み、認証済みAPIへ送ります。旧Go LaunchAgentは任意の手動legacy経路で、同じ機器ポートをブラウザ経路と同時利用しません。コードの存在は実機受入・医院での切替や撤去の完了を意味しません。

その根拠となるのが、旧環境（Windows 7端末）からサルベージされた Drワン一式（`old_db/clinics/jouto/lab-drwan/`）に対する **REA (Reverse Engineer Anything)** および静的調査の解析結果です。

---

## 2. 対象機器と通信仕様サマリー

城東の旧シリアル設定（`mdcon*.cmd`）と記録された過去観測、新カルテのコード参照を分けた一覧です。旧COM番号をMacのポートと同一視せず、他医院へ一般化しません。下表のSourceTypeは現行wire値です。

| 機器名 | 機種概要 | 通信ポート | 通信設定 | 新カルテの SourceType | 備考 |
|---|---|---|---|---|---|
| **富士ドライケム NX600** | 生化学自動分析装置 | 旧COM6 | 9600 bps, 8N1 | `fuji_nx600` | 城東過去観測16項目。日常Web Serial対象、実機受入・医院公開は別確認 |
| **富士ドライケム AU10V** | 免疫反応測定装置（ホルモン/T4等） | 旧COM7 | 9600 bps, 8N1 | `fuji_au10v` | 城東過去観測。日常Web Serial対象、実機受入・医院公開は別確認 |
| **アークレイ ポケットケム PU-4010** | 尿分析装置 | 旧COM3 | 2400 bps, 設定7E1／現場観測8E1 (※) | `arkray_pu4010` | decoder-only。差が未review、日常選択・agent運用対象外 |
| **IDEXX VetLab Station** | 血液検査総合ステーション | 旧COM5 | 9600 bps, 8N1 | `idexx_vetlab` | decoder/persist/PIMS応答コードあり。日常対象外、医院ケーブルの応答送信・常時運用は未承認 |
| *富士ドライケム FDC-3500* | 旧生化学装置 | 旧COM1 | 9600 bps, 7E1 | — | 旧設定の参照。新規運用の採用なし |

※ 通信ハンドシェイクおよび電文フォーマットの詳細は [`drwan-binary-analysis.md`](./drwan-binary-analysis.md) を参照。

---

## 3. 新カルテ（AnimalEkarte）における実装構成

以下は`backend/internal/medicalrecord/`のコード参照です。decoder/persistの存在、プロトコル忠実度、実機運用承認は別の状態です。API型の正本は[api.yaml](../../../backend/docs/api.yaml)、適合差と停止条件は[D07技術契約](../../design/domains/examinations/technical-spec.md#examination-device-contract)へ委ねます。

* **カタログ定義**: [`backend/internal/medicalrecord/lab_device_item_catalog.go`](../../../backend/internal/medicalrecord/lab_device_item_catalog.go)
  * 城東の初期観測カタログ36項目（NX600 16 + AU10V 1 + 尿 8 + IDEXX 11）。旧全項目・全医院の網羅や具体基準値の医師承認ではない
* **電文デコーダ**:
  * [`lab_device_fuji.go`](../../../backend/internal/medicalrecord/lab_device_fuji.go): 富士ドライケム（NX600 / AU10V）の固定長・セパレータ電文パース
  * [`lab_device_urine.go`](../../../backend/internal/medicalrecord/lab_device_urine.go): 尿分析装置 PU-4010 の電文パース
  * [`lab_device_idexx.go`](../../../backend/internal/medicalrecord/lab_device_idexx.go), [`lab_device_idexx_pims.go`](../../../backend/internal/medicalrecord/lab_device_idexx_pims.go): IDEXX VetLab の PIMS シリアルセッション管理（STX/ETX、チェックサム、ACK/NAK応答）
* **受信・永続化サービス**: [`lab_device_receive_service.go`](../../../backend/internal/medicalrecord/lab_device_receive_service.go)
  * テナント（`clinic_id`）分離。本日診療中カルテカードで明示選択した有効待機の患者へ自動persistする。医院ごとの有効待機は1件で、待機なし・期限切れは未紐付け。電文の検体ID・旧CLID/PTID・氏名・同日受診だけで患者を自動決定しない。後付けはjob＋petの現在医院・関連を照合する（[D07受信](../domains/examinations/lifecycle.md#fr-examinations-receive)）
* **マッピング・LLM推論補助**: [`lab_device_mapping_draft.go`](../../../backend/internal/medicalrecord/lab_device_mapping_draft.go), [`lab_device_mapping_llm.go`](../../../backend/internal/medicalrecord/lab_device_mapping_llm.go)
  * clinic＋source＋機器原コードと院内マスタの対応を管理・推薦。名称一致やLLM候補だけで対応を確定しない（[D07技術契約](../../design/domains/examinations/technical-spec.md#examination-device-contract)）

---

## 4. ドキュメント一覧

1. **[`drwan-binary-analysis.md`](./drwan-binary-analysis.md)**
   旧システム「Drワン」のバイナリ構造、VB6 ランタイム特性、通信設定コマンド（`mdcon*.cmd`）、項目変換テーブル（`Drimke.tbl`）、MDBスキーマの解析まとめ。
2. **[`rea-reproduction-guide.md`](./rea-reproduction-guide.md)**
   **REA (Reverse Engineer Anything)** を使って追加のバイナリ解析や新機器調査を行う際の手順書。

---

## 5. セキュリティ・プライバシー原則（重要）

* **PII（個人情報・患者情報）非開示の徹底**:
  * 旧MDB（`mkan.mdb`）や受信バッファ（`djusin.*`）に含まれる飼い主氏名、ペット名、電話番号、住所、生の診療検査値は、ドキュメントやコミットログに一切転記しません。
  * 記録するのは「通信プロトコル仕様」「ボーレート」「項目コード体系」「テーブル/カラム定義」などの**構造情報のみ**に限定します。

---

## 6. 周辺連携（外注検査・画像機器）と未確認事項

以下は旧資産に記録された痕跡です。実利用・形式・対応範囲と移行初日の最低継続要件は未確認であり、新カルテの専用連携採用・対象外・Phase2の決定にしません。

1. **外注検査センター連携（大阪血清）**:
   * `大阪血清050611.exe`（VB5製）と更新処理の痕跡。媒体・輸送方式・format・検査種類・患者/依頼ID・結果版/訂正キー・重複/再配送・未知患者・失敗通知/担当、現在の利用と移行初日必須性は未確認。機密formatや窓口への追加調査は別作業。
2. **画像診断機器連携（PicJoint）**:
   * `picset.cmd`にPicJoint共有先とCR/US/CT/ES等の設定痕跡。これだけでは実画像形式、DICOM/PACS、患者同定、自動監視・取込を証明しない。旧viewer/共有参照、既存添付への移送、専用連携を分け、実形式/サイズ/series、患者・撮影日・画像ID、差替/ACL/旧path、医院別利用と初日必須性を未確認として残す。
3. **医院別の未確認領域（八王子等）**:
   * 解析記録上、八王子（`old_db/clinics/hachioji/`）の機器PC複製は未回収。機種・台数・配線・外注/画像利用は未確認。城東のCOM/36項目/単位/旧ID体系を医院横断で適用しない。資料回収/ヒアリング/実機確認は別作業。


<a id="analysis-contract-boundaries"></a>
## 7. 解析資料と確定済み移行先の境界

業務・FR/ACは[D07](../domains/examinations/lifecycle.md#examination-document-scope)、型・writer/互換差は[D07設計](../../design/domains/examinations/technical-spec.md#examination-device-contract)、概念は[検査オントロジー](../ontology/examinations.md#examination-state-and-receive-concepts)、接続・運用条件は[接続正本](../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)と[legacy手順の適用範囲](../../ops/deploy/lab-devices/LAB_DEVICE_AGENT_MACOS.md)へ委ねます。以下は既存条件との接続と調査不足の区別で、新しいFR/AC・実装・実機合格ではありません。

| 観点 | 維持する境界・確認先 |
|:---|:---|
| 旧IDと患者指定 | 旧CLID/PTID・検査sessionと新clinic/owner/pet/record/exam/jobの対応は未検証。移行専用の対応照合と通常の有効待機受信を混同しない。[受信と同定](../domains/examinations/lifecycle.md#fr-examinations-receive) |
| コード・原値・基準 | 旧内部番号をsource/exam codeにせずシリアルの原ラベル/suffixを使う。原値文字列・定性/不等式/欠測・単位・保存時の範囲/版を維持し、未知はneeds_review。旧単位/種/当時範囲が不明なら現在マスタで補完せず医師確認へ。[mapping](../domains/examinations/lifecycle.md#fr-examinations-mapping)・[master](../domains/examinations/lifecycle.md#fr-examinations-master) |
| 日時・再送・訂正 | 測定日時と受信監査時刻、1frame=1jobからのN検査、同医院/source/fingerprint、未紐付け、受信とpersistの成否を分ける。全由来検査のdetachとfixture終端revertは別。使用/confirmed/official保護・原版/理由/actor/履歴を保持する。[再送/補償](../domains/examinations/lifecycle.md#fr-examinations-idempotency)・[訂正](../domains/examinations/lifecycle.md#fr-examinations-revision)・[失敗](../domains/examinations/lifecycle.md#fr-examinations-recovery) |
| PIMSと依頼送信 | I/sへのセッション応答は測定受信の維持であり、検査依頼・患者登録・取消/受付リスト送信の採用を意味しない。日常は受信専用。旧運用の依頼IDや送受信の実態は別確認。[接続正本](../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md#idexx追記2026-08-1920)・[計画/結果の区別](../ontology/examinations.md#clinical-results-and-certificates-boundary) |
| 認証・運用担当 | 認証APIの現在clinic・対象照合とlab-import操作別grantを保持。legacyのloopback/Host/Origin/PNA/consumer境界を日常必須credentialや同一OSユーザー隔離保証へ広げない。具体基準値は医師判断、設定/停止/再送/復旧担当と実医院grantは別検証。新しい必須役職を追加しない。[権限・運用](../../design/domains/examinations/technical-spec.md#権限全writer)・[SECURITY](../../../SECURITY.md) |
| 画像添付と専用連携 | 既存JPEG/PNG/GIF/PDFの撮影・選択・D&D・閲覧、医院/患者/添付ID・保存版/認可・既存容量制限を維持。旧PicJoint参照・DICOM/PACS・自動連携の採用とは別。[既存画像契約](../domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-image-drop-selected-print) |

<a id="migration-and-phase-unknowns"></a>
### 移行最低互換とPhaseの未確認事項

大阪血清・PicJointについて、現利用と医院別の移行初日に必要な結果/画像参照、依頼送信の要否、旧履歴の対象期間・ID対応・訂正系列・移送/旧参照継続・受入証跡は未確認です。必要な継続業務と、新たな専用自動連携の採否/Phaseを別に判断する課題として残し、旧運用/手入力・旧viewerで継続可能とも自動化必須とも推測しません。通常受信がDrワンMDB/rawを入力にしない契約は維持し、別の移行設計を本資料から採用も全面禁止もしません。

[D03のPhase2記述](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-conversation-summary)は会話要約節の音声入力・外部サービス連携の文脈です。既決の検査機器連携全体をPhase2へ延期する根拠にせず、外注・画像をPhase1へ追加採用する根拠にもしません。旧資料の痕跡や解析手順は機器接続・データ移行・外部問い合わせ・実装の実行許可ではありません。

# 検査機器連携（Lab Device Integration）

本ディレクトリは、動物病院の院内検査機器（生化学自動分析装置、血球計算機、内分泌測定機、尿検査機など）と電子カルテシステム（AnimalEkarte）の自動データ連携に関する仕様、旧システム（Drワン）のバイナリ解析結果、および新カルテの実装方針を記録する専用ドキュメントです。

---

## 1. 概要と背景

動物病院では日々の診療において、血液検査や尿検査などの院内機器が稼働しています。  
従来（旧城東医院等）は、Windows端末上で動作する専用クライアント「**Drワン**」がRS-232C（シリアル通信）経由で各機器から測定データを受信し、ローカルデータベース（Access MDB）および旧カルテ（NOAH）に展開していました。

新電子カルテシステム（AnimalEkarte）では、レガシーなWindowsクライアント環境に依存せず、**現場の受信エージェントおよびGoバックエンドが機器からの生電文を直接解釈・永続化するアーキテクチャ**を採用しています。

その根拠となるのが、旧環境（Windows 7端末）からサルベージされた Drワン一式（`old_db/clinics/jouto/lab-drwan/`）に対する **REA (Reverse Engineer Anything)** および静的調査の解析結果です。

---

## 2. 対象機器と通信仕様サマリー

旧システムのシリアル通信設定（`mdcon*.cmd`）および新カルテのアダプタ定義に基づく機器一覧です：

| 機器名 | 機種概要 | 通信ポート | 通信設定 | 新カルテの SourceType | 備考 |
|---|---|---|---|---|---|
| **富士ドライケム NX600** | 生化学自動分析装置 | COM6 相当 | 9600 bps, 8N1 | `FUJI_NX600` | 現場確認済（16項目） |
| **富士ドライケム AU10V** | 免疫反応測定装置（ホルモン/T4等） | COM7 相当 | 9600 bps, 8N1 | `FUJI_AU10V` | 現場確認済 |
| **アークレイ ポケットケム PU-4010** | 尿分析装置 | COM3 相当 | 2400 bps, 7E1 (※) | `ARKRAY_PU4010` | 現場8E1受信、設定7E1 |
| **IDEXX VetLab Station** | 血液検査総合ステーション | COM5 相当 | 9600 bps, 8N1 | `IDEXX_VETLAB` | PIMSシリアル・双方向セッション |
| *富士ドライケム FDC-3500* | 旧生化学装置 | COM1 相当 | 9600 bps, 7E1 | — | 旧機器（参照用） |

※ 通信ハンドシェイクおよび電文フォーマットの詳細は [`drwan-binary-analysis.md`](./drwan-binary-analysis.md) を参照。

---

## 3. 新カルテ（AnimalEkarte）における実装構成

バックエンドの `backend/internal/medicalrecord/` 配下に、解析結果に基づいたクリーンなパーサー・サービス層が実装されています：

* **カタログ定義**: [`backend/internal/medicalrecord/lab_device_item_catalog.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_item_catalog.go)
  * 城東観測カタログ（NX600 16 + AU10V 1 + 尿 8 + IDEXX 11 = 全36項目）
* **電文デコーダ**:
  * [`lab_device_fuji.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_fuji.go): 富士ドライケム（NX600 / AU10V）の固定長・セパレータ電文パース
  * [`lab_device_urine.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_urine.go): 尿分析装置 PU-4010 の電文パース
  * [`lab_device_idexx.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_idexx.go), [`lab_device_idexx_pims.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_idexx_pims.go): IDEXX VetLab の PIMS シリアルセッション管理（STX/ETX、チェックサム、ACK/NAK応答）
* **受信・永続化サービス**: [`lab_device_receive_service.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_receive_service.go)
  * テナント（`clinic_id`）分離、当日受診（TodayVisit）自動紐付け、未紐付けバッファ管理
* **マッピング・LLM推論補助**: [`lab_device_mapping_service.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_mapping_service.go), [`lab_device_mapping_llm.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_mapping_llm.go)
  * 機器の項目コードと院内マスタ（カルテ検査項目）の対応関係を管理・推薦

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

## 6. 周辺連携（外注検査・画像機器）と今後の拡張

本ディレクトリの主スコープである「院内リアルタイムシリアル検査機器」に加え、旧環境から以下の周辺連携の痕跡が確認されています：

1. **外注検査センター連携（大阪血清）**:
   * `大阪血清050611.exe`（VB5製）。外部委託検査データの更新バッチ。
2. **画像診断機器連携（PicJoint）**:
   * `picset.cmd`。院内ファイルサーバ（`\\DKSERVER\PicJoint\`）経由での CR（レントゲン）、US（エコー）、CT、ES（内視鏡）等の画像ファイル連携。
3. **医院別の未対応領域（八王子等）**:
   * 八王子医院（`old_db/clinics/hachioji/`）の検査機器資産は未回収のため、他社機器の導入状況に応じて本手順（REA）による追加調査を予定。


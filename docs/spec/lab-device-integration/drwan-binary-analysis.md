# Drワン バイナリおよび通信仕様 解析レポート

**解析対象資産**: `old_db/clinics/jouto/lab-drwan/`（城東医院 Windows 7 端末複製）  
**解析ツール**: REA (Reverse Engineer Anything) v6.7.0, 静的解析ツール群  
**方針**: 個人情報（飼主名・ペット名・住所・生検査値など）は一切出さず、通信パラメータ、バイナリ構造、テーブル仕様、コード帯のみを記録する。

---

## 1. バイナリ・アーティファクト解析結果 (REA)

旧クライアントの実行ファイル群（`C->ProgramFiles->DRワン/` 配下）を REA の `inspect-artifact` および `inspect-managed-artifact` で検証した結果です。

### 1.1 実行ファイル仕様
* **対象バイナリ**:
  * `Wanrsin.exe`（3,227,648 bytes, SHA256: `667d340e51...`）: IDEXX VetLab Station 専用受信・通信管理
  * `WanView.exe`（4,046,848 bytes）: 検査結果表示・カルテ閲覧
  * `jyusin.exe`（3,801,088 bytes）: 汎用シリアル受信デーモン
  * `WanCC.exe`, `mwsmenu.exe`, `sbd110.exe` 等
  * `大阪血清050611.exe`（39,424 bytes）: 外注検査会社（大阪血清）データ更新バッチ（Visual Basic 5.0 製）
* **REA 判定結果**:
  * **フォーマット**: Windows PE32 (GUI) Intel 80386 (32-bit x86)
  * **マネージド分類**: `not-managed`（`.NET` アセンブリではなく Win32 ネイティブ）
  * **ランタイム依存**: **Visual Basic 6.0 (`MSVBVM60.DLL`)**、一部ツールは **Visual Basic 5.0 (`MSVBVM50.DLL`, `VB5JP.DLL`)**
  * **COM/OCX コンポーネント**: `COMDLG32.OCX`, `CrDraw.ocx`, `CrView.ocx`, `Wandsp.ocx` などを利用。

---

## 2. 機器別シリアル通信設定（`mdcon*.cmd`）

`C->ProgramFiles->DRワン/mdcon*.cmd` のスクリプトに、実稼働していた各検査機器のRS-232C通信パラメータが完全に記録されていました。

### 行フォーマット仕様
* 1行目: 機器データ格納サブフォルダ
* 2行目: COM ポート番号
* 3行目: ボーレート (bps)
* 4行目: パリティ（`N`=None, `E`=Even, `O`=Odd）
* 5行目: データビット（7 or 8）
* 6行目: ストップビット（1 or 2）
* 8行目: Drワン画面上の表示名

### 機器設定一覧表

| ファイル | 機器フォルダ | COM | 設定 | 画面表示名 | 接続機器 / 備考 |
|---|---|---|---|---|---|
| `mdcon2.cmd` | `\PU4010\` | COM3 | 2400 bps, 7E1 | 尿検査 | アークレイ ポケットケム PU-4010 (※現場受信は 8E1) |
| `mdcon3.cmd` | `\FDC7000\`| COM4 | 9600 bps, 8N1 | ドライケム | 富士ドライケム 7000 (予備/過去) |
| `mdcon4.cmd` | `\IDEXX\` | COM5 | 9600 bps, 8N1 | ＩＤＥＸＸ | IDEXX VetLab Station（末尾フラグ `60000 / 1`） |
| `mdcon5.cmd` | `\NX600\` | COM6 | 9600 bps, 8N1 | ドライケム(新) | 富士ドライケム NX600（現行中核機） |
| `mdcon6.cmd` | `\AU10V\` | COM7 | 9600 bps, 8N1 | ホルモン | 富士ドライケム AU10V（内分泌・免疫） |
| `mdcon0.cmd` | `\FDC3500\`| COM1 | 9600 bps, 7E1 | FDC-3500 | 富士ドライケム 3500（旧機種） |
| `mdconM.cmd` | `\MANUAL\` | — | — | 手入力項目 | 院内手動入力用 |

---

## 3. 項目マスタ変換ルール（`Drimke.tbl` & `ike`）

機器から送信される「文字ラベル」と、カルテ側の「内部項目コード」を紐付けるテーブルです。

### 3.1 `Drimke.tbl` の構造
テキスト形式（Shift-JIS）で「機器側ラベル」「フラグ」「内部コード」が定義されています。

```text
GGT-PS      0   100
GOT-P       0   110
GPT-P       0   120
BUN-P       0   150
GLU-P       0   160
...
IRBC        0   910
IWBC        0   950
IRETIC      0   1030
INEU        0   1040
IBUN        0   1310
IALT        0   1370
INa         0   1610
```

### 3.2 項目コード帯の分類

| コード帯 | 対象カテゴリ | プレフィックス / 例 | レコード数規模 (dkensa) |
|---|---|---|---|
| **100–360** | 富士ドライケム 生化学 | GGT, AST/GOT, ALT/GPT, BUN, GLU, CRE | 約 41.6 万行 |
| **400–420** | 電解質 | Na, K, Cl | 約 8.7 万行 |
| **510–580** | 汎用 CBC (血球計算) | WBC, RBC, HGB, HCT, PLT | 約 30.1 万行 |
| **590 / 700** | PCV / 黄疸指数 | 手入力・遠心分離測定 | 約 14.2 万行 |
| **910–1270** | **IDEXX 血球 (ProCyte/LaserCyte)** | `I*` (IRBC, IWBC, IHCT, IRETIC, INEU 等) | **約 134.3 万行** |
| **1310–1630**| **IDEXX 生化学 (Catalyst)** | `I*` (IBUN, IALT, IGLU, INa, IK 等) | **約 34.6 万行** |
| **2010–2080**| 免疫・ホルモン (AU10V) | T4, COR (コルチゾール), v-LIP 等 | 約 7,100 行 |
| **3310–3470**| 尿検査 (PU-4010) | GLU, PRO, BIL, PH, BLD, KET, URO 等 | 約 19.5 万行 |

---

## 4. 旧データ保存構造（JET Access MDB）

データ根（`DRワン/ワンdat/`）の MDB ファイル群の構成です。

| MDB ファイル | テーブル名 | 主な役割 |
|---|---|---|
| **`mkan.mdb`** | `mkanja` | 患者マスタ（旧NOAH の CLID / PTID とのリンク） |
| | `hiduke` | 検査セッション（`hidhdpo`: セッション主キー, 日付整数 `YYYYMMDD`） |
| | `dkensa` | 項目別測定結果（`dkehdpo`: セッションID, `dkekocd`: 項目コード, `dkekent`: 測定値文字列） |
| | `dgdspkou` | 表示順序テーブル |
| **`ikensa.mdb`** | `ike` | 項目マスタ（コード、名称、単位、基準値範囲） |

---

## 5. IDEXX VetLab Station の双方向プロトコル解析

`Wanrsin.exe` の逆アセンブルおよび通信キャプチャ検証により、IDEXX 機器との双方向ハンドシェイク仕様が判明しています。

* **接続方式**: RS-232C (COM5, 9600 bps, 8N1)
* **セッション維持**:  
  単なる一方通行のデータ受信ではなく、IDEXX VetLab Station 側から定期的なポール信号が送られます。
  * `I` を受信時 ➔ `ACK` + `A` + `IM` を返送
  * `s` を受信時 ➔ `ACK` + `A` + `SM` を返送
  * これに応答しないと PIMS セッションが切断され、データ送信が停止します。
* **データフレーム仕様**:
  * フォーマット: `[STX] '0' [Source:1B] 0x80 [80 | PayloadLen] [Payload...] [Checksum:1B] [ETX]`
  * チェックサム計算: `(sum(payload bytes) & 0x7F) | 0x80`

新カルテの Go 実装（[`lab_device_idexx_pims.go`](file:///Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/medicalrecord/lab_device_idexx_pims.go)）はこの双方向ステートマシンを完全再現しています。

---

## 6. 新カルテ（AnimalEkarte）への反映成果

本解析により、以下の成果が AnimalEkarte に反映されました：

1. **ゼロ・レガシー化**:
   * Windows 7 や VB6 アプリケーションを完全に撤去し、Mac / Linux / Docker コンテナ上で動作する軽量 Go エージェントで直接シリアル電文を受信可能にした。
2. **完全な型定義カタログ**:
   * `backend/internal/medicalrecord/lab_device_item_catalog.go` に、実データに基づく 36 項目の初期観測カタログを確定。
3. **安全なマッピング**:
   * 富士ドライケム（100–420帯）と IDEXX（910/1310帯）でコード体系が分離されている歴史的経緯を踏まえ、新カルテでは `SourceType` 別に独立したマッピングを適用。

---

## 7. その他の外部連携資産（外注検査・画像診断機器）

Drワン資産の追加調査により、リアルタイムシリアル通信以外にも以下の連携機能が存在することが確認されました。

### 7.1 外注検査機関連携（大阪血清）
* **対象バイナリ**: `大阪血清050611.exe`
* **アーキテクチャ**: Visual Basic 5.0（`MSVBVM50.DLL`, `VB5JP.DLL`）
* **役割**: 院内機器では測定できない外部委託検査（アレルギー検査、特殊病理、PCR等）のデータフロッピー・CDやオンラインファイルを Access MDB / カルテに更新（`CmdUpdate`）するバッチ。

### 7.2 画像診断機器連携（`picset.cmd` と PicJoint 共有）
* **設定ファイル**: `C->ProgramFiles->DRワン/picset.cmd`
* **連携先共有フォルダ**: `\\DKSERVER\PicJoint\`
* **対象モダリティ（画像種別コード）**:
  * **CR**: レントゲン（X線撮影装置）
  * **US**: 超音波診断装置（エコー）
  * **CT**: コンピュータ断層撮影装置
  * **ES**: 内視鏡（エンドスコープ）
  * **DC**: デジタルカメラ（患部写真）
  * **XC / MS**: その他特殊診断機器
* **新カルテへの含意**:  
  検査結果（数値）だけでなく、各画像機器から出力される DICOM / JPEG 画像がファイルサーバ経由でカルテに紐付けられていた運用痕跡を示しています。新カルテでの画像管理・メディア添付機能の仕様検討に活用可能です。

---

## 8. 医院別状況（城東 vs 八王子）

* **城東医院 (`jouto`)**:
  * 現場 Windows 7 端末から Drワン一式（Program Files, MDBデータ, 設定スクリプト）が完全に複製されており、本ドキュメントの解析が完了しています。
* **八王子医院 (`hachioji`)**:
  * 現在のレポジトリ資産（`old_db/clinics/hachioji/`）には検査機器連携PCの複製がまだ存在しません（`README.md` では `—` 状態）。
  * 八王子で異なる検査機器（他社血球計や生化学機）が使われている場合は、現場端末のサルベージまたはヒアリングを実施し、本手順書に沿って REA で追加解析を行う必要があります。


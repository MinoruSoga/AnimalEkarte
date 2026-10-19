# Drワン バイナリおよび通信仕様 解析レポート

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | reference.lab-device-integration.drwan-analysis／旧資産解析記録 |
| 決定状態 | 旧解析記録を継承。2026-10-19の承認は記述訂正・正本接続・未確認事項の明記のみ。新規連携の採用なし |
| 実装状態 | 記載コードへの参照。全コードの適合・配備・撤去完了は未確認 |
| 実機検証状態 | 今回未実施。過去観測と現在の運用承認を分ける |
| 文書確認日 | 2026-10-19 |


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

解析記録は`C->ProgramFiles->DRワン/mdcon*.cmd`から読めたRS-232C設定を示します。過去設定と現場受信観測の差は保持し、現在の配線・全機種/全医院の運用承認を証明するものではありません。

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

以下は`Wanrsin.exe`と通信キャプチャについて旧解析資料に記録されたセッション観測です。今回バイナリ/生キャプチャ/実機を再解析しておらず、網羅的なプロトコル忠実度・運用承認とは分けます。

* **接続方式**: RS-232C (COM5, 9600 bps, 8N1)
* **セッション維持**:
  単なる一方通行のデータ受信ではなく、IDEXX VetLab Station 側から定期的なポール信号が送られます。
  * `I` を受信時 ➔ `ACK` + `A` + `IM` を返送
  * `s` を受信時 ➔ `ACK` + `A` + `SM` を返送
  * これに応答しないと PIMS セッションが切断され、データ送信が停止します。
* **データフレーム仕様**:
  * フォーマット: `[STX] '0' [Source:1B] 0x80 [80 | PayloadLen] [Payload...] [Checksum:1B] [ETX]`
  * チェックサム計算: `(sum(payload bytes) & 0x7F) | 0x80`

新カルテにはPIMS応答コード（[`lab_device_idexx_pims.go`](../../../backend/internal/medicalrecord/lab_device_idexx_pims.go)）が存在します。コード参照を完全再現・実機合格の保証にしません。[接続正本](../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)のとおり医院VetLabケーブルでの応答送信/常時運用は未承認です。I/sへの応答は検査依頼送信の採用ではありません。

---

## 6. 新カルテ（AnimalEkarte）のコード参照と検証境界

旧解析記録に関連するコード参照を以下に示します。現在の日常経路・未確認事項は[解析と確定仕様の境界](README.md#analysis-contract-boundaries)へ従い、コード存在を医院切替/撤去済みとしません。

1. **旧環境と新経路の分離**:
   * 新カルテの通常受信はDrワン/MDBを利用しません。日常はNX600/AU10VのブラウザWeb Serial、旧Go agentは任意手動です。旧Win7/VB5/6の実撤去・医院切替は未確認で、[接続正本](../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)の現行Win7記載と状態確認を分けます。
2. **城東の初期観測カタログ**:
   * [`lab_device_item_catalog.go`](../../../backend/internal/medicalrecord/lab_device_item_catalog.go)は36項目を保持。全旧項目・全医院への網羅、医師による具体基準値承認や実機受入を意味しません。
3. **旧番号と新マッピングの区別**:
   * 富士100–420帯とIDEXX910/1310帯は旧内部番号の解析記録です。新カルテはclinic/source/原ラベルで対応し、旧番号をsourceやexam codeへ流用しません。原値・単位・保存時範囲/版・未知コード・訂正の既決保護と旧項目対応の未検証を[境界](README.md#analysis-contract-boundaries)へ残します。

---

## 7. その他の外部連携資産（外注検査・画像診断機器）

旧資産の解析記録には、リアルタイムシリアル以外の実行ファイル・設定の痕跡があります。これだけで媒体・形式・実利用・患者同定・新カルテの対応範囲を確定しません。

### 7.1 外注検査機関連携（大阪血清）
* **対象バイナリ**: `大阪血清050611.exe`
* **アーキテクチャ**: Visual Basic 5.0（`MSVBVM50.DLL`, `VB5JP.DLL`）
* **観測と未確認**: 更新処理（`CmdUpdate`）の痕跡が記録されています。フロッピー/CD/オンライン等の実媒体、検査種類、format、依頼/患者ID、訂正/再送/失敗、現在の実利用・移行初日必須性はこの記録だけでは未確認です。専用連携の採用・対象外・Phaseを決めません（[未確認事項](README.md#migration-and-phase-unknowns)）。

### 7.2 画像診断機器連携（`picset.cmd` と PicJoint 共有）
* **設定ファイル**: `C->ProgramFiles->DRワン/picset.cmd`
* **連携先共有フォルダ**: `\\<院内画像サーバ>\<画像共有>\`（説明用の汎用表記。旧設定に共有先の記載を観測した事実を保持し、実際のホスト名・共有名は記載しない）
* **対象モダリティ（画像種別コード）**:
  * **CR**: レントゲン（X線撮影装置）
  * **US**: 超音波診断装置（エコー）
  * **CT**: コンピュータ断層撮影装置
  * **ES**: 内視鏡（エンドスコープ）
  * **DC**: デジタルカメラ（患部写真）
  * **XC / MS**: その他特殊診断機器
* **新カルテへの含意**:
  共有先と種別コードの設定痕跡です。実ファイルのDICOM/JPEG等の形式、患者/カルテへの同定、PACS・自動取込、実利用範囲は未確認です。[既存添付と専用連携の区別](README.md#analysis-contract-boundaries)と[移行最低互換の未確認事項](README.md#migration-and-phase-unknowns)へ接続し、設定だけで新機能を採用しません。

---

## 8. 医院別状況（城東 vs 八王子）

* **城東医院 (`jouto`)**:
  * 本資料は城東Windows 7端末複製のProgram Files/MDB構造/設定を解析した記録です。今回複製の完全性や生データ・実機を再検証しておらず、医院切替・全機器の運用承認とは別です。
* **八王子医院 (`hachioji`)**:
  * 現在のレポジトリ資産（`old_db/clinics/hachioji/`）には検査機器連携PCの複製がまだ存在しません（`README.md` では `—` 状態）。
  * 八王子の機器構成・接続・外注/画像利用は未確認です。城東の条件を一般化せず、資料回収/ヒアリング等を別調査として残します。この記載は回収・機器操作・患者データ取込の実行許可ではありません。

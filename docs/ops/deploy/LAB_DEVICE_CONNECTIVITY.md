# 検査機器連携 — 実装入口

**この文書:** AnimalEkarte が院内検査機器の結果を受けて保持するときの実装契約。操作手順ではない。
**外部 sibling checkout 前提:** 操作手順と機器調査資料はこの repository には含まれない。checkout を `AnimalEkarte/` と `old_db/` が同じ親を持つ配置にした場合だけ、`../../../../old_db/docs/lab-go/hospital-field-pack/手元テスト手順.md` を参照できる。外部資料が無い場合も、本書の要約（Drワンを使わない、検査用 Mac 1台、3種の source type、raw payload を残さない）を実装契約とする。

**Drワンは組み込まない。** `source_type=drwan` は閉じる。`mkan.mdb` を読まない。

**日常経路（2026-08-19）:** ファイルアップロードしない。検査用 Mac の `/lab-device` が有線シリアルを読む。UI は1画面（待機 + 未紐付け欄）。正本は `todo.md` 城東節と Linear BRT-94〜100。城東3種（`fuji_nx600` / `fuji_au10v` / `arkray_pu4010`）は AE-LAB-0〜4 済み。

外部 sibling 資料（存在する場合のみ）— シリアル枠: `../../../../old_db/docs/lab-go/go-impl/device-serial-adapter.md`
3台マスタ: `../../../../old_db/docs/lab-go/go-impl/device-item-master.md`
城東 Win7 Drワン解析（値なし）: `../../../../old_db/clinics/jouto/research/out/30_drwan_win7/30_report.md`
IDEXX PIMS セッション: `../../../../old_db/docs/lab-go/go-impl/idexx-pims-serial-session.md`

IDEXX は **JOU-LAB-X**。3台のスロットに混ぜない。受信専用では PIMS が切れる。

---

## 切り分け

- **現行カルテ + Drワンは医院の Windows 7。**
- **本リポジトリは新カルテ。Drワン製品も `mkan.mdb` も使わない。**
- 常時つなぐのは検査機器用 Mac 1 台。院内の他 Mac はつながない。

```
現行: 機器 --COM--> Windows 7 --> Drワン --> 現行カルテ
あと: 同じ口 --USB-Serial--> 検査機器用 Mac の /lab-device --> lab-imports
```

Win7 の COM 番号を Mac の `/dev/cu.*` だと思わない。口はつなぐたびに増えた行。

コード入口: `backend/internal/model/lab_import.go`、`backend/internal/medicalrecord/lab_import_service.go`、`backend/internal/medicalrecord/lab_import_handler.go`、`backend/internal/model/lab_device_receive.go`、`backend/internal/medicalrecord/lab_device_receive.go`

接続文字列と生ペイロードはログに出さない。受信ファイルをリポジトリに置かない。

---

## 画面と API

- 日常経路は `/lab-device`（`LabDeviceBoard`）。権限は `lab-import`。確認ダイアログは無い。ペット検索はせず、本日診療中のカルテカードを選ぶ。受信結果は日別に一覧する。
- 医院セットアップで口→機器プロファイルを1回許可する。以後は `/lab-device` を開いたまま自動再オープンする。［読む］は無い。TTL の数値 UI は無い。
- 診察端末の検査画面は未紐付けバナーから1クリックで `attach` する。値は編集しない。
- 保持確認は `/examinations`（城東3種はペット確定後の persist。fixture は commit）
- `fixture` だけ commit 可。`drwan` は preview 200 + `blocked_reasons`、commit 400。`GetJob` は `drwan` を 400

| メソッド | パス | いま |
| --- | --- | --- |
| `POST` | `/api/v1/lab-imports/preview` | 書き込みなし。`drwan` は 200 + blocked |
| `POST` | `/api/v1/lab-imports` | `fixture` のみ。他は 400 |
| `GET` | `/api/v1/lab-imports/:job_id` | ジョブ |
| `GET` | `/api/v1/lab-imports/:job_id/events` | 監査 |
| `POST` | `/api/v1/lab-imports/:job_id/attach` | カルテへ紐付け（`lab-import:edit`） |
| `POST` | `/api/v1/lab-imports/:job_id/detach` | 紐付け解除（`lab-import:edit`） |
| `POST` | `/api/v1/lab-imports/:job_id/revert` | 打ち消し |

ペット紐付けは電文から自動でやらない。

---

## 院内機器（城東・確定）

Win7 `mdcon*.cmd` と 2026-08-18/19 の Mac 受信で confirmed。COM4 は触らない。

| COM | Drワン画面名 | 機器 | 速さ | 新カルテ |
| --- | --- | --- | --- | --- |
| COM6 | ドライケム(新) | 富士 DRI-CHEM NX600 | 9600 8N1 | `fuji_nx600`。`/lab-device` 既定スロット |
| COM7 | ホルモン | 富士 DRI-CHEM IMMUNO AU10V | 9600 8N1 | `fuji_au10v`。既定スロット |
| COM3 | 尿検査 | アークレイ PU-4010 | 現場 2400 8E1（`mdcon2` は 2400 7E1） | decoder-only。7E1/8E1差が未reviewのためagent既定slot・運用support対象外 |
| COM5 | ＩＤＥＸＸ | VetLab Station の PIMS（先に ProCyte / Catalyst） | 9600 8N1 | **別 source_type。未実装。** 本体 Ethernet に生打ちしない |
| COM4 | ドライケム | 富士 DRI-CHEM 7000V | 9600 8N1 | 触らない |

キャプチャの置き場（生値は Git に上げない）:

- 3台: `old_db/docs/lab-go/hospital-field-pack/captures/2026-08-18-jouto/`
- IDEXX: Desktop `lab-capture-20260819/idexx/` と試験票 `exams/2026-08-19-jouto-idexx.md`

---

## 3台アダプタ（実装済み）

既定スロット JSON は `lab_device_receive.go`。ACK / ENQ はしない。`value_raw` は文字列のまま。未知コードは落とさず `needs_review`。ASTM / HL7 と決めない。同じ指紋は `duplicate`。

| `source_type` | 枠 | 項目コード例 |
| --- | --- | --- |
| `fuji_nx600` | STX…ETX、9600 8N1、項目 36 バイト | `Na-P` `BUN-P` `GPT-P` |
| `fuji_au10v` | 同じヘッダ、本体 1 項目、88 バイト | `vf-SAA`（不等式あり） |
| `arkray_pu4010` | 先に `b & 0x7F`。STX 複数 + ETX。2400 8E1 | `GLU` `PRO` `PH` |

詳細オフセットは `device-serial-adapter.md`。`-P` は削らない。

---

## IDEXX（追記・2026-08-19/20）

つなぐ相手は **IDEXX VetLab Station**。ProCyte Dx / Catalyst One は VetLab 配下。公式のカルテ口は ⚙ → 設定 → **顧客情報管理システム**（ネットワーク接続またはシリアル接続）。城東の現行は **シリアル接続**。検査待ち／受付済みリストは公式ではネットワーク接続が必要で、シリアルでは出ない。

### 物理

- 現行: VetLab PIMS → Win7 COM5 → Drワン
- 新カルテ: 同じ PIMS シリアルを検査機器用 Mac が読む（USB-Serial、FTDI）
- ProCyte / Catalyst 本体の Ethernet、検査器同士の LAN には刺さない
- VetLab 本体へ `nc` しない。直接接続は PIMS 側 IP を VetLab に入れる向き
- PIMS は公式上一系統。試すと現行 COM5 が止まる

### 受信（城東 2026-08-19）

- 9600 8N1。`cu.usbserial`
- 枠は STX（`0x02`）… ETX（`0x03`）。CR/LF なし
- 短フレーム（`STX` `10` … `I` または `s` … `ETX`）が繰り返す。測定ではない。セッション維持用の問い合わせ。I は 21 バイトあり、長さだけでは捨てられない
- 長フレーム（約 2 KB）に血球ラベルと単位（`WBC` `RBC` `HCT` `HGB` `PLT` `NEU` `LYM` `MONO` `EOS` `BASO` `RETIC`、単位 `K/uL` `g/dL` `fL` 等）
- 同じ長フレームがストリーム開放まで繰り返す。指紋で 1 測定にする
- 規格名はまだ付けない
- Mac 受信専用および単独 ACK では PIMS オフラインのまま。Drワンは I→ACK+A+IM、s→ACK+A+SM。Source/Port 候補は城東 `mdcon4.cmd` の `2`/`2`（`CByte` → `0x02`）。組み立ては `lab_device_idexx_pims.go`（シリアルには繋がない。本番 VetLab へは送らない）

### Drワン内部との関係（読まない）

Win7 の `Drimke.tbl` は機器ラベル → 内部コード（`IRBC`→910、`IBUN`→1310）。`ike` / `dkensa` の 910–1270 が血球、1310–1630 が生化学。
これは **Drワンが書いた番号**。AnimalEkarte は `mkan.mdb` を読まず、**シリアルに出たラベル**（`WBC` 等）をマスタキーにする。内部番号を `source_type` や `exam_code` にしない。

`mkan.mdb` には患者氏名・住所がある。解析以外で開かない。Git に置かない。

### 実装してよいこと / いけないこと

| する | しない |
| --- | --- |
| 別 `source_type`（`idexx_vetlab`） | 既定3スロットに IDEXX フレームを足す |
| 短 I/s は測定にしない。長フレームは 1 指紋 | 復元途中の IM/SM を本番 VetLab へ送る |
| I に ACK+A+IM、s に ACK+A+SM（`CollectIDEXXPIMSReplies`、host `0x02`）。agent は `--pims-reply` のときだけ同じ usbserial に書く | 既定 agent で常時接続したことにする。`--pims-reply` を医院 VetLab ケーブルで使う |
| ラベル＋`value_raw`＋単位 | 910 などの内部コードを persist。本体へ ASTM / `nc` |
| 保存 raw または Drワン確立直後の再生でデコード | 患者検体を常時接続試験に使う |

---

## 手入力

Drワン `mdconM` に総蛋白・黄疸指数・フィラリア・FIV/FeLV・PCV・CRP・UPC・鏡検などがある。電文に無い行は検査画面の手入力（または `manual`）。機器アダプタで埋めない。

---

## やってはいけないこと

- `source_type=drwan` を commit 可能にする
- `mkan.mdb` / `Link/*.txt` / `djusin.*` を lab-imports の入力にする
- COM4 / 7000V
- IDEXX 検査器本体への直結
- 生ペイロード・飼主名・接続文字列をログ / Git に出す
- 電文が無い規格名（ASTM / HL7）をパーサ名にする
- IDEXX を `fuji_nx600` 等の既定スロットに混ぜる

---

## 改訂

| 日付 | 内容 |
| --- | --- |
| 2026-08-17 | 初版。医院疎通と実装入口 |
| 2026-08-18 | 3台 Mac 受信。アダプタ仕様は old_db |
| 2026-08-19 | `/lab-device` 直読。AE-LAB-0〜4 |
| 2026-08-20 | 城東 `mdcon` 確定。IDEXX PIMS シリアルと Drワン内部コード帯を追記。MDB は入力にしない |

*患者データ・認証情報・機器識別子の生値は記入しない。*

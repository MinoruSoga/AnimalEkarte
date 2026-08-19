# 検査機器連携 — 医院疎通と実装入口

**2026-08-19 方針:** ファイルアップロードしない。検査用 Mac の新カルテ待機ページが有線シリアルを読む。UI は1画面（待機 + 未紐付け欄）。正本は `todo.md` の城東節と Linear BRT-94〜100、仕様は `old_db/docs/lab-go/go-impl/device-serial-adapter.md`。

**現場手順（08-18 疎通当日）:**  
`/Users/minoru/Dev/Case/AnimalHospital/old_db/docs/lab-go/hospital-field-pack/00-明日の現場手順.md`

08-18 の医院作業は **機器 ↔ 持っていく Mac の取得・閲覧** だった。実装の正本は上記 08-19 方針。

---

## 切り分け

- **現行カルテ + Drワンは医院の Windows 7。**
- **本リポジトリは新カルテ。Drワンを使わない。**
- 院内の全 Mac を機器につなぐのではない。あとで常時つなぐのは検査機器用 Mac 1 台。

```
現行: 機器 --COM--> Windows 7 --> Drワン --> 現行カルテ
今日: 機器 --空き USB/LAN または移した線--> 持っていく Mac（ファイル閲覧）
あと: 同じ口を検査機器用 Mac の `/lab-device` が読み、lab-imports が保持する
```

コードの `source_type=drwan` は仮名。Drワン製品を組み込む計画ではない。

院内機器（詳細は old_db `hospital-field-pack/07-機器調査.md`）:

| COM | 機器（推定・当日照合） | 今日 | 新カルテ（あとで） |
| --- | --- | --- | --- |
| COM6 | 富士 DRI-CHEM NX600 | **Mac 受信の最優先** | 電文を見てから直読 |
| COM7 | 富士 DRI-CHEM IMMUNO AU10V | 次点 | 電文を見てから直読 |
| COM3 | 尿（ペーパー読取） | 本体特定 | 直読 + 手入力 |
| COM5 | IDEXX ProCyte Dx + Catalyst One | やらない | VetLab PIMS。生 ASTM は打たない |
| COM4 | 富士 DRI-CHEM 7000V | 触らない | 触らない |

---

## 画面と API

- 日常経路は `/lab-device`（`LabDeviceBoard`）。権限は `lab-import`。確認ダイアログは無い。ペット検索はせず、本日診療中のカルテカードを選ぶ。受信結果は日別に一覧する。
- 医院セットアップで口→機器プロファイルを1回許可する。以後は `/lab-device` を開いたまま自動再オープンする。［読む］は無い。TTL の数値 UI は無い。
- 診察端末の検査画面は未紐付けバナーから1クリックで `attach` する。値は編集しない。
- 保持確認は `/examinations`（城東3種はペット確定後の persist。fixture は commit）
- `fixture` だけ commit 可。`drwan` は preview 200 + `blocked_reasons`、commit 400。`GetJob` は `drwan` を 400
- preview は今日の医院合格条件ではない

入口: `backend/internal/model/lab_import.go`、`lab_import_service.go`、`lab_import_handler.go`

本番デフォルトは閉じる。接続文字列と生ペイロードはログに出さない。受信ファイルをリポジトリに置かない。

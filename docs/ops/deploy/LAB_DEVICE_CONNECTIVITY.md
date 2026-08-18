# 検査機器連携 — 医院疎通と実装入口

**現場手順:**  
`/Users/minoru/Dev/Case/AnimalHospital/old_db/docs/lab-go/hospital-field-pack/00-明日の現場手順.md`

今日の医院作業は **機器 ↔ 持っていく Mac の取得・閲覧**。  
このリポジトリのアダプタ本直しは、受信ファイルが残ってから。

---

## 切り分け

- **現行カルテ + Drワンは医院の Windows 7。**
- **本リポジトリは新カルテ。Drワンを使わない。**
- 院内の全 Mac を機器につなぐのではない。あとで常時つなぐのは検査機器用 Mac 1 台。

```
現行: 機器 --COM--> Windows 7 --> Drワン --> 現行カルテ
今日: 機器 --空き USB/LAN または移した線--> 持っていく Mac（ファイル閲覧）
あと: 同じ口を検査機器用 Mac が読み、lab-imports が保持する
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

- 検査取込 UI は無い。保持確認は commit 後の `/examinations`
- `fixture` だけ commit 可。`drwan` は preview 200 + `blocked_reasons`、commit 400
- preview は今日の医院合格条件ではない

入口: `backend/internal/model/lab_import.go`、`lab_import_service.go`、`lab_import_handler.go`

本番デフォルトは閉じる。接続文字列と生ペイロードはログに出さない。受信ファイルをリポジトリに置かない。

# 検査機器連携 — 医院疎通と実装入口

**現場手順:**  
`/Users/minoru/Dev/Case/AnimalHospital/old_db/docs/lab-go/hospital-field-pack/00-明日の現場手順.md`  
**当日の Mac 操作（前夜準備含む）:** 同パック `08-画面操作.md` B0–B2。当日は preview API のみ。

---

## 切り分け

- **現行カルテ + Drワンは医院の Windows 7。** 観察だけ。Mac では動かない。  
- **本リポジトリは MacBook 上の新カルテ。Drワンを使わない。** 実機 COM は Win7 側にある。

```
現行: 機器 --COM--> Windows 7 --> Drワン --> 現行カルテ
新  : MacBook 上の lab-imports API（Drワンなし。実機は線を移すまで届かない）
```

コードの `source_type=drwan` は仮名。Drワン製品を組み込む計画ではない。

院内機器（Drive 写真 + 現行 Drワン画面。詳細は old_db `hospital-field-pack/07-機器調査.md`）:

| COM | 機器（推定・当日照合） | 新カルテ |
| --- | --- | --- |
| COM6 | 富士 DRI-CHEM NX600 | 将来候補: シリアル直読（当日はやらない） |
| COM7 | 富士 DRI-CHEM IMMUNO AU10V | 将来候補: シリアル直読（当日はやらない） |
| COM5 | IDEXX ProCyte Dx + Catalyst One | VetLab Station 経由。機器へ生 ASTM は打たない |
| COM3 | 尿（ペーパー読取） | 実機特定 + 手入力混在（当日は現物確認） |
| COM4 | 富士 DRI-CHEM 7000V | 現行未使用。触らない |

---

## 画面と API

- 検査取込 UI は無い。確認は `/examinations`
- `fixture` だけ commit 可。`drwan` は preview 200 + `blocked_reasons`、commit 400

入口: `backend/internal/model/lab_import.go`、`lab_import_service.go`、`lab_import_handler.go`

医院テストで機器を読む場合も本番デフォルトは閉じる。接続文字列と生ペイロードはログに出さない。

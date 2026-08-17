# 検査機器連携 — 医院疎通と実装入口

**現場手順:**  
`/Users/minoru/Dev/Case/AnimalHospital/old_db/docs/lab-go/hospital-field-pack/00-明日の現場手順.md`

---

## 切り分け

- **現行カルテ**は検査機器連携に **Drワン** を使う。明日は規格特定の参照として見る。  
- **本リポジトリ（新カルテ）は Drワンを使わない。** 機器と有線で直接つなぐ。

```
現行（参照）: 機器 --有線--> PC --> Drワン --> 現行カルテ
新カルテ    : 機器 --有線--> PC --> lab-imports API
```

コードの `source_type=drwan` は仮名。Drワン製品を組み込む計画ではない。

院内機器（Drive 写真 + 現行 Drワン画面。詳細は old_db `hospital-field-pack/07-機器調査.md`）:

| COM | 機器 | 新カルテ |
| --- | --- | --- |
| COM6 | 富士 DRI-CHEM NX600 | 優先。シリアル直接 |
| COM7 | 富士 DRI-CHEM IMMUNO AU10V | 優先。シリアル直接 |
| COM5 | IDEXX ProCyte Dx + Catalyst One | VetLab Station 経由。機器へ生 ASTM は打たない |
| COM3 | 尿（ペーパー読取） | 実機特定 + 手入力混在 |
| COM4 | 富士 DRI-CHEM 7000V | 現行未使用。触らない |

---

## 画面と API

- 検査取込 UI は無い。確認は `/examinations`
- `fixture` だけ commit 可。`drwan` は preview 200 + `blocked_reasons`、commit 400

入口: `backend/internal/model/lab_import.go`、`lab_import_service.go`、`lab_import_handler.go`

医院テストで機器を読む場合も本番デフォルトは閉じる。接続文字列と生ペイロードはログに出さない。

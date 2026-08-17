# 検査機器連携 — 医院疎通と実装入口

**現場手順の正本（記入欄つき）:**  
`/Users/minoru/Dev/Case/AnimalHospital/old_db/docs/lab-go/hospital-field-pack/00-明日の現場手順.md`

このファイルは **AnimalEkarte 側** の入口。実装・API・画面の有無だけをここに置く。

---

## 明日の対象

```
検査機器 --有線--> 今の PC --> Drワン（現行ツール）--> 本リポジトリの lab-imports API
```

旧 NOAH の取込画面は使わない。Drワンは現行スタックのツールで、`source_type=drwan`。

---

## 画面

- 検査取込 UI は **無い**（権限リソース `lab-import` のラベルだけ）
- 結果確認は `/examinations`（`examinations` view）

---

## いまの API 境界

| source_type | preview | commit |
| --- | --- | --- |
| `fixture` | 可 | 可 |
| `drwan` | 200 + `blocked_reasons` | **400** |
| `manual` | blocked | 400 |

入口コード:

- `backend/internal/model/lab_import.go`
- `backend/internal/medicalrecord/lab_import_service.go`
- `backend/internal/medicalrecord/lab_import_handler.go`

医院テストで `drwan` を開ける場合も、本番デフォルトは閉じる。接続文字列と生ペイロードはログに出さない。

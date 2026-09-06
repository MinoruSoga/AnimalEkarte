# マスタ設定 仕様書 インデックス (Master Settings Index)

本ディレクトリには、Animal Ekarte の核となる各マスタデータの詳細定義、および管理画面での操作仕様が格納されています。

---

## 🏗️ 基礎・共通マスタ

| マスタ名 | 仕様書 | 概要 |
|:---|:---|:---|
| **医院情報** | [../19-clinic-settings.md](../19-clinic-settings.md) | 院名、住所、登録番号、消費税率（通常/軽減）。 |
| **動物種別** | [master-animal-species.md](./master-animal-species.md) | 犬、猫、エキゾチック等の大分類。 |
| **職種** | [master-occupation.md](./master-occupation.md) | スタッフの役割（獣医、看護師等）の定義。 |
| **スタッフ** | [master-staff.md](./master-staff.md) | 氏名、職種、所属院、LINE 予約用プロフィール。 |
| **権限グループ** | [master-permission-group.md](./master-permission-group.md) | ロールごとの詳細なアクセス権限（CRUD）マトリックス。 |

---

## 🩺 医療・臨床マスタ

| マスタ名 | 仕様書 | 概要 |
|:---|:---|:---|
| **診療項目** | [master-treatment.md](./master-treatment.md) | 診察、処置、手術等の名称と標準価格。 |
| **主訴種別** | [master-chief-complaint.md](./master-chief-complaint.md) | 診察理由の大分類（消化器、皮膚等）。 |
| **薬剤** | [master-medicine.md](./master-medicine.md) | 薬品名、剤形、単価の定義（在庫紐付けは API のみで設定画面に UI 未実装）。 |
| **診断・病名** | [master-diagnosis.md](./master-diagnosis.md) | 疾患カテゴリと正式病名の体系的定義。 |
| **検査項目** | [master-examinations.md](./master-examinations.md) | 検査プランの名称・価格の定義（診療項目マスタと同一UIを共有。課税区分は保存されない）。 |
| **問診/定型文**| [master-interview.md](./master-interview.md) | カルテ入力を効率化する各種テンプレート。 |
| **ケージ** | [master-cage.md](./master-cage.md) | 入院室の番号、サイズ、収容タイプ。 |
| **入院プラン** | [master-hospitalization-plan.md](./master-hospitalization-plan.md) | 入院・宿泊の単価・対象体格・料金単位の定義。 |

---

## ✂️ 専門サービスマスタ

| マスタ名 | 仕様書 | 概要 |
|:---|:---|:---|
| **トリミングコース種別** | [master-trimming-course-type.md](./master-trimming-course-type.md) | シャンプー・カット等、コースカテゴリの分類を管理。 |
| **トリミング** | [master-trimming.md](./master-trimming.md) | コース・オプションの定義（価格は対象サイズ区分のみ、犬種別ではない）。 |
| **予約区分** | [master-reservation-type.md](./master-reservation-type.md) | 診察、手術等の予約枠と LINE 公開設定。 |
| **シフトパターン**| [master-shift-template.md](./master-shift-template.md) | よく使う勤務時間と休憩時間のテンプレート。 |

---

## 💰 会計・外部連携マスタ

| マスタ名 | 仕様書 | 概要 |
|:---|:---|:---|
| **締め時間** | [closing-time-settings.md](./closing-time-settings.md) | AM/PM 境界、日界、休診日の管理。 |
| **支払方法** | [payment-methods.md](./payment-methods.md) | 現金、カード、QR 等の決済手段。 |
| **保険** | [master-insurance.md](./master-insurance.md) | ペット保険名称・補償率の管理（マスタ補償率の会計への自動設定は未実装。会計画面で選択した割合による控除計算は実装済み）。 |
| **割引キャンペーン** | [master-campaigns.md](./master-campaigns.md) | 会計割引ルールの期間・対象カテゴリ/商品設定。 |
| **販売商品** | [master-merchandise.md](./master-merchandise.md) | 療法食、ケア用品等の販売品。 |
| **LINE ページ** | [28-line-reservation.md §2](../28-line-reservation.md) | 飼い主向け予約画面の案内文言カスタマイズ（LINE 予約設定仕様書へ統合）。 |

---

## 技術的な共通事項

- **Notion スタイル編集**: 各マスタは `MasterSidePanel`（`SidePeekPanel` ベース）と `PropertyRow`/`PropertyInput` を使用した、リストを離れない直感的な編集体験を提供します。
- **認可ガード**: 個別のマスタに対し、`ResourceMasterMedical` や `ResourceMasterStaff` 等の独立した権限チェックがバックエンドで実行されます。
- **名称一意のエラー表示**: `(clinic_id, name)` UNIQUE のマスタで重複保存したとき、トーストは種別ラベルだけでなく入力した実名を含む（診療項目 5 タブの正本は [master-treatment.md §2.3](./master-treatment.md)）。

---

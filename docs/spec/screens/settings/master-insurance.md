# 保険マスタ 仕様書 (Insurance Management)

## 概要
- **画面の目的**: ペット保険（アニコム、アイペット等）の定義、および保険が補償する側の割合（「補償率」）の設定管理。
- **URLパターン**: `/settings/insurance`
- **アクセス権限**: 保険マスタ管理権限が必要（`ResourceMasterInsurance`）

---

## 画面構成

### 1. 保険会社一覧
- **項目**: 名称（例：アニコム 50%）、補償率、連絡先、有効/無効（ステータス）。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **保険名称**: ペット編集画面で保険選択時に表示される名称。
- **補償率(%)**: `coverage_rate`（0〜100 の整数。`validateInsuranceForm`）。UIラベルは「補償率」であり、保険が補償する側の割合を表す（飼主の窓口負担割合ではない）。入力に HTML `max` は付けない。101 以上はインラインエラーを出し、HTML5 制約で保存ボタンを無音ブロックしない。
- **連絡先**: 保険会社の電話番号等（`contact_phone`）。
- **備考**: 自由記述。

---

## 主要な機能

### 1. 会計との連動（未実装）
`coverage_rate` は保険マスタ・飼主/ペット編集画面（`frontend/src/features/owners/`）でのみ参照され、会計・請求計算のバックエンドサービス（`backend/internal/service/accounting*.go`, `billing*.go`）からは参照されていない。会計詳細画面で保険を選択すると自動的に「保険負担分」を差し引く機能は現時点で実装されていない。

---

## 技術仕様

### 使用コンポーネント
- **`InsuranceSettings`**: メインページ。
- **`InsuranceSidePanel`**: 補償率・連絡先の編集、`StatusToggleButton` による有効/無効切り替え。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/insurances` | 取扱い保険の一覧取得 | `master-insurance` | `view` |
| GET | `/api/v1/masters/insurances/:id` | 特定の保険会社情報の取得 | `master-insurance` | `view` |
| POST | `/api/v1/masters/insurances` | 新規保険会社の登録 | `master-insurance` | `create` |
| PATCH | `/api/v1/masters/insurances/:id` | 設定内容の更新 | `master-insurance` | `edit` |
| DELETE | `/api/v1/masters/insurances/:id` | 保険会社の削除 | `master-insurance` | `delete` |
| PATCH | `/api/v1/masters/insurances/reorder` | 表示順序の一括保存（BE実装済みだが本画面からは未呼出） | `master-insurance` | `edit` |

---


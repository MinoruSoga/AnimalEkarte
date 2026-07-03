# 保険マスタ 仕様書 (Insurance Management)

## 概要
- **画面の目的**: ペット保険（アニコム、アイペット等）の定義、および窓口精算時の「負担割合」設定の管理。
- **URLパターン**: `/settings/insurance`
- **アクセス権限**: 会計マスタ管理権限が必要（`ResourceMasterInsurance`）

---

## 画面構成

### 1. 保険会社一覧
- **項目**: 名称（例：アニコム 50%）、負担割合、有効/無効。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **保険名称**: 飼主選択時に表示される名称。
- **負担割合**: 
    - 飼主が窓口で支払うべき割合を設定。
    - 残りの金額は「保険請求分」として会計上で自動算出されます。
- **備考**: 加入条件や入力上の注意点。

---

## 主要な機能

### 1. 会計精算との自動連動
会計詳細画面において特定の保険が選択されると、請求総額から「保険負担分」を瞬時に差し引き、飼主への最終的な請求額を決定します。

### 2. インボイス・明細表示
領収書および明細書において、保険適用後の金額と適用前の総額が、制度に則った形式で正しく印字されます。

---

## 技術仕様

### 使用コンポーネント
- **`PropInput`**: 負担割合（%）の数値入力。
- **`NotionStatusPill`**: 保険会社の取扱い停止・再開の切り替え。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/insurances` | 取扱い保険の一覧取得 | `master-insurance` | `view` |
| GET | `/api/v1/masters/insurances/:id` | 特定の保険会社情報の取得 | `master-insurance` | `view` |
| POST | `/api/v1/masters/insurances` | 新規保険会社の登録 | `master-insurance` | `create` |
| PATCH | `/api/v1/masters/insurances/:id` | 設定内容の更新 | `master-insurance` | `edit` |
| DELETE | `/api/v1/masters/insurances/:id` | 保険会社の削除 | `master-insurance` | `delete` |
| PATCH | `/api/v1/masters/insurances/reorder` | 表示順序の一括保存 | `master-insurance` | `edit` |

---


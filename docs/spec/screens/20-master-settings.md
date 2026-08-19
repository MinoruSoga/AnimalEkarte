# マスタ設定トップ 仕様書 (Master Settings Index)

## 概要
- **画面の目的**: 診療、トリミング、入院、スタッフ、権限、および外部連携など、システム全体で共通利用される定義データ（マスタ）を一元的に管理するポータル。
- **URLパターン**: `/settings` (トップ)
- **アクセス権限**: `/settings` トップに `RequirePermission` 無し（認証 Layout）。各カードは resource で表示フィルタ、遷移先は個別ガード

---

## 画面構成

### セクション別マスタ一覧
管理対象を 8 つの論理的なセクションに分類し、直感的なアクセスを実現しています。

| セクション名 | 管理対象（リンク） |
|:---|:---|
| **基本設定** | [医院情報](./19-clinic-settings.md)、[動物種類マスタ](./settings/master-animal-species.md) |
| **カルテ** | [診療項目](./settings/master-treatment.md)、[検査マスタ](./settings/master-examinations.md)（診療項目の検査タブ。サイドバー「マスタ設定」→「カルテ関連」→「検査マスタ」。`/settings/treatment-items?tab=examination`）、[診断マスタ](./settings/master-diagnosis.md)、[問診テンプレート](./settings/master-interview.md)、[主訴種別](./settings/master-chief-complaint.md)、[薬剤マスタ](./settings/master-medicine.md)、検査機器マスタ（サイドバー「マスタ設定」→「カルテ関連」→「検査機器マスタ」。`/settings/lab-device-item-masters`。機器名と検査種別は `lab_devices`。登録は `lab-import:create`、編集は `lab-import:edit`。日常の送信画面には出さない） |
| **予約管理マスタ** | [予約区分マスタ](./settings/master-reservation-type.md) |
| **入院・ケージ管理** | [入院プラン](./settings/master-hospitalization-plan.md)、[ケージマスタ](./settings/master-cage.md) |
| **トリミング関連** | [トリミングマスタ](./settings/master-trimming.md)、[トリミングコース種別](./settings/master-trimming-course-type.md) |
| **会計・商品** | [商品マスタ](./settings/master-merchandise.md)、[保険マスタ](./settings/master-insurance.md)、[支払方法](./settings/payment-methods.md)、[割引キャンペーン](./settings/master-campaigns.md)、[締め時間設定](./settings/closing-time-settings.md) |
| **スタッフ・権限** | [スタッフ管理](./settings/master-staff.md)、[職種マスタ](./settings/master-occupation.md)、[権限グループ設定](./settings/master-permission-group.md) |
| **シフト管理** | [シフトパターン](./settings/master-shift-template.md) |

> 補足: [LINE予約設定](./28-line-reservation.md)（`/line-reservation`）と [Lステップ連携設定](./31-lstep-integration.md) はサイドバーの別セクションから遷移する。

---

## 共通の操作性

### 1. カード型リスト (`CardRow`)
各マスタの項目は Notion ライクなカード形式で表示され、以下の情報を提供します。
- **アイコン**: 視覚的な識別のための lucide アイコン。
- **名称と説明**: そのマスタが何に使用されるかの要約。
- **遷移シェブロン**: 右端の `ChevronRight` アイコン。

### 2. サイドパネル編集 (`SidePeekPanel`)
大部分のマスタ画面では、一覧を離れずに詳細を編集できるサイドパネル方式を採用しています。これにより、作業のコンテキストを維持したまま複数の項目を連続して編集することが可能です。

---

## 技術仕様

### 権限制御 (RBAC)
ポータル画面自体は全ユーザーが閲覧可能ですが、個別のマスタへのアクセスおよび操作権限は、`ResourceMasterMedical`, `ResourceMasterStaff`, `ResourceMasterTrimming` 等のリソースキーに基づき、バックエンドのハンドラー層で厳格に認可チェック（`RequirePermission`）が適用されます。

### 使用コンポーネント
- **`MasterSettingsIndex`**: メインコンテナ。
- **`PermissionFilteredSection`**: セクション別のグループ表示部品。セクション内の全マスタに閲覧権限がない場合はセクションごと非表示になる。
- **`PermissionFilteredCard`** / **`CardRow`**: 個別マスタへのリンク・サマリ表示部品。

---

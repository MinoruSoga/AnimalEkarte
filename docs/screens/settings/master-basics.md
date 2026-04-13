# 共通基礎マスタ設定 仕様書

## 概要
本ドキュメントでは、項目数が少なく、共通のCRUDパターン（一覧 + サイドパネル）で管理される基礎的なマスタ設定を定義します。

---

## 1. 動物種マスタ (`AnimalSpeciesSettings`)
- **URL**: `/settings/animal-species`
- **目的**: 犬、猫、うさぎ等の種類を定義。
- **項目**: 名称、ステータス、表示順。

## 2. 職能マスタ (`OccupationSettings`)
- **URL**: `/settings/occupations`
- **目的**: 獣医師、愛玩動物看護師、トリマー等の職種を定義。
- **項目**: 名称、ステータス、説明、表示順。

## 3. 予約区分マスタ (`ReservationTypeSettings`)
- **URL**: `/settings/reservation-type`
- **目的**: 診療、手術、ワクチン、検診等の予約カテゴリを定義。
- **項目**: 
  - 名称、ステータス、表示順
  - **カラー**: カレンダー表示等で使用するカラーコード。
  - **グループ**: `ReservationTypeGroup` への紐付け。
  - **LINE表示設定**: LINE予約ページでの表示名、説明文、画像URL等。

## 4. 入院プランマスタ (`HospitalizationSettings`)
- **URL**: `/settings/hospitalization`
- **目的**: 入院料（小型/中型/大型別など）の基本単価を定義。
- **項目**: 名称、ステータス、単価(税込)、備考。

## 共通UI・操作
- **一覧**: `DataTable` による表示、キーワード検索、ステータスフィルタ。
- **編集**: `MasterSidePanel` によるスライドイン編集。
- **D&D**: 全ての基礎マスタにおいて、並び替え（D&D）をサポート。

## API連携
| カテゴリ | エンドポイント |
|:---|:---|
| 動物種 | `/api/v1/masters/animal-species` |
| 職能 | `/api/v1/masters/occupations` |
| 予約区分 | `/api/v1/masters/reservation-types` |
| 入院プラン | `/api/v1/masters/hospitalization-plans` |

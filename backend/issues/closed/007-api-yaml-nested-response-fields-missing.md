---
status: open
---

# [api.yaml] Pet/Reservation/MedicalRecord スキーマにネストレスポンスフィールドが未定義

## 背景

001〜002チケットでバックエンドに Preload を追加し、ネストオブジェクトをレスポンスに
含める実装を行う。それに伴い api.yaml のスキーマも更新が必要。
また MedicalRecord は既に FindAll で Owner/Pet/Doctor/Inquiry を Preload しているが
スキーマ上に定義されていない。

## 問題

### Pet schema（api.yaml L133）

```yaml
# 現状: animal_species と insurance は定義済み
# ❌ 欠落: owner ネスト
Pet:
  properties:
    owner_id:
      type: integer   # IDのみ
    animal_species:
      $ref: '#/components/schemas/AnimalSpecies'  # ✅ あり
    insurance:
      $ref: '#/components/schemas/Insurance'       # ✅ あり
    # owner: ← ない（001チケット実装後に追加が必要）
```

### ReservationAppointment schema（api.yaml L218）

```yaml
# 現状: ID のみで pet/service_type/doctor ネストがない
ReservationAppointment:
  properties:
    pet_id:
      type: integer
    service_type_id:
      type: integer
    doctor_id:
      type: integer
    # pet:          ← ない（002チケット実装後に追加が必要）
    # service_type: ← ない
    # doctor:       ← ない
```

### MedicalRecord schema（api.yaml L270付近）

```yaml
# 現状: ID のみで owner/pet/doctor/inquiry ネストがない
MedicalRecord:
  properties:
    owner_id:
      type: integer
    pet_id:
      type: integer
    doctor_id:
      type: integer
    # owner:   ← ない（FindAll で既に Preload 済みだが yaml に未定義）
    # pet:     ← ない
    # doctor:  ← ない
    # inquiry: ← ない
```

## 修正方針

各スキーマにネストフィールドを追加する（001/002チケットの実装完了後に対応）:

### Pet schema に owner を追加

```yaml
owner:
  type: object
  description: 飼主情報（Preload時のみ含まれる）
  properties:
    id:
      type: integer
      format: int64
    owner_name:
      type: string
    phone:
      type: string
```

### ReservationAppointment schema にネストを追加

```yaml
pet:
  $ref: '#/components/schemas/Pet'
  description: ペット情報（Preload時のみ含まれる）
service_type:
  $ref: '#/components/schemas/ServiceType'
  description: サービス種別（Preload時のみ含まれる）
doctor:
  $ref: '#/components/schemas/Staff'
  description: 担当医（Preload時のみ含まれる）
```

### MedicalRecord schema にネストを追加

```yaml
owner:
  $ref: '#/components/schemas/Owner'
  description: 飼主情報（Preload時のみ含まれる）
pet:
  $ref: '#/components/schemas/Pet'
  description: ペット情報（Preload時のみ含まれる）
doctor:
  $ref: '#/components/schemas/Staff'
  description: 担当医（Preload時のみ含まれる）
inquiry:
  type: object
  description: 問診情報（Preload時のみ含まれる）
```

## 依存チケット

- **001**: Pet owner ネストは 001 実装後に追加
- **002**: Reservation ネストは 002 実装後に追加

## 完了条件

- [ ] `Pet` schema に `owner` ネストオブジェクト定義を追加
- [ ] `ReservationAppointment` schema に `pet`, `service_type`, `doctor` ネスト定義を追加
- [ ] `MedicalRecord` schema に `owner`, `pet`, `doctor`, `inquiry` ネスト定義を追加
- [ ] Swagger UI でネストフィールドが正しく表示される

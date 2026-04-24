# TASK-084: chief_complaint — レイヤー間ファイル命名不統一

## 優先度

LOW

---

## 概要

`chief_complaint` ドメインのファイル命名が Handler 層と Service/Repository 層で不統一。
他の全ドメインは全レイヤーで同一プレフィックスを使用しているが、このドメインのみ例外。

---

## 問題

| レイヤー | ファイル名 | プレフィックス |
|---------|----------|--------------|
| Handler | `chief_complaint_handler.go` | `chief_complaint_` |
| Handler | `chief_complaint_request.go` | `chief_complaint_` |
| Handler | `chief_complaint_response.go` | `chief_complaint_` |
| Service | `chief_complaint_type_service.go` | `chief_complaint_type_` ← **不統一** |
| Repository | `chief_complaint_type_repository.go` | `chief_complaint_type_` ← **不統一** |

### 他ドメインとの比較

```
# ✅ 統一されているドメイン例
handler/cage_handler.go
service/cage_service.go
repository/cage_repository.go

handler/exam_type_handler.go
service/exam_type_service.go
repository/exam_type_repository.go

# ❌ 不統一（chief_complaint）
handler/chief_complaint_handler.go
service/chief_complaint_type_service.go    ← _type_ が余分
repository/chief_complaint_type_repository.go  ← _type_ が余分
```

---

## 修正方針

ファイル名のみリネーム。内部の struct 名・interface 名・メソッド名は変更不要。

```bash
# Service
git mv backend/internal/service/chief_complaint_type_service.go \
       backend/internal/service/chief_complaint_service.go

# Repository
git mv backend/internal/repository/chief_complaint_type_repository.go \
       backend/internal/repository/chief_complaint_repository.go
```

---

## 修正ファイル

| 変更前 | 変更後 |
|--------|--------|
| `service/chief_complaint_type_service.go` | `service/chief_complaint_service.go` |
| `repository/chief_complaint_type_repository.go` | `repository/chief_complaint_repository.go` |

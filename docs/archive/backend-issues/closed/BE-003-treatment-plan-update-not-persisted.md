# BE-003: 治療プラン更新API - DB未保存（PATCH エンドポイント不完全/missing）

## 🔴 根本原因推定（2026-03-16 調査中）

### 問題パターン（BE-003, BE-004, BE-005 共通）
フロントエンドから PATCH リクエストを送信 → バックエンド 200/201 で応答 → **しかし DB 更新/作成がない**

| Issue | エンドポイント（期待） | リソース | 症状 |
|-------|---------------------|----------|------|
| BE-003 | `PATCH /api/v1/medical-records/:id/treatment-plans/:planId` | treatment_plans | 更新されない |
| BE-004 | `POST /api/v1/medical-records/:id/estimates` | estimates | 作成されない |
| BE-005 | `PATCH /api/v1/medical-records/:id/inquiries` | inquiries | 更新されない |

### バックエンド実装状況（確認済み）
✅ **treatment_plan_handler.go**
- `UpdateTreatmentPlan()` 実装済み（109-137 行）
- route 登録済み：`rg.PATCH("/:id/treatment-plans/:planId", h.UpdateTreatmentPlan)` （155 行）
- service call: `h.svc.TreatmentPlan.Update(ctx, planID, input)` （131 行）

❓ **estimate_handler.go / inquiry_handler.go**
- 要確認：PATCH/POST エンドポイントが正しく registered されているか
- 要確認：service layer で実際に DB に書き込まれているか

### 推정된 원인들（우선순위순）

#### 1️⃣ 프론트엔드 요청 구조 불일치 (높음)
```javascript
// 프론트: 잘못된 경로
PATCH /api/v1/medical-records/17/treatment-plans
// → plan ID가 없어서 route match 안 함

// 백엔드 기대값
PATCH /api/v1/medical-records/:id/treatment-plans/:planId
// → plan ID가 필수
```

#### 2️⃣ 서비스 층 에러 처리 (중간)
service 메서드가 실제로 DB 에 저장하지 않거나, DB save가 실패해도 에러를 반환하지 않는 경우

#### 3️⃣ 트랜잭션 롤백 (중간)
GORM 트랜잭션 내에서 에러 발생 → 자동 롤백 → 그러나 핸들러는 200으로 응답

#### 4️⃣ 주의해야 할 점 (낮음)
- 의료 기록 ID 검증이 누락된 경우
- 권한 확인이 없는 경우
- 동시성 문제 (race condition)

## 재현 절차
1. 의료 기록 편집 화면 열기 (ID: 17)
2. "진료/치료 계획" 탭 클릭
3. 치료 방침 필드 편집
4. 저장 버튼 클릭
5. "진료기록 업데이트됨" 알림 표시
6. DB 확인 → **데이터 없음**

## 테스트 환경
- 기록ID: 17
- 펫ID: 15
- 테스트 시간: 2026-03-16

## DB 확인 결과
```sql
SELECT COUNT(*) FROM treatment_plans WHERE medical_record_id = 17;
-- 결과: 0행 (존재하지 않음)
```

## 다음 단계（우선순위）

### A. 백엔드 조사
1. **estimate_handler.go**
   - route 확인: `POST /:id/estimates` 가 정말로 handler를 호출하는가?
   - service call 확인: estimate.Create() 가 실제로 저장하는가?
   - error handling 강화: db.Error 를 slog로 출력

2. **inquiry_handler.go** (또는 inquiry_template_handler.go)
   - route 확인: `PATCH /:id/inquiries` 가 존재하는가?
   - Update() 에서 실제로 DB Update를 수행하는가?

3. **모든 PATCH/POST 핸들러**
   - service 에서 `s.repo.Update/Create()` 호출 후 error 를 확인하고 반환하는가?
   - slog 출력으로 실행 흐름을 추적하는가?

### B. 프론트엔드 조사
1. API 요청 생성 로직 확인
   - 정확한 URL path 확인
   - request body 구조 확인
   - response status 및 body 확인

2. React Query / 네트워크 탭 검사
   - 실제로 보내진 요청을 확인
   - 응답 status code 확인 (200 vs 422 vs 500)

## 참고
- 같은 패턴: **BE-004, BE-005** 도 동일한 원인일 가능성 높음
- 우선 하나 (BE-003, BE-004 또는 BE-005 중 선택) 를 깊게 파고 원인 특정 후, 나머지에 적용

---

## 🔍 実装コード確認結果（2026-03-16）

### UpdateTreatmentPlan flow analysis

#### handler: treatment_plan_handler.go:109-137
- ✅ `strconv.ParseUint(c.Param("planId"), 10, 64)` で planId をパース
- ✅ `c.ShouldBindJSON(&req)` でリクエストをバインド
- ✅ `service.UpdateTreatmentPlanInput` に変換（全フィールドポインタ型）
- ✅ `h.svc.TreatmentPlan.Update(c.Request.Context(), planID, input)` を呼び出し
- ✅ エラー時は `RespondError(c, err)` で返却
- ✅ 成功時は `c.JSON(http.StatusOK, toTreatmentPlanResponse(plan))`

#### service: treatment_plan_service.go:90-100
- ✅ `buildTreatmentPlanUpdateFields(input)` で `map[string]any` を構築
- ✅ `len(fields) == 0` チェック → `WrapInvalidInput` で拒否
- ✅ `s.repo.Update(ctx, id, fields)` で DB 更新
- ✅ 成功後 `s.repo.FindByID(ctx, id)` で再取得して返却
- ✅ `slog.InfoContext(ctx, "treatment plan updated", ...)` でログ出力

#### repository: treatment_plan_repository.go:68-80
- ✅ `r.db.WithContext(ctx).Model(&model.TreatmentPlan{}).Where("id = ?", id).Updates(fields)` で DB UPDATE
- ✅ `result.RowsAffected == 0` チェック → `WrapNotFound` で 404 返却
- ✅ エラー時は `apperrors.Wrap(result.Error, "update treatment plan")` でラップ

#### ルート登録: treatment_plan_handler.go:152-157 + medical_record_handler.go:199
- ✅ `rg.PATCH("/:id/treatment-plans/:planId", h.UpdateTreatmentPlan)` で登録
- ✅ `h.RegisterTreatmentPlanMedicalRecordRoutes(records)` で medical-records サブリソースとして登録

### 根本原因（特定済み）

**バックエンドは完全に正常動作する。問題はフロントエンド側の API 呼び出し先が間違っている。**

1. **FE が間違ったエンドポイントにデータを送信している**
   - `useMedicalRecordForm.ts:101-116` は `PATCH /v1/medical-records/:id` に `plan` フィールドを送信
   - バックエンドの `updateMedicalRecordRequest` (medical_record_request.go:17-25) に `plan` フィールドは存在しない
   - Gin の `ShouldBindJSON` は未知のフィールドを無視する → `plan` は捨てられる → 200 で返却

2. **正しい API エンドポイントが使われていない**
   - 正しいパス: `PATCH /api/v1/medical-records/:id/treatment-plans/:planId`
   - FE にこの API を呼ぶコードが存在しない

3. **treatment_plans は medical_records の直接フィールドではなく子テーブル**
   - DB 再設計で分離済みだが FE が追従していない

### 修正方針

#### FE 修正
1. `features/medical-records/api/` に treatment-plan 用の API 関数を追加:
   - `create-treatment-plan.ts`: `POST /v1/medical-records/:id/treatment-plans`
   - `update-treatment-plan.ts`: `PATCH /v1/medical-records/:id/treatment-plans/:planId`
   - `get-treatment-plans.ts`: `GET /v1/medical-records/:id/treatment-plans`
2. `useMedicalRecordForm.ts` の `handleSave` から `plan` フィールドを削除
3. 治療プランタブに独自の保存フローを実装

#### 修正が不要なもの
- ❌ バックエンド（handler/service/repository）— 全て正常
- ❌ ルート登録 — 正常
- ❌ DB スキーマ — 正常

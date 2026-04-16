# Section 14 Master Settings CRUD - Test Execution Guide
## Browser Functional Test with Chrome DevTools MCP

**Date**: 2026-04-12  
**Target Environment**: http://localhost:3003 (local development)  
**Test Account**: admin@example.com / password  
**Total Test Cases**: 25 (5 masters × 5 CRUD operations each)

---

## Quick Reference: Test Endpoints

| Master | Route | GET Endpoint | POST/PATCH/DELETE |
|--------|-------|--------------|-------------------|
| 動物種 | /settings/animal-species | GET /api/v1/masters/animal-species | /api/v1/masters/animal-species/{id} |
| サービス種別 | /settings/reservation-type | GET /api/v1/masters/reservation-types | /api/v1/masters/reservation-types/{id} |
| スタッフ | /settings/staff | GET /api/v1/masters/staffs | /api/v1/masters/staffs/{id} |
| トリミング | /settings/trimming | GET /api/v1/masters/trimming-courses | /api/v1/masters/trimming-courses/{id} |
| 診断病名 | /settings/diagnosis | GET /api/v1/masters/diagnosis-names | /api/v1/masters/diagnosis-names/{id} |

---

## Test Case 1: Animal Species Master (`/settings/animal-species`)

### 1.1 READ Test
```
Test ID: TC-1.1-READ
Step 1: Navigate to http://localhost:3003/settings/animal-species
Expected: Page loads, displays list of animal species
Step 2: Count items in list
Expected Result: 6 items displayed (犬, 猫, 鳥, ウサギ, ハムスター, その他)
Network Check: GET /api/v1/masters/animal-species → HTTP 200
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 1.2 CREATE Test
```
Test ID: TC-1.2-CREATE
Precondition: Initial count = 6 (from 1.1)
Step 1: Click "+ 新規登録" button
Expected: Right side panel opens with "新規作成" title
Step 2: Focus on name field (should auto-focus)
Step 3: Type "テスト_動物種_20260412"
Step 4: Click "保存" button
Expected Result:
  - Network: POST /api/v1/masters/animal-species
    Payload: { name: "テスト_動物種_20260412" }
    Response: HTTP 201, returns { id: <new_id>, name: "テスト_動物種_20260412", ... }
  - UI: Toast shows "登録しました"
  - List: Count increases 6→7, new item visible at bottom
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 1.3 UPDATE Test
```
Test ID: TC-1.3-UPDATE
Precondition: Test item created in TC-1.2 with ID = <id>
Step 1: Click edit icon (pencil) on the test item
Expected: Side panel opens with current data "テスト_動物種_20260412"
Step 2: Clear name field, type "テスト_動物種_編集"
Step 3: Click "保存" button
Expected Result:
  - Network: PATCH /api/v1/masters/animal-species/<id>
    Payload: { name: "テスト_動物種_編集" }
    Response: HTTP 200, returns updated object
  - UI: Toast shows "編集しました"
  - List: Item name immediately updates to "テスト_動物種_編集"
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 1.4 DELETE Test
```
Test ID: TC-1.4-DELETE
Precondition: Test item exists with name "テスト_動物種_編集"
Step 1: Click edit icon or delete icon on the test item
Step 2: If edit panel: click trash icon. If list: click delete icon
Expected: Confirmation dialog shows "このデータを削除してもよろしいですか？"
Step 3: Click "削除する" button
Expected Result:
  - Network: DELETE /api/v1/masters/animal-species/<id>
    Response: HTTP 204 No Content
  - UI: Toast shows "削除しました"
  - List: Count decreases 7→6, test item removed
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 1.5 FK Constraint Error Test
```
Test ID: TC-1.5-FK-ERROR
Precondition: Animal species exists with existing pet references (e.g., "犬" id=1)
Step 1: Try to delete species "犬" (which has pets assigned)
Expected: Confirmation dialog shown
Step 2: Click "削除する"
Expected Result:
  - Network: DELETE /api/v1/masters/animal-species/1
    Response: HTTP 409 Conflict
    Body: { error: "この動物種はペット情報で使用中のため削除できません" }
  - UI: Toast shows error message (or generic "削除に失敗しました")
  - List: Count unchanged, item still visible
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

---

## Test Case 2: Reservation Type Master (`/settings/reservation-type`)

### 2.1 READ Test
```
Test ID: TC-2.1-READ
Step 1: Navigate to http://localhost:3003/settings/reservation-type
Expected: Page loads, displays list of reservation types
Step 2: Count items in list
Expected Result: 25 items displayed
Network Check: GET /api/v1/masters/reservation-types → HTTP 200
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 2.2 CREATE Test
```
Test ID: TC-2.2-CREATE
Precondition: Initial count = 25
Step 1: Click "+ 新規登録" button
Step 2: Type "テスト_予約_20260412"
Step 3: Click "保存" button
Expected Result:
  - Network: POST /api/v1/masters/reservation-types → HTTP 201
  - UI: Toast "登録しました", count 25→26
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 2.3 UPDATE Test
```
Test ID: TC-2.3-UPDATE
Precondition: Test item created
Step 1: Edit name to "テスト_予約_編集"
Step 2: Click "保存"
Expected Result:
  - Network: PATCH /api/v1/masters/reservation-types/<id> → HTTP 200
  - UI: List updates immediately
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 2.4 DELETE Test
```
Test ID: TC-2.4-DELETE
Precondition: Test item exists
Step 1: Click delete
Step 2: Confirm deletion
Expected Result:
  - Network: DELETE → HTTP 204
  - List: Count 26→25
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 2.5 FK Constraint Error Test
```
Test ID: TC-2.5-FK-ERROR
Precondition: Reservation type with existing reservations (e.g., "一般診療")
Step 1: Try to delete reservation type with existing appointments
Expected Result:
  - Network: DELETE → HTTP 409 Conflict
  - Error message shown
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

---

## Test Case 3: Staff Master (`/settings/staff`)

### 3.1 READ Test
```
Test ID: TC-3.1-READ
Step 1: Navigate to http://localhost:3003/settings/staff
Expected: Page loads, displays staff list
Step 2: Count items
Expected Result: 16 items
Network Check: GET /api/v1/masters/staffs → HTTP 200
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 3.2 CREATE Test
```
Test ID: TC-3.2-CREATE
Precondition: Initial count = 16
Step 1: Click "+ 新規登録"
Step 2: Type "テストスタッフ_20260412"
Step 3: Click "保存"
Expected Result:
  - Network: POST → HTTP 201
  - List: 16→17
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 3.3 UPDATE Test
```
Test ID: TC-3.3-UPDATE
Precondition: Test staff created
Step 1: Edit name to "テストスタッフ_編集"
Step 2: Click "保存"
Expected Result:
  - Network: PATCH → HTTP 200
  - List: Updates immediately
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 3.4 DELETE Test
```
Test ID: TC-3.4-DELETE
Precondition: Test staff exists
Step 1: Click delete, confirm
Expected Result:
  - Network: DELETE → HTTP 204
  - List: 17→16
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 3.5 FK Constraint Error Test
```
Test ID: TC-3.5-FK-ERROR
Precondition: Staff with existing appointments (e.g., doctor with scheduled reservations)
Step 1: Try to delete staff with appointments
Expected Result:
  - Network: DELETE → HTTP 409
  - Error message shown
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

---

## Test Case 4: Trimming Master (`/settings/trimming`)

### 4.1 READ Test
```
Test ID: TC-4.1-READ
Step 1: Navigate to http://localhost:3003/settings/trimming
Expected: Page loads
Step 2: Count items
Expected Result: 5 courses
Network Check: GET /api/v1/masters/trimming-courses → HTTP 200
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 4.2 CREATE Test
```
Test ID: TC-4.2-CREATE
Precondition: Initial count = 5
Step 1: Click "+ 新規登録"
Step 2: Type "テストコース_20260412"
Step 3: Click "保存"
Expected Result:
  - Network: POST → HTTP 201
  - List: 5→6
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 4.3 UPDATE Test
```
Test ID: TC-4.3-UPDATE
Precondition: Test course created
Step 1: Edit name to "テストコース_編集"
Step 2: Click "保存"
Expected Result:
  - Network: PATCH → HTTP 200
  - List: Updates
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 4.4 DELETE Test
```
Test ID: TC-4.4-DELETE
Precondition: Test course exists
Step 1: Click delete, confirm
Expected Result:
  - Network: DELETE → HTTP 204
  - List: 6→5
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 4.5 FK Constraint Error Test
```
Test ID: TC-4.5-FK-ERROR
Precondition: Course with existing trimming records
Step 1: Try to delete course with records
Expected Result:
  - Network: DELETE → HTTP 409
  - Error: "このトリミングコースはトリミング記録で使用中のため削除できません"
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

---

## Test Case 5: Diagnosis Master (`/settings/diagnosis`)

### 5.1 READ Test
```
Test ID: TC-5.1-READ
Step 1: Navigate to http://localhost:3003/settings/diagnosis
Expected: Page loads
Step 2: Count items
Expected Result: 20 diagnoses
Network Check: GET /api/v1/masters/diagnosis-names → HTTP 200
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 5.2 CREATE Test
```
Test ID: TC-5.2-CREATE
Precondition: Initial count = 20
Step 1: Click "+ 新規登録"
Step 2: Type "テスト病名_20260412"
Step 3: Click "保存"
Expected Result:
  - Network: POST → HTTP 201
  - List: 20→21
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 5.3 UPDATE Test
```
Test ID: TC-5.3-UPDATE
Precondition: Test diagnosis created
Step 1: Edit name to "テスト病名_編集"
Step 2: Click "保存"
Expected Result:
  - Network: PATCH → HTTP 200
  - List: Updates
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 5.4 DELETE Test
```
Test ID: TC-5.4-DELETE
Precondition: Test diagnosis exists
Step 1: Click delete, confirm
Expected Result:
  - Network: DELETE → HTTP 204
  - List: 21→20
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

### 5.5 FK Constraint Error Test
```
Test ID: TC-5.5-FK-ERROR
Precondition: Diagnosis with existing medical records
Step 1: Try to delete diagnosis with records
Expected Result:
  - Network: DELETE → HTTP 409
  - Error: "この診断名は診療記録で使用中のため削除できません"
Status: [  ] PASS [  ] FAIL [  ] PARTIAL
Notes: ___________________________________
```

---

## Summary Checklist

### All Tests Completed
- [ ] TC-1.1 through TC-1.5 (Animal Species - 5 tests)
- [ ] TC-2.1 through TC-2.5 (Reservation Types - 5 tests)
- [ ] TC-3.1 through TC-3.5 (Staff - 5 tests)
- [ ] TC-4.1 through TC-4.5 (Trimming - 5 tests)
- [ ] TC-5.1 through TC-5.5 (Diagnosis - 5 tests)

### HTTP Status Code Verification
- [ ] All GET requests returned HTTP 200
- [ ] All POST requests returned HTTP 201
- [ ] All PATCH requests returned HTTP 200
- [ ] All successful DELETE requests returned HTTP 204
- [ ] All FK constraint violations returned HTTP 409

### UI/UX Verification
- [ ] All toast messages displayed correctly
- [ ] All list counts updated correctly (±1 for create/delete)
- [ ] All edits reflected immediately in list
- [ ] All side panels opened/closed correctly
- [ ] All confirmation dialogs displayed

### Error Handling Verification
- [ ] FK constraint errors returned HTTP 409
- [ ] Error messages displayed in toast or modal
- [ ] List state not modified on error

### Test Results Summary
```
Total Test Cases: 25
PASS:     [ ]
FAIL:     [ ]
PARTIAL:  [ ]
BLOCKED:  [ ]

Critical Issues Found: [ ]
Minor Issues Found:    [ ]
```

---

## Post-Test Cleanup

After all tests complete:
1. Verify all test records deleted (lists returned to initial counts)
2. Verify no orphaned data in database
3. Document any bugs discovered in FUNCTIONAL_TEST_REPORT.md
4. Create BUG-XXX tickets for any NG tests

---

## Notes for Test Executor

- Use consistent timestamp format (YYYY-MM-DD HH:MM:SS) for test data
- Allow 2-3 seconds between operations for API responses
- Watch browser console for JavaScript errors (mcp__chrome-devtools__list_console_messages)
- Capture network requests in DevTools for verification
- If any test fails, screenshot the error before proceeding
- Don't modify application code during testing


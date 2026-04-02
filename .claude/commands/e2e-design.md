---
description: E2E テスト設計・シナリオ生成
---

# /e2e-design [feature]

E2E テストシナリオを設計・生成します。

## 使用法

```bash
# Owner 機能の E2E テスト設計
/e2e-design owners

# Medical Record 機能
/e2e-design medical-records
```

## 生成内容

- ユーザーフロー分析
- クリティカルパス特定
- テストシナリオ設計
- Playwright スクリプト生成

## テストシナリオ例

```
Scenario 1: Owner 登録→ Pet 追加→ Vaccination 記録
  1. Login
  2. Click "新規Owner"
  3. Fill form and submit
  4. Verify success
  5. Add Pet
  6. Record Vaccination
  7. Verify record
```

## 出力形式

```
E2E Tests Generated:
- owner-flow.spec.ts
- medical-flow.spec.ts

Total scenarios: XX
Estimated execution time: XXm
```

## 使用エージェント

`test-strategist` (Sonnet) を自動起動

---
description: ハーネスの現在状態を表示・リセット
argument-hint: "[reset]"
---

# ハーネス状態確認

`.claude/logs/harness-active.json` の内容を表示し、ハーネスの進行状況を確認する。

## 引数なしの場合: 状態表示

```bash
cat .claude/logs/harness-active.json 2>/dev/null || echo "ハーネス未実行"
```

表示形式:

```
HARNESS STATUS
==============
Task:      BE-042
Iteration: 2 / 3
Started:   2026-06-10T09:00:00Z
Changed files:
  - backend/internal/handler/patient.go
  - backend/internal/service/patient.go

Iteration results:
  [1] FAIL — P15(clinic_id), P10(error wrap)
  [2] In progress...
```

ハーネス未実行の場合:
```
No active harness session.
Run: /harness <task>
```

## reset の場合: ハーネス状態をクリア

```bash
rm -f .claude/logs/harness-active.json
```

確認メッセージ:
```
HARNESS RESET: State file removed.
Next /harness run will start fresh from iteration 1.
```

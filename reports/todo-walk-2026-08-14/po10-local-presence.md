# PO-10 presence（local のみ · 2026-08-14）

| 項目 | 値 |
|------|-----|
| env | **local** docker `ekarte_db` |
| 方針 | 値・clinic_id を台帳に残さない。集計のみ |
| STG/PROD | **未実施**（承認 window が必要） |

## 集計（legacy channel secret）

| 指標 | 件数 |
|------|------|
| settings 行 | 3 |
| `line_channel_secret` present の行 | **0** |
| `liff_id` present の行 | 2 |
| secret present の clinic 数 | **0** |

```
po10_local: secret_present_clinics=0 liff_present_rows=2 total_rows=3 drop=FORBIDDEN_UNTIL_POLICY operator=agent_readonly opaque_ref=reports/todo-walk-2026-08-14/po10-local-presence.md
```

**注:** local ゼロは STG/PROD の代替にしない。DROP / guard 除去はしない。

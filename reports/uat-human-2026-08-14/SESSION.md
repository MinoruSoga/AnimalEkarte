# UAT human session log

| 項目 | 値 |
|------|-----|
| 日付 | 2026-MM-DD |
| プラン | A 30分 / **B 90分** / C 3時間 |
| 実施 build SHA | （`git log -1 --format=%h` · pull 後に記入） |
| フル UAT 証跡 build | `1386e1db0`（`reports/uat-2026-08-14/`） |
| 正本 | [fable exec-session](../fable-po-confirm-answer-2026-08-14-exec-session.md) · [todo-po.md](../../todo-po.md) |

## 送付（M1–M5）

| 通 | Status | 備考 |
|----|--------|------|
| M1 #201 | pending | |
| M2 #256 U13 | pending | |
| M3 #258 | pending | |
| M4 #299 / PS109 | pending | |
| M5 PO-008 | pending | |

## H レーン結果

E-6 途中行（値は実施後に置換）:

```text
#254 human lane: H1=[PENDING_STG] H2=[PENDING_STG] H3=[ ] H4=[CARRY] H5=[ ] H6=[CARRY] H7=[ ] build=[ ] uat_full=FAIL0@1386e1db0 residual_disposition=[PENDING] final_signoff=[PENDING] opaque_ref=[ ]
```

| ID | result | evidence | note |
|----|--------|----------|------|
| H3 | | H3-audit.md | |
| H5 | | H5-shift-templates.md | |
| H7 | | H7-spot-check.md | |
| H4 | CARRY / … | H4-closing.md | 90分既定は持ち越し可 |
| H6 | CARRY / … | H6-identity-links.md | 90分既定は持ち越し可 |
| H1 | PENDING_STG | — | merge 後 |
| H2 | PENDING_STG | — | merge 後 |

result ∈ `PASS` | `FAIL_BUG` | `ACCEPT_DISPOSITION` | `CARRY` | `PENDING_STG`

## 持ち越し

- 

## 禁止行為チェック

- [ ] #254/#253 close していない
- [ ] green 前 merge していない
- [ ] 値を発明・代筆していない
- [ ] シナリオ md に結果を書いていない
- [ ] disposition なしスキップをしていない

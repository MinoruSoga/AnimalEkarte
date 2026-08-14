# UAT environment status

| 項目 | 値 |
|------|-----|
| 確認日時 | 2026-08-14 |
| チェックスクリプト | `docs/ops/testing/scripts/check-uat-env.sh` |
| 結果 | **READY**（FAIL 0） |

## 観測

| チェック | 結果 |
|----------|------|
| frontend :3003 | 200 |
| backend :8080/health | 200 |
| E2E_LOGIN_* | set（値は非公開） |
| POST /api/v1/login | 200 |
| LIFF_MOCK / VITE_LIFF_MOCK | local 用 true |
| Playwright package | present |
| Chrome :9222 | listening（DevTools MCP 用） |
| アーキテクチャ文書 | TEST_ARCHITECTURE 他 present |

## 次の実施

1. 受入正本: `docs/ops/testing/scenarios/`
2. 項目単位: `FIELD-LEVEL-PROTOCOL.md` + `FORM-FIELD-INVENTORY.md`
3. 実行: browser-test / Playwright MCP / 再現スクリプト
4. 証跡: `reports/uat-YYYY-MM-DD/`

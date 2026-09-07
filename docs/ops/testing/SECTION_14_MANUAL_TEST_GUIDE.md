# ブラウザ操作・探索テストガイド (Manual Test Guide)

> **目的**: L4 [scenarios](scenarios/README.md) を補う focused exploratory checks を示す。網羅的受入の正本ではない。
> **最新更新**: 2026-09-01

## 1. 安全境界

- approved isolated UAT tenant または disposable local DB だけで行う。production と未承認 shared clinic は禁止。
- create/update/delete と external send は target、許可 clinic、cleanup を事前定義する。
- real LINE/L-step/file send は human STG lane が明示承認した場合だけ。未承認なら mock または BLOCKED。
- 下記で canonical scenario/API が示されない期待は exploratory (`要実測`) であり、文書の記述だけで acceptance requirement に昇格させない。

## 2. focused checks

### 2.1 outpatient

- `/reservations`: 15-minute drag snapping を探索確認する。
- home/reception: 予約の表示と status transition を対象 scenario の期待に照合する。
- medical record: `PatientInfoCard`、SOAPS、vital graph を該当 [scenario inventory](scenarios/README.md) と照合する。
- next visit: 推奨日保存後の best-effort next-visit **tag sync evidence** を確認する。reminder delivery は別の後日 `next_visit_reminder` batch であり、保存時に外部 reminder reservation が即作成されるとは期待しない。対象日に scheduled batch trigger log を別確認する。

### 2.2 accounting

- persisted billing totals を authoritative contract とし、frontend `frontend/src/lib/calculations.ts` と `frontend/src/features/accounting/tax-breakdown.ts` の表示計算を照合する。
- `/accounting/close`: actual register cash **total** を入力し、theoretical cash との差額を確認する。denomination-by-denomination input は存在しない。
- `/accounting/reports`: monthly sales/payment/daily trend と export を該当 API/spec に照合する。spec がない差異は `要実測` と記録する。

### 2.3 L-step / CRM

- `/settings/lstep/tags`: tag name、owner count/detail、automatic/manual classification を確認する。CPM stage distribution の画面ではない。
- `/lstep/checkup-sync`: CPM stage と適用可能な owner/pet/LINE filters を確認する。
- owner chat/file send: mock で UI boundary を確認する。real external send は human-approved lane のみ。

## 3. security/exploratory guards

- authorization: system admin と clinic staff permission group/capability の差を確認する。
- master delete/FK response、navigation blocking、audit evidence は対応する API/screen spec が特定できる場合だけ acceptance expectation とする。未特定なら `要実測`。
- warning dialog の存在だけを safety control とみなさない。canonical spec が lock/Undo/physical block を要求する場合はそれを確認する。

## 4. identity provisioning

Product account model は `admin/doctor/staff` role enum ではない。

| identity | provisioning requirement | use |
|:--|:--|:--|
| system admin | explicit UAT account with `is_system_admin` | system-level administration only |
| clinic staff | explicit account/staff attached to target clinic with approved permission group/capabilities | clinical/accounting/master operations permitted by capabilities |
| restricted clinic staff | least-privilege permission group | negative authorization checks |

CSV `002_master` は permission groups/rules を含むが account を含まない。別 phase `003_login` は non-production の許可環境で合成 account を作るが、上表の専用 UAT identities/capabilities を満たすとは限らない。[UAT setup](UAT-ENV-SETUP.md) と [staff provisioning](../deploy/STAFF_ACCOUNT_PROVISIONING.md) に従う。credential 値を文書・report・chat に書かない。

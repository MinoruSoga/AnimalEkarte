# Runbook: line_reservation_settings.break_hours 形状監査

> **目的**: R1-3デプロイ前のbreak_hoursデータ形状監査手順を定義する。
> **読者**: 運用担当(deploy実施者)。
> **タイミング**: R1-3デプロイ前のSTG/本番監査実行時。

> BE-refactor.md R1-3（D10/F-2）フォローアップ。deploy 前に **1回** 実施する STG/prod 監査手順。

## 背景（なぜ必要か）

R1-3 で `parseBusinessHoursForDate` の `break_hours` JSON unmarshal 失敗を **fail-closed（予約拒否）** に変更した
（従来は `breaks=nil` にフォールバックし休憩時間の重複検証を silent にスキップしていた=本物のバグ）。
併せて保存時ガード `validateBreakHoursShape`（`line_reservation_setting_service.go`）を追加し、
以降の**新規保存**では不正形状（`{}`・非HHMM 等）を 400 で拒否する。

しかしこのガード導入**以前**に保存された既存行に、構文は valid JSON だが形状が不正な `break_hours`
（例: `{}`、`[{"start":"900"}]`、`[{"start":"25:00"}]`）が残っている場合、その clinic の LINE 予約作成が
**deploy 直後から警告なく全件拒否**され続ける可用性障害になり得る。本監査でそれを**事前検出**する。

## 対象の「不正形状」の定義（`minutesSinceMidnight` と一致）

有効な `break_hours` は `[{"start":"HHMM","end":"HHMM"}, ...]`（jsonb array）で、各 start/end は:
- 長さ 4 の文字列、
- `HH` ∈ 00–23、`MM` ∈ 00–59。

`NULL` / `[]`（休憩なし）は有効。上記に反するものが「fail-closed で予約拒否される」不正データ。

## 監査 SQL（read-only・STG/prod で実行）

```sql
-- 不正形状の break_hours を持つ line_reservation_settings を列挙する（読み取り専用・変更なし）
SELECT
  id,
  clinic_id,
  break_hours
FROM line_reservation_settings
WHERE break_hours IS NOT NULL
  AND jsonb_typeof(break_hours) IS DISTINCT FROM 'null'
  AND break_hours::text <> '[]'
  AND (
    -- (a) そもそも配列でない（例: {} や "文字列"）→ []BreakPeriod への unmarshal 失敗
    jsonb_typeof(break_hours) <> 'array'
    -- (b) 配列だがエントリの形状/値が不正
    OR EXISTS (
      SELECT 1
      FROM jsonb_array_elements(
        CASE WHEN jsonb_typeof(break_hours) = 'array'
             THEN break_hours ELSE '[]'::jsonb END
      ) AS e
      WHERE jsonb_typeof(e) <> 'object'
         OR NOT (e ? 'start') OR NOT (e ? 'end')
         OR jsonb_typeof(e->'start') <> 'string'
         OR jsonb_typeof(e->'end')   <> 'string'
         -- HHMM: HH=00..23, MM=00..59, ちょうど4桁（minutesSinceMidnight と一致）
         OR (e->>'start') !~ '^([01][0-9]|2[0-3])[0-5][0-9]$'
         OR (e->>'end')   !~ '^([01][0-9]|2[0-3])[0-5][0-9]$'
    )
  )
ORDER BY clinic_id, id;
```

**期待結果**: 0 行なら deploy 安全。1 行以上ならその clinic の LINE 予約が deploy 後に fail-closed になる。

## 判定と是正

| 監査結果 | 対応 |
|---|---|
| 0 行 | ✅ R1-3 を安全に deploy 可能。 |
| 1 行以上 | ⚠️ deploy 前に該当 clinic に正しい `break_hours` を再設定してもらう（管理画面の予約設定保存で `validateBreakHoursShape` を通した値に上書き）。または該当行を `'[]'`（休憩なし）へ是正してから deploy。**data 修正は明示承認 + 通常の管理フロー経由**で行い、本 runbook から直接 UPDATE はしない。 |

是正 UPDATE 例（**承認後のみ・要バックアップ**、休憩なしへ倒す最小是正）:

```sql
-- 承認後のみ。対象 id を監査 SQL の結果から明示指定すること（無条件 UPDATE 禁止）
-- UPDATE line_reservation_settings SET break_hours = '[]'::jsonb WHERE id = <明示ID>;
```

## 検証（このロジックはコードでテスト済み）

- 保存側ガード: `line_reservation_setting_service_test.go`（object/非string/非HHMM→400・空→許容・正常→保存）
- 予約作成側 fail-closed: `reservation_validators_test.go`（`TestValidateBusinessRules_MalformedBreakHours_FailsClosed` ほか）

上記 SQL の HHMM 正規表現 `^([01][0-9]|2[0-3])[0-5][0-9]$` は `minutesSinceMidnight`
（`timeslot_engine.go`: len==4・HH≤23・MM≤59）と等価。

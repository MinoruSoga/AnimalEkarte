-- 005_add_emg_period.sql
-- Issue #150: レジ締めに EMG（緊急）区分を追加する。
--
-- cash_register_closes.period を varchar(2) → varchar(3) に拡張し、
-- CHECK 制約に 'emg' を許可する。
-- 既存の 'am' / 'pm' レコードは列拡張・制約緩和のみのため影響を受けない。
--
-- 制約名について:
--   001_init.sql は period をインライン CHECK (period IN ('am','pm')) で定義しており、
--   PostgreSQL はこれを cash_register_closes_period_check と自動命名する。
--   ただし制約名は環境差・手動変更でズレうるため、固定名の DROP には依存しない。
--   period 列を参照する CHECK 制約を pg_constraint から名前非依存で検出し、
--   すべて削除してから新しい制約を貼り直す（再実行安全・冪等）。

ALTER TABLE cash_register_closes
    ALTER COLUMN period TYPE varchar(3);

-- period 列を参照する既存 CHECK 制約を名前に依存せず全削除する。
DO $$
DECLARE
    constraint_name text;
BEGIN
    FOR constraint_name IN
        SELECT con.conname
        FROM pg_constraint con
        WHERE con.conrelid = 'cash_register_closes'::regclass
          AND con.contype = 'c'
          AND EXISTS (
              SELECT 1
              FROM unnest(con.conkey) AS cols(attnum)
              JOIN pg_attribute att
                ON att.attrelid = con.conrelid
               AND att.attnum = cols.attnum
              WHERE att.attname = 'period'
          )
    LOOP
        EXECUTE format('ALTER TABLE cash_register_closes DROP CONSTRAINT %I', constraint_name);
    END LOOP;
END $$;

ALTER TABLE cash_register_closes
    ADD CONSTRAINT cash_register_closes_period_check
    CHECK (period IN ('am', 'pm', 'emg'));

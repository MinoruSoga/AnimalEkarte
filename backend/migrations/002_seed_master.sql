-- =============================================================================
-- Animal Ekarte - マスタデータ v1.0
-- PostgreSQL 18
-- 冪等性保証: ON CONFLICT DO NOTHING
-- 内容: システム共通マスタデータ（clinic_id なし・全クリニック共通）
-- 依存: 001_init.sql
-- 実行順: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. companies（本部情報: 1件）
-- -----------------------------------------------------------------------------
INSERT INTO companies (name) VALUES
    ('ノア動物病院')
ON CONFLICT DO NOTHING;

-- -----------------------------------------------------------------------------
-- 2. animal_species（ペット種類: 6件、システム共通・clinic_idなし）
-- -----------------------------------------------------------------------------
INSERT INTO animal_species (id, name, is_active, sort_order) VALUES
    (1, '犬',         true, 1),
    (2, '猫',         true, 2),
    (3, '鳥',         true, 3),
    (4, 'うさぎ',     true, 4),
    (5, 'ハムスター', true, 5),
    (6, 'その他',     true, 6)
ON CONFLICT DO NOTHING;

SELECT setval(pg_get_serial_sequence('animal_species', 'id'), (SELECT MAX(id) FROM animal_species));

-- =============================================================================
-- 006_fix_hospital_settings_permissions: 医院マスタ権限修正
-- =============================================================================
-- BUG-377 / BUG-378: 医院マスタ (hospital-settings) の作成・削除権限を
-- permission_group_rules から剥奪する。
--
-- 目的: Frontend の usePermission() と Backend の RequirePermission() を
--       単一の真実とし、is_system_admin 二重ガードを廃止する。
--       is_system_admin=true は hasPermission() でバイパスされるため、
--       能力は維持される（admin@noavet.jp 等は従来通り作成・削除可）。
-- -----------------------------------------------------------------------------
UPDATE permission_group_rules
   SET can_create = false,
       can_delete = false
 WHERE resource = 'hospital-settings';

-- =============================================================================
-- 005_seed_appointments_monthly: デモ予約データ投入（向こう1ヶ月分）
-- =============================================================================
-- 2026-04-15 〜 2026-05-15 のデモ予約データを 3 院に投入
-- 冪等: notes に [DEMO-MONTHLY] を含む既存レコードがある場合は全件スキップ
-- 日曜休診、平日 9:00-17:00 / 土曜 9:00-12:00
-- ======================================================================

DO $$
DECLARE
    v_start_date date := '2026-04-15';
    v_end_date   date := '2026-05-15';
    v_day        date;
    v_clinic     bigint;
    v_owner      bigint;
    v_pet        bigint;
    v_type       bigint;
    v_doctor     bigint;
    v_start_ts   timestamptz;
    v_end_ts     timestamptz;
    v_duration   int;
    v_hour       int;
    v_minute     int;
    v_status     text;
    v_visit_type text;
    v_is_desig   boolean;
    v_notes      text;
    v_daily_cnt  int;
    v_slot_idx   int;
    v_dow        int;
    v_today      date := '2026-04-15';

    -- clinic 3 (八王子): owner-pet 候補
    arr_c3_pets   bigint[] := ARRAY[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28];
    arr_c3_owners bigint[] := ARRAY[1,1,2,3,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,18,19,19,20,21,21,22,22];
    arr_c3_types  bigint[] := ARRAY[1,2,3,5,6,7,9,11];   -- 診察/ワクチン/狂犬病/フィラリア/健診/トリミング
    arr_c3_docs   bigint[] := ARRAY[1,2,3,4,33];

    -- clinic 4 (城東)
    arr_c4_pets   bigint[] := ARRAY[29,30,31,32,33,34,35,36,37,38];
    arr_c4_owners bigint[] := ARRAY[23,23,24,25,26,27,28,28,29,30];
    arr_c4_types  bigint[] := ARRAY[26,27,28,29,30,31,33];
    arr_c4_docs   bigint[] := ARRAY[4,16,17,18,34];

    -- clinic 5 (敷島)
    arr_c5_pets   bigint[] := ARRAY[39,40,41,42,43,44,45,46,47,48];
    arr_c5_owners bigint[] := ARRAY[31,32,33,34,34,35,36,37,38,38];
    arr_c5_types  bigint[] := ARRAY[45,46,47,48,49,50,52];
    arr_c5_docs   bigint[] := ARRAY[4,26,27,28,35];

    arr_notes text[] := ARRAY[
        '皮膚の経過観察', '定期健診', 'ワクチン接種希望', '食欲低下', '耳を痒がる',
        '嘔吐が続く', '初回接種', '体重管理', '再診', 'トリミング希望',
        '咳が出る', '爪切り希望', '健康診断', '下痢気味', '目の充血',
        'シニア健診', '歯石除去相談', '予防薬処方', '便の状態相談', '元気がない'
    ];
BEGIN
    -- 冪等ガード: 既に投入済みならスキップ
    IF EXISTS (SELECT 1 FROM appointments WHERE notes LIKE '%[DEMO-MONTHLY]%') THEN
        RAISE NOTICE '005_seed_appointments_monthly: already applied, skipping';
        RETURN;
    END IF;

    -- 再現性のためシード固定
    PERFORM setseed(0.42);

    v_day := v_start_date;
    WHILE v_day <= v_end_date LOOP
        v_dow := EXTRACT(DOW FROM v_day)::int;  -- 0=Sun ... 6=Sat

        -- 日曜休診
        IF v_dow <> 0 THEN
            -- 各院ループ
            FOR v_clinic IN SELECT unnest(ARRAY[3::bigint, 4::bigint, 5::bigint]) LOOP
                -- 1日あたりの件数: 八王子 4-6件、他院 2-4件、土曜は半分
                v_daily_cnt := CASE
                    WHEN v_clinic = 3 AND v_dow = 6 THEN 2 + floor(random()*2)::int
                    WHEN v_clinic = 3              THEN 4 + floor(random()*3)::int
                    WHEN v_dow = 6                 THEN 1 + floor(random()*2)::int
                    ELSE                                2 + floor(random()*3)::int
                END;

                FOR v_slot_idx IN 1..v_daily_cnt LOOP
                    -- 時間帯: 土曜 9-11、平日 9-16 台
                    IF v_dow = 6 THEN
                        v_hour := 9 + floor(random()*3)::int;
                    ELSE
                        v_hour := 9 + floor(random()*8)::int;  -- 9-16
                        IF v_hour = 12 THEN v_hour := 14; END IF;  -- 昼休み回避
                    END IF;
                    v_minute := (floor(random()*4)::int) * 15;  -- 00/15/30/45

                    v_start_ts := (v_day::timestamp + make_interval(hours => v_hour, mins => v_minute)) AT TIME ZONE 'Asia/Tokyo';

                    -- 予約種別を選択
                    IF v_clinic = 3 THEN
                        v_type  := arr_c3_types[1 + floor(random() * array_length(arr_c3_types,1))::int];
                        v_doctor:= arr_c3_docs[1 + floor(random() * array_length(arr_c3_docs,1))::int];
                        v_slot_idx := v_slot_idx;  -- noop
                    ELSIF v_clinic = 4 THEN
                        v_type  := arr_c4_types[1 + floor(random() * array_length(arr_c4_types,1))::int];
                        v_doctor:= arr_c4_docs[1 + floor(random() * array_length(arr_c4_docs,1))::int];
                    ELSE
                        v_type  := arr_c5_types[1 + floor(random() * array_length(arr_c5_types,1))::int];
                        v_doctor:= arr_c5_docs[1 + floor(random() * array_length(arr_c5_docs,1))::int];
                    END IF;

                    -- 所要時間を reservation_types から取得
                    SELECT duration_minutes INTO v_duration FROM reservation_types WHERE id = v_type;
                    v_end_ts := v_start_ts + make_interval(mins => COALESCE(v_duration, 15));

                    -- owner/pet を選択
                    IF v_clinic = 3 THEN
                        v_slot_idx := 1 + floor(random() * array_length(arr_c3_pets,1))::int;
                        v_pet   := arr_c3_pets[v_slot_idx];
                        v_owner := arr_c3_owners[v_slot_idx];
                    ELSIF v_clinic = 4 THEN
                        v_slot_idx := 1 + floor(random() * array_length(arr_c4_pets,1))::int;
                        v_pet   := arr_c4_pets[v_slot_idx];
                        v_owner := arr_c4_owners[v_slot_idx];
                    ELSE
                        v_slot_idx := 1 + floor(random() * array_length(arr_c5_pets,1))::int;
                        v_pet   := arr_c5_pets[v_slot_idx];
                        v_owner := arr_c5_owners[v_slot_idx];
                    END IF;

                    -- ステータス: 今日より過去=completed混在、今日=checked_in/in_consultation、未来=confirmed/pending
                    IF v_day < v_today THEN
                        v_status := (ARRAY['completed','completed','accounting','cancelled'])[1 + floor(random()*4)::int];
                    ELSIF v_day = v_today THEN
                        v_status := (ARRAY['confirmed','checked_in','in_consultation','completed'])[1 + floor(random()*4)::int];
                    ELSE
                        v_status := CASE WHEN random() < 0.15 THEN 'pending' ELSE 'confirmed' END;
                    END IF;

                    v_visit_type := CASE WHEN random() < 0.3 THEN 'first' ELSE 'revisit' END;
                    v_is_desig   := random() < 0.3;
                    v_notes      := arr_notes[1 + floor(random() * array_length(arr_notes,1))::int] || ' [DEMO-MONTHLY]';

                    BEGIN
                        INSERT INTO appointments (
                            clinic_id, start_time, end_time, owner_id, pet_id,
                            visit_type, reservation_type_id, doctor_id, is_designated,
                            status, notes, source
                        ) VALUES (
                            v_clinic, v_start_ts, v_end_ts, v_owner, v_pet,
                            v_visit_type::visit_type, v_type, v_doctor, v_is_desig,
                            v_status::reservation_status, v_notes, 'manual'::reservation_source
                        );
                    EXCEPTION
                        WHEN unique_violation THEN
                            -- 同一医師・同一時刻の衝突はスキップ（デモ用途で許容）
                            NULL;
                    END;
                END LOOP;
            END LOOP;
        END IF;

        v_day := v_day + 1;
    END LOOP;

    RAISE NOTICE '005_seed_appointments_monthly: inserted appointments for % to %', v_start_date, v_end_date;
END $$;

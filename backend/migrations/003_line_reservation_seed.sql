-- ============================================================
-- LINE予約システム シードデータ
-- 既存 service_types / staffs の予約用カラムに初期値を設定
-- ============================================================

-- ============================================================
-- A. service_types の予約用カラム更新
-- 既存レコードの名前で照合してUPDATE（clinic_id条件は外してグローバル適用）
-- ============================================================

-- 顧客向けメニュー（reservation_visible = true）
UPDATE service_types SET duration_minutes = 15,  short_name = '診察',     reservation_visible = true,  is_internal = false WHERE name = '一般診察';
UPDATE service_types SET duration_minutes = 15,  short_name = '再診',     reservation_visible = true,  is_internal = false WHERE name = '一般診察(再診)';
UPDATE service_types SET duration_minutes = 15,  short_name = 'ワクチン', reservation_visible = true,  is_internal = false WHERE name = 'ワクチン接種';
UPDATE service_types SET duration_minutes = 15,  short_name = '狂犬病',   reservation_visible = true,  is_internal = false WHERE name = '狂犬病';
UPDATE service_types SET duration_minutes = 15,  short_name = 'フィラリア予防', reservation_visible = true, is_internal = false WHERE name = 'フィラリア予防';
UPDATE service_types SET duration_minutes = 15,  short_name = '健康診断', reservation_visible = true,  is_internal = false WHERE name = '健康診断';
UPDATE service_types SET duration_minutes = 15,  short_name = 'お手入れ', reservation_visible = true,  is_internal = false WHERE name = 'お手入れ';
UPDATE service_types SET duration_minutes = 60,  short_name = '手術',     reservation_visible = true,  is_internal = false WHERE name = '手術';
UPDATE service_types SET duration_minutes = 15,  short_name = 'ホテルお預け', reservation_visible = true, is_internal = false WHERE name = 'ホテルお預け';
UPDATE service_types SET duration_minutes = 15,  short_name = 'ホテルお迎え', reservation_visible = true, is_internal = false WHERE name = 'ホテルお迎え';
UPDATE service_types SET duration_minutes = 15,  short_name = '面会',     reservation_visible = true,  is_internal = false WHERE name = '面会';
UPDATE service_types SET duration_minutes = 15,  short_name = '健診報告', reservation_visible = true,  is_internal = false WHERE name = '健康診断結果報告';
UPDATE service_types SET duration_minutes = 15,  short_name = 'クイック', reservation_visible = true,  is_internal = false WHERE name = 'クイックシャンプー';
UPDATE service_types SET duration_minutes = 60,  short_name = '',         reservation_visible = true,  is_internal = false WHERE name = '室内ドックラン（アジリティー開放）';
UPDATE service_types SET duration_minutes = 15,  short_name = '相談',     reservation_visible = false, is_internal = false WHERE name = '無料相談（猫専門病院のみ）';

-- 管理者専用メニュー（is_internal = true, reservation_visible = false）
UPDATE service_types SET duration_minutes = 60,  short_name = '休憩枠',      reservation_visible = false, is_internal = true WHERE name = '休憩枠';
UPDATE service_types SET duration_minutes = 30,  short_name = 'OPE準備のため', reservation_visible = false, is_internal = true WHERE name = 'OPE準備のため';
UPDATE service_types SET duration_minutes = 60,  short_name = '予約不可',    reservation_visible = false, is_internal = true WHERE name = '予約不可' AND duration_minutes = 15;
UPDATE service_types SET duration_minutes = 30,  short_name = '引継ぎ',      reservation_visible = false, is_internal = true WHERE name = 'カンファレンス';

-- ============================================================
-- B. staffs の予約用カラム更新
-- ============================================================

-- 獣医師: staff_type = 'doctor', reservation_visible = true
UPDATE staffs SET staff_type = 'doctor', reservation_visible = true
WHERE name IN ('林 文明', '三井 隆行', '菊島 弘貴', '鈴木 諒平', '加藤 茉里', '稲村 一真', '原 駿太朗');

-- ============================================================
-- C. reservation_settings の初期データ（クリニックが存在する場合に INSERT）
-- ============================================================
INSERT INTO reservation_settings (
    clinic_id,
    status,
    business_hours,
    break_hours,
    daily_limit,
    booking_window_max_days,
    booking_window_min_days,
    calendar_months,
    phone_number,
    time_slot_mode,
    time_slot_interval_minutes,
    no_staff_mode,
    show_no_staff_option,
    national_holiday_closed,
    additional_fields
)
SELECT
    c.id,
    'stopped',
    '{"start":"0900","end":"1900"}',
    '[{"start":"1200","end":"1300"}]',
    1,
    30,
    2,
    2,
    '0552334126',
    'minimize_gaps',
    15,
    'first_available',
    true,
    false,
    '[
        {"key":"phone","label":"電話番号","required":true,"placeholder":"例) 090-1234-5678"},
        {"key":"owner_name","label":"飼い主名","required":true,"placeholder":""},
        {"key":"pet_info","label":"ペットの名前と種類","required":true,"placeholder":"例) ポチ（柴犬）"},
        {"key":"symptoms","label":"診察内容","required":false,"placeholder":""}
    ]'
FROM clinics c
ON CONFLICT (clinic_id) DO NOTHING;

-- ============================================================
-- C-2. テスト環境用 LINE クレデンシャルを設定（docs/LINE-SETUP.md 参照）
-- clinic_id=3: テスト-八王子（@642hdxoh）
-- clinic_id=4: テスト-城東（@151lnsqa）
-- ※ 本番環境では管理画面「LINE予約 基本設定」から入力すること
-- ============================================================
UPDATE reservation_settings SET
    line_channel_id     = '2009755544',
    line_channel_secret = '5344ef84eb7072b5894f7e087db28827',
    liff_id             = '2009755581-w5NOA3EW',
    line_access_token   = 'pwMi3yP6jhRa0xbmnR0IPEcE5l+OIp21a7ia3hmoiuFSCvqkR5Tmmfm6fLoSTB1Bt7uQjAe9NN7fZ+LBDtNKLGnrqBrjDmhTnws9PVxQKLyinomNzUAb61KADX7NJmFBfEsLQQ9VmlU+tMJcWh+zswdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 3;

UPDATE reservation_settings SET
    line_channel_id     = '2009755545',
    line_channel_secret = '25e4661a8de553953a4b34c6ac7a91cb',
    liff_id             = '2009755586-nvKfG3Cp',
    line_access_token   = 'CUAtYMry8doD9ALCF/6Y0JocVqRrxC8IbOCMyRyxwDw5EJhyujJ4lQ8mVGrt7WawTi+voAxZ79mKAg+4qlsUPBVU6VMZdk7wEA42NZXQAl8gBr2tSYmZpbRzAiLfuGhxuba+koBHVk8yTuaCCjLBzAdB04t89/1O/w1cDnyilFU='
WHERE clinic_id = 4;

-- ============================================================
-- D. staff_excluded_service_types の初期データ
-- 獣医師が非対応のコース（お手入れ, 手術, ホテルお預け, ホテルお迎え, 面会, クイックシャンプー）
-- ============================================================
INSERT INTO staff_excluded_service_types (staff_id, service_type_id)
SELECT s.id, st.id
FROM staffs s
CROSS JOIN service_types st
WHERE s.name IN ('林 文明', '三井 隆行', '菊島 弘貴', '鈴木 諒平', '加藤 茉里', '稲村 一真', '原 駿太朗')
  AND st.name IN ('お手入れ', '手術', 'ホテルお預け', 'ホテルお迎え', '面会', 'クイックシャンプー')
ON CONFLICT (staff_id, service_type_id) DO NOTHING;

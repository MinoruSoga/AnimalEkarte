-- PU-4010 の既定スロットが 9600 のまま保存された station 行を 2400 8(E)1 へ直す。
-- 電文正本 old_db/docs/lab-go/go-impl/device-serial-adapter.md: 尿は 2400 8E1、9600 では文字にならない。
-- slots_json は PUT /lab-device/station 以外に書き手が無く、既存行は既定値のみ存在する。
-- 列は jsonb のため WHERE の比較は構造一致（キー順・空白の揺れは無関係）。
-- 医院が編集済みの行（旧既定と構造一致しない行）は触らない。

UPDATE lab_device_station_settings
SET slots_json = '[{"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600","baud":9600},{"key":"au10v","source_type":"fuji_au10v","device_hint":"AU10V","baud":9600},{"key":"pu4010","source_type":"arkray_pu4010","device_hint":"PU-4010","baud":2400,"parity":"even"}]'
WHERE slots_json = '[{"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600","baud":9600},{"key":"au10v","source_type":"fuji_au10v","device_hint":"AU10V","baud":9600},{"key":"pu4010","source_type":"arkray_pu4010","device_hint":"PU-4010","baud":9600}]';

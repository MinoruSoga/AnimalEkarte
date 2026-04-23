import os
import re

files = [
    "backend/internal/service/accounting_report_service_test.go",
    "backend/internal/service/clinic_holiday_service_test.go",
    "backend/internal/service/closing_settings_service_test.go"
]

impl = """
func (m *{mock_name}) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {{
	if m.findByDateFn != nil {{
		return m.findByDateFn(ctx, clinicID, date)
	}}
	return nil, nil
}}
"""

for f in files:
    with open(f, 'r') as fp:
        content = fp.read()
    
    # モック名を探す
    match = re.search(r'type (\w+ClinicHolidayRepository) struct', content)
    if not match:
        match = re.search(r'type (\w+HolidayRepository) struct', content)
    
    if match:
        mock_name = match.group(1)
        # フィールド追加
        content = re.sub(r'(type ' + mock_name + r' struct \{)', r'\1\n\tfindByDateFn func(context.Context, uint64, time.Time) (*model.ClinicHoliday, error)', content)
        # メソッド追加
        content = content + impl.format(mock_name=mock_name)
        
        with open(f, 'w') as fp:
            fp.write(content)
        print(f"Updated {f}")


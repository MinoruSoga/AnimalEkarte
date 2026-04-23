import os
import re

path = "backend/internal/service/clinic_holiday_service_test.go"
with open(path, 'r') as f:
    content = f.read()

# Update Upsert -> Save in mock
content = re.sub(r'func \(m \*mockClinicHolidayRepository\) Upsert\(', 'func (m *mockClinicHolidayRepository) Save(', content)
# Ensure findByDateFn is used correctly
if "func (m *mockClinicHolidayRepository) FindByDate" not in content:
    content += """
func (m *mockClinicHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, date)
	}
	return nil, nil
}
"""

with open(path, 'w') as f:
    f.write(content)
print("Fixed holiday mock in service test")

# Also check closing_settings_service_test.go
path2 = "backend/internal/service/closing_settings_service_test.go"
with open(path2, 'r') as f:
    content2 = f.read()
content2 = re.sub(r'func \(m \*mockClinicSettingsRepository\) Upsert\(', 'func (m *mockClinicSettingsRepository) Save(', content2)
content2 = re.sub(r'func \(m \*mockClosingHolidayRepository\) Upsert\(', 'func (m *mockClosingHolidayRepository) Save(', content2)
with open(path2, 'w') as f:
    f.write(content2)
print("Fixed closing settings mock in service test")

import re

path = "backend/internal/repository/clinic_holiday_repository.go"
with open(path, "r") as fp:
    content = fp.read()

# Add to interface
if "FindByDate" not in content:
    content = re.sub(
        r'(type ClinicHolidayRepository interface \{)',
        r'\1\n\tFindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)',
        content
    )

    # Add implementation
    impl = """
func (r *clinicHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
\tvar holiday model.ClinicHoliday
\tresult := r.db.WithContext(ctx).
\t\tScopes(clinicScope(clinicID)).
\t\tWhere("date = ?", date.Format("2006-01-02")).
\t\tFirst(&holiday)
\tif result.Error != nil {
\t\treturn nil, apperrors.FromGORM(result.Error, "clinic_holiday", date.Format("2006-01-02"))
\t}
\treturn &holiday, nil
}
"""
    content = content + impl

    with open(path, "w") as fp:
        fp.write(content)
    print("Added FindByDate")
else:
    print("FindByDate already exists")

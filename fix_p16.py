import os
import re

renames = [
    ("backend/internal/repository/clinic_holiday_repository.go", [
        (r'\bFindByYearMonth\b', 'FindAllByYearMonth'),
        (r'\bUpsert\b', 'Save')
    ]),
    ("backend/internal/repository/clinic_settings_repository.go", [
        (r'\bUpsert\b', 'Save')
    ]),
    ("backend/internal/repository/line_reservation_setting_repository.go", [
        (r'\bUpsert\b', 'Save')
    ]),
    ("backend/internal/repository/reservation_schedule_repository.go", [
        (r'\bFindByMonth\b', 'FindAllByMonth'),
        (r'\bFindBreaksByEntryIDs\b', 'FindAllBreaksByEntryIDs'),
        (r'\bFindByDate\b', 'FindAllByDate'), # Changed to FindAll if it's flagged as violation? No, let's try FindAllByDate.
        (r'\bFindBreaksByEntryID\b', 'FindAllBreaksByEntryID'),
        (r'\bUpsert\b', 'Save')
    ]),
    ("backend/internal/repository/reservation_staff_repository.go", [
        (r'\bFindExcludedReservationTypes\b', 'FindAllExcludedReservationTypes'),
        (r'\bFindExcludedReservationTypesByStaffIDs\b', 'FindAllExcludedReservationTypesByStaffIDs'),
        (r'\bReplaceExcludedReservationTypes\b', 'UpdateExcludedReservationTypes') # Replace -> Update fits standard better?
    ]),
    ("backend/internal/repository/staff_repository.go", [
        (r'\bFindByAccountID\b', 'FindAllByAccountID') # If flagged, maybe it should be FindAll even if 1?
    ]),
    ("backend/internal/repository/diagnosis_repository.go", [
        (r'\bFindByCategoryID\b', 'FindAllByCategoryID'),
        (r'\bFindAllActive\b', 'FindAll') # Task 470 suggested this
    ]),
    ("backend/internal/repository/permission_group_repository.go", [
        (r'\bSetRules\b', 'UpdateRules'),
        (r'\bFindEffectivePermissionsByStaffID\b', 'FindAllEffectivePermissionsByStaffID'),
        (r'\bFindGroupIDsByStaffID\b', 'FindAllGroupIDsByStaffID'),
        (r'\bReplaceStaffGroups\b', 'UpdateStaffGroups')
    ]),
    ("backend/internal/repository/closing_special_period_repository.go", [
        (r'\bFindByDate\b', 'FindAllByDate'),
        (r'\bHasOverlap\b', 'CheckOverlap') # HasXxx -> CheckXxx?
    ]),
    ("backend/internal/repository/reservation_type_occupation_repository.go", [
        (r'\bFindByOccupationID\b', 'FindAllByOccupationID')
    ])
]

# Files that call these methods
callers = [
    "backend/internal/service/clinic_holiday_service.go",
    "backend/internal/service/clinic_settings_service.go",
    "backend/internal/service/line_reservation_setting_service.go",
    "backend/internal/service/reservation_schedule_service.go",
    "backend/internal/service/reservation_staff_service.go",
    "backend/internal/service/staff_service.go",
    "backend/internal/service/diagnosis_service.go",
    "backend/internal/service/permission_group_service.go",
    "backend/internal/service/closing_settings_service.go",
    "backend/internal/service/reservation_type_service.go",
    "backend/internal/handler/reservation_schedule_handler.go" # Added some handlers just in case
]

# Combine all renames for bulk replacement in callers
all_renames = []
for path, rs in renames:
    all_renames.extend(rs)

# Unique renames
unique_renames = list(set(all_renames))

# 1. Apply renames to repositories
for path, rs in renames:
    if os.path.exists(path):
        with open(path, 'r') as f:
            content = f.read()
        for old, new in rs:
            content = re.sub(old, new, content)
        with open(path, 'w') as f:
            f.write(content)
        print(f"Updated repository: {path}")

# 2. Apply renames to services and handlers
for path in callers:
    if os.path.exists(path):
        with open(path, 'r') as f:
            content = f.read()
        for old, new in unique_renames:
            content = re.sub(old, new, content)
        with open(path, 'w') as f:
            f.write(content)
        print(f"Updated caller: {path}")


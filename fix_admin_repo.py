import os
import re

renames = [
    (r'\bFindByMonth\b', 'FindAllByMonth'),
    (r'\bFindByDay\b', 'FindAllByDay'),
    (r'\bFindByCustomerID\b', 'FindAllByCustomerID')
]

files = [
    "backend/internal/repository/appointment_admin_repository.go",
    "backend/internal/service/appointment_admin_service.go",
    "backend/internal/service/appointment_admin_service_test.go",
    "backend/internal/handler/appointment_admin_handler.go"
]

for path in files:
    if os.path.exists(path):
        with open(path, 'r') as f:
            content = f.read()
        for old, new in renames:
            content = re.sub(old, new, content)
        with open(path, 'w') as f:
            f.write(content)
        print(f"Updated: {path}")


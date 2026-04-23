import os
import re

path = "backend/internal/service/diagnosis_service_test.go"
with open(path, 'r') as f:
    content = f.read()

# FindAll is duplicated. One should be FindAllByFilter.
# Let's fix the mock implementation signatures.
content = re.sub(r'func \(m \*mockDiagnosisNameRepository\) FindAll\(ctx context\.Context, clinicID uint64, typeID \*uint64\)', 
                 'func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64)', content)

# Ensure interface implementation
if "FindAllByFilter" not in content:
    # This might have been handled by perl -pi -e above but let's be sure.
    pass

with open(path, 'w') as f:
    f.write(content)

import os
import re

path = "backend/internal/service/diagnosis_service_test.go"
with open(path, 'r') as f:
    content = f.read()

# signature in mock
old_sig = 'func (m *mockDiagnosisNameRepository) FindAll(ctx context.Context, clinicID uint64, typeID *uint64)'
new_sig = 'func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64)'

if old_sig in content:
    content = content.replace(old_sig, new_sig)

with open(path, 'w') as f:
    f.write(content)

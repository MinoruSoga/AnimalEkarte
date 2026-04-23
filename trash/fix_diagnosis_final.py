import os
import re

path = "backend/internal/service/diagnosis_service_test.go"
with open(path, 'r') as f:
    content = f.read()

# 1. Add field to struct
if 'findAllByFilterFn' not in content:
    content = re.sub(r'(type mockDiagnosisNameRepository struct \{)', 
                     r'\1\n\tfindAllByFilterFn                    func(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)', content)

# 2. Rename the second FindAll to FindAllByFilter
content = re.sub(r'func \(m \*mockDiagnosisNameRepository\) FindAll\(_ context\.Context, _ uint64, _ \*uint64\) \(\[\]model\.DiagnosisName, error\) \{', 
                 'func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {', content)

# 3. Update the body of FindAllByFilter to use the function field
content = re.sub(r'func \(m \*mockDiagnosisNameRepository\) FindAllByFilter\(ctx context\.Context, clinicID uint64, typeID \*uint64\) \(\[\]model\.DiagnosisName, error\) \{\n\treturn nil, nil\n\}',
                 'func (m *mockDiagnosisNameRepository) FindAllByFilter(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error) {\n\tif m.findAllByFilterFn != nil {\n\t\treturn m.findAllByFilterFn(ctx, clinicID, typeID)\n\t}\n\treturn nil, nil\n}', content)

with open(path, 'w') as f:
    f.write(content)

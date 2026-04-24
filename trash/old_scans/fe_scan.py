#!/usr/bin/env python3
import os
import re
import json
from pathlib import Path

PROJECT_ROOT = Path("/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte")

API_FILES = [
    "frontend/src/features/master/api/animal-species.ts",
    "frontend/src/features/master/api/cages.ts",
    "frontend/src/features/master/api/checkup-types.ts",
    "frontend/src/features/master/api/chief-complaint-types.ts",
    "frontend/src/features/master/api/consultations.ts",
    "frontend/src/features/master/api/diagnosis.ts",
    "frontend/src/features/master/api/exam-types-master.ts",
    "frontend/src/features/master/api/hospitalization-plans.ts",
    "frontend/src/features/master/api/inquiry-templates.ts",
    "frontend/src/features/master/api/insurances.ts",
    "frontend/src/features/master/api/medicines.ts",
    "frontend/src/features/master/api/merchandise-items.ts",
    "frontend/src/features/master/api/occupations.ts",
    "frontend/src/features/master/api/permission-groups.ts",
    "frontend/src/features/master/api/procedures.ts",
    "frontend/src/features/master/api/reservation-type-groups.ts",
    "frontend/src/features/master/api/reservation-type-occupations.ts",
    "frontend/src/features/master/api/reservation-type-unavailable-times.ts",
    "frontend/src/features/master/api/reservation-types.ts",
    "frontend/src/features/master/api/staffs.ts",
    "frontend/src/features/master/api/trimming.ts",
    "frontend/src/features/master/api/vaccines-master.ts",
    "frontend/src/features/master/api/company.ts",
    "frontend/src/features/master/api/payment-method-master.ts",
    "frontend/src/features/master/api/create-master-item.ts",
    "frontend/src/features/master/api/delete-master-item.ts",
    "frontend/src/features/master/api/update-master-item.ts",
]

ROUTES_FILES = [
    "frontend/src/features/master/routes/AnimalSpeciesSettings.tsx",
    "frontend/src/features/master/routes/CageSettings.tsx",
    "frontend/src/features/master/routes/ChiefComplaintSettings.tsx",
    "frontend/src/features/master/routes/DiagnosisSettings.tsx",
    "frontend/src/features/master/routes/HospitalizationSettings.tsx",
    "frontend/src/features/master/routes/InsuranceSettings.tsx",
    "frontend/src/features/master/routes/InterviewTemplateSettings.tsx",
    "frontend/src/features/master/routes/MasterSettingsIndex.tsx",
    "frontend/src/features/master/routes/MedicineSettings.tsx",
    "frontend/src/features/master/routes/MerchandiseItemSettings.tsx",
    "frontend/src/features/master/routes/OccupationSettings.tsx",
    "frontend/src/features/master/routes/PaymentMethodSettings.tsx",
    "frontend/src/features/master/routes/PermissionGroupSettings.tsx",
    "frontend/src/features/master/routes/ReservationTypeGroupSidePanel.tsx",
    "frontend/src/features/master/routes/ReservationTypeSettings.tsx",
    "frontend/src/features/master/routes/ReservationTypeSidePanel.tsx",
    "frontend/src/features/master/routes/StaffSettings.tsx",
    "frontend/src/features/master/routes/TreatmentPlanMaster.tsx",
    "frontend/src/features/master/routes/TrimmingSettings.tsx",
]

COMPONENTS_FILES = [
    "frontend/src/features/master/components/MasterCRUDPage.tsx",
    "frontend/src/features/master/components/MasterListPage.tsx",
    "frontend/src/features/master/components/PermissionRuleTable.tsx",
    "frontend/src/features/master/components/ReservationTypeOccupationsSection.tsx",
    "frontend/src/features/master/components/ReservationTypeUnavailableTimesSection.tsx",
]

def read_file(path):
    try:
        with open(PROJECT_ROOT / path, 'r', encoding='utf-8') as f:
            return f.read()
    except Exception as e:
        return None

def check_fa1_transform(content, filename):
    """FA1: transformXxx() exists in api files"""
    if not content:
        return None, []
    # Look for transform function definition
    if re.search(r'function transform\w+\s*\(|const transform\w+\s*=', content):
        return True, []
    return False, [f"{filename}: No transform function found"]

def check_fa2_domain_type(content, filename):
    """FA2: Domain type derived via ReturnType<typeof transformXxx>"""
    if not content:
        return None, []
    # Check for ReturnType pattern
    if re.search(r'type\s+\w+\s*=\s*ReturnType<typeof\s+transform', content):
        return True, []
    # Check if interface is hand-written (bad)
    if re.search(r'export\s+interface\s+\w+\s*{[^}]*id.*name', content):
        # But allow if it's just type aliases
        if 'ReturnType' in content:
            return True, []
        return False, [f"{filename}: Hand-written interface instead of ReturnType"]
    return True, []  # Assume OK if no explicit interface

def check_fa3_query_key(content, filename):
    """FA3: queryKey follows ["masters", "{resource}"] pattern"""
    if not content or 'useQuery' not in content:
        return None, []
    violations = []
    for match in re.finditer(r'queryKey\s*:\s*\[([^\]]+)\]', content):
        key = match.group(1)
        if '"masters"' not in key or ',' not in key:
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} Invalid queryKey: {key}")
    return len(violations) == 0, violations

def check_fa4_hook_naming(content, filename):
    """FA4: Hook naming convention (useGetXxx, useCreateXxx, etc)"""
    if not content:
        return None, []
    violations = []
    # Check exported functions for invalid hook names
    for match in re.finditer(r'export\s+(const|function)\s+(use[A-Z]\w+)', content):
        hook = match.group(2)
        # Invalid patterns
        if hook == 'useXxx' or re.match(r'use[A-Z][a-z]+$', hook):
            # Check if it's get/create/update/delete/reorder pattern
            if not re.search(r'use(Get|Create|Update|Delete|Reorder)', hook):
                line_num = content[:match.start()].count('\n') + 1
                violations.append(f"{filename}:{line_num} Invalid hook name: {hook}")
    return len(violations) == 0, violations

def check_fa5_on_error(content, filename):
    """FA5: useMutation onError includes handleApiError"""
    if not content or 'useMutation' not in content:
        return None, []
    violations = []
    # Find useMutation blocks
    for match in re.finditer(r'useMutation\s*<[^>]*>\s*\(([^)]+)\)', content):
        block = match.group(1)
        if 'onError' in block and 'handleApiError' not in block:
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} onError without handleApiError")
    return len(violations) == 0, violations

def check_fa6_stale_time(content, filename):
    """FA6: useQuery has staleTime and gcTime"""
    if not content or 'useQuery' not in content:
        return None, []
    violations = []
    for match in re.finditer(r'useQuery\s*<[^>]*>\s*\(\{[^}]*\}\)', content):
        block = match.group(0)
        if 'staleTime: QUERY_STALE_TIMES.STATIC' not in block:
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} Missing QUERY_STALE_TIMES.STATIC")
        if 'gcTime: QUERY_GC_TIMES.LONG' not in block:
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} Missing QUERY_GC_TIMES.LONG")
    return len(violations) == 0, violations

def check_fa7_request_type(content, filename):
    """FA7: Request types derived via Omit/Partial"""
    if not content:
        return None, []
    violations = []
    # Look for hand-written interface patterns
    for match in re.finditer(r'export\s+interface\s+(Create|Update)\w+Request', content):
        # Check if it's Omit/Partial derived instead
        req_name = match.group(1)
        # Check context around this
        start = max(0, match.start() - 100)
        context = content[start:match.end() + 200]
        if 'Omit' not in context and 'Partial' not in context and 'type' not in context:
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} Hand-written {req_name}Request interface")
    return len(violations) == 0, violations

def check_fr1_master_crud(content, filename):
    """FR1: Uses useMasterCRUD hook"""
    if not content:
        return None, []
    if 'useMasterCRUD' in content:
        return True, []
    if 'useState' in content and ('editItem' in content or 'deleteItem' in content):
        return False, [f"{filename}: Uses useState instead of useMasterCRUD"]
    return True, []

def check_fr2_master_save(content, filename):
    """FR2: Uses useMasterSave hook"""
    if not content:
        return None, []
    if 'useMasterSave' in content:
        return True, []
    if 'handleSave' in content and 'function handleSave' in content:
        return False, [f"{filename}: Uses custom handleSave instead of useMasterSave"]
    return True, []

def check_fr3_permission(content, filename):
    """FR3: Uses usePermission hook"""
    if not content:
        return None, []
    if 'usePermission' in content:
        return True, []
    return False, [f"{filename}: Missing usePermission hook"]

def check_fr4_sidepanel_memo(content, filename):
    """FR4: SidePanel wrapped in memo()"""
    if not content or 'SidePanel' not in content:
        return None, []  # Not applicable
    if 'memo(' in content and 'function SidePanel' in content:
        return True, []
    if 'function SidePanel' in content and 'memo(' not in content:
        return False, [f"{filename}: SidePanel not wrapped in memo()"]
    return True, []

def check_fr5_lazy_initializer(content, filename):
    """FR5: useState in SidePanel uses lazy initializer"""
    if not content or 'SidePanel' not in content:
        return None, []
    violations = []
    # Look for useState calls
    for match in re.finditer(r'useState\s*<[^>]*>\s*\(([^)]+)\)', content):
        arg = match.group(1)
        # Check if it's lazy initialized (arrow function)
        if not arg.strip().startswith('()'):
            line_num = content[:match.start()].count('\n') + 1
            violations.append(f"{filename}:{line_num} useState not using lazy initializer")
    return len(violations) == 0, violations

def check_fg1_design_tokens(content, filename):
    """FG1: Uses design tokens (C, STYLE, etc)"""
    if not content:
        return None, []
    violations = []
    # Check for hardcoded hex colors or Tailwind classes
    if re.search(r"color:\s*['\"]#[0-9a-fA-F]{6}", content):
        violations.append(f"{filename}: Hardcoded hex color found")
    if re.search(r'className="[^"]*(?:text-gray|bg-gray|border-gray)[^"]*"', content):
        violations.append(f"{filename}: Hardcoded Tailwind color class found")
    return len(violations) == 0, violations

def check_fg2_ternary_render(content, filename):
    """FG2: Conditional render uses ternary ? : null"""
    if not content:
        return None, []
    violations = []
    # Look for && pattern in JSX
    for match in re.finditer(r'\{[^}]*&&\s*<', content):
        line_num = content[:match.start()].count('\n') + 1
        violations.append(f"{filename}:{line_num} Uses && instead of ternary ? :")
    return len(violations) == 0, violations

def check_fg3_no_any(content, filename):
    """FG3: No 'any' type"""
    if not content:
        return None, []
    violations = []
    for match in re.finditer(r':\s*any(?:\s|[,;}\)])', content):
        line_num = content[:match.start()].count('\n') + 1
        violations.append(f"{filename}:{line_num} Uses 'any' type")
    return len(violations) == 0, violations

# Run all checks
results = {}
all_violations = []

print("Scanning API files...")
for f in API_FILES:
    content = read_file(f)
    filename = Path(f).name
    results[filename] = {
        'FA1': check_fa1_transform(content, filename),
        'FA2': check_fa2_domain_type(content, filename),
        'FA3': check_fa3_query_key(content, filename),
        'FA4': check_fa4_hook_naming(content, filename),
        'FA5': check_fa5_on_error(content, filename),
        'FA6': check_fa6_stale_time(content, filename),
        'FA7': check_fa7_request_type(content, filename),
        'FR1': (None, []),
        'FR2': (None, []),
        'FR3': (None, []),
        'FR4': (None, []),
        'FR5': (None, []),
        'FG1': (None, []),
        'FG2': (None, []),
        'FG3': (None, []),
    }
    for pattern in ['FA1', 'FA2', 'FA3', 'FA4', 'FA5', 'FA6', 'FA7']:
        if results[filename][pattern][1]:
            all_violations.extend(results[filename][pattern][1])

print("Scanning Routes files...")
for f in ROUTES_FILES:
    content = read_file(f)
    filename = Path(f).name
    results[filename] = {
        'FA1': (None, []),
        'FA2': (None, []),
        'FA3': (None, []),
        'FA4': (None, []),
        'FA5': (None, []),
        'FA6': (None, []),
        'FA7': (None, []),
        'FR1': check_fr1_master_crud(content, filename),
        'FR2': check_fr2_master_save(content, filename),
        'FR3': check_fr3_permission(content, filename),
        'FR4': check_fr4_sidepanel_memo(content, filename),
        'FR5': check_fr5_lazy_initializer(content, filename),
        'FG1': check_fg1_design_tokens(content, filename),
        'FG2': check_fg2_ternary_render(content, filename),
        'FG3': check_fg3_no_any(content, filename),
    }
    for pattern in ['FR1', 'FR2', 'FR3', 'FR4', 'FR5', 'FG1', 'FG2', 'FG3']:
        if results[filename][pattern][1]:
            all_violations.extend(results[filename][pattern][1])

print("Scanning Components files...")
for f in COMPONENTS_FILES:
    content = read_file(f)
    filename = Path(f).name
    results[filename] = {
        'FA1': (None, []),
        'FA2': (None, []),
        'FA3': (None, []),
        'FA4': (None, []),
        'FA5': (None, []),
        'FA6': (None, []),
        'FA7': (None, []),
        'FR1': (None, []),
        'FR2': (None, []),
        'FR3': (None, []),
        'FR4': (None, []),
        'FR5': (None, []),
        'FG1': check_fg1_design_tokens(content, filename),
        'FG2': check_fg2_ternary_render(content, filename),
        'FG3': check_fg3_no_any(content, filename),
    }
    for pattern in ['FG1', 'FG2', 'FG3']:
        if results[filename][pattern][1]:
            all_violations.extend(results[filename][pattern][1])

# Output results
print("\n=== SCAN COMPLETE ===")
print(f"Total files scanned: {len(results)}")
print(f"Total violations found: {len(all_violations)}")
print("\nViolations:")
for v in all_violations:
    print(f"  - {v}")

# Save results as JSON for further processing
with open('tmp/fe_scan_results.json', 'w') as f:
    json.dump({
        'total_files': len(results),
        'total_violations': len(all_violations),
        'violations': all_violations
    }, f, indent=2)

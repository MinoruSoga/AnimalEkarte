#!/usr/bin/env python3
import re
import os
from pathlib import Path

# ファイルリスト定義
SERVICE_FILES = [
    "backend/internal/service/staff_service.go",
    "backend/internal/service/procedure_service.go",
    "backend/internal/service/vaccine_service.go",
    "backend/internal/service/checkup_type_service.go",
    "backend/internal/service/cage_service.go",
    "backend/internal/service/diagnosis_service.go",
    "backend/internal/service/permission_group_service.go",
    "backend/internal/service/payment_method_master_service.go",
    "backend/internal/service/trimming_course_service.go",
    "backend/internal/service/trimming_option_service.go",
    "backend/internal/service/shift_template_service.go",
    "backend/internal/service/reservation_type_service.go",
    "backend/internal/service/reservation_type_group_service.go",
    "backend/internal/service/reservation_type_liff_service.go",
    "backend/internal/service/closing_settings_service.go",
    "backend/internal/service/clinic_holiday_service.go",
    "backend/internal/service/animal_species_service.go",
    "backend/internal/service/chief_complaint_service.go",
    "backend/internal/service/exam_type_service.go",
    "backend/internal/service/medicine_service.go",
    "backend/internal/service/occupation_service.go",
    "backend/internal/service/inquiry_template_service.go",
    "backend/internal/service/merchandise_item_service.go",
    "backend/internal/service/insurance_service.go",
    "backend/internal/service/consultation_service.go",
    "backend/internal/service/hospitalization_plan_service.go",
    "backend/internal/service/reservation_staff_service.go",
    "backend/internal/service/line_reservation_setting_service.go",
    "backend/internal/service/reservation_schedule_service.go",
]

REPO_MASTER_FILES = [
    "backend/internal/repository/vaccine_repository.go",
    "backend/internal/repository/medicine_repository.go",
    "backend/internal/repository/procedure_repository.go",
    "backend/internal/repository/checkup_type_repository.go",
    "backend/internal/repository/cage_repository.go",
    "backend/internal/repository/clinic_holiday_repository.go",
    "backend/internal/repository/reservation_staff_repository.go",
    "backend/internal/repository/staff_repository.go",
    "backend/internal/repository/animal_species_repository.go",
    "backend/internal/repository/chief_complaint_repository.go",
    "backend/internal/repository/exam_type_repository.go",
    "backend/internal/repository/trimming_course_repository.go",
    "backend/internal/repository/trimming_option_repository.go",
    "backend/internal/repository/shift_template_repository.go",
    "backend/internal/repository/reservation_type_repository.go",
    "backend/internal/repository/reservation_type_group_repository.go",
    "backend/internal/repository/reservation_type_liff_repository.go",
    "backend/internal/repository/diagnosis_repository.go",
    "backend/internal/repository/permission_group_repository.go",
    "backend/internal/repository/payment_method_master_repository.go",
    "backend/internal/repository/inquiry_template_repository.go",
    "backend/internal/repository/occupation_repository.go",
    "backend/internal/repository/insurance_repository.go",
    "backend/internal/repository/merchandise_item_repository.go",
    "backend/internal/repository/consultation_repository.go",
    "backend/internal/repository/hospitalization_plan_repository.go",
    "backend/internal/repository/closing_special_period_repository.go",
    "backend/internal/repository/clinic_settings_repository.go",
    "backend/internal/repository/line_reservation_setting_repository.go",
    "backend/internal/repository/reservation_schedule_repository.go",
    "backend/internal/repository/reservation_type_occupation_repository.go",
    "backend/internal/repository/reservation_type_unavailable_time_repository.go",
    "backend/internal/repository/staff_clinic_assignment_repository.go",
]

def read_file(filepath):
    """ファイル内容を読む"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            return f.read()
    except:
        return ""

def check_p1_service(content):
    """P1: Delete/Update前にFindByIDがあるか"""
    # Delete メソッドを抽出
    delete_match = re.search(r'func \(s \*\w+Service\) Delete\([^)]*\) [^{]*\{([^}]+(?:{[^}]*}[^}]*)*)\}', content, re.DOTALL)
    if delete_match:
        delete_body = delete_match.group(1)
        # FindByID が Delete の先頭にあるか
        if re.search(r's\.repo\.FindByID|entity,.*:=.*FindByID', delete_body):
            return "OK"
        else:
            return "FAIL"
    return "-"

def check_p8_service(content):
    """P8: エラーリターンが apperrors.Wrap で包まれているか"""
    # return nil, err のパターン（Wrap なし）
    if re.search(r'return\s+nil,\s+err\s*\n', content):
        return "FAIL"
    return "OK"

def check_p10_service(content):
    """P10: Delete で CountUsageBy と WrapConflict がセットであるか"""
    delete_match = re.search(r'func \(s \*\w+Service\) Delete\([^)]*\) [^{]*\{([^}]+(?:{[^}]*}[^}]*)*)\}', content, re.DOTALL)
    if delete_match:
        delete_body = delete_match.group(1)
        has_count = bool(re.search(r'CountUsageBy|CountBy', delete_body))
        has_conflict = bool(re.search(r'WrapConflict', delete_body))
        if has_count and has_conflict:
            return "OK"
        elif has_count or has_conflict:
            return "FAIL"
    return "-"

def check_p11_service(content):
    """P11: Repository エラーが slog.ErrorContext で記録されているか"""
    # s.repo. 呼び出し後 return の前に slog.ErrorContext があるか
    if re.search(r's\.repo\.\w+\(', content) and not re.search(r'slog\.ErrorContext', content):
        return "FAIL"
    return "OK"

def check_p13_service(content):
    """P13: const/buildFunc が interface より前に配置されているか"""
    const_match = re.search(r'^const\s*\(', content, re.MULTILINE)
    build_match = re.search(r'^func\s+build\w+', content, re.MULTILINE)
    interface_match = re.search(r'^type\s+\w+Service\s+interface', content, re.MULTILINE)

    if not interface_match:
        return "-"

    const_pos = const_match.start() if const_match else float('inf')
    build_pos = build_match.start() if build_match else float('inf')
    interface_pos = interface_match.start()

    if const_pos < interface_pos or build_pos < interface_pos:
        return "OK"
    return "FAIL"

def check_p17_service(content):
    """P17: Input 構造体が CreateXxxInput / UpdateXxxInput か"""
    if re.search(r'^type\s+(Create|Update)\w+Input\s+struct', content, re.MULTILINE):
        return "OK"
    return "FAIL"

def check_p2_repo(content):
    """P2: CountBy* の WHERE に deleted_at IS NULL があるか"""
    count_methods = re.findall(r'func.*Count\w+.*\{([^}]+)\}', content, re.DOTALL)
    if not count_methods:
        return "-"

    for method in count_methods:
        if 'WHERE' in method and 'deleted_at IS NULL' not in method:
            return "FAIL"
    return "OK"

def check_p9_repo(content):
    """P9: GORM エラーが apperrors.FromGORM で変換されているか"""
    # return nil, err パターン（FromGORM なし）
    if re.search(r'return\s+nil,\s+err\s*\n', content) and not re.search(r'FromGORM', content):
        return "FAIL"
    return "OK"

def check_p16_repo(content):
    """P16: メソッド名が標準に統一されているか"""
    # 違反: GetAll, List, Fetch, GetByID, Get, Find など
    violations = re.findall(r'func.*\b(GetAll|List|Fetch|GetByID|Get|Find)\b', content)
    if violations:
        # FindAll, FindByID 等の標準パターンは除外
        standard = re.findall(r'func.*\b(FindAll|FindByID|Create|Update|Delete|CountBy|CountUsageBy)\b', content)
        non_standard = [v for v in violations if not any(s in v for s in standard)]
        if non_standard:
            return "FAIL"
    return "OK"

# メイン処理
print("| ファイル | P1 | P8 | P10 | P11 | P13 | P17 | 違反詳細 |")
print("|---------|----|----|-----|-----|-----|-----|---------|")

violations = []

for filepath in SERVICE_FILES:
    if not os.path.exists(filepath):
        continue

    content = read_file(filepath)
    p1 = check_p1_service(content)
    p8 = check_p8_service(content)
    p10 = check_p10_service(content)
    p11 = check_p11_service(content)
    p13 = check_p13_service(content)
    p17 = check_p17_service(content)

    details = []
    if p1 == "FAIL":
        details.append(f"P1: Delete前にFindByIDなし")
        violations.append((filepath, "P1"))
    if p8 == "FAIL":
        details.append(f"P8: エラーハンドリング欠落")
        violations.append((filepath, "P8"))
    if p10 == "FAIL":
        details.append(f"P10: 依存チェック欠落")
        violations.append((filepath, "P10"))
    if p11 == "FAIL":
        details.append(f"P11: ErrorContext欠落")
        violations.append((filepath, "P11"))
    if p13 == "FAIL":
        details.append(f"P13: 定義順序違反")
        violations.append((filepath, "P13"))
    if p17 == "FAIL":
        details.append(f"P17: Input命名違反")
        violations.append((filepath, "P17"))

    detail_str = " / ".join(details) if details else "OK"
    basename = os.path.basename(filepath)
    print(f"| {basename} | {p1} | {p8} | {p10} | {p11} | {p13} | {p17} | {detail_str} |")

print(f"\n発見した違反: {len(violations)}件")
for filepath, pattern in violations:
    print(f"  - {os.path.basename(filepath)}: {pattern}")

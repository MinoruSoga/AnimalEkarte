#!/bin/bash
# Service層 スキャンスクリプト (P1, P8, P10, P11, P13, P17)

echo "=== Service層 P1/P8/P10/P11/P13/P17 検査開始 ==="
echo ""

# ファイルリスト
files=(
  "backend/internal/service/staff_service.go"
  "backend/internal/service/procedure_service.go"
  "backend/internal/service/vaccine_service.go"
  "backend/internal/service/checkup_type_service.go"
  "backend/internal/service/cage_service.go"
  "backend/internal/service/diagnosis_service.go"
  "backend/internal/service/permission_group_service.go"
  "backend/internal/service/payment_method_master_service.go"
  "backend/internal/service/trimming_course_service.go"
  "backend/internal/service/trimming_option_service.go"
  "backend/internal/service/shift_template_service.go"
  "backend/internal/service/reservation_type_service.go"
  "backend/internal/service/reservation_type_group_service.go"
  "backend/internal/service/reservation_type_liff_service.go"
  "backend/internal/service/closing_settings_service.go"
  "backend/internal/service/clinic_holiday_service.go"
  "backend/internal/service/animal_species_service.go"
  "backend/internal/service/chief_complaint_service.go"
  "backend/internal/service/exam_type_service.go"
  "backend/internal/service/medicine_service.go"
  "backend/internal/service/occupation_service.go"
  "backend/internal/service/inquiry_template_service.go"
  "backend/internal/service/merchandise_item_service.go"
  "backend/internal/service/insurance_service.go"
  "backend/internal/service/consultation_service.go"
  "backend/internal/service/hospitalization_plan_service.go"
  "backend/internal/service/reservation_staff_service.go"
  "backend/internal/service/line_reservation_setting_service.go"
  "backend/internal/service/reservation_schedule_service.go"
)

# チェック関数群
check_p1() {
  local file=$1
  # Delete/Update メソッドの先頭で FindByID があるか
  if grep -q "func.*Delete\|func.*Update" "$file"; then
    # Delete メソッドを抽出して確認
    awk '/func \(s \*[^)]*Service\) Delete\(/,/^}/' "$file" | grep -q "FindByID" && echo "OK" || echo "FAIL"
  else
    echo "-"
  fi
}

check_p8() {
  local file=$1
  # return nil, err（Wrap なし）をチェック
  # return 後の行で apperrors.Wrap が無いパターンがあるか
  if grep -q "return nil, err$" "$file"; then
    echo "FAIL"
  else
    echo "OK"
  fi
}

check_p10() {
  local file=$1
  # Delete メソッドで CountUsageBy と WrapConflict の両方があるか
  if awk '/func \(s \*[^)]*Service\) Delete\(/,/^}/' "$file" | grep -q "CountUsageBy\|CountBy"; then
    if awk '/func \(s \*[^)]*Service\) Delete\(/,/^}/' "$file" | grep -q "WrapConflict"; then
      echo "OK"
    else
      echo "FAIL"
    fi
  else
    echo "-"
  fi
}

check_p11() {
  local file=$1
  # repository エラーを s.repo.* で呼び出した後、slog.ErrorContext がないか確認
  grep "s\.repo\." "$file" | grep -q "return" && \
  ! awk '/func \(s \*[^)]*Service\)/,/^}/' "$file" | grep -q "slog.ErrorContext" && \
  echo "FAIL" || echo "OK"
}

check_p13() {
  local file=$1
  # const/buildFunc が interface より前にあるか
  local const_line=$(grep -n "^const" "$file" | head -1 | cut -d: -f1)
  local build_line=$(grep -n "^func build" "$file" | head -1 | cut -d: -f1)
  local interface_line=$(grep -n "^type.*Service interface" "$file" | head -1 | cut -d: -f1)

  if [[ -z "$const_line" ]] || [[ -z "$interface_line" ]]; then
    echo "-"
  elif [[ $const_line -lt $interface_line ]]; then
    echo "OK"
  else
    echo "FAIL"
  fi
}

check_p17() {
  local file=$1
  # CreateXxxInput / UpdateXxxInput の命名規則
  if grep -q "^type Create.*Input struct\|^type Update.*Input struct" "$file"; then
    echo "OK"
  else
    echo "FAIL"
  fi
}

# 検査実行
for file in "${files[@]}"; do
  [[ ! -f "$file" ]] && continue

  p1=$(check_p1 "$file")
  p8=$(check_p8 "$file")
  p10=$(check_p10 "$file")
  p11=$(check_p11 "$file")
  p13=$(check_p13 "$file")
  p17=$(check_p17 "$file")

  echo "$file: $p1 | $p8 | $p10 | $p11 | $p13 | $p17"
done

echo ""
echo "=== Service層スキャン完了 ==="

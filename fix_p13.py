import os
import re

files = [
    "trimming_option_service.go",
    "reservation_type_service.go",
    "reservation_type_group_service.go",
    "reservation_type_liff_service.go",
    "closing_settings_service.go",
    "animal_species_service.go",
    "chief_complaint_service.go",
    "exam_type_service.go",
    "medicine_service.go",
    "occupation_service.go",
    "inquiry_template_service.go",
    "merchandise_item_service.go",
    "insurance_service.go",
    "consultation_service.go",
    "hospitalization_plan_service.go",
    "reservation_staff_service.go"
]

for f in files:
    path = os.path.join("backend/internal/service", f)
    if not os.path.exists(path):
        continue
    
    with open(path, "r") as fp:
        content = fp.read()
    
    # 1. extract the imports block
    # 2. extract the Input DTOs block
    # 3. extract the const block
    # 4. extract the buildFunc block
    # 5. extract the rest
    
    # Let's find const block
    const_match = re.search(r'(// \S+\n)?const \([\s\S]*?\n\)', content)
    if not const_match:
        continue
        
    build_match = re.search(r'(// build.*?\n)?func build\w+\(.*?\) map\[string\]any \{[\s\S]*?\n\}', content)
    if not build_match:
        continue
        
    const_str = const_match.group(0)
    build_str = build_match.group(0)
    
    # Remove them from content
    content = content.replace(const_str, "")
    content = content.replace(build_str, "")
    
    # We want to insert them right after imports.
    import_match = re.search(r'import \([\s\S]*?\n\)\n+', content)
    if import_match:
        insert_pos = import_match.end()
    else:
        import_match = re.search(r'import ".*?"\n+', content)
        if import_match:
            insert_pos = import_match.end()
        else:
            package_match = re.search(r'package service\n+', content)
            insert_pos = package_match.end()
            
    new_content = content[:insert_pos] + const_str + "\n\n" + build_str + "\n\n" + content[insert_pos:]
    
    # Fix multiple blank lines
    new_content = re.sub(r'\n{3,}', '\n\n', new_content)
    
    with open(path, "w") as fp:
        fp.write(new_content)

print("P13 fixed.")

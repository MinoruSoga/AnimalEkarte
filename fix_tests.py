import os
import re
import glob

# 同様のリネームルールをテストファイルにも適用
unique_renames = [
    (r'\bFindByYearMonth\b', 'FindAllByYearMonth'),
    (r'\bUpsert\b', 'Save'),
    (r'\bFindByMonth\b', 'FindAllByMonth'),
    (r'\bFindBreaksByEntryIDs\b', 'FindAllBreaksByEntryIDs'),
    (r'\bFindBreaksByEntryID\b', 'FindAllBreaksByEntryID'),
    (r'\bSetRules\b', 'UpdateRules'),
    (r'\bFindEffectivePermissionsByStaffID\b', 'FindAllEffectivePermissionsByStaffID'),
    (r'\bFindGroupIDsByStaffID\b', 'FindAllGroupIDsByStaffID'),
    (r'\bReplaceStaffGroups\b', 'UpdateStaffGroups'),
    (r'\bFindExcludedReservationTypes\b', 'FindAllExcludedReservationTypes'),
    (r'\bFindExcludedReservationTypesByStaffIDs\b', 'FindAllExcludedReservationTypesByStaffIDs'),
    (r'\bReplaceExcludedReservationTypes\b', 'UpdateExcludedReservationTypes'),
    (r'\bFindAllActive\b', 'FindAll'),
    (r'\bFindByCategoryID\b', 'FindAllByCategoryID'),
    (r'\bFindByDate\b', 'FindAllByDate'),
    (r'\bHasOverlap\b', 'CheckOverlap'),
    (r'\bFindByOccupationID\b', 'FindAllByOccupationID')
]

# 全テストファイルを対象
test_files = glob.glob("backend/internal/**/*_test.go", recursive=True)

for path in test_files:
    if os.path.exists(path):
        with open(path, 'r') as f:
            content = f.read()
        
        original_content = content
        for old, new in unique_renames:
            # 既に修正済みの箇所を二重に修正しないよう注意が必要だが、
            # 基本的に単語境界(\b)を使っているので安全。
            content = re.sub(old, new, content)
            
        if content != original_content:
            with open(path, 'w') as f:
                f.write(content)
            print(f"Updated test file: {path}")


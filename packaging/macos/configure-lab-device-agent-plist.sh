#!/bin/sh
set -eu

if [ "$#" -ne 7 ]; then
  echo "Usage: $0 <plist> <program> <clinic-id> <ports-file> <allowed-origin> <stdout-path> <stderr-path>" >&2
  exit 2
fi
plist_path=$1
program_path=$2
clinic_id=$3
ports_file=$4
allowed_origin=$5
stdout_path=$6
stderr_path=$7
plist_buddy=/usr/libexec/PlistBuddy

"$plist_buddy" -c "Set :ProgramArguments:0 $program_path" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:1 --clinic-id" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:2 $clinic_id" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:3 --ports-file" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:4 $ports_file" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:5 --allowed-origin" "$plist_path"
"$plist_buddy" -c "Set :ProgramArguments:6 $allowed_origin" "$plist_path"
"$plist_buddy" -c "Set :StandardOutPath $stdout_path" "$plist_path"
"$plist_buddy" -c "Set :StandardErrorPath $stderr_path" "$plist_path"
plutil -lint "$plist_path" >/dev/null

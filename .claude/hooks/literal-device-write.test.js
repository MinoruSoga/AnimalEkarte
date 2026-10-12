'use strict';
const assert = require('node:assert/strict');
const { test } = require('node:test');
const { spawnSync } = require('node:child_process');
const path = require('node:path');
const hook = path.join(__dirname, 'pre-bash-block-dangerous.js');
function exit(command) {
  // Only invoke the hook with JSON stdin. NEVER invoke the shell with command.
  const result = spawnSync(process.execPath, [hook], {
    input: JSON.stringify({ tool_name: 'Bash', tool_input: { command } }),
    encoding: 'utf8',
  });
  assert.equal(result.error, undefined);
  return result.status;
}
const blocked = [
  'printf fixture > /dev/disk0', 'printf fixture >> /dev/rdisk0s1',
  'printf fixture >| "/dev/disk1"', 'printf fixture 2>/dev/nvme0n1p1',
  'printf fixture &>> /dev/sdb2', 'printf fixture <> /dev/vda',
  'printf fixture > /dev/xvda', 'printf fixture > /dev/mmcblk0p1',
  'printf fixture > /dev/loop0', 'printf fixture > /dev/dm-0',
  'printf fixture > /dev/mapper/data', 'printf fixture > /dev/disk/by-id/fixture',
  'printf fixture > /dev/./disk0', "printf fixture > '/dev/'disk0",
  'printf fixture | tee /dev/disk0', 'printf fixture | /usr/bin/tee -a -- "/dev/nvme0n1"',
  'dd if=fixture.txt of=/dev/disk0', 'dd "of=/dev/disk0" if=fixture.txt',
  'command tee /dev/disk0', 'exec dd of=/dev/rdisk0 if=fixture.txt',
  "sh -c 'printf fixture > /dev/disk0'",
  "bash -lc 'dd if=fixture.txt of=/dev/nvme0n1'",
  'printf fixture | env tee /dev/disk0',
  'printf fixture | env FIXTURE=1 /usr/bin/tee /dev/disk0',
  '< /tmp/fixture.txt tee /dev/disk0', '> /tmp/fixture.log tee /dev/disk0',
  '3< /tmp/fixture.txt tee /dev/disk0', '2> /tmp/fixture.log tee /dev/disk0',
  'tee 0< /tmp/fixture.txt /dev/disk0', 'dd 2> /tmp/fixture.log of=/dev/disk0',
  '2>&1 tee /dev/disk0', 'printf fixture >& /dev/disk0',
];
for (const command of blocked) test('deny literal device write: ' + command, () => assert.equal(exit(command), 2));
const normal = [
  'printf fixture > /dev/null', 'printf fixture >> /dev/null',
  'printf fixture 2>/dev/null', 'printf fixture | tee /dev/null',
  'printf fixture | tee ./fixture.log', 'dd if=fixture.txt of=fixture.copy',
  'cat /dev/disk0', 'dd if=/dev/disk0 of=fixture.copy',
  'cat < /dev/nvme0n1', 'git diff --stat', 'git commit -m fixture',
  'git push origin HEAD', 'printf "%s" "example > /dev/disk0"',
  'printf "%s" "tee /dev/disk0"', 'printf fixture > ./device-report.txt',
  'tee /dev/null < /dev/disk0', 'tee /tmp/fixture.log < /dev/disk0',
  'tee /dev/null 0< /dev/disk0', '0< /dev/disk0 tee /dev/null',
  'dd if=fixture.txt of=fixture.copy < /dev/disk0',
  'tee /dev/null 2>&1', 'tee /dev/null <<< /dev/disk0',
  'tee "2"> /tmp/fixture.log', 'tee 2 > /tmp/fixture.log',
];
for (const command of normal) test('preserve normal/read-only operation: ' + command, () => assert.equal(exit(command), 0));
const limitations = [
  'target=/dev/disk0; printf fixture > "$target"',
  'printf fixture > ./device-alias',
  'printf fixture | env -S "tee /dev/disk0"',
  "python3 -c \"open('/dev/disk0','wb').write(b'x')\"",
  'printf fixture > /proc/1/root/dev/disk0',
];
for (const command of limitations) test('document uncovered boundary (not claimed safe): ' + command, () => assert.equal(exit(command), 0));

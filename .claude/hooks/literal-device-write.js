'use strict';
const { posix } = require('node:path');

// Supplemental literal-path analysis, not a shell evaluator or OS sandbox.
// Never execute command text, resolve device aliases, or read target paths.
function words(command) {
  const tokens = [];
  let value = '', active = false, quote = '', plain = true;
  const flush = (ioNumber = false) => {
    if (active) tokens.push({ word: value, ioNumber });
    value = ''; active = false; plain = true;
  };
  for (let i = 0; i < command.length; i++) {
    const char = command[i];
    if (quote) {
      if (char === quote) { quote = ''; continue; }
      if (char === '\\' && quote === '"' && /["\\$]/.test(command[i + 1] || '')) {
        value += command[++i]; continue;
      }
      value += char; continue;
    }
    if (char === "'" || char === '"') { quote = char; active = true; plain = false; continue; }
    if (char === '\\' && i + 1 < command.length) { active = true; plain = false; value += command[++i]; continue; }
    if (char === '#' && !active) {
      while (i < command.length && command[i] !== '\n') i++;
      flush(); tokens.push({ op: ';' }); continue;
    }
    if (char === '\n' || /[;&|()<>\s]/.test(char)) {
      flush(/[<>]/.test(char) && plain && /^\d+$/.test(value));
      if (char === '\n') tokens.push({ op: ';' });
      else if (!/\s/.test(char)) {
        const match = command.slice(i).match(/^(?:&>>|&>|>>|>\||<>|<<<|<<-|<<|<&|>&|&&|\|\||[;&|()<>])/);
        if (match) { tokens.push({ op: match[0] }); i += match[0].length - 1; }
      }
      continue;
    }
    active = true; value += char;
  }
  flush();
  return tokens;
}
function rawDevice(target) {
  if (!target.startsWith('/')) return false;
  const path = posix.normalize(target);
  return /^\/dev\/(?:r?disk\d+(?:s\d+)*|(?:sd|hd|vd|xvd)[a-z]+\d*|nvme\d+n\d+(?:p\d+)?|mmcblk\d+(?:p\d+)?|(?:loop|dm-|md|ada|da|nvd)\d+(?:p\d+|s\d+)?|disk\/by-[^/]+\/[^/]+|mapper\/[^/]+)$/.test(path);
}
function commandName(tokens) {
  let i = 0;
  while (tokens[i]?.word && /^[A-Za-z_]\w*=/.test(tokens[i].word)) i++;
  if (tokens[i]?.word === 'command' || tokens[i]?.word === 'exec') i++;
  if (tokens[i]?.word === 'env') {
    i++;
    while (tokens[i]?.word && /^[A-Za-z_]\w*=/.test(tokens[i].word)) i++;
  }
  return { name: posix.basename(tokens[i]?.word || ''), index: i };
}
function commandWords(tokens) {
  const result = [];
  for (let i = 0; i < tokens.length; i++) {
    if (tokens[i].ioNumber) continue;
    if (['>', '>>', '>|', '&>', '&>>', '<>', '<', '<<', '<<-', '<<<', '<&', '>&'].includes(tokens[i].op)) {
      if (tokens[i + 1]?.word !== undefined) i++;
      continue;
    }
    if (tokens[i].word !== undefined) result.push(tokens[i]);
  }
  return result;
}
function literalDeviceWrite(command, depth = 0) {
  if (depth > 3) return false;
  const tokens = words(command);
  for (let i = 0; i < tokens.length; i++) {
    if (['>', '>>', '>|', '&>', '&>>', '<>', '>&'].includes(tokens[i].op) &&
        tokens[i + 1]?.word && rawDevice(tokens[i + 1].word)) return true;
  }
  let segment = [];
  const check = () => {
    const positional = commandWords(segment);
    const { name, index } = commandName(positional);
    const args = positional.slice(index + 1).map(t => t.word);
    if (name === 'tee' && args.some(rawDevice)) return true;
    if (name === 'dd' && args.some(arg => arg.startsWith('of=') && rawDevice(arg.slice(3)))) return true;
    if (['sh', 'bash', 'zsh', 'dash'].includes(name)) {
      const c = args.findIndex(arg => /^-[A-Za-z]*c[A-Za-z]*$/.test(arg));
      if (c !== -1 && args[c + 1] && literalDeviceWrite(args[c + 1], depth + 1)) return true;
    }
    return false;
  };
  for (const token of tokens) {
    if (token.op && [';', '&', '|', '&&', '||', '(', ')'].includes(token.op)) {
      if (check()) return true;
      segment = [];
    } else segment.push(token);
  }
  return check();
}
module.exports = { literalDeviceWrite };

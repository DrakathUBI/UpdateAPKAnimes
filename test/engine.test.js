import test from 'node:test';
import assert from 'node:assert/strict';
import { MacroEngine, parseArgs } from '../src/engine.js';

test('parseArgs', () => {
  const out = parseArgs(['node', 'x', '--play', 'a.json', '--silent', '--start-at', 'L1']);
  assert.equal(out.play, 'a.json');
  assert.equal(out.silent, true);
  assert.equal(out.startAt, 'L1');
});

test('goto + if_then_else + calc + runtime', async () => {
  const logs = [];
  const e = new MacroEngine({ silent: false, logger: m => logs.push(m) });
  const script = {
    actions: [
      { type: 'set_variable', label: 'L0', params: { name: 'a', value: '2' } },
      { type: 'calculation', params: { expression: '${a}^3', target: 'res' } },
      { type: 'if_then_else', params: { left: '${res}', operator: 'equals', right: '8', thenLabel: 'OK', elseLabel: 'NO' } },
      { type: 'goto', label: 'NO', params: { label: 'END' } },
      { type: 'set_variable', label: 'OK', params: { name: 'ok', value: 'yes' } },
      { type: 'stop', label: 'END', params: {} }
    ]
  };

  const r = await e.runScript(script);
  assert.equal(r.vars.ok, 'yes');
  assert.equal(r.stopped, true);
  assert.ok(logs.length > 0);
});

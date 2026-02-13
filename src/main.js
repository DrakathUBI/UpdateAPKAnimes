#!/usr/bin/env node
import fs from 'node:fs';
import { MacroEngine, parseArgs } from './engine.js';

async function main() {
  const args = parseArgs(process.argv);
  if (!args.play) {
    console.log('Uso: node src/main.js --play <macro.json> [--start-at Label] [--silent]');
    process.exit(1);
  }
  const macro = JSON.parse(fs.readFileSync(args.play, 'utf8'));
  const engine = new MacroEngine({ silent: args.silent });
  const result = await engine.runScript(macro, { startAt: args.startAt });
  if (!args.silent) console.log(`Variáveis finais em: ${result.outFile}`);
}

main().catch(err => {
  console.error(err.stack || err.message);
  process.exit(1);
});

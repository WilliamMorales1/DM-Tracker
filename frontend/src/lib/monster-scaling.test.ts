import { test, expect } from 'vitest';
import { crToNumber, numberToCr, scaleMonster } from './monster-scaling';
test('crToNumber parses fractional and whole challenge ratings', () => {
  expect(crToNumber('1/8')).toBe(0.125);
  expect(crToNumber('1/4')).toBe(0.25);
  expect(crToNumber('1/2')).toBe(0.5);
  expect(crToNumber('5')).toBe(5);
  expect(crToNumber('')).toBe(0);
  expect(crToNumber('not-a-number')).toBe(0);
});
test('numberToCr formats fractional CRs and passes whole numbers through', () => {
  expect(numberToCr(0.125)).toBe('1/8');
  expect(numberToCr(0.25)).toBe('1/4');
  expect(numberToCr(0.5)).toBe('1/2');
  expect(numberToCr(5)).toBe('5');
});
function baseMonster(overrides = {}) {
  return {
    id: 1,
    name: 'Goblin',
    type: 'humanoid',
    size: 'Small',
    alignment: 'neutral evil',
    cr: '1/4',
    source: 'MM',
    hp: 7,
    hpFormula: '2d6',
    ac: 15,
    acDesc: 'leather armor, shield',
    speed: '30 ft.',
    str: 8,
    dex: 14,
    con: 10,
    int: 10,
    wis: 8,
    cha: 8,
    savingThrows: '',
    skills: '',
    damageResistances: '',
    damageImmunities: '',
    vulnerabilities: '',
    conditionImmunities: '',
    senses: '',
    passivePerception: 9,
    languages: '',
    environment: '',
    imageUrl: '',
    tokenUrl: '',
    traits: null,
    actions: null,
    reactions: null,
    legendaryActions: null,
    bonusActions: null,
    spellcasting: null,
    lairActions: null,
    regionalEffects: null,
    notes: '',
    ...overrides,
  };
}
test('scaleMonster returns the same object unchanged when target CR matches current CR', () => {
  const goblin = baseMonster();
  expect(scaleMonster(goblin, crToNumber(goblin.cr))).toBe(goblin);
});
test('scaleMonster raises HP and relabels name/cr when scaled up', () => {
  const goblin = baseMonster();
  const scaled = scaleMonster(goblin, 10);
  expect(scaled.cr).toBe('10');
  expect(scaled.name).toMatch(/CR 10, scaled/);
  expect(scaled.hp).toBeGreaterThan(goblin.hp);
});
test('scaleMonster lowers HP when scaled down', () => {
  const monster = baseMonster({ cr: '10', hp: 210 });
  const scaled = scaleMonster(monster, 1);
  expect(scaled.hp).toBeLessThan(monster.hp);
});
test('scaleMonster never returns HP below 1', () => {
  const monster = baseMonster({ cr: '30', hp: 800 });
  const scaled = scaleMonster(monster, 0);
  expect(scaled.hp).toBeGreaterThanOrEqual(1);
});
test('scaleMonster rescales to-hit and damage numbers in entry text', () => {
  const monster = baseMonster({
    cr: '1/4',
    actions: [
      {
        name: 'Scimitar',
        text: 'Melee Weapon Attack: +4 to hit. Hit: 5 (1d6 + 2) slashing damage.',
      },
    ],
  });
  const scaled = scaleMonster(monster, 20);
  const text = scaled.actions?.[0].text ?? '';
  expect(text).not.toBe(monster.actions?.[0].text);
  expect(text).toMatch(/to hit/);
  expect(text).toMatch(/slashing damage/);
});
test('scaleMonster keeps HP proportional for creatures below their CR band', () => {
  const boss = baseMonster({ cr: '25', hp: 500, ac: 20 });
  const scaled = scaleMonster(boss, 23);
  expect(scaled.hp).toBe(Math.round((500 * 513) / 603));
  expect(scaled.ac).toBe(20);
});
test('scaleMonster never drops a below-band creature far under the target band', () => {
  const boss = baseMonster({ cr: '21', hp: 345 });
  const scaled = scaleMonster(boss, 20);
  expect(scaled.hp).toBeGreaterThan(300);
});

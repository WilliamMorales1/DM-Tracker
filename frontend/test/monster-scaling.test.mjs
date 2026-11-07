import { test } from 'node:test';
import assert from 'node:assert/strict';
import { crToNumber, numberToCr, scaleMonster } from './monster-scaling.mjs';
test('crToNumber parses fractional and whole challenge ratings', () => {
    assert.equal(crToNumber('1/8'), 0.125);
    assert.equal(crToNumber('1/4'), 0.25);
    assert.equal(crToNumber('1/2'), 0.5);
    assert.equal(crToNumber('5'), 5);
    assert.equal(crToNumber(''), 0);
    assert.equal(crToNumber('not-a-number'), 0);
});
test('numberToCr formats fractional CRs and passes whole numbers through', () => {
    assert.equal(numberToCr(0.125), '1/8');
    assert.equal(numberToCr(0.25), '1/4');
    assert.equal(numberToCr(0.5), '1/2');
    assert.equal(numberToCr(5), '5');
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
    assert.equal(scaleMonster(goblin, crToNumber(goblin.cr)), goblin);
});
test('scaleMonster raises HP and relabels name/cr when scaled up', () => {
    const goblin = baseMonster();
    const scaled = scaleMonster(goblin, 10);
    assert.equal(scaled.cr, '10');
    assert.match(scaled.name, /CR 10, scaled/);
    assert.ok(scaled.hp > goblin.hp, `expected scaled HP (${scaled.hp}) > base HP (${goblin.hp})`);
});
test('scaleMonster lowers HP when scaled down', () => {
    const monster = baseMonster({ cr: '10', hp: 210 });
    const scaled = scaleMonster(monster, 1);
    assert.ok(scaled.hp < monster.hp, `expected scaled HP (${scaled.hp}) < base HP (${monster.hp})`);
});
test('scaleMonster never returns HP below 1', () => {
    const monster = baseMonster({ cr: '30', hp: 800 });
    const scaled = scaleMonster(monster, 0);
    assert.ok(scaled.hp >= 1);
});
test('scaleMonster rescales to-hit and damage numbers in entry text', () => {
    const monster = baseMonster({
        cr: '1/4',
        actions: [{ name: 'Scimitar', text: 'Melee Weapon Attack: +4 to hit. Hit: 5 (1d6 + 2) slashing damage.' }],
    });
    const scaled = scaleMonster(monster, 20);
    const text = scaled.actions?.[0].text ?? '';
    assert.notEqual(text, monster.actions?.[0].text);
    assert.match(text, /to hit/);
    assert.match(text, /slashing damage/);
});
test('scaleMonster keeps HP proportional for creatures below their CR band', () => {
    const boss = baseMonster({ cr: '25', hp: 500, ac: 20 });
    const scaled = scaleMonster(boss, 23);
    assert.equal(scaled.hp, Math.round((500 * 513) / 603));
    assert.equal(scaled.ac, 20);
});
test('scaleMonster never drops a below-band creature far under the target band', () => {
    const boss = baseMonster({ cr: '21', hp: 345 });
    const scaled = scaleMonster(boss, 20);
    assert.ok(scaled.hp > 300, `expected HP > 300, got ${scaled.hp}`);
});

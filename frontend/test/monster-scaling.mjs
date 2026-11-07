export const CR_OPTIONS = ['0', '1/8', '1/4', '1/2', ...Array.from({ length: 30 }, (_, i) => String(i + 1))];
export function crToNumber(cr) {
    cr = (cr || '0').trim();
    if (cr.includes('/')) {
        const [n, d] = cr.split('/').map(Number);
        return d ? n / d : 0;
    }
    const n = Number(cr);
    return Number.isFinite(n) ? n : 0;
}
export function numberToCr(n) {
    if (n === 0.125)
        return '1/8';
    if (n === 0.25)
        return '1/4';
    if (n === 0.5)
        return '1/2';
    return String(n);
}
function crToPb(crNumber) {
    if (crNumber < 5)
        return 2;
    return Math.ceil(crNumber / 4) + 1;
}
// DMG p274 "Monster Statistics by Challenge Rating" - HP and DPR ranges
// (same source data as 5etools' ScaleCreatureConsts).
const CR_HP_RANGES = {
    '0': [1, 6],
    '0.125': [7, 35],
    '0.25': [36, 49],
    '0.5': [50, 70],
    '1': [71, 85],
    '2': [86, 100],
    '3': [101, 115],
    '4': [116, 130],
    '5': [131, 145],
    '6': [146, 160],
    '7': [161, 175],
    '8': [176, 190],
    '9': [191, 205],
    '10': [206, 220],
    '11': [221, 235],
    '12': [236, 250],
    '13': [251, 265],
    '14': [266, 280],
    '15': [281, 295],
    '16': [296, 310],
    '17': [311, 325],
    '18': [326, 340],
    '19': [341, 355],
    '20': [356, 400],
    '21': [401, 445],
    '22': [446, 490],
    '23': [491, 535],
    '24': [536, 580],
    '25': [581, 625],
    '26': [626, 670],
    '27': [671, 715],
    '28': [716, 760],
    '29': [761, 805],
    '30': [806, 850],
};
const CR_DPR_RANGES = {
    '0': [0, 1],
    '0.125': [2, 3],
    '0.25': [4, 5],
    '0.5': [6, 8],
    '1': [9, 14],
    '2': [15, 20],
    '3': [21, 26],
    '4': [27, 32],
    '5': [33, 38],
    '6': [39, 44],
    '7': [45, 50],
    '8': [51, 56],
    '9': [57, 62],
    '10': [63, 68],
    '11': [69, 74],
    '12': [75, 80],
    '13': [81, 86],
    '14': [87, 92],
    '15': [93, 98],
    '16': [99, 104],
    '17': [105, 110],
    '18': [111, 116],
    '19': [117, 122],
    '20': [123, 140],
    '21': [141, 158],
    '22': [159, 176],
    '23': [177, 194],
    '24': [195, 212],
    '25': [213, 230],
    '26': [231, 248],
    '27': [249, 266],
    '28': [267, 284],
    '29': [285, 302],
    '30': [303, 320],
};
// DMG p274 suggested Armor Class per CR - used as a delta reference rather
// than an absolute replacement, so a monster's own AC design is preserved.
const CR_AC = {
    '0': 13,
    '0.125': 13,
    '0.25': 13,
    '0.5': 13,
    '1': 13,
    '2': 13,
    '3': 13,
    '4': 14,
    '5': 15,
    '6': 15,
    '7': 15,
    '8': 16,
    '9': 16,
    '10': 17,
    '11': 17,
    '12': 17,
    '13': 18,
    '14': 18,
    '15': 18,
    '16': 18,
    '17': 19,
    '18': 19,
    '19': 19,
    '20': 19,
    '21': 19,
    '22': 19,
    '23': 19,
    '24': 19,
    '25': 19,
    '26': 19,
    '27': 19,
    '28': 19,
    '29': 19,
    '30': 19,
};
function rangeOf(table, crNumber) {
    const key = numberToCr(crNumber);
    return table[key] ?? table[String(Math.round(crNumber))] ?? [0, 0];
}
// Scales `value` by the ratio of the target CR's band midpoint to the source
// CR's. Position-in-band extrapolation breaks for creatures whose stats sit
// outside their CR band (e.g. WoW bosses with HP far below the band).
function scaleByMidpoint(table, crIn, crOut, value) {
  const midIn = midpoint(rangeOf(table, crIn));
  const midOut = midpoint(rangeOf(table, crOut));
  return midIn > 0 ? (value * midOut) / midIn : value;
}
function midpoint([a, b]) {
    return (a + b) / 2;
}
function fmtSigned(n) {
    return n >= 0 ? `+${n}` : `${n}`;
}
// Rescales an `NdM + K` (or `NdM - K`) hit-die formula so its average lands
// on `newAvg`, keeping the same die type and scaling the die count by the
// same ratio the average changed by.
function scaleFormula(formula, oldAvg, newAvg) {
    const m = /^(\d+)d(\d+)\s*([+-]\s*\d+)?$/.exec((formula || '').trim());
    if (!m || oldAvg <= 0)
        return formula;
    const n = Number(m[1]);
    const sides = Number(m[2]);
    const dieAvg = (sides + 1) / 2;
    const ratio = newAvg / oldAvg;
    const newN = Math.max(1, Math.round(n * ratio));
    const newMod = Math.round(newAvg - newN * dieAvg);
    return newMod === 0 ? `${newN}d${sides}` : `${newN}d${sides} ${fmtSigned(newMod)}`;
}
const reToHit = /([+-]\d+) to hit/g;
const reDC = /DC (\d+)/g;
const reDamage = /(\d+)\s*\((\d+)d(\d+)(?:\s*([+-])\s*(\d+))?\)/g;
function scaleText(text, pbDelta, dprRatio) {
    let out = text.replace(reToHit, (_, bonus) => `${fmtSigned(Number(bonus) + pbDelta)} to hit`);
    out = out.replace(reDC, (_, dc) => `DC ${Math.max(1, Number(dc) + pbDelta)}`);
    out = out.replace(reDamage, (_, avg, n, sides) => {
        const dieAvg = (Number(sides) + 1) / 2;
        const newN = Math.max(1, Math.round(Number(n) * dprRatio));
        const newAvgTotal = Math.max(newN, Math.round(Number(avg) * dprRatio));
        const newMod = newAvgTotal - Math.round(newN * dieAvg);
        const modPart = newMod === 0 ? '' : ` ${fmtSigned(newMod)}`;
        return `${newAvgTotal} (${newN}d${sides}${modPart})`;
    });
    return out;
}
function scaleEntries(entries, pbDelta, dprRatio) {
    if (!entries)
        return entries;
    return entries.map((e) => ({ name: e.name, text: scaleText(e.text, pbDelta, dprRatio) }));
}
export function scaleMonster(base, targetCrNumber) {
    const crInNumber = crToNumber(base.cr);
    if (crInNumber === targetCrNumber)
        return base;
    const pbIn = crToPb(crInNumber);
    const pbOut = crToPb(targetCrNumber);
    const pbDelta = pbOut - pbIn;
    const newHp = Math.max(1, Math.round(scaleByMidpoint(CR_HP_RANGES, crInNumber, targetCrNumber, base.hp)));
    const newAc = base.ac + Math.round((CR_AC[numberToCr(targetCrNumber)] ?? base.ac) - (CR_AC[numberToCr(crInNumber)] ?? base.ac));
    const dprIn = midpoint(rangeOf(CR_DPR_RANGES, crInNumber));
    const dprOut = midpoint(rangeOf(CR_DPR_RANGES, targetCrNumber));
    const dprRatio = dprIn > 0 ? dprOut / dprIn : 1;
    return {
        ...base,
        name: `${base.name} (CR ${numberToCr(targetCrNumber)}, scaled)`,
        cr: numberToCr(targetCrNumber),
        hp: newHp,
        hpFormula: scaleFormula(base.hpFormula, base.hp, newHp),
        ac: newAc,
        traits: scaleEntries(base.traits, pbDelta, dprRatio),
        actions: scaleEntries(base.actions, pbDelta, dprRatio),
        reactions: scaleEntries(base.reactions, pbDelta, dprRatio),
        legendaryActions: scaleEntries(base.legendaryActions, pbDelta, dprRatio),
        spellcasting: scaleEntries(base.spellcasting, pbDelta, dprRatio),
    };
}

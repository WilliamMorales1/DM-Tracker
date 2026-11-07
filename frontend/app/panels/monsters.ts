import { api } from '../api.js';
import { h, mount } from '../dom.js';
import { CR_OPTIONS, crToNumber, scaleMonster } from '../monster-scaling.js';
import { setTitle } from '../title.js';
import { registerPanel } from '../panel.js';
import { autocomplete } from '../autocomplete.js';
// baseType strips a parenthetical suffix so "aberration (beholder)" filters
// together with plain "aberration".
function baseType(type) {
    return type.replace(/\s*\(.*\)\s*$/, '');
}
function statRow(label, value) {
    if (!value)
        return null;
    return h('div', {}, [h('strong', {}, [`${label}. `]), value]);
}
function entryBlock(title, entries) {
    if (!entries || entries.length === 0)
        return null;
    return h('div', { className: 'mb-3' }, [
        h('h5', { className: 'text-yellow-400 border-b border-gray-600 pb-1 font-semibold' }, [title]),
        ...entries.map((e) => h('p', { className: 'mb-2' }, [
            ...(e.name ? [h('strong', { className: 'italic' }, [`${e.name}. `])] : []),
            e.text,
        ])),
    ]);
}
function abilityScore(label, score) {
    const mod = Math.floor((score - 10) / 2);
    const modText = mod >= 0 ? `+${mod}` : `${mod}`;
    return h('div', { className: 'text-center min-w-[56px]' }, [
        h('div', { className: 'text-sm text-gray-400' }, [label]),
        h('div', { className: 'font-bold' }, [String(score)]),
        h('div', { className: 'text-sm text-gray-400' }, [`(${modText})`]),
    ]);
}
function monsterStatBlock(m, onScale, baseCr) {
    const headerLine = [m.size, m.type, m.alignment].filter(Boolean).join(', ');
    return h('div', { className: 'card' }, [
        h('div', { className: 'card-body' }, [
            h('div', { className: 'flex flex-wrap gap-4' }, [
                h('div', { className: 'flex-1 min-w-[280px]' }, [
                    h('div', { className: 'flex items-center gap-2 mb-0' }, [
                        m.tokenUrl
                            ? h('div', { className: 'group relative z-10 w-12 h-12 shrink-0' }, [
                                h('img', {
                                    src: m.tokenUrl,
                                    alt: '',
                                    className: 'absolute inset-0 w-full h-full rounded-full border border-gray-600 object-cover cursor-zoom-in group-hover:shadow-xl group-hover:border-gray-400 group-hover:scale-[5] transition-transform duration-150 origin-center',
                                    onerror: (e) => {
                                        ;
                                        e.target.style.display = 'none';
                                    },
                                }),
                            ])
                            : null,
                        h('h3', { className: 'text-yellow-400 mb-0 text-xl font-bold' }, [m.name]),
                        h('select', {
                            className: 'form-select inline-block w-auto ml-2',
                            onchange: onScale,
                        }, CR_OPTIONS.map((cr) => h('option', { value: cr, selected: cr === m.cr }, [`CR ${cr}${cr === baseCr ? ' (actual)' : ''}`]))),
                    ]),
                    h('div', { className: 'italic text-gray-400 mb-2' }, [headerLine || 'N/A']),
                    h('div', {}, [h('strong', {}, ['Armor Class. ']), `${m.ac}${m.acDesc ? ` (${m.acDesc})` : ''}`]),
                    h('div', {}, [h('strong', {}, ['Hit Points. ']), `${m.hp}${m.hpFormula ? ` (${m.hpFormula})` : ''}`]),
                    h('div', { className: 'mb-2' }, [h('strong', {}, ['Speed. ']), m.speed || 'N/A']),
                    h('div', { className: 'flex gap-3 my-3 border-t border-b border-gray-600 py-2' }, [
                        abilityScore('STR', m.str),
                        abilityScore('DEX', m.dex),
                        abilityScore('CON', m.con),
                        abilityScore('INT', m.int),
                        abilityScore('WIS', m.wis),
                        abilityScore('CHA', m.cha),
                    ]),
                    statRow('Saving Throws', m.savingThrows),
                    statRow('Skills', m.skills),
                    statRow('Damage Vulnerabilities', m.vulnerabilities),
                    statRow('Damage Resistances', m.damageResistances),
                    statRow('Damage Immunities', m.damageImmunities),
                    statRow('Condition Immunities', m.conditionImmunities),
                    h('div', {}, [
                        h('strong', {}, ['Senses. ']),
                        [m.senses, `passive Perception ${m.passivePerception}`].filter(Boolean).join(', '),
                    ]),
                    statRow('Languages', m.languages || 'N/A'),
                    h('div', { className: 'mb-2' }, [h('strong', {}, ['Challenge. ']), `${m.cr || 'N/A'}`]),
                    statRow('Environment', m.environment),
                    h('div', { className: 'text-sm text-gray-400' }, [m.source ? `Source: ${m.source}` : '']),
                ]),
                m.imageUrl
                    ? h('img', {
                        src: m.imageUrl,
                        alt: m.name,
                        className: 'rounded border border-gray-600 max-w-[320px] max-h-[420px] object-contain',
                    })
                    : null,
            ]),
            h('hr', { className: 'border-gray-600 my-3' }, []),
            entryBlock('Traits', m.traits),
            entryBlock('Actions', m.actions),
            entryBlock('Bonus Actions', m.bonusActions),
            entryBlock('Reactions', m.reactions),
            entryBlock('Legendary Actions', m.legendaryActions),
            entryBlock('Lair Actions', m.lairActions),
            entryBlock('Regional Effects', m.regionalEffects),
            entryBlock('Spellcasting', m.spellcasting),
            m.notes ? h('div', { className: 'mt-3 text-gray-400' }, [h('strong', {}, ['Notes. ']), m.notes]) : null,
        ]),
    ]);
}
// Wraps the stat block with a CR scaler (mirroring 5etools' scale-creature
// control): picking a target CR recomputes HP/AC/to-hit/DC/damage via
// monster-scaling module and re-renders the block without refetching.
function monsterDetailView(base) {
    const baseCr = base.cr || '0';
    const body = h('div', {}, []);
    function render(m) {
        mount(body, monsterStatBlock(m, onScale, baseCr));
    }
    function onScale(e) {
        const value = e.target.value;
        render(value === baseCr ? base : scaleMonster(base, crToNumber(value)));
    }
    render(base);
    return body;
}
export async function monstersPanel() {
    const list = (await api.monsters());
    if (list.length === 0) {
        return h('div', { className: 'text-gray-400' }, ['No monsters found. Run: go run ./cmd/ingest-5etools']);
    }
    const detail = h('div', { className: 'mt-3' }, []);
    const types = Array.from(new Set(list.map((m) => baseType(m.type)).filter(Boolean))).sort((a, b) => a.localeCompare(b));
    const sources = Array.from(new Set(list.map((m) => m.source).filter(Boolean))).sort((a, b) => a.localeCompare(b));
    let typeFilter = '';
    let sourceFilter = '';
    let selectSeq = 0;
    let selectedName = null;
    const url = new URL(window.location.href);
    const savedSearch = url.searchParams.get('monsterQuery') || '';
    async function selectMonster(m) {
        const seq = ++selectSeq;
        detail.innerHTML = '';
        detail.append(h('div', { className: 'text-gray-400' }, ['Loading...']));
        const full = await api.monster(m.id);
        if (seq !== selectSeq)
            return;
        mount(detail, monsterDetailView(full));
        selectedName = full.name;
        const url = new URL(window.location.href);
        url.searchParams.set('monster', full.name);
        window.history.replaceState(null, '', url);
        setTitle(selectedName, 'Monsters');
    }
    const { input: nameInput, refresh: refreshSuggestions } = autocomplete({
        placeholder: 'Type a monster name...',
        suggestionClass: 'monster-suggestions',
        label: (m) => m.name,
        search: (query) => {
            const q = query.toLowerCase();
            return list
                .filter((m) => (typeFilter === '' || baseType(m.type) === typeFilter) && (sourceFilter === '' || m.source === sourceFilter) && (q === '' || m.name.toLowerCase().includes(q)))
                .slice()
                .sort((a, b) => a.name.localeCompare(b.name));
        },
        renderItem: (m) => [
            h('span', {}, [m.name]),
            h('span', { className: 'text-gray-400 text-sm whitespace-nowrap' }, [
                [m.type, m.cr ? `CR ${m.cr}` : null].filter(Boolean).join(' · '),
            ]),
        ],
        onSelect: selectMonster,
    });
    nameInput.value = savedSearch;
    nameInput.addEventListener('input', () => {
        const url = new URL(window.location.href);
        if (nameInput.value) url.searchParams.set('monsterQuery', nameInput.value);
        else url.searchParams.delete('monsterQuery');
        url.searchParams.delete('monster');
        window.history.replaceState(null, '', url);
    });
    const typeSelect = h('select', {
        className: 'form-select w-auto',
        'aria-label': 'Filter monsters by type',
        onchange: (e) => {
            typeFilter = e.target.value;
            refreshSuggestions();
        },
    }, [h('option', { value: '' }, ['All Types']), ...types.map((t) => h('option', { value: t }, [t]))]);
    const sourceSelect = h('select', {
        className: 'form-select w-auto',
        'aria-label': 'Filter monsters by source',
        onchange: (e) => {
            sourceFilter = e.target.value;
            refreshSuggestions();
        },
    }, [h('option', { value: '' }, ['All Sources']), ...sources.map((s) => h('option', { value: s }, [s]))]);
    const root = h('div', {}, [h('div', { className: 'flex gap-2 mb-2' }, [nameInput, typeSelect, sourceSelect]), detail]);
    const savedMonster = url.searchParams.get('monster');
    if (savedMonster) {
        const match = list.find((monster) => monster.name === savedMonster);
        if (match) void selectMonster(match);
    }
    return registerPanel(root, { getTitle: () => selectedName });
}

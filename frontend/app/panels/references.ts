import { api } from '../api.js';
import { h } from '../dom.js';
import { dataTable } from '../dataTable.js';
import { setTitle } from '../title.js';
import { registerPanel } from '../panel.js';
async function actionsView() {
    const actions = await api.actions();
    const skillAreas = await api.skillAreas();
    const columns = [
        { header: 'Skill', render: (s) => s.skill, sortValue: (s) => s.skill },
        { header: 'Areas', render: (s) => s.areas, sortValue: (s) => s.areas },
    ];
    const children = [];
    for (const a of actions) {
        children.push(h('div', { className: 'mb-3' }, [
            h('div', { className: 'flex items-baseline gap-3 mb-2' }, [
                h('h2', { className: 'text-base font-bold text-yellow-400 uppercase tracking-wide' }, [a.name]),
                h('span', { className: 'text-xs font-semibold text-gray-400 uppercase tracking-widest' }, [a.tag]),
            ]),
            h('p', { className: 'text-gray-200 leading-relaxed mb-3' }, [a.description]),
        ]));
    }
    children.push(dataTable(columns, skillAreas, 'No skill areas found.'));
    return h('div', {}, children);
}
async function conditionsView() {
    const conditions = (await api.conditions());
    const exhaustionLevels = (await api.exhaustionLevels());
    const columns = [
        { header: 'Condition', render: (c) => c.name, sortValue: (c) => c.name },
        {
            header: 'Effects',
            render: (c) => {
                if (c.effects) {
                    return h('ul', { className: 'list-disc list-outside ml-5 space-y-1' }, c.effects.split('\n').map((b) => h('li', {}, [b])));
                }
                return c.description;
            },
            sortValue: (c) => c.name,
        },
    ];
    const exhaustionColumns = [
        { header: 'Level', render: (e) => e.level, sortValue: (e) => e.level },
        { header: 'Effect', render: (e) => e.effect, sortValue: (e) => e.level },
    ];
    const exhaustion = conditions.find((c) => c.name === 'Exhaustion');
    return h('div', {}, [
        dataTable(columns, conditions, 'No conditions found.'),
        ...(exhaustion
            ? [
                h('h2', { className: 'text-base font-bold text-yellow-400 uppercase tracking-wide mt-4 mb-2' }, [
                    'Exhaustion Levels',
                ]),
                dataTable(exhaustionColumns, exhaustionLevels, 'No exhaustion levels found.'),
                ...exhaustion.descriptionAfter
                    .split('\n\n')
                    .map((para) => h('p', { className: 'text-gray-200 leading-relaxed mt-3' }, [para])),
            ]
            : []),
    ]);
}
async function rangesView() {
    return h('div', {}, [
        h('img', {
            src: '/images/dnd-ranges.jpg',
            alt: 'D&D ranges',
            className: 'max-w-full h-auto rounded',
        }),
    ]);
}
const REFS = [
    { id: 'actions', title: 'Actions', load: actionsView },
    { id: 'conditions', title: 'Conditions', load: conditionsView },
    { id: 'ranges', title: 'Ranges', load: rangesView },
];
export async function refsPanel() {
    const content = h('div', {}, []);
    let selectedTitle = null;
    async function showRef(id) {
        const ref = REFS.find((r) => r.id === id);
        content.innerHTML = '';
        selectedTitle = ref?.title ?? null;
        setTitle(selectedTitle, 'References');
        if (!ref)
            return;
        content.append(await ref.load());
    }
    const select = h('select', { className: 'form-select w-auto', 'aria-label': 'Select a reference' }, [
        h('option', { value: '' }, ['Select a reference...']),
        ...REFS.map((r) => h('option', { value: r.id }, [r.title])),
    ]);
    select.addEventListener('change', () => showRef(select.value));
    select.value = REFS[0].id;
    await showRef(REFS[0].id);
    const root = h('div', {}, [h('div', { className: 'mb-4' }, [select]), content]);
    return registerPanel(root, { getTitle: () => selectedTitle });
}

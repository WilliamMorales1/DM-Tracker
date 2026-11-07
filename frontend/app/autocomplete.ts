import { h } from './dom.js';
export function autocomplete(opts) {
    const minLen = opts.minQueryLength ?? 0;
    const placement = opts.placement ?? 'below';
    const suggestions = h('div', {
        className: `${opts.suggestionClass} list-group shadow border border-gray-600 fixed z-[2000] max-h-[280px] overflow-y-auto hidden`,
    });
    document.body.append(suggestions);
    function hide() {
        suggestions.classList.add('hidden');
        suggestions.innerHTML = '';
    }
    function position() {
        const rect = input.getBoundingClientRect();
        suggestions.style.left = `${rect.left}px`;
        suggestions.style.width = `${rect.width}px`;
        if (placement === 'above') {
            suggestions.style.bottom = `${window.innerHeight - rect.top + 4}px`;
            suggestions.style.top = '';
        }
        else {
            suggestions.style.top = `${rect.bottom + 4}px`;
            suggestions.style.bottom = '';
        }
    }
    function show(query) {
        if (query.length < minLen) {
            hide();
            return;
        }
        const matches = opts.search(query);
        if (matches.length === 0) {
            hide();
            return;
        }
        suggestions.innerHTML = '';
        for (const item of matches) {
            suggestions.append(h('button', {
                type: 'button',
                className: 'list-group-item list-group-item-action flex justify-between items-center gap-2',
                onmousedown: (e) => {
                    e.preventDefault();
                    input.value = opts.label(item);
                    hide();
                    opts.onSelect(item);
                },
            }, opts.renderItem(item)));
        }
        position();
        suggestions.classList.remove('hidden');
    }
    const input = h('input', {
        type: 'text',
        className: 'form-control',
        placeholder: opts.placeholder,
        'aria-label': opts.ariaLabel ?? opts.placeholder,
        autocomplete: 'off',
        oninput: (e) => {
            const value = e.target.value;
            opts.onInputChange?.(value);
            show(value);
        },
        onfocus: (e) => show(e.target.value),
        onblur: () => hide(),
    });
    return {
        input,
        refresh: () => {
            if (document.activeElement === input)
                show(input.value);
        },
    };
}

import { h } from './dom';
export function autocomplete(opts) {
  const minLen = opts.minQueryLength ?? 0;
  const placement = opts.placement ?? 'below';
  const suggestions = h('div', {
    className: `${opts.suggestionClass} list-group shadow border border-gray-600 fixed z-[2000] max-h-[280px] overflow-y-auto hidden`,
  });
  document.body.append(suggestions);
  let items = [];
  let active = -1;
  function hide() {
    suggestions.classList.add('hidden');
    suggestions.innerHTML = '';
    items = [];
    active = -1;
  }
  function choose(item) {
    input.value = opts.label(item);
    hide();
    opts.onSelect(item);
  }
  function setActive(index) {
    const buttons = suggestions.children;
    buttons[active]?.classList.remove('active');
    active = index;
    const el = buttons[active];
    el?.classList.add('active');
    el?.scrollIntoView({ block: 'nearest' });
  }
  function position() {
    const rect = input.getBoundingClientRect();
    suggestions.style.left = `${rect.left}px`;
    suggestions.style.width = `${rect.width}px`;
    if (placement === 'above') {
      suggestions.style.bottom = `${window.innerHeight - rect.top + 4}px`;
      suggestions.style.top = '';
    } else {
      suggestions.style.top = `${rect.bottom + 4}px`;
      suggestions.style.bottom = '';
    }
  }
  // Per-frame tracking also covers layout shifts that fire no scroll/resize event.
  function track() {
    if (suggestions.classList.contains('hidden')) return;
    if (!input.isConnected) {
      hide();
      return;
    }
    position();
    requestAnimationFrame(track);
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
    items = matches;
    active = -1;
    for (const item of matches) {
      suggestions.append(
        h(
          'button',
          {
            type: 'button',
            className:
              'list-group-item list-group-item-action flex justify-between items-center gap-2',
            onmousedown: (e) => {
              e.preventDefault();
              choose(item);
            },
          },
          opts.renderItem(item),
        ),
      );
    }
    const wasHidden = suggestions.classList.contains('hidden');
    position();
    suggestions.classList.remove('hidden');
    if (wasHidden) requestAnimationFrame(track);
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
    onkeydown: (e) => {
      let open = !suggestions.classList.contains('hidden') && items.length > 0;
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
        if (!open) {
          show(input.value);
          open = !suggestions.classList.contains('hidden');
          if (!open) return;
        }
        e.preventDefault();
        const step = e.key === 'ArrowDown' ? 1 : -1;
        setActive(
          active === -1
            ? step === 1
              ? 0
              : items.length - 1
            : (active + step + items.length) % items.length,
        );
      } else if (e.key === 'Enter' && open && active >= 0) {
        e.preventDefault();
        choose(items[active]);
      } else if (e.key === 'Escape' && open) {
        e.preventDefault();
        hide();
      }
    },
    onblur: () => hide(),
  });
  return {
    input,
    refresh: () => {
      if (document.activeElement === input) show(input.value);
    },
  };
}

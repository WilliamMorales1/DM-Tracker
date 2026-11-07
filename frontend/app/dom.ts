export function h(tag, attrs = {}, children = []) {
    const el = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
        if (v === undefined)
            continue;
        if (k.startsWith('on') && typeof v === 'function') {
            el.addEventListener(k.slice(2).toLowerCase(), v);
        }
        else if (k === 'className') {
            el.setAttribute('class', String(v));
        }
        else if (k === 'style' && typeof v === 'object') {
            Object.assign(el.style, v);
        }
        else if (typeof v === 'boolean') {
            if (v)
                el.setAttribute(k, '');
        }
        else {
            el.setAttribute(k, String(v));
        }
    }
    for (const c of children) {
        if (c === null || c === undefined)
            continue;
        el.append(typeof c === 'string' ? document.createTextNode(c) : c);
    }
    return el;
}
export function svg(tag, attrs = {}, children = []) {
    const el = document.createElementNS('http://www.w3.org/2000/svg', tag);
    for (const [k, v] of Object.entries(attrs)) {
        el.setAttribute(k, String(v));
    }
    for (const c of children) {
        el.append(typeof c === 'string' ? document.createTextNode(c) : c);
    }
    return el;
}
export function clear(el) {
    el.innerHTML = '';
}
const SANITIZE_DISALLOWED_TAGS = new Set(['SCRIPT', 'STYLE', 'IFRAME', 'OBJECT', 'EMBED', 'FORM', 'LINK', 'META']);
// Strips script-capable elements/attributes from HTML before it's assigned to
// innerHTML - notes are locally authored, but rendered markdown can still
// carry raw HTML (e.g. pasted from elsewhere) that shouldn't get to run.
export function sanitizeHtml(html) {
    const template = document.createElement('template');
    template.innerHTML = html;
    const walker = document.createTreeWalker(template.content, NodeFilter.SHOW_ELEMENT);
    const toRemove = [];
    let node = walker.nextNode();
    while (node) {
        if (SANITIZE_DISALLOWED_TAGS.has(node.tagName)) {
            toRemove.push(node);
        }
        else {
            for (const attr of Array.from(node.attributes)) {
                const name = attr.name.toLowerCase();
                const value = attr.value.trim().toLowerCase();
                if (name.startsWith('on') || ((name === 'href' || name === 'src') && value.startsWith('javascript:'))) {
                    node.removeAttribute(attr.name);
                }
            }
        }
        node = walker.nextNode();
    }
    for (const el of toRemove)
        el.remove();
    return template.innerHTML;
}
export function mount(parent, child) {
    clear(parent);
    parent.append(child);
}
// Wires a form's submit handler to run an async action with a spinner in
// `status` while busy, the submit button disabled, and errors shown on failure.
export function onSubmitAsync(form, submitBtn, status, busyText, run) {
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        submitBtn.setAttribute('disabled', '');
        status.innerHTML = '';
        status.append(h('span', { className: 'spinner-border spinner-border-sm text-warning me-2' }), h('span', { className: 'text-secondary small' }, [busyText]));
        try {
            await run();
            status.innerHTML = '';
        }
        catch (err) {
            status.innerHTML = '';
            status.append(h('span', { className: 'text-danger small' }, [String(err)]));
        }
        finally {
            submitBtn.removeAttribute('disabled');
        }
    });
}

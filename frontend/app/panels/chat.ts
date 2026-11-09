import { api } from '../api.js';
import { h, onSubmitAsync } from '../dom.js';
function renderAnswer(text) {
    const wrap = h('div', {}, []);
    for (const para of text.split('\n\n')) {
        const p = h('p', {}, []);
        const lines = para.split('\n');
        lines.forEach((line, j) => {
            p.append(document.createTextNode(line));
            if (j < lines.length - 1)
                p.append(h('br'));
        });
        wrap.append(p);
    }
    return wrap;
}
const MODEL_STORAGE_KEY = 'tracker-chat-model';
export async function chatPanel() {
    const history = h('div', { id: 'chat-history' }, []);
    const status = h('div', { className: 'mt-2' }, []);
    let models = [];
    let modelsError = '';
    try {
        models = await api.models();
    }
    catch (err) {
        modelsError = String(err);
    }
    const savedModel = localStorage.getItem(MODEL_STORAGE_KEY);
    // Cloud models answer in seconds; local CPU inference times out on the agent prompt.
    const defaultModel = models.find((m) => m.endsWith(':cloud') || m.endsWith('-cloud') || m.startsWith('ollama.')) ?? models[0] ?? '';
    const initialModel = savedModel && models.includes(savedModel) ? savedModel : defaultModel;
    const modelSelect = h('select', {
        className: 'form-select w-auto',
        'aria-label': 'Chat model',
        disabled: models.length === 0,
    }, models.map((m) => h('option', { value: m }, [m])));
    modelSelect.value = initialModel;
    modelSelect.addEventListener('change', () => localStorage.setItem(MODEL_STORAGE_KEY, modelSelect.value));
    const input = h('input', {
        name: 'q',
        type: 'text',
        className: 'form-control',
        placeholder: 'Ask anything about the campaign...',
        'aria-label': 'Chat question',
        required: true,
    });
    const submitBtn = h('button', {
        className: 'btn btn-primary',
        type: 'submit',
        disabled: models.length === 0,
    }, ['Ask']);
    const form = h('form', { className: 'mt-3' }, [
        h('div', { className: 'flex gap-2 items-stretch' }, [
            modelSelect,
            h('div', { className: 'input-group flex-grow-1' }, [input, submitBtn]),
        ]),
        status,
    ]);
    if (models.length === 0) {
        status.append(h('span', { className: 'text-secondary small' }, [
            modelsError
                ? `No LLM models: ${modelsError}`
                : 'No chat models found. Pull one in Lemonade (`lemonade pull`) or Ollama (`ollama pull`).',
        ]));
    }
    onSubmitAsync(form, submitBtn, status, 'Thinking...', async () => {
        const q = input.value.trim();
        if (!q || !modelSelect.value)
            return;
        const res = await api.chat(q, modelSelect.value);
        history.append(h('div', {}, [
            h('div', { className: 'chat-msg user flex justify-end' }, [
                h('div', { className: 'chat-bubble' }, [res.question]),
            ]),
            h('div', { className: 'chat-msg agent flex justify-start' }, [
                h('div', { className: 'chat-bubble' }, [renderAnswer(res.answer)]),
            ]),
        ]));
        input.value = '';
    });
    return h('div', {}, [history, form]);
}

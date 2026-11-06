import { api } from "../api.js";
import { h, onSubmitAsync } from "../dom.js";
function escapeRegExp(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
function highlight(text, query) {
  const terms = query
    .split(/\s+/)
    .map((t) => t.trim())
    .filter(Boolean)
    .map(escapeRegExp);
  if (terms.length === 0) return [text];
  const pattern = `(${terms.join("|")})`;
  const parts = text.split(new RegExp(pattern, "gi"));
  const isMatch = new RegExp(`^${pattern}$`, "i");
  return parts.map((part) =>
    isMatch.test(part) ? h("mark", {}, [part]) : part,
  );
}
const PAGE_SIZE = 10;
export function searchPanel() {
  const status = h("div", {}, []);
  const results = h("div", { className: "mt-3" }, []);
  const pager = h(
    "div",
    {
      className: "d-flex justify-content-center align-items-center gap-2 mt-3",
    },
    [],
  );
  let current = [];
  let currentQuery = "";
  let page = 0;
  function renderCard(r) {
    const header = [
      h("strong", { className: "text-yellow-400" }, [r.chapterTitle]),
    ];
    // Score < 0 = full-lore browse mode (no query), so no match badge.
    if (r.score >= 0) {
      header.push(
        h("span", { className: "badge bg-gray-600 text-white" }, [
          `${(r.score * 100).toFixed(0)}% match`,
        ]),
      );
    }
    return h("div", { className: "card mb-3" }, [
      h("div", { className: "card-header" }, header),
      h(
        "div",
        { className: "card-body whitespace-pre-wrap text-sm" },
        highlight(r.content, currentQuery),
      ),
    ]);
  }
  function render() {
    results.innerHTML = "";
    pager.innerHTML = "";
    if (current.length === 0) {
      results.append(
        h("p", { className: "text-gray-400" }, ["No results found."]),
      );
      return;
    }
    const pages = Math.ceil(current.length / PAGE_SIZE);
    if (page >= pages) page = pages - 1;
    if (page < 0) page = 0;
    const start = page * PAGE_SIZE;
    for (const r of current.slice(start, start + PAGE_SIZE)) {
      results.append(renderCard(r));
    }
    if (pages <= 1) return;
    const prev = h(
      "button",
      { className: "btn btn-sm btn-outline-secondary", type: "button" },
      ["‹ Prev"],
    );
    prev.disabled = page === 0;
    prev.addEventListener("click", () => {
      page--;
      render();
    });
    const next = h(
      "button",
      { className: "btn btn-sm btn-outline-secondary", type: "button" },
      ["Next ›"],
    );
    next.disabled = page === pages - 1;
    next.addEventListener("click", () => {
      page++;
      render();
    });
    pager.append(
      prev,
      h("span", { className: "text-gray-400 text-sm" }, [
        ` Page ${page + 1} of ${pages} `,
      ]),
      next,
    );
  }
  const input = h("input", {
    name: "q",
    type: "text",
    className: "form-control",
    placeholder: "Search lore semantically... (e.g. 'drow priestess tactics')",
    "aria-label": "Lore search query",
  });
  const submitBtn = h(
    "button",
    { className: "btn btn-warning", type: "submit" },
    ["Search"],
  );
  const form = h("form", {}, [
    h("div", { className: "input-group mb-3" }, [input, submitBtn]),
    status,
  ]);
  async function load(q) {
    currentQuery = q;
    current = await api.search(q);
    page = 0;
    render();
  }
  onSubmitAsync(form, submitBtn, status, "Searching...", async () => {
    await load(input.value.trim());
  });
  // Show all blocks in order on first load.
  void load("");
  return h("div", {}, [form, results, pager]);
}

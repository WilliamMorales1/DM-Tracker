import { h } from "./dom.js";
export function dataTable(columns, rows, emptyMessage) {
  if (rows.length === 0) {
    return h("p", { className: "p-3 text-gray-400" }, [emptyMessage]);
  }
  let filter = "";
  let sortCol = null;
  let sortAsc = true;
  const tbody = h("tbody", {}, []);
  const theadRow = h("tr", {}, []);
  const wrapper = h("div", { className: "overflow-x-auto" }, []);
  function getRows() {
    let result = rows;
    if (filter.trim()) {
      const f = filter.toLowerCase();
      result = result.filter((row) =>
        columns.some((c) =>
          String(c.sortValue ? c.sortValue(row) : "")
            .toLowerCase()
            .includes(f),
        ),
      );
    }
    if (sortCol !== null) {
      const col = columns[sortCol];
      if (col.sortValue) {
        result = [...result].sort((a, b) => {
          const av = col.sortValue(a);
          const bv = col.sortValue(b);
          const cmp = av < bv ? -1 : av > bv ? 1 : 0;
          return sortAsc ? cmp : -cmp;
        });
      }
    }
    return result;
  }
  function renderHead() {
    theadRow.innerHTML = "";
    columns.forEach((c, i) => {
      const th = h(
        "th",
        {
          className: c.sortValue
            ? "cursor-pointer whitespace-nowrap"
            : "whitespace-nowrap",
          onclick: c.sortValue
            ? () => {
                if (sortCol === i) {
                  sortAsc = !sortAsc;
                } else {
                  sortCol = i;
                  sortAsc = true;
                }
                renderHead();
                renderBody();
              }
            : undefined,
        },
        [c.header + (sortCol === i ? (sortAsc ? " ▲" : " ▼") : "")],
      );
      theadRow.append(th);
    });
  }
  function highlight(el, term) {
    if (!term) return;
    const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT);
    const textNodes = [];
    let node;
    while ((node = walker.nextNode())) textNodes.push(node);
    const re = new RegExp(term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"), "ig");
    for (const tn of textNodes) {
      const text = tn.textContent ?? "";
      if (!re.test(text)) continue;
      re.lastIndex = 0;
      const frag = document.createDocumentFragment();
      let lastIndex = 0;
      let m;
      while ((m = re.exec(text))) {
        frag.append(text.slice(lastIndex, m.index));
        frag.append(h("mark", {}, [m[0]]));
        lastIndex = m.index + m[0].length;
      }
      frag.append(text.slice(lastIndex));
      tn.replaceWith(frag);
    }
  }
  function renderBody() {
    tbody.innerHTML = "";
    const f = filter.trim();
    for (const row of getRows()) {
      const tr = h("tr", {}, []);
      for (const c of columns) {
        const td = h("td", {}, [c.render(row)]);
        if (f) highlight(td, f);
        tr.append(td);
      }
      tbody.append(tr);
    }
  }
  const filterInput = h("input", {
    type: "text",
    className: "form-control mb-2",
    placeholder: "Filter columns...",
    oninput: (e) => {
      filter = e.target.value;
      renderBody();
    },
  });
  renderHead();
  renderBody();
  const table = h("table", { className: "table table-hover w-full" }, [
    h("thead", {}, [theadRow]),
    tbody,
  ]);
  wrapper.append(filterInput, table);
  return wrapper;
}

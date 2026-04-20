const STORAGE_KEY = "ot-calculator-payload-v2";

const state = {
  entries: []
};

function uid(prefix) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

function defaultRow() {
  return {
    id: uid("row"),
    date: "",
    period: "AM",
    kind: "OT",
    employeeId: "A",
    startTime: "",
    endTime: ""
  };
}

function saveState() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
}

function normalizeLegacy(parsed) {
  if (Array.isArray(parsed.entries)) {
    return parsed.entries.map((e) => ({ period: "AM", ...e }));
  }

  const rows = [];
  if (Array.isArray(parsed.otEntries)) {
    parsed.otEntries.forEach((e) => rows.push({ period: "AM", ...e, kind: "OT" }));
  }
  if (Array.isArray(parsed.breakEntries)) {
    parsed.breakEntries.forEach((e) => rows.push({ period: "AM", ...e, kind: "BREAK" }));
  }
  return rows;
}

function loadState() {
  const raw = localStorage.getItem(STORAGE_KEY) || localStorage.getItem("ot-calculator-payload-v1");
  if (!raw) return;
  try {
    const parsed = JSON.parse(raw);
    state.entries = normalizeLegacy(parsed);
  } catch {
    state.entries = [];
  }
}

function kindSelect(value) {
  return `<select data-field="kind"><option value="OT" ${value === "OT" ? "selected" : ""}>OT</option><option value="BREAK" ${value === "BREAK" ? "selected" : ""}>Break</option></select>`;
}

function employeeSelect(value) {
  return `<select data-field="employeeId"><option value="A" ${value === "A" ? "selected" : ""}>A</option><option value="B" ${value === "B" ? "selected" : ""}>B</option></select>`;
}

function rowHtml(row) {
  return `
    <td><input data-field="date" type="text" inputmode="numeric" placeholder="YYYY-MM-DD" value="${row.date || ""}"></td>
    <td><select data-field="period"><option value="AM" ${row.period === "AM" ? "selected" : ""}>AM</option><option value="PM" ${row.period === "PM" ? "selected" : ""}>PM</option></select></td>
    <td>${kindSelect(row.kind)}</td>
    <td>${employeeSelect(row.employeeId)}</td>
    <td><input data-field="startTime" type="text" inputmode="numeric" placeholder="HH:MM" value="${row.startTime || ""}"></td>
    <td><input data-field="endTime" type="text" inputmode="numeric" placeholder="HH:MM" value="${row.endTime || ""}"></td>
    <td><button type="button" data-action="delete">Delete</button></td>
  `;
}

function bindRows(tbodyEl, rows) {
  tbodyEl.innerHTML = "";
  rows.forEach((row, index) => {
    const tr = document.createElement("tr");
    tr.innerHTML = rowHtml(row);

    tr.querySelectorAll("[data-field]").forEach((el) => {
      el.addEventListener("change", async (e) => {
        const field = e.target.dataset.field;
        state.entries[index][field] = e.target.value.trim();
        saveState();
        await recalculate();
      });
    });

    tr.querySelector('[data-action="delete"]').addEventListener("click", async () => {
      state.entries.splice(index, 1);
      render();
      saveState();
      await recalculate();
    });

    tbodyEl.appendChild(tr);
  });
}

function render() {
  bindRows(document.getElementById("entry-body"), state.entries);
}

function isComplete(row) {
  return row.employeeId && row.date && row.startTime && row.endTime;
}

function toPayload() {
  const otEntries = [];
  const breakEntries = [];

  state.entries.filter(isComplete).forEach((row) => {
    const base = {
      id: row.id,
      employeeId: row.employeeId,
      date: row.date,
      period: row.period,
      startTime: row.startTime,
      endTime: row.endTime
    };
    if (row.kind === "BREAK") {
      breakEntries.push(base);
    } else {
      otEntries.push(base);
    }
  });

  return { otEntries, breakEntries };
}

function renderDaily(data) {
  const root = document.getElementById("daily-results");
  root.innerHTML = "";

  ["A", "B"].forEach((emp) => {
    const h3 = document.createElement("h3");
    h3.textContent = `Employee ${emp}`;
    root.appendChild(h3);

    const byDate = data[emp] || {};
    const dates = Object.keys(byDate).sort();
    if (dates.length === 0) {
      const p = document.createElement("p");
      p.textContent = "No records";
      root.appendChild(p);
      return;
    }

    dates.forEach((dateKey) => {
      const day = byDate[dateKey];
      const rate20H = Math.floor(day.rate20Minutes / 60);
      const rate20M = day.rate20Minutes % 60;
      const rate15H = Math.floor(day.rate15Minutes / 60);
      const rate15M = day.rate15Minutes % 60;
      const div = document.createElement("div");
      div.className = "day-card";
      div.innerHTML = `
        <strong>Day ${day.dateLabel}:</strong><br>
        2.0x : (${(day.rate20Segments || []).join(" + ") || "-"}) = ${rate20H}hr:${String(rate20M).padStart(2, "0")}min(${Number(day.rate20RoundedHours * 2).toFixed(1)}hr)<br>
        1.5x : (${(day.rate15Segments || []).join(" + ") || "-"}) = ${rate15H}hr:${String(rate15M).padStart(2, "0")}min (${Number(day.rate15RoundedHours * 1.5).toFixed(1)}hr)<br>
        <strong>Total: ${Number(day.totalWeighted).toFixed(1)}hr</strong>
      `;
      root.appendChild(div);
    });
  });
}

function renderMonthly(data) {
  const root = document.getElementById("monthly-results");
  root.innerHTML = "";

  ["A", "B"].forEach((emp) => {
    const h3 = document.createElement("h3");
    h3.textContent = `Employee ${emp}`;
    root.appendChild(h3);

    const byMonth = data[emp] || {};
    const months = Object.keys(byMonth).sort();
    if (months.length === 0) {
      const p = document.createElement("p");
      p.textContent = "No records";
      root.appendChild(p);
      return;
    }

    months.forEach((monthKey) => {
      const m = byMonth[monthKey];
      const div = document.createElement("div");
      div.className = "month-card";
      div.innerHTML = `
        <strong>${monthKey}</strong><br>
        2.0x : ${m.rate20RoundedHours}hr<br>
        1.5x : ${m.rate15RoundedHours}hr<br>
        <strong>Total weighted: ${Number(m.totalWeighted).toFixed(1)}hr</strong>
      `;
      root.appendChild(div);
    });
  });
}

async function recalculate() {
  try {
    const resp = await fetch("/api/calculate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(toPayload())
    });
    if (!resp.ok) {
      throw new Error(await resp.text());
    }
    const data = await resp.json();
    renderDaily(data.dailySummary || {});
    renderMonthly(data.monthlySummary || {});
  } catch (err) {
    document.getElementById("daily-results").innerHTML = `<p>Error: ${String(err)}</p>`;
    document.getElementById("monthly-results").innerHTML = "";
  }
}

function init() {
  loadState();
  if (state.entries.length === 0) state.entries.push(defaultRow());

  document.getElementById("add-entry").addEventListener("click", async () => {
    state.entries.push(defaultRow());
    render();
    saveState();
    await recalculate();
  });

  render();
  saveState();
  recalculate();
}

document.addEventListener("DOMContentLoaded", init);

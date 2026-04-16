const STORAGE_KEY = "ot-calculator-payload-v1";

const state = {
  otEntries: [],
  breakEntries: []
};

function uid(prefix) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

function defaultRow(prefix) {
  return {
    id: uid(prefix),
    employeeId: "A",
    date: "",
    startTime: "",
    endTime: ""
  };
}

function saveState() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
}

function loadState() {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return;
  try {
    const parsed = JSON.parse(raw);
    state.otEntries = Array.isArray(parsed.otEntries) ? parsed.otEntries : [];
    state.breakEntries = Array.isArray(parsed.breakEntries) ? parsed.breakEntries : [];
  } catch {
    state.otEntries = [];
    state.breakEntries = [];
  }
}

function employeeSelect(value) {
  return `<select data-field="employeeId"><option value="A" ${value === "A" ? "selected" : ""}>A</option><option value="B" ${value === "B" ? "selected" : ""}>B</option></select>`;
}

function rowHtml(row) {
  return `
    <td>${employeeSelect(row.employeeId)}</td>
    <td><input data-field="date" type="date" value="${row.date || ""}"></td>
    <td><input data-field="startTime" type="time" value="${row.startTime || ""}"></td>
    <td><input data-field="endTime" type="time" value="${row.endTime || ""}"></td>
    <td><button type="button" data-action="delete">Delete</button></td>
  `;
}

function bindRows(tbodyEl, rows, key) {
  tbodyEl.innerHTML = "";
  rows.forEach((row, index) => {
    const tr = document.createElement("tr");
    tr.innerHTML = rowHtml(row);

    tr.querySelectorAll("[data-field]").forEach((el) => {
      el.addEventListener("change", async (e) => {
        const field = e.target.dataset.field;
        state[key][index][field] = e.target.value;
        saveState();
        await recalculate();
      });
    });

    tr.querySelector('[data-action="delete"]').addEventListener("click", async () => {
      state[key].splice(index, 1);
      render();
      saveState();
      await recalculate();
    });

    tbodyEl.appendChild(tr);
  });
}

function render() {
  bindRows(document.getElementById("ot-body"), state.otEntries, "otEntries");
  bindRows(document.getElementById("break-body"), state.breakEntries, "breakEntries");
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
      const div = document.createElement("div");
      div.className = "day-card";
      div.innerHTML = `
        <strong>Day ${day.dateLabel}</strong><br>
        (${(day.rate20Segments || []).join(" + ") || "-"}) ${day.rate20Minutes} mins -> ${day.rate20RoundedHours} hr x2.0 = ${(day.rate20RoundedHours * 2).toFixed(1)}<br>
        (${(day.rate15Segments || []).join(" + ") || "-"}) ${day.rate15Minutes} mins -> ${day.rate15RoundedHours} hr x1.5 = ${(day.rate15RoundedHours * 1.5).toFixed(1)}<br>
        <strong>Total: ${Number(day.totalWeighted).toFixed(1)}</strong>
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
      div.innerHTML = `<strong>${monthKey}</strong> - 1.5x hrs: ${m.rate15RoundedHours}, 2.0x hrs: ${m.rate20RoundedHours}, total weighted: ${Number(m.totalWeighted).toFixed(1)}`;
      root.appendChild(div);
    });
  });
}

async function recalculate() {
  try {
    const resp = await fetch("/api/calculate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(state)
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
  if (state.otEntries.length === 0) state.otEntries.push(defaultRow("ot"));
  if (state.breakEntries.length === 0) state.breakEntries.push(defaultRow("br"));

  document.getElementById("add-ot").addEventListener("click", async () => {
    state.otEntries.push(defaultRow("ot"));
    render();
    saveState();
    await recalculate();
  });

  document.getElementById("add-break").addEventListener("click", async () => {
    state.breakEntries.push(defaultRow("br"));
    render();
    saveState();
    await recalculate();
  });

  render();
  saveState();
  recalculate();
}

document.addEventListener("DOMContentLoaded", init);

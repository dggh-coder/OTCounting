const API_BASE = window.__API_BASE__ || "";

const state = {
  staff: [],
  selectedStaff: "",
  date: "",
  period: "",
  rows: []
};

function endpoint(path) {
  return API_BASE ? `${API_BASE}${path}` : path;
}

function rowTemplate() {
  return { type: "00", startTime: "", endTime: "", inputBy: "" };
}

function formatPeriodLabel(p) {
  if (p === "00") return "Morning (早)";
  if (p === "01") return "Noon (中)";
  return "Evening (晚)";
}

async function loadStaff() {
  const select = document.getElementById("staff-select");
  select.innerHTML = "<option value=''>-- Select --</option>";
  try {
    const resp = await fetch(endpoint("/api/staff"));
    const data = await resp.json();
    state.staff = data.staff || [];
    state.staff.forEach((s) => {
      const opt = document.createElement("option");
      opt.value = s.staffid;
      opt.textContent = `${s.displayname} (${s.staffid})`;
      select.appendChild(opt);
    });
  } catch {
    const msg = document.getElementById("select-msg");
    msg.textContent = "Cannot load staff list.";
  }
}

function renderRows() {
  const body = document.getElementById("entry-body");
  body.innerHTML = "";
  state.rows.forEach((r, idx) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td><select data-k="type"><option value="00" ${r.type === "00" ? "selected" : ""}>OT</option><option value="01" ${r.type === "01" ? "selected" : ""}>Break</option></select></td>
      <td><input data-k="startTime" placeholder="HH:MM" value="${r.startTime}"></td>
      <td><input data-k="endTime" placeholder="HH:MM" value="${r.endTime}"></td>
      <td><input data-k="inputBy" placeholder="staffid (optional)" value="${r.inputBy}"></td>
      <td><button data-action="del" type="button">Delete</button></td>
    `;
    tr.querySelectorAll("[data-k]").forEach((el) => {
      el.addEventListener("change", (e) => {
        state.rows[idx][e.target.dataset.k] = e.target.value.trim();
      });
    });
    tr.querySelector("[data-action='del']").addEventListener("click", () => {
      state.rows.splice(idx, 1);
      renderRows();
    });
    body.appendChild(tr);
  });
}

function showStep(stepId) {
  ["step-select", "step-period", "step-input"].forEach((id) => {
    document.getElementById(id).classList.toggle("hidden", id !== stepId);
  });
}

function resetToStart(message = "") {
  state.period = "";
  state.rows = [];
  document.getElementById("input-msg").textContent = "";
  document.getElementById("select-msg").textContent = message;
  showStep("step-select");
}

async function confirmInput() {
  const msg = document.getElementById("input-msg");
  msg.textContent = "";
  if (state.rows.length === 0) {
    msg.textContent = "Please add at least one row.";
    return;
  }
  const entries = state.rows.map((r) => ({
    type: r.type,
    startTime: r.startTime,
    endTime: r.endTime,
    inputBy: r.inputBy || null
  }));
  const payload = {
    otstaffid: state.selectedStaff,
    date: state.date,
    period: state.period,
    entries
  };
  const resp = await fetch(endpoint("/api/ot/input"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (!resp.ok) {
    msg.textContent = await resp.text();
    return;
  }
  resetToStart("Saved successfully.");
}

function bindEvents() {
  document.getElementById("to-period").addEventListener("click", () => {
    state.selectedStaff = document.getElementById("staff-select").value;
    state.date = document.getElementById("work-date").value;
    if (!state.selectedStaff || !state.date) {
      document.getElementById("select-msg").textContent = "Please select both staff and date.";
      return;
    }
    document.getElementById("select-msg").textContent = "";
    showStep("step-period");
  });

  document.getElementById("back-select").addEventListener("click", () => showStep("step-select"));

  document.querySelectorAll("[data-period]").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.period = btn.dataset.period;
      state.rows = [rowTemplate()];
      document.getElementById("context").textContent = `${state.selectedStaff} / ${state.date} / ${formatPeriodLabel(state.period)}`;
      renderRows();
      showStep("step-input");
    });
  });

  document.getElementById("add-row").addEventListener("click", () => {
    state.rows.push(rowTemplate());
    renderRows();
  });
  document.getElementById("cancel-input").addEventListener("click", () => showStep("step-period"));
  document.getElementById("confirm").addEventListener("click", confirmInput);
}

function init() {
  bindEvents();
  loadStaff();
  showStep("step-select");
}

document.addEventListener("DOMContentLoaded", init);

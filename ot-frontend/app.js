const API_BASE = window.__API_BASE__ || "";

const state = {
  staff: [],
  selectedStaff: "",
  date: "",
  rows: [],
  existing: []
};

function endpoint(path) {
  return API_BASE ? `${API_BASE}${path}` : path;
}

function rowTemplate() {
  return { type: "00", startTime: "", endTime: "" };
}

function switchTab(tabName) {
  const names = ["ot", "staff", "process"];
  names.forEach((n) => {
    const section = document.getElementById(`tab-${n}`);
    const button = document.getElementById(`tab-btn-${n}`);
    if (section) section.classList.toggle("hidden", n !== tabName);
    if (button) button.classList.toggle("active", n === tabName);
  });
}

function renderStaffList() {
  const root = document.getElementById("staff-list");
  root.innerHTML = "";
  if (state.staff.length === 0) {
    root.textContent = "No staff found.";
    return;
  }
  state.staff.forEach((s) => {
    const div = document.createElement("div");
    div.className = "staff-item";
    div.textContent = `ID: ${s.staffid} | Eng: ${s.nameeng || ""} | Chi: ${s.namechi || ""} | Display: ${s.displayname || ""} | Domain: ${s.domainname || ""}`;
    root.appendChild(div);
  });
}

function fillStaffSelects() {
  const otSelect = document.getElementById("staff-select");
  const processSelect = document.getElementById("process-staff-filter");
  if (otSelect) otSelect.innerHTML = "<option value=''>-- Select --</option>";
  if (processSelect) processSelect.innerHTML = "<option value=''>All Staff (last 10 days)</option>";

  state.staff.forEach((s) => {
    if (otSelect) {
      const opt1 = document.createElement("option");
      opt1.value = s.staffid;
      opt1.textContent = `${s.displayname || s.staffid} (${s.staffid})`;
      otSelect.appendChild(opt1);
    }
    if (processSelect) {
      const opt2 = document.createElement("option");
      opt2.value = s.staffid;
      opt2.textContent = `${s.displayname || s.staffid} (${s.staffid})`;
      processSelect.appendChild(opt2);
    }
  });
}

async function loadStaff() {
  try {
    const resp = await fetch(endpoint("/api/staff"));
    const data = await resp.json();
    state.staff = data.staff || [];
    fillStaffSelects();
    renderStaffList();
  } catch {
    document.getElementById("select-msg").textContent = "Cannot load staff list.";
  }
}

async function saveStaff() {
  const msg = document.getElementById("staff-msg");
  msg.textContent = "";
  msg.style.color = "#b00020";
  const staffid = document.getElementById("staff-id").value.trim();
  const nameeng = document.getElementById("staff-nameeng").value.trim();
  const namechi = document.getElementById("staff-namechi").value.trim();
  const displayname = document.getElementById("staff-displayname").value.trim();
  const domainname = document.getElementById("staff-domainname").value.trim();
  if (!staffid) {
    msg.textContent = "Staff No (ID) is required.";
    return;
  }

  const resp = await fetch(endpoint("/api/staff/input"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ staffid, nameeng, namechi, displayname, domainname })
  });
  if (!resp.ok) {
    msg.textContent = await resp.text();
    return;
  }

  ["staff-id", "staff-nameeng", "staff-namechi", "staff-displayname", "staff-domainname"].forEach((id) => {
    document.getElementById(id).value = "";
  });
  msg.style.color = "#0a7a2f";
  msg.textContent = "Staff saved.";
  await loadStaff();
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

function renderExisting() {
  const body = document.getElementById("existing-body");
  body.innerHTML = "";
  if (state.existing.length === 0) {
    const tr = document.createElement("tr");
    tr.innerHTML = `<td colspan="4">No existing records</td>`;
    body.appendChild(tr);
    return;
  }
  state.existing.forEach((r) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${r.type === "00" ? "OT" : "Break"}</td>
      <td>${r.startTime}</td>
      <td>${r.endTime}</td>
      <td><button data-action="delete-existing" data-id="${r.id}" type="button">Delete</button></td>
    `;
    tr.querySelector("[data-action='delete-existing']").addEventListener("click", async (e) => {
      await deleteExistingRecord(e.target.dataset.id);
    });
    body.appendChild(tr);
  });
}

async function deleteExistingRecord(id) {
  const resp = await fetch(endpoint(`/api/ot/entry?id=${encodeURIComponent(id)}`), { method: "DELETE" });
  if (!resp.ok) {
    document.getElementById("input-msg").textContent = await resp.text();
    return;
  }
  await loadExistingRecords();
}

function showStep(stepId) {
  ["step-select", "step-input"].forEach((id) => {
    document.getElementById(id).classList.toggle("hidden", id !== stepId);
  });
}

function resetToStart(message = "") {
  state.rows = [];
  state.existing = [];
  document.getElementById("input-msg").textContent = "";
  document.getElementById("select-msg").textContent = message;
  showStep("step-select");
}

async function loadExistingRecords() {
  const url = endpoint(`/api/ot/entries?otstaffid=${encodeURIComponent(state.selectedStaff)}&date=${encodeURIComponent(state.date)}`);
  const resp = await fetch(url);
  const data = await resp.json();
  state.existing = (data.entries || []).map((e) => ({ id: e.id, type: e.type, startTime: e.startTime, endTime: e.endTime }));
  renderExisting();
}

async function confirmInput() {
  const msg = document.getElementById("input-msg");
  msg.textContent = "";
  const timePattern = /^([01]\d|2[0-3]):[0-5]\d$/;
  for (const r of state.rows) {
    if (!r.startTime || !r.endTime || !timePattern.test(r.startTime) || !timePattern.test(r.endTime)) {
      msg.textContent = "Start and End must be HH:MM and cannot be empty.";
      return;
    }
  }

  const existingAsEntries = state.existing.map((e) => ({ type: e.type, startTime: e.startTime, endTime: e.endTime }));
  const newEntries = state.rows.map((r) => ({ type: r.type, startTime: r.startTime, endTime: r.endTime }));
  const allEntries = [...existingAsEntries, ...newEntries];
  if (allEntries.length === 0) {
    msg.textContent = "No records to confirm.";
    return;
  }

  const payload = { otstaffid: state.selectedStaff, date: state.date, entries: allEntries };
  const resp = await fetch(endpoint("/api/ot/input"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload)
  });
  if (!resp.ok) {
    msg.textContent = await resp.text();
    return;
  }

  state.rows = [];
  await loadExistingRecords();
  renderRows();
  msg.style.color = "#0a7a2f";
  msg.textContent = "Confirmed and recalculated.";
}

function renderProcessTexts(rows) {
  const root = document.getElementById("process-root");
  root.innerHTML = "";
  if (!rows.length) {
    root.textContent = "No process text rows.";
    return;
  }

  const grouped = {};
  rows.forEach((r) => {
    if (!grouped[r.otstaffid]) grouped[r.otstaffid] = {};
    if (!grouped[r.otstaffid][r.date_label]) grouped[r.otstaffid][r.date_label] = [];
    grouped[r.otstaffid][r.date_label].push(r);
  });

  Object.keys(grouped).sort().forEach((staffid) => {
    const h3 = document.createElement("h3");
    h3.textContent = `Staff ${staffid}`;
    root.appendChild(h3);
    const byDate = grouped[staffid];
    Object.keys(byDate).sort((a, b) => b.localeCompare(a)).forEach((date) => {
      const card = document.createElement("div");
      card.className = "day-card";
      card.innerHTML = `<strong>${date}</strong>`;
      byDate[date].forEach((row) => {
        const p = document.createElement("p");
        p.textContent = `2.0: ${row.process20txt} | 1.5: ${row.process15txt}`;
        card.appendChild(p);
      });
      root.appendChild(card);
    });
  });
}

async function loadProcessTexts() {
  const staffid = document.getElementById("process-staff-filter").value;
  const url = staffid ? endpoint(`/api/ot/process-texts?otstaffid=${encodeURIComponent(staffid)}`) : endpoint("/api/ot/process-texts");
  const resp = await fetch(url);
  const data = await resp.json();
  renderProcessTexts(data.rows || []);
}

function bindEvents() {
  document.querySelectorAll(".tab-btn").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const tab = btn.dataset.tab;
      switchTab(tab);
      if (tab === "process") {
        await loadProcessTexts();
      }
    });
  });
  document.getElementById("save-staff")?.addEventListener("click", saveStaff);
  document.getElementById("refresh-process")?.addEventListener("click", loadProcessTexts);

  document.getElementById("to-period")?.addEventListener("click", async () => {
    state.selectedStaff = document.getElementById("staff-select").value;
    state.date = document.getElementById("work-date").value;
    if (!state.selectedStaff || !state.date) {
      document.getElementById("select-msg").textContent = "Please select both staff and date.";
      return;
    }
    document.getElementById("select-msg").textContent = "";
    state.rows = [rowTemplate()];
    document.getElementById("context").textContent = `${state.selectedStaff} / ${state.date}`;
    await loadExistingRecords();
    renderRows();
    showStep("step-input");
  });

  document.getElementById("add-row")?.addEventListener("click", () => {
    state.rows.push(rowTemplate());
    renderRows();
  });
  document.getElementById("cancel-input")?.addEventListener("click", () => resetToStart());
  document.getElementById("confirm")?.addEventListener("click", confirmInput);
}

function init() {
  bindEvents();
  loadStaff();
  showStep("step-select");
  switchTab("ot");
}

document.addEventListener("DOMContentLoaded", init);

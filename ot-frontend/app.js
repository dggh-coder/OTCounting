const API_BASE = window.__API_BASE__ || "";

const state = {
  staff: [],
  selectedStaff: "",
  date: "",
  rows: []
};

function endpoint(path) {
  return API_BASE ? `${API_BASE}${path}` : path;
}

function rowTemplate() {
  return { type: "00", startTime: "", endTime: "" };
}

function switchTab(tabName) {
  const isOT = tabName === "ot";
  document.getElementById("tab-ot").classList.toggle("hidden", !isOT);
  document.getElementById("tab-staff").classList.toggle("hidden", isOT);
  document.getElementById("tab-btn-ot").classList.toggle("active", isOT);
  document.getElementById("tab-btn-staff").classList.toggle("active", !isOT);
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
    div.textContent = `ID: ${s.staffid} | Eng: ${s.nameeng} | Chi: ${s.namechi} | Display: ${s.displayname} | Domain: ${s.domainname}`;
    root.appendChild(div);
  });
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
    renderStaffList();
  } catch {
    const msg = document.getElementById("select-msg");
    msg.textContent = "Cannot load staff list.";
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

  document.getElementById("staff-id").value = "";
  document.getElementById("staff-nameeng").value = "";
  document.getElementById("staff-namechi").value = "";
  document.getElementById("staff-displayname").value = "";
  document.getElementById("staff-domainname").value = "";
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

function showStep(stepId) {
  ["step-select", "step-input"].forEach((id) => {
    document.getElementById(id).classList.toggle("hidden", id !== stepId);
  });
}

function resetToStart(message = "") {
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
  const timePattern = /^([01]\d|2[0-3]):[0-5]\d$/;
  for (const r of state.rows) {
    if (!r.startTime || !r.endTime || !timePattern.test(r.startTime) || !timePattern.test(r.endTime)) {
      msg.textContent = "Start and End must be HH:MM and cannot be empty.";
      return;
    }
  }
  const entries = state.rows.map((r) => ({ type: r.type, startTime: r.startTime, endTime: r.endTime }));
  const payload = {
    otstaffid: state.selectedStaff,
    date: state.date,
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
  document.getElementById("tab-btn-ot").addEventListener("click", () => switchTab("ot"));
  document.getElementById("tab-btn-staff").addEventListener("click", () => switchTab("staff"));
  document.getElementById("save-staff").addEventListener("click", saveStaff);

  document.getElementById("to-period").addEventListener("click", () => {
    state.selectedStaff = document.getElementById("staff-select").value;
    state.date = document.getElementById("work-date").value;
    if (!state.selectedStaff || !state.date) {
      document.getElementById("select-msg").textContent = "Please select both staff and date.";
      return;
    }
    document.getElementById("select-msg").textContent = "";
    state.rows = [rowTemplate()];
    document.getElementById("context").textContent = `${state.selectedStaff} / ${state.date}`;
    renderRows();
    showStep("step-input");
  });

  document.getElementById("add-row").addEventListener("click", () => {
    state.rows.push(rowTemplate());
    renderRows();
  });
  document.getElementById("cancel-input").addEventListener("click", () => showStep("step-select"));
  document.getElementById("confirm").addEventListener("click", confirmInput);
}

function init() {
  bindEvents();
  loadStaff();
  showStep("step-select");
  switchTab("ot");
}

document.addEventListener("DOMContentLoaded", init);

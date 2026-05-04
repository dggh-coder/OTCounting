
async function loadWelcomeUser() {
  const el = document.getElementById("welcome-title");
  if (!el) return;
  try {
    const resp = await fetch(endpoint("/api/user"), { cache: "no-store" });
    if (!resp.ok) {
      el.textContent = "";
      return;
    }
    const data = await resp.json();
    const username = String(data?.user || "").trim();
    el.textContent = username ? `Welcome ${username}` : "";
  } catch (_) {
    el.textContent = "";
  }
}
const API_BASE = window.__API_BASE__ || "";

const state = {
  staff: [],
  groups: [],
  nextGroupId: 1,
  reportRowsCount: 0,
  auditRowsCount: 0,
  reportMonthYear: new Date().getFullYear(),
  reportMonthValue: "",
  auditRows: []
};

function endpoint(path) { return API_BASE ? `${API_BASE}${path}` : path; }
function rowTemplate() { return { type: "00", startTime: "", endTime: "" }; }
function createGroup() { return { id: state.nextGroupId++, staff: "", date: "", remarks: "", remarksReadonly: false, hasPeriodRecord: false, expanded: false, locked: false, rows: [], existing: [], msg: "" }; }



function switchTopView(viewName){
  const showingDriverShell = viewName === "driver" || viewName === "staffmgmt";
  ["home","driver"].forEach((n)=>{
    const shouldShow = (n === "driver" && showingDriverShell) || n === viewName;
    document.getElementById(`view-${n}`)?.classList.toggle("hidden", !shouldShow);
  });
  document.querySelectorAll('.top-link').forEach((b)=>b.classList.toggle('active', b.dataset.view===viewName));
  if (viewName === "driver") {
    setDriverMode("ot");
  } else if (viewName === "staffmgmt") {
    setDriverMode("staff");
  }
}

function startClock(){
  const el=document.getElementById('live-clock');
  if(!el) return;
  const tick=()=>{ el.textContent = new Date().toLocaleTimeString(); };
  tick();
  setInterval(tick,1000);
}
function switchTab(tabName) {
  ["summary", "ot", "report", "staff"].forEach((n) => {
    const section = document.getElementById(`tab-${n}`);
    const button = document.getElementById(`tab-btn-${n}`);
    if (section) section.classList.toggle("hidden", n !== tabName);
    if (button) button.classList.toggle("active", n === tabName);
  });
}

function setDriverMode(mode) {
  const summaryBtn = document.getElementById("tab-btn-summary");
  const otBtn = document.getElementById("tab-btn-ot");
  const reportBtn = document.getElementById("tab-btn-report");
  const staffBtn = document.getElementById("tab-btn-staff");
  if (summaryBtn) summaryBtn.classList.toggle("hidden", mode !== "ot");
  if (otBtn) otBtn.classList.toggle("hidden", mode !== "ot");
  if (reportBtn) reportBtn.classList.toggle("hidden", mode !== "ot");
  if (staffBtn) staffBtn.classList.toggle("hidden", mode !== "staff");
  switchTab(mode === "ot" ? "summary" : mode);
}




function switchReportTab(tabName) {
  ["monthly", "audit"].forEach((n) => {
    document.getElementById(`report-tab-${n}`)?.classList.toggle("hidden", n !== tabName);
    document.getElementById(`report-tab-btn-${n}`)?.classList.toggle("active", n === tabName);
  });
}

function fillReportStaffOptions() {
  const sel = document.getElementById('report-staff');
  if (!sel) return;
  const options = state.staff
    .filter((s) => (s.staffgroup || '').trim().toLowerCase() === 'driver')
    .sort((a,b)=> (a.displayname||a.staffid).localeCompare(b.displayname||b.staffid))
    .map((s)=>`<option value="${s.staffid}">${s.displayname || s.staffid} (${s.staffid})</option>`);
  sel.innerHTML = options.length
    ? `<option value="" selected>-- Select Driver --</option>${options.join('')}`
    : '<option value="">No driver</option>';
}

function fillAuditStaffOptions() {
  const sel = document.getElementById('audit-staff');
  if (!sel) return;
  sel.innerHTML = document.getElementById('report-staff')?.innerHTML || '<option value="">No driver</option>';
}


function monthLabel(v) {
  if (!v) return 'Select Month';
  const [y, m] = v.split('-');
  const names = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  return `${names[Number(m)-1]} ${y}`;
}

function renderReportMonthPicker() {
  document.getElementById('report-year-label').textContent = String(state.reportMonthYear);
  const root = document.getElementById('report-month-grid');
  if (!root) return;
  const names = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  root.innerHTML = names.map((n, idx) => {
    const val = `${state.reportMonthYear}-${String(idx+1).padStart(2,'0')}`;
    const active = state.reportMonthValue === val ? 'active' : '';
    return `<button type="button" class="month-cell ${active}" data-month="${val}">${n}</button>`;
  }).join('');
  root.querySelectorAll('[data-month]').forEach((btn)=>btn.addEventListener('click', () => {
    state.reportMonthValue = btn.getAttribute('data-month');
    document.getElementById('report-month-trigger').textContent = monthLabel(state.reportMonthValue);
    document.getElementById('report-month-picker').classList.add('hidden');
    renderReportMonthPicker();
  }));
}
function resetMonthlyReportPage() {
  const staff = document.getElementById('report-staff');
  const monthTrigger = document.getElementById('report-month-trigger');
  const msg = document.getElementById('report-msg');
  const context = document.getElementById('report-context');
  const body = document.getElementById('report-body');
  if (staff) staff.value = '';
  if (monthTrigger) monthTrigger.textContent = 'Select Month';
  state.reportMonthValue = '';
  state.reportMonthYear = new Date().getFullYear();
  if (msg) msg.textContent = '';
  if (context) context.textContent = '';
  if (body) body.innerHTML = '<tr><td colspan="3">No data</td></tr>';
  state.reportRowsCount = 0;
}

function resetAuditReportPage() {
  const staff = document.getElementById('audit-staff');
  if (staff) staff.value = '';
  const start = document.getElementById('audit-start-date');
  const end = document.getElementById('audit-end-date');
  if (start) start.value = '';
  if (end) end.value = '';
  document.getElementById('audit-msg').textContent = '';
  document.getElementById('audit-context').textContent = '';
  document.getElementById('audit-body').innerHTML = '<tr><td colspan="3">No data</td></tr>';
  state.auditRowsCount = 0;
  state.auditRows = [];
}

async function loadMonthlyReport() {
  const msg = document.getElementById('report-msg');
  const body = document.getElementById('report-body');
  const context = document.getElementById('report-context');
  const staffSel = document.getElementById('report-staff');
  const staffID = document.getElementById('report-staff')?.value || '';
  const month = state.reportMonthValue || '';
  if (!staffID || !month) { msg.textContent = 'Please select staff and month.'; return; }
  msg.textContent = '';
  const yyyymm = month.replace('-', '');
  const staffName = staffSel?.selectedOptions?.[0]?.textContent?.split(' (')[0] || staffID;
  context.textContent = `${staffName} ${yyyymm} :`;
  const [resp, summaryResp] = await Promise.all([
    fetch(endpoint(`/api/ot/driver-monthly-report?otstaffid=${encodeURIComponent(staffID)}&yyyymm=${encodeURIComponent(yyyymm)}`), { cache: 'no-store' }),
    fetch(endpoint(`/api/ot/driver-monthly-summary?yyyymm=${encodeURIComponent(yyyymm)}`), { cache: 'no-store' })
  ]);
  if (!resp.ok) { msg.textContent = await resp.text(); return; }
  if (!summaryResp.ok) { msg.textContent = await summaryResp.text(); return; }
  const data = await resp.json();
  const summaryData = await summaryResp.json();
  const summaryRow = (summaryData.rows || []).find((r) => String(r.otstaffid) === String(staffID)) || { totalhrs20: 0, totalhrs15: 0 };
  const rows = data.rows || [];
  state.reportRowsCount = rows.length;
  let detailRows = '';
  let lastDate = '';
  rows.forEach((r) => {
    if (r.date !== lastDate) {
      detailRows += `<tr class="report-remark-row"><td colspan="3">${r.date} Justification: ${r.remarks || '-'}</td></tr>`;
      lastDate = r.date;
    }
    detailRows += `<tr><td>${r.date}</td><td>${r.startTime}</td><td>${r.endTime}</td></tr>`;
  });
  if (!detailRows) detailRows = '<tr><td colspan="3">No data</td></tr>';
  const totalRows = `<tr class="report-total-row"><td colspan="3">2.0 Total: ${summaryRow.totalhrs20} hrs</td></tr><tr class="report-total-row"><td colspan="3">1.5 Total: ${summaryRow.totalhrs15} hrs</td></tr>`;
  body.innerHTML = `${detailRows}${totalRows}`;
}

function exportMonthlyReport(kind = 'csv') {
  const staffSel = document.getElementById('report-staff');
  const staffID = staffSel?.value || '';
  const month = state.reportMonthValue || '';
  if (!staffID || !month) {
    document.getElementById('report-msg').textContent = 'Please select staff and month before export.';
    return;
  }
  if (!state.reportRowsCount) {
    window.alert('No result for export.');
    return;
  }
  const yyyymm = month.replace('-', '');
  const staffName = staffSel?.selectedOptions?.[0]?.textContent?.split(' (')[0] || staffID;
  const suffix = kind === 'xlsx' ? 'export-xlsx' : 'export';
  const url = endpoint(`/api/ot/driver-monthly-report/${suffix}?otstaffid=${encodeURIComponent(staffID)}&yyyymm=${encodeURIComponent(yyyymm)}&staffname=${encodeURIComponent(staffName)}`);
  window.open(url, '_blank');
}

function renderAuditRows(rows, summaryRows) {
  const body = document.getElementById('audit-body');
  const detailsByDate = {};
  (rows || []).forEach((r) => {
    if (!detailsByDate[r.date]) detailsByDate[r.date] = [];
    detailsByDate[r.date].push(r);
  });

  const summaryByDate = {};
  (summaryRows || []).forEach((r) => {
    if (!summaryByDate[r.date]) {
      summaryByDate[r.date] = { periods: { "00": {}, "01": {}, "02": {} }, total20: 0, total15: 0 };
    }
    const day = summaryByDate[r.date];
    if (["00", "01", "02"].includes(r.period)) day.periods[r.period] = r;
    day.total20 += Number(r.totalhrs20 || 0);
    day.total15 += Number(r.totalhrs15 || 0);
  });

  const allDates = Object.keys(detailsByDate).sort();
  let html = '';
  allDates.forEach((date) => {
    (detailsByDate[date] || []).forEach((r) => {
      html += `<tr><td>${r.date}</td><td>${r.startTime}</td><td>${r.endTime}</td></tr>`;
    });

    const day = summaryByDate[date] || { periods: { "00": {}, "01": {}, "02": {} }, total20: 0, total15: 0 };
    const p20 = ["00", "01", "02"].map((p) => `[<strong>${(day.periods[p]?.process20txt || '').trim()}</strong>]`).join(' + ');
    const p15 = ["00", "01", "02"].map((p) => `[<strong>${(day.periods[p]?.process15txt || '').trim()}</strong>]`).join(' + ');
    html += `<tr class="audit-total-row"><td colspan="3">2.0 OT: ${p20} = ${day.total20} hrs</td></tr>`;
    html += `<tr class="audit-total-row"><td colspan="3">1.5 OT: ${p15} = ${day.total15} hrs</td></tr>`;
  });

  body.innerHTML = html || '<tr><td colspan="3">No data</td></tr>';
}

async function loadAuditReport() {
  const msg = document.getElementById('audit-msg');
  const staffSel = document.getElementById('audit-staff');
  const staffID = staffSel?.value || '';
  const startDate = document.getElementById('audit-start-date')?.value.trim() || '';
  const endDate = document.getElementById('audit-end-date')?.value.trim() || '';
  if (!staffID || !startDate || !endDate) { msg.textContent = 'Please select staff, start date and end date.'; return; }
  if (endDate < startDate) { msg.textContent = 'End Date must be greater than or equal to Start Date.'; return; }
  msg.textContent = '';
  document.getElementById('audit-context').textContent = `${staffSel.selectedOptions?.[0]?.textContent || staffID} ${startDate} ~ ${endDate}`;
  const resp = await fetch(endpoint(`/api/ot/driver-audit-report?otstaffid=${encodeURIComponent(staffID)}&startDate=${encodeURIComponent(startDate)}&endDate=${encodeURIComponent(endDate)}`), { cache: 'no-store' });
  if (!resp.ok) { msg.textContent = await resp.text(); return; }
  const data = await resp.json();
  state.auditRows = data.rows || [];
  state.auditRowsCount = state.auditRows.length;
  renderAuditRows(state.auditRows, data.summaryRows || []);
}

function exportAuditReport(kind = 'csv') {
  const staffSel = document.getElementById('audit-staff');
  const staffID = staffSel?.value || '';
  const startDate = document.getElementById('audit-start-date')?.value.trim() || '';
  const endDate = document.getElementById('audit-end-date')?.value.trim() || '';
  if (!staffID || !startDate || !endDate) {
    document.getElementById('audit-msg').textContent = 'Please select staff, start date and end date before export.';
    return;
  }
  if (!state.auditRowsCount) {
    window.alert('No result for export.');
    return;
  }
  const staffName = staffSel?.selectedOptions?.[0]?.textContent?.split(' (')[0] || staffID;
  const suffix = kind === 'xlsx' ? 'export-xlsx' : 'export';
  const url = endpoint(`/api/ot/driver-audit-report/${suffix}?otstaffid=${encodeURIComponent(staffID)}&startDate=${encodeURIComponent(startDate)}&endDate=${encodeURIComponent(endDate)}&staffname=${encodeURIComponent(staffName)}`);
  window.open(url, '_blank');
}
function currentYYYYMM() {
  const now = new Date();
  return `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, "0")}`;
}

function previousYYYYMM(yyyymm) {
  const y = Number(yyyymm.slice(0, 4));
  const m = Number(yyyymm.slice(4, 6));
  const d = new Date(y, m - 2, 1);
  return `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, "0")}`;
}

function renderDriverSummary(rows, yyyymm, rootId, labelId, cardClassName = "") {
  const label = document.getElementById(labelId);
  if (label) label.textContent = `Summary Month · ${yyyymm}`;
  const root = document.getElementById(rootId);
  if (!root) return;
  root.innerHTML = "";
  if (!rows.length) { root.innerHTML = "<p class='summary-empty'>No driver data found for this month.</p>"; return; }
  rows.forEach((r) => {
    const hrs20 = Number(r.totalhrs20 || 0);
    const hrs15 = Number(r.totalhrs15 || 0);
    const card = document.createElement("article");
    card.className = `summary-card ${cardClassName}`.trim();
    card.innerHTML = `<div class="summary-head"><h3 class="summary-name">${r.displayname || r.otstaffid}</h3><span class="summary-id">ID ${r.otstaffid}</span></div>
    <div class="summary-metrics">
      <div class="metric-pill metric-pill--20"><span class="metric-label">2.0x OT</span><strong><span class="metric-value">${hrs20}</span> hrs</strong></div>
      <div class="metric-pill metric-pill--15"><span class="metric-label">1.5x OT</span><strong><span class="metric-value">${hrs15}</span> hrs</strong></div>
    </div>`;
    root.appendChild(card);
  });
}

async function loadDriverSummary() {
  const yyyymm = currentYYYYMM();
  const prev = previousYYYYMM(yyyymm);
  const [respCurrent, respPrev] = await Promise.all([
    fetch(endpoint(`/api/ot/driver-monthly-summary?yyyymm=${encodeURIComponent(yyyymm)}`), { cache: 'no-store' }),
    fetch(endpoint(`/api/ot/driver-monthly-summary?yyyymm=${encodeURIComponent(prev)}`), { cache: 'no-store' })
  ]);
  if (!respCurrent.ok) throw new Error(await respCurrent.text());
  if (!respPrev.ok) throw new Error(await respPrev.text());
  const dataCurrent = await respCurrent.json();
  const dataPrev = await respPrev.json();
  renderDriverSummary(dataCurrent.rows || [], yyyymm, 'driver-summary-grid', 'summary-month-label');
  renderDriverSummary(dataPrev.rows || [], prev, 'driver-summary-grid-prev', 'summary-prev-month-label', 'summary-card--prev');
}
function fillStaffOptions(selected) {
  return ["<option value=''>-- Select --</option>"]
    .concat(state.staff.filter((s) => (s.staffgroup || "").trim().toLowerCase() === "driver")
      .map((s) => `<option value="${s.staffid}" ${selected === s.staffid ? "selected" : ""}>${s.displayname || s.staffid} (${s.staffid})</option>`)).join("");
}

function initDatePickers(scopeEl) {
  if (!window.flatpickr || !scopeEl) return;
  scopeEl.querySelectorAll("input.js-date-picker").forEach((el) => {
    if (el._flatpickr) return;
    flatpickr(el, {
      dateFormat: "Y-m-d",
      allowInput: true,
      clickOpens: !el.disabled,
      disableMobile: true
    });
  });
}

function renderStaffList() {
  const root = document.getElementById("staff-list");
  root.innerHTML = "";
  if (state.staff.length === 0) { root.textContent = "No staff found."; return; }

  const table = document.createElement("table");
  table.className = "staff-table";
  table.innerHTML = `<thead><tr><th>ID</th><th>Eng</th><th>Chi</th><th>Display</th><th>Domain</th><th>Group</th><th>Action</th></tr></thead><tbody></tbody>`;
  const tbody = table.querySelector("tbody");

  state.staff.forEach((s) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `<td>${s.staffid || "-"}</td><td>${s.nameeng || "-"}</td><td>${s.namechi || "-"}</td><td>${s.displayname || "-"}</td><td>${s.domainname || "-"}</td><td>${s.staffgroup || "-"}</td><td><button class="btn-danger" data-action="delete-staff" data-staffid="${s.staffid}" type="button">Delete</button></td>`;
    tr.querySelector("[data-action='delete-staff']").addEventListener("click", async (e) => deleteStaff(e.target.dataset.staffid));
    tbody.appendChild(tr);
  });

  root.appendChild(table);
}

function renderGroups() {
  const root = document.getElementById("ot-groups");
  root.innerHTML = "";
  state.groups.forEach((g) => {
    const sec = document.createElement("section"); sec.className = `card ot-group-card${g.expanded ? " is-expanded" : ""}`;
    const statusIsDraft = g.rows.length > 0;
    sec.innerHTML = `<button class="group-remove" data-action="remove" type="button" aria-label="Remove OT Input #${g.id}">×</button>
    <div class="ot-group-header">
      <h2>OT Input</h2>
      ${statusIsDraft ? `<span class="status-pill status-pill--draft">Draft</span>` : `<span class="status-pill">Saved</span>`}
    </div>
    <div class="row ot-group-form">
      <label>Staff<select data-k="staff" ${g.locked ? "disabled" : ""}>${fillStaffOptions(g.staff)}</select></label>
      <label>Date<input data-k="date" class="js-date-picker" type="text" value="${g.date}" placeholder="YYYY-MM-DD" ${g.locked ? "disabled" : ""}></label>
      <button class="btn-primary" data-action="next" type="button" ${g.locked ? "disabled" : ""}>Next</button>
      ${g.expanded ? `<div class="remarks-wrap"><label class="remarks-field">Justification<input data-k="remarks" class="${g.remarksReadonly ? "remarks-locked" : ""}" type="text" value="${g.remarks}" ${g.remarksReadonly ? "disabled" : ""}></label>${g.hasPeriodRecord ? `<button class="btn-ghost remarks-toggle-btn" data-action="toggle-remarks" type="button" title="${g.remarksReadonly ? "Edit Justification" : "Save Justification"}">${g.remarksReadonly ? "✎" : "✓"}</button>` : ""}</div>` : ""}
    </div>
    <div class="msg select-msg">${g.msg || ""}</div>
    <div class="period-area ${g.expanded ? "" : "hidden"}">
      <h3>Existing Records (Read-only, can delete)</h3>
      <table><thead><tr><th>Type</th><th>Start (HH:MM)</th><th>End (HH:MM)</th><th></th></tr></thead><tbody class="existing-body"></tbody></table>
      <div class="section-head"><h3>New Records</h3><button class="btn-ghost" data-action="add-row" type="button">Add Row</button></div>
      <table><thead><tr><th>Type</th><th>Start (HH:MM)</th><th>End (HH:MM)</th><th></th></tr></thead><tbody class="entry-body"></tbody></table>
      <div class="actions"><button class="btn-primary" data-action="confirm" type="button" ${g.rows.length === 0 ? "disabled" : ""}>Confirm</button></div>
      <div class="msg input-msg"></div>
    </div>`;

    sec.querySelectorAll("[data-k]").forEach((el) => el.addEventListener("change", (e) => { g[e.target.dataset.k] = e.target.value.trim(); }));
    sec.querySelector("[data-action='remove']").addEventListener("click", () => {
      state.groups = state.groups.filter((x) => x.id !== g.id); if (!state.groups.length) state.groups = [createGroup()]; renderGroups();
    });
    sec.querySelector("[data-action='next']").addEventListener("click", async () => {
      if (!g.staff || !g.date) { g.msg = "Please select both staff and date."; renderGroups(); return; }
      const dup = state.groups.find((x) => x.id !== g.id && x.staff === g.staff && x.date === g.date);
      if (dup) { g.msg = "Same Staff + Date already exists on this page."; renderGroups(); return; }
      g.msg = ""; g.expanded = true; g.locked = true; if (g.rows.length === 0) g.rows = [rowTemplate()];
      renderGroups();
      try {
        await loadExistingRecords(g);
      } catch (err) {
        g.msg = err?.message || "Failed to load existing records.";
      }
      renderGroups();
    });

    const period = sec.querySelector(".period-area");
    const existingBody = sec.querySelector(".existing-body");
    existingBody.innerHTML = g.existing.length ? "" : `<tr><td colspan="4">No existing records</td></tr>`;
    g.existing.forEach((r) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${r.type === "00" ? "OT" : "Break"}</td><td>${r.startTime}</td><td>${r.endTime}</td><td><button type="button">Delete</button></td>`;
      tr.querySelector("button").addEventListener("click", async () => {
        if (!window.confirm("Delete this existing record?")) return;
        await deleteExistingRecord(g, r.id);
      });
      existingBody.appendChild(tr);
    });
    const entryBody = sec.querySelector(".entry-body"); entryBody.innerHTML = "";
    g.rows.forEach((r, idx) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `<td><select data-k="type"><option value="00" ${r.type === "00" ? "selected" : ""}>OT</option><option value="01" ${r.type === "01" ? "selected" : ""}>Break</option></select></td><td><input data-k="startTime" placeholder="HH:MM" value="${r.startTime}"></td><td><input data-k="endTime" placeholder="HH:MM" value="${r.endTime}"></td><td><button type="button">Delete</button></td>`;
      tr.querySelectorAll("[data-k]").forEach((el) => el.addEventListener("change", (e) => { g.rows[idx][e.target.dataset.k] = e.target.value.trim(); }));
      tr.querySelector("button").addEventListener("click", () => {
        g.rows.splice(idx, 1); renderGroups();
      });
      entryBody.appendChild(tr);
    });
    sec.querySelector("[data-action='add-row']").addEventListener("click", () => { g.rows.push(rowTemplate()); renderGroups(); });
    sec.querySelector("[data-action='confirm']").addEventListener("click", async () => { await confirmInput(g, sec.querySelector('.input-msg')); });
    sec.querySelector("[data-action='toggle-remarks']")?.addEventListener("click", async () => {
      if (g.remarksReadonly) {
        g.remarksReadonly = false; renderGroups(); return;
      }
      await saveRemarksOnly(g);
      g.remarksReadonly = true;
      renderGroups();
    });
    initDatePickers(sec);

    root.appendChild(sec);
  });
}

async function loadStaff() { const resp = await fetch(endpoint('/api/staff'), { cache: 'no-store' }); const data = await resp.json(); state.staff = data.staff || []; renderStaffList(); renderGroups(); }
async function loadExistingRecords(g) { const resp = await fetch(endpoint(`/api/ot/entries?otstaffid=${encodeURIComponent(g.staff)}&date=${encodeURIComponent(g.date)}`)); if(!resp.ok){throw new Error(await resp.text());} const data = await resp.json(); const entries=data.entries||[]; g.existing = entries.map((e)=>({id:e.id,type:e.type,startTime:e.startTime,endTime:e.endTime})); g.hasPeriodRecord = !!data.exists; g.remarks = typeof data.remarks === "string" ? data.remarks : (entries.length ? (entries[0].remarks || "") : g.remarks); g.remarksReadonly = g.hasPeriodRecord; }
async function saveRemarksOnly(g){ const resp = await fetch(endpoint('/api/ot/remarks'),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({otstaffid:g.staff,date:g.date,remarks:g.remarks})}); if(!resp.ok){throw new Error(await resp.text());} }
async function deleteExistingRecord(g,id){ const resp=await fetch(endpoint(`/api/ot/entry?id=${encodeURIComponent(id)}`),{method:'DELETE'}); if(resp.ok){await loadExistingRecords(g); renderGroups();}}
async function confirmInput(g,msgEl){ msgEl.textContent=''; const p=/^([01]\d|2[0-3]):[0-5]\d$/; for(const r of g.rows){ if(!r.startTime||!r.endTime||!p.test(r.startTime)||!p.test(r.endTime)){msgEl.textContent='Start and End must be HH:MM and cannot be empty.';return;}}
 const all=[...g.existing.map((e)=>({type:e.type,startTime:e.startTime,endTime:e.endTime})),...g.rows.map((r)=>({type:r.type,startTime:r.startTime,endTime:r.endTime}))];
 if(!all.length){msgEl.textContent='No records to confirm.';return;} const payload={otstaffid:g.staff,date:g.date,remarks:g.remarks,entries:all}; const resp=await fetch(endpoint('/api/ot/input'),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)}); if(!resp.ok){msgEl.textContent=await resp.text(); return;} g.rows=[]; await loadExistingRecords(g); renderGroups(); }

async function saveStaff(){/* unchanged simplified */
  const msg=document.getElementById('staff-msg'); msg.textContent=''; msg.style.color='#b00020';
  const staffid=document.getElementById('staff-id').value.trim(); const nameeng=document.getElementById('staff-nameeng').value.trim(); const namechi=document.getElementById('staff-namechi').value.trim(); const displayname=document.getElementById('staff-displayname').value.trim(); const domainname=document.getElementById('staff-domainname').value.trim(); const staffgroup=document.getElementById('staff-staffgroup').value.trim(); if(!staffid){msg.textContent='Staff No (ID) is required.';return;}
  const resp=await fetch(endpoint('/api/staff/input'),{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({staffid,nameeng,namechi,displayname,domainname,staffgroup})}); if(!resp.ok){msg.textContent=await resp.text();return;} ['staff-id','staff-nameeng','staff-namechi','staff-displayname','staff-domainname','staff-staffgroup'].forEach((id)=>document.getElementById(id).value=''); msg.style.color='#0a7a2f'; msg.textContent='Staff saved.'; await loadStaff();
}
async function deleteStaff(staffid){const msg=document.getElementById('staff-msg'); msg.textContent=''; if(!window.confirm(`Delete staff ${staffid}?`)) return; const resp=await fetch(endpoint(`/api/staff?staffid=${encodeURIComponent(staffid)}`),{method:'DELETE'}); if(!resp.ok){msg.textContent=await resp.text();return;} msg.style.color='#0a7a2f'; msg.textContent=`Staff ${staffid} deleted.`; await loadStaff();}

function resetStaffInputForm() {
  ['staff-id','staff-nameeng','staff-namechi','staff-displayname','staff-domainname','staff-staffgroup']
    .forEach((id) => {
      const input = document.getElementById(id);
      if (input) input.value = '';
    });
}

function bindEvents(){
  document.querySelectorAll('[data-tab]').forEach((btn)=>btn.addEventListener('click', async function(){
    const tab = this.getAttribute('data-tab');
    if (!tab) return;
    switchTab(tab);
    await reloadActiveSubPage(tab);
  }));
  document.querySelectorAll('.top-link').forEach((btn)=>btn.addEventListener('click', function(){
    const view = this.getAttribute('data-view');
    const targetTab = this.getAttribute('data-tab-target');
    switchTopView(view);
    if (view === 'driver' && targetTab) switchTab(targetTab);
    void reloadView(view, targetTab);
  }));
  document.querySelectorAll('[data-report-tab]').forEach((btn)=>btn.addEventListener('click', function(){
    const tab = this.getAttribute('data-report-tab');
    switchReportTab(tab);
    if (tab === 'monthly') resetMonthlyReportPage();
    if (tab === 'audit') resetAuditReportPage();
  }));

  document.getElementById('report-month-trigger')?.addEventListener('click', () => {
    document.getElementById('report-month-picker').classList.toggle('hidden');
    renderReportMonthPicker();
  });
  document.addEventListener('click', (e) => {
    const wrap = document.querySelector('.month-picker-wrap');
    if (!wrap || wrap.contains(e.target)) return;
    document.getElementById('report-month-picker')?.classList.add('hidden');
  });
  document.getElementById('report-year-prev')?.addEventListener('click', () => { state.reportMonthYear -= 1; renderReportMonthPicker(); });
  document.getElementById('report-year-next')?.addEventListener('click', () => { state.reportMonthYear += 1; renderReportMonthPicker(); });
  document.getElementById('report-search')?.addEventListener('click',loadMonthlyReport);
  document.getElementById('report-export-csv')?.addEventListener('click', () => exportMonthlyReport('csv'));
  document.getElementById('report-export-xlsx')?.addEventListener('click', () => exportMonthlyReport('xlsx'));
  document.getElementById('audit-search')?.addEventListener('click', loadAuditReport);
  document.getElementById('audit-export-csv')?.addEventListener('click', () => exportAuditReport('csv'));
  document.getElementById('audit-export-xlsx')?.addEventListener('click', () => exportAuditReport('xlsx'));
  document.getElementById('save-staff')?.addEventListener('click',saveStaff);
  document.getElementById('add-group')?.addEventListener('click',()=>{state.groups.push(createGroup()); renderGroups();});
}

async function reloadActiveSubPage(tabName) {
  if (tabName === 'summary') {
    await loadDriverSummary();
    return;
  }
  if (tabName === 'report') {
    fillReportStaffOptions();
    fillAuditStaffOptions();
    switchReportTab('monthly');
    resetMonthlyReportPage();
    initDatePickers(document.getElementById('report-tab-audit'));
    return;
  }
  if (tabName === 'ot') {
    state.groups = [createGroup()];
    renderGroups();
    return;
  }
  if (tabName === 'staff') {
    resetStaffInputForm();
    document.getElementById('staff-msg').textContent = '';
    await loadStaff();
  }
}

async function reloadView(view, targetTab) {
  if (view === 'driver') {
    const activeTab = targetTab || (document.querySelector('.tab-btn.active')?.getAttribute('data-tab')) || 'ot';
    await reloadActiveSubPage(activeTab);
    return;
  }
  if (view === 'staffmgmt') {
    switchTab('staff');
    await reloadActiveSubPage('staff');
  }
}

function init(){ state.groups=[createGroup()]; bindEvents(); loadStaff(); switchTab('ot'); switchTopView('home'); startClock();
  loadWelcomeUser(); }
document.addEventListener('DOMContentLoaded', init);

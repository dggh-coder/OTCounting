const API_BASE = window.__API_BASE__
  || (window.location.port === "8080" ? "" : `${window.location.protocol}//${window.location.hostname}:8080`);

const state = {
  staff: [],
  groups: [],
  nextGroupId: 1
};

function endpoint(path) { return API_BASE ? `${API_BASE}${path}` : path; }
function rowTemplate() { return { type: "00", startTime: "", endTime: "" }; }
function createGroup() { return { id: state.nextGroupId++, staff: "", date: "", remarks: "", expanded: false, locked: false, rows: [], existing: [], msg: "" }; }



function switchTopView(viewName){
  ["home","driver"].forEach((n)=>{
    document.getElementById(`view-${n}`)?.classList.toggle("hidden", n!==viewName);
  });
  document.querySelectorAll('.top-link').forEach((b)=>b.classList.toggle('active', b.dataset.view===viewName));
}

function startClock(){
  const el=document.getElementById('live-clock');
  if(!el) return;
  const tick=()=>{ el.textContent = new Date().toLocaleTimeString(); };
  tick();
  setInterval(tick,1000);
}
function switchTab(tabName) {
  ["ot", "staff"].forEach((n) => {
    const section = document.getElementById(`tab-${n}`);
    const button = document.getElementById(`tab-btn-${n}`);
    if (section) section.classList.toggle("hidden", n !== tabName);
    if (button) button.classList.toggle("active", n === tabName);
  });
}

function fillStaffOptions(selected) {
  return ["<option value=''>-- Select --</option>"]
    .concat(state.staff.filter((s) => (s.staffgroup || "").trim().toLowerCase() === "driver")
      .map((s) => `<option value="${s.staffid}" ${selected === s.staffid ? "selected" : ""}>${s.displayname || s.staffid} (${s.staffid})</option>`)).join("");
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
    const sec = document.createElement("section"); sec.className = "card ot-group-card";
    sec.innerHTML = `<button class="group-remove" data-action="remove" type="button" aria-label="Remove OT Input #${g.id}">×</button>
    <div class="ot-group-header">
      <h2>OT Input</h2>
      ${g.locked ? `<span class="status-pill">Locked</span>` : `<span class="status-pill status-pill--draft">Draft</span>`}
    </div>
    <div class="row ot-group-form">
      <label>Staff<select data-k="staff" ${g.locked ? "disabled" : ""}>${fillStaffOptions(g.staff)}</select></label>
      <label>Date<input data-k="date" type="date" value="${g.date}" ${g.locked ? "disabled" : ""}></label>
      <button class="btn-primary" data-action="next" type="button" ${g.locked ? "disabled" : ""}>Next</button>
      ${g.expanded ? `<label class="remarks-field">Remarks<input data-k="remarks" type="text" value="${g.remarks}" placeholder="optional"></label>` : ""}
    </div>
    <div class="msg select-msg">${g.msg || ""}</div>
    <div class="period-area ${g.expanded ? "" : "hidden"}">
      <h3>Existing Records (Read-only, can delete)</h3>
      <table><thead><tr><th>Type</th><th>Start (HH:MM)</th><th>End (HH:MM)</th><th></th></tr></thead><tbody class="existing-body"></tbody></table>
      <div class="section-head"><h3>New Records</h3><button class="btn-ghost" data-action="add-row" type="button">Add Row</button></div>
      <table><thead><tr><th>Type</th><th>Start (HH:MM)</th><th>End (HH:MM)</th><th></th></tr></thead><tbody class="entry-body"></tbody></table>
      <div class="actions"><button class="btn-primary" data-action="confirm" type="button">確認 Confirm</button></div>
      <div class="msg input-msg"></div>
    </div>`;

    sec.querySelectorAll("[data-k]").forEach((el) => el.addEventListener("change", (e) => { g[e.target.dataset.k] = e.target.value.trim(); }));
    sec.querySelector("[data-action='remove']").addEventListener("click", () => {
      if (!window.confirm("Remove this OT Input block?")) return;
      state.groups = state.groups.filter((x) => x.id !== g.id); if (!state.groups.length) state.groups = [createGroup()]; renderGroups();
    });
    sec.querySelector("[data-action='next']").addEventListener("click", async () => {
      if (!g.staff || !g.date) { g.msg = "Please select both staff and date."; renderGroups(); return; }
      const dup = state.groups.find((x) => x.id !== g.id && x.staff === g.staff && x.date === g.date);
      if (dup) { g.msg = "Same Staff + Date already exists on this page."; renderGroups(); return; }
      g.msg = ""; g.expanded = true; g.locked = true; if (g.rows.length === 0) g.rows = [rowTemplate()];
      await loadExistingRecords(g); renderGroups();
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
        if (!window.confirm("Delete this new row?")) return;
        g.rows.splice(idx, 1); renderGroups();
      });
      entryBody.appendChild(tr);
    });
    sec.querySelector("[data-action='add-row']").addEventListener("click", () => { g.rows.push(rowTemplate()); renderGroups(); });
    sec.querySelector("[data-action='confirm']").addEventListener("click", async () => { await confirmInput(g, sec.querySelector('.input-msg')); });

    root.appendChild(sec);
  });
}

async function loadStaff() { const resp = await fetch(endpoint('/api/staff')); const data = await resp.json(); state.staff = data.staff || []; renderStaffList(); renderGroups(); }
async function loadExistingRecords(g) { const resp = await fetch(endpoint(`/api/ot/entries?otstaffid=${encodeURIComponent(g.staff)}&date=${encodeURIComponent(g.date)}`)); const data = await resp.json(); g.existing = (data.entries||[]).map((e)=>({id:e.id,type:e.type,startTime:e.startTime,endTime:e.endTime})); }
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

function bindEvents(){
  document.querySelectorAll('.tab-btn').forEach((btn)=>btn.addEventListener('click', function(){switchTab(this.getAttribute('data-tab'));}));
  document.querySelectorAll('.top-link').forEach((btn)=>btn.addEventListener('click', function(){switchTopView(this.getAttribute('data-view'));}));
  document.getElementById('save-staff')?.addEventListener('click',saveStaff);
  document.getElementById('add-group')?.addEventListener('click',()=>{state.groups.push(createGroup()); renderGroups();});
}
function init(){ state.groups=[createGroup()]; bindEvents(); loadStaff(); switchTab('ot'); switchTopView('home'); startClock(); }
document.addEventListener('DOMContentLoaded', init);

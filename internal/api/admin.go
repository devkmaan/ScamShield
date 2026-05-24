package api

const adminHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>ScamShield Console</title>
  <style>
    :root {
      color-scheme: light;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      --ink: #17202a;
      --muted: #667085;
      --line: #d8dee8;
      --panel: #ffffff;
      --bg: #f4f7fb;
      --accent: #0f766e;
      --danger: #b42318;
      --warn: #b54708;
      --ok: #027a48;
    }
    * { box-sizing: border-box; }
    body { margin: 0; background: var(--bg); color: var(--ink); }
    header { background: #101828; color: #fff; padding: 16px 24px; display: flex; align-items: center; justify-content: space-between; gap: 16px; }
    header h1 { margin: 0; font-size: 20px; font-weight: 800; letter-spacing: 0; }
    header .status { display: flex; gap: 10px; align-items: center; font-size: 13px; color: #d0d5dd; }
    main { max-width: 1440px; margin: 0 auto; padding: 18px; }
    nav { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 14px; }
    nav button, .btn { border: 1px solid var(--line); background: #fff; color: var(--ink); border-radius: 6px; padding: 9px 12px; font-weight: 700; cursor: pointer; }
    nav button.active, .btn.primary { background: var(--accent); color: #fff; border-color: var(--accent); }
    .grid { display: grid; grid-template-columns: repeat(12, 1fr); gap: 14px; }
    .panel { background: var(--panel); border: 1px solid var(--line); border-radius: 8px; padding: 16px; min-width: 0; }
    .span-3 { grid-column: span 3; }
    .span-4 { grid-column: span 4; }
    .span-5 { grid-column: span 5; }
    .span-6 { grid-column: span 6; }
    .span-7 { grid-column: span 7; }
    .span-8 { grid-column: span 8; }
    .span-12 { grid-column: span 12; }
    h2 { margin: 0 0 12px; font-size: 16px; }
    h3 { margin: 0 0 8px; font-size: 13px; color: var(--muted); text-transform: uppercase; letter-spacing: 0; }
    label { display: block; font-size: 12px; font-weight: 800; color: #344054; margin: 10px 0 6px; }
    textarea, input, select { width: 100%; border: 1px solid #cbd5e1; border-radius: 6px; padding: 10px; font: inherit; background: #fff; color: var(--ink); }
    textarea { min-height: 112px; resize: vertical; }
    .row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; }
    .metric { font-size: 26px; font-weight: 850; }
    .muted { color: var(--muted); }
    .tab { display: none; }
    .tab.active { display: block; }
    .result, pre { white-space: pre-wrap; overflow: auto; background: #101828; color: #edf2f7; border-radius: 8px; padding: 12px; font-size: 13px; max-height: 420px; }
    table { width: 100%; border-collapse: collapse; font-size: 13px; }
    th, td { border-bottom: 1px solid #edf2f7; padding: 9px 8px; text-align: left; vertical-align: top; }
    th { color: #475467; font-size: 12px; }
    .pill { display: inline-flex; align-items: center; border-radius: 999px; padding: 3px 8px; font-size: 12px; font-weight: 800; border: 1px solid var(--line); }
    .CRITICAL, .HIGH_RISK { color: var(--danger); background: #fff1f0; border-color: #fecdca; }
    .CAUTION { color: var(--warn); background: #fffaeb; border-color: #fedf89; }
    .LOW { color: var(--ok); background: #ecfdf3; border-color: #abefc6; }
    @media (max-width: 920px) {
      .span-3, .span-4, .span-5, .span-6, .span-7, .span-8, .span-12 { grid-column: span 12; }
      header { align-items: flex-start; flex-direction: column; }
    }
  </style>
</head>
<body>
  <header>
    <h1>ScamShield</h1>
    <div class="status"><span id="ready">checking</span><span id="clock"></span></div>
  </header>
  <main>
    <nav>
      <button class="active" data-tab="dashboard">Dashboard</button>
      <button data-tab="check">Risk Check</button>
      <button data-tab="whatsapp">WhatsApp</button>
      <button data-tab="recovery">Recovery</button>
      <button data-tab="merchant">Merchant Risk</button>
      <button data-tab="model">Model Lab</button>
      <button data-tab="events">Events</button>
    </nav>

    <section id="dashboard" class="tab active">
      <div class="grid">
        <div class="panel span-3"><h3>Decisions</h3><div class="metric" id="mDecisions">0</div></div>
        <div class="panel span-3"><h3>High Risk</h3><div class="metric" id="mHigh">0</div></div>
        <div class="panel span-3"><h3>Review Queue</h3><div class="metric" id="mReview">0</div></div>
        <div class="panel span-3"><h3>Evidence</h3><div class="metric" id="mEvidence">0</div></div>
        <div class="panel span-8"><h2>Recent Decisions</h2><div id="decisionTable"></div></div>
        <div class="panel span-4"><h2>Top Risk Payees</h2><div id="merchantTable"></div></div>
      </div>
    </section>

    <section id="check" class="tab">
      <div class="grid">
        <div class="panel span-5">
          <h2>Manual Risk Check</h2>
          <label>Input Type</label>
          <select id="checkType"><option>TEXT</option><option>URL</option><option>UPI_ID</option><option>QR</option><option>SCREENSHOT</option></select>
          <label>Message, URL, UPI ID, or QR payload</label>
          <textarea id="checkText">Your SBI KYC is blocked. Click https://sbi-verify-support.com and share OTP immediately.</textarea>
          <div class="row" style="margin-top:12px"><button class="btn primary" onclick="runCheck()">Analyze</button><button class="btn" onclick="seedSimulation()">Generate Demo Data</button></div>
        </div>
        <div class="panel span-7"><h2>Decision</h2><pre id="checkResult">No decision yet.</pre></div>
      </div>
    </section>

    <section id="whatsapp" class="tab">
      <div class="grid">
        <div class="panel span-5">
          <h2>WhatsApp Webhook Simulator</h2>
          <label>User Phone</label><input id="waUser" value="919999999999" />
          <label>Forwarded Message</label><textarea id="waText">KYC update urgent. Click https://paytm-verify-help.com and share OTP</textarea>
          <div class="row" style="margin-top:12px"><button class="btn primary" onclick="sendWhatsApp()">Send Webhook</button><button class="btn" onclick="loadOutbox()">Load Outbox</button></div>
        </div>
        <div class="panel span-7"><h2>Bot Replies</h2><pre id="outboxResult">No replies loaded.</pre></div>
      </div>
    </section>

    <section id="recovery" class="tab">
      <div class="grid">
        <div class="panel span-5">
          <h2>Recovery Case</h2>
          <label>User ID</label><input id="recoveryUser" value="demo-user" />
          <label>Loss Context</label><textarea id="recoveryText">I already paid 5000 after a Telegram task commission message to taskbonus@paytm.</textarea>
          <div class="row" style="margin-top:12px"><button class="btn primary" onclick="startRecovery()">Create Recovery Draft</button><button class="btn" onclick="saveEvidence()">Save Evidence</button></div>
        </div>
        <div class="panel span-7"><h2>Reports and Evidence</h2><pre id="recoveryResult">No recovery case yet.</pre></div>
      </div>
    </section>

    <section id="merchant" class="tab">
      <div class="grid">
        <div class="panel span-5">
          <h2>Merchant / Payee Risk</h2>
          <label>UPI ID</label><input id="payee" value="refund-care@okaxis" />
          <label>Alias</label><input id="payeeAlias" value="Marketplace Refund Agent" />
          <div class="row" style="margin-top:12px"><button class="btn primary" onclick="observePayee()">Observe</button><button class="btn" onclick="reportPayee()">Report Scam</button></div>
        </div>
        <div class="panel span-7"><h2>Payee Profile</h2><pre id="payeeResult">No payee loaded.</pre></div>
      </div>
    </section>

    <section id="model" class="tab">
      <div class="grid">
        <div class="panel span-5">
          <h2>Model Lab</h2>
          <label>Text</label><textarea id="modelText">Part time job daily task. Pay registration fee and earn commission.</textarea>
          <div class="row" style="margin-top:12px"><button class="btn primary" onclick="scoreText()">Score Text</button><button class="btn" onclick="modelMeta()">Metadata</button></div>
        </div>
        <div class="panel span-7"><h2>Model Output</h2><pre id="modelResult">No score yet.</pre></div>
      </div>
    </section>

    <section id="events" class="tab">
      <div class="panel"><div class="row" style="justify-content:space-between"><h2>Event Log</h2><button class="btn" onclick="loadEvents()">Refresh</button></div><pre id="eventsResult">No events loaded.</pre></div>
    </section>
  </main>

  <script>
    const $ = (id) => document.getElementById(id);
    const show = (obj) => JSON.stringify(obj, null, 2);
    async function api(path, options = {}) {
      const res = await fetch(path, options);
      const text = await res.text();
      let data;
      try { data = text ? JSON.parse(text) : {}; } catch { data = text; }
      if (!res.ok) throw new Error(show(data));
      return data;
    }
    document.querySelectorAll('nav button').forEach(btn => btn.addEventListener('click', () => {
      document.querySelectorAll('nav button').forEach(b => b.classList.remove('active'));
      document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
      btn.classList.add('active');
      $(btn.dataset.tab).classList.add('active');
      refresh();
    }));
    async function refresh() {
      $('clock').textContent = new Date().toLocaleTimeString();
      try {
        const ready = await api('/ready');
        $('ready').textContent = ready.status + ' / events ' + ready.eventCount;
      } catch { $('ready').textContent = 'offline'; }
      const active = document.querySelector('.tab.active').id;
      if (active === 'dashboard') await loadDashboard();
      if (active === 'events') await loadEvents();
    }
    async function loadDashboard() {
      const s = await api('/v1/admin/summary');
      $('mDecisions').textContent = s.decisionCount;
      $('mHigh').textContent = s.highRiskCount;
      $('mReview').textContent = s.humanReviewCount;
      $('mEvidence').textContent = s.evidenceCount;
      $('decisionTable').innerHTML = table(s.recentDecisions || [], ['riskLevel','score','scamType','inputType','createdAt']);
      $('merchantTable').innerHTML = table(s.topRiskMerchants || [], ['riskScore','complaintCount','payeeHash']);
    }
    function table(items, fields) {
      if (!items.length) return '<p class="muted">No rows yet.</p>';
      return '<table><thead><tr>' + fields.map(f => '<th>' + f + '</th>').join('') + '</tr></thead><tbody>' +
        items.map(row => '<tr>' + fields.map(f => '<td>' + cell(f, row[f]) + '</td>').join('') + '</tr>').join('') + '</tbody></table>';
    }
    function cell(field, value) {
      if (value == null) return '';
      if (field === 'riskLevel') return '<span class="pill ' + value + '">' + value + '</span>';
      if (typeof value === 'number') return value.toFixed ? value.toFixed(2) : value;
      return String(value).slice(0, 96);
    }
    async function runCheck() {
      const body = { inputType: $('checkType').value, userId: 'web-user' };
      const value = $('checkText').value;
      if (body.inputType === 'URL') body.url = value;
      else if (body.inputType === 'UPI_ID') body.upiId = value;
      else if (body.inputType === 'QR') body.qrPayload = value;
      else body.text = value;
      const data = await api('/v1/check', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(body) });
      $('checkResult').textContent = show(data);
      await loadDashboard();
    }
    async function seedSimulation() {
      const data = await api('/v1/simulate/stream', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({count:12}) });
      $('checkResult').textContent = show(data);
      await loadDashboard();
    }
    async function sendWhatsApp() {
      const payload = { entry:[{ changes:[{ value:{ messages:[{ id:'wamid.web.' + Date.now(), from:$('waUser').value, type:'text', text:{ body:$('waText').value } }] } }] }] };
      const data = await api('/webhooks/whatsapp', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) });
      $('outboxResult').textContent = show(data);
      setTimeout(loadOutbox, 350);
    }
    async function loadOutbox() {
      const data = await api('/v1/outbox?userId=' + encodeURIComponent($('waUser').value));
      $('outboxResult').textContent = show(data);
    }
    async function startRecovery() {
      const data = await api('/v1/recovery/start', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({userId:$('recoveryUser').value, text:$('recoveryText').value, inputType:'TEXT'}) });
      $('recoveryResult').textContent = show(data);
    }
    async function saveEvidence() {
      const data = await api('/v1/evidence', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({userId:$('recoveryUser').value, mediaType:'TEXT', source:'WEB_CONSOLE', content:$('recoveryText').value}) });
      const all = await api('/v1/evidence');
      $('recoveryResult').textContent = show({created:data, all});
    }
    async function observePayee() {
      const data = await api('/internal/payee/observe', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({rawPayee:$('payee').value, alias:$('payeeAlias').value, source:'WEB_CONSOLE'}) });
      $('payeeResult').textContent = show(data);
    }
    async function reportPayee() {
      const data = await api('/internal/payee/report', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({rawPayee:$('payee').value, alias:$('payeeAlias').value, comment:'Reported from console'}) });
      $('payeeResult').textContent = show(data);
      await loadDashboard();
    }
    async function scoreText() {
      const data = await api('/internal/model/score-text', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({text:$('modelText').value, languageHint:'hinglish'}) });
      $('modelResult').textContent = show(data);
    }
    async function modelMeta() {
      const data = await api('/internal/model/metadata');
      $('modelResult').textContent = show(data);
    }
    async function loadEvents() {
      const data = await api('/v1/events');
      $('eventsResult').textContent = show(data);
    }
    refresh();
    setInterval(refresh, 10000);
  </script>
</body>
</html>`

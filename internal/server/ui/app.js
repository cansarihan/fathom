"use strict";

const $ = (id) => document.getElementById(id);
const token = new URLSearchParams(location.search).get("token") || "";
const auth = (p) => (token ? p + (p.includes("?") ? "&" : "?") + "token=" + encodeURIComponent(token) : p);

let targets = new Map(); // name -> state
let rate = [];           // aggregate exec/sec samples for the echogram

function fmt(n) {
  if (n >= 1e6) return (n / 1e6).toFixed(1) + "M";
  if (n >= 1e3) return (n / 1e3).toFixed(1) + "k";
  return String(n);
}

function escapeHTML(s) {
  return (s || "").replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
}

function renderTargets() {
  const box = $("targets");
  box.innerHTML = "";
  let totalExecs = 0, totalRate = 0, corpus = 0;
  for (const t of targets.values()) {
    totalExecs += t.execs || 0;
    totalRate += t.status === "running" ? (t.rate || 0) : 0;
    corpus += t.corpus || 0;
    const card = document.createElement("div");
    card.className = "card" + (t.crashed ? " crashed" : "");
    card.innerHTML =
      '<div style="display:flex;justify-content:space-between;align-items:center">' +
      '<span class="name">' + escapeHTML(t.name) + '</span>' +
      '<span class="pill ' + t.status + '">' + t.status + '</span></div>' +
      '<div class="spark">' + (t.spark || "") + '</div>' +
      '<div class="row"><span>execs <b>' + fmt(t.execs || 0) + '</b></span>' +
      '<span>exec/s <b>' + fmt(t.rate || 0) + '</b></span>' +
      '<span>corpus <b>' + (t.corpus || 0) + '</b></span></div>';
    box.appendChild(card);
  }
  $("s-execs").textContent = fmt(totalExecs);
  $("s-rate").textContent = fmt(totalRate);
  $("s-corpus").textContent = corpus;
  $("s-targets").textContent = targets.size;
}

function renderCrashes(crashes) {
  const ol = $("crashes");
  ol.innerHTML = "";
  $("empty").style.display = crashes.length ? "none" : "block";
  $("contact-count").textContent = crashes.length;
  $("s-crashes").textContent = crashes.length;
  for (const c of crashes) {
    const li = document.createElement("li");
    li.innerHTML = '<div class="sig">' + escapeHTML(c.signature) + '</div>' +
      '<div class="by">' + escapeHTML(c.target) + ' · ×' + c.count + '</div>';
    ol.appendChild(li);
  }
}

function pushSpark(t, value) {
  t.history = t.history || [];
  t.history.push(value);
  if (t.history.length > 24) t.history.shift();
  const ramp = "▁▂▃▄▅▆▇█";
  const max = Math.max(1, ...t.history);
  t.spark = t.history.map((v) => ramp[Math.min(7, Math.floor((v / max) * 7))]).join("");
}

async function start() {
  $("run").disabled = true;
  targets = new Map();
  rate = [];
  $("crashes").innerHTML = "";
  $("status").textContent = "running";

  const res = await fetch(auth("/api/v1/campaigns"), { method: "POST" });
  const campaign = await res.json();
  for (const t of campaign.targets || []) targets.set(t.target, { name: t.target, status: t.status });
  renderTargets();

  const es = new EventSource(auth("/api/v1/campaigns/" + campaign.id + "/stream"));
  es.onmessage = (ev) => {
    let msg;
    try { msg = JSON.parse(ev.data); } catch { return; }
    if (msg.type === "update") onUpdate(msg);
    else if (msg.type === "done") { es.close(); finish(campaign.id); }
  };
  es.onerror = () => { es.close(); finish(campaign.id); };
}

function onUpdate(msg) {
  const t = targets.get(msg.target) || { name: msg.target, status: "running" };
  const ev = (msg.update && msg.update.event) || {};
  if (ev.kind === "progress") {
    t.status = "running";
    t.execs = ev.execs;
    t.rate = ev.execs_per_sec;
    t.corpus = ev.corpus_total;
    pushSpark(t, ev.execs_per_sec);
    rate.push(ev.execs_per_sec);
    if (rate.length > 600) rate.shift();
  } else if (ev.kind === "crash") {
    t.crashed = true;
    t.status = "crash";
  }
  targets.set(msg.target, t);
  renderTargets();
}

async function finish(id) {
  try {
    const res = await fetch(auth("/api/v1/campaigns/" + id));
    const campaign = await res.json();
    for (const t of campaign.targets || []) {
      const cur = targets.get(t.target) || { name: t.target };
      Object.assign(cur, { status: t.status, execs: t.execs, rate: t.execs_per_sec, corpus: t.corpus, crashed: t.crashed });
      targets.set(t.target, cur);
    }
    renderTargets();
    renderCrashes(campaign.crashes || []);
    $("status").textContent = campaign.status === "error" ? "error" : "done";
    if (campaign.err) $("meta").textContent = "error: " + campaign.err;
  } finally {
    $("run").disabled = false;
  }
}

// The echogram scrolls the aggregate exec/sec as a depth sounder trace: brighter
// and taller where the engine runs hot, fading as it scrolls left.
const canvas = $("echogram");
const ctx = canvas.getContext("2d");
let W = 0, H = 0, dpr = 1;

function resize() {
  dpr = Math.max(1, window.devicePixelRatio || 1);
  const r = canvas.getBoundingClientRect();
  W = r.width; H = r.height;
  canvas.width = Math.floor(W * dpr);
  canvas.height = Math.floor(H * dpr);
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
}

function drawEchogram() {
  ctx.clearRect(0, 0, W, H);
  // faint depth gridlines
  ctx.strokeStyle = "rgba(21,48,60,0.8)";
  ctx.lineWidth = 1;
  for (let i = 1; i < 4; i++) {
    const y = (H * i) / 4;
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke();
  }
  const samples = rate.slice(-Math.floor(W));
  const max = Math.max(1, ...samples);
  const step = W / Math.max(1, samples.length);
  // filled trace
  ctx.beginPath();
  ctx.moveTo(0, H);
  samples.forEach((v, i) => {
    const x = i * step;
    const y = H - (v / max) * (H - 12);
    ctx.lineTo(x, y);
  });
  ctx.lineTo(samples.length * step, H);
  const grad = ctx.createLinearGradient(0, 0, 0, H);
  grad.addColorStop(0, "rgba(57,217,138,0.35)");
  grad.addColorStop(1, "rgba(57,217,138,0.02)");
  ctx.fillStyle = grad;
  ctx.fill();
  // bright leading edge
  ctx.beginPath();
  samples.forEach((v, i) => {
    const x = i * step;
    const y = H - (v / max) * (H - 12);
    if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
  });
  ctx.strokeStyle = "#39d98a";
  ctx.lineWidth = 1.6;
  ctx.shadowColor = "#39d98a";
  ctx.shadowBlur = 8;
  ctx.stroke();
  ctx.shadowBlur = 0;
  requestAnimationFrame(drawEchogram);
}

async function init() {
  resize();
  window.addEventListener("resize", resize);
  requestAnimationFrame(drawEchogram);
  $("run").addEventListener("click", start);
  try {
    const st = await (await fetch(auth("/api/v1/status"))).json();
    $("meta").textContent = (st.dir || "") + " · " + (st.duration || "") + "/target · fathom " + (st.version || "");
  } catch (e) { /* best effort */ }
  start();
}

init();

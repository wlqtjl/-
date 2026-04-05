/* ===== app.js — 我在 · WoZai Digital Soul ===== */
(function () {
  "use strict";

  /* ---------- DOM refs ---------- */
  const $ = (s) => document.querySelector(s);
  const screens = {
    home: $("#s-home"),
    auth: $("#s-auth"),
    list: $("#s-list"),
    create: $("#s-create"),
    chat: $("#s-chat"),
  };
  const ov = $("#ov"),
    ovTxt = $("#ov-txt"),
    toastEl = $("#toast"),
    ap = $("#ap");

  /* ---------- Token helpers ---------- */
  function getAccessToken() {
    return localStorage.getItem("wz_access");
  }
  function getRefreshToken() {
    return localStorage.getItem("wz_refresh");
  }
  function saveTokens(t) {
    localStorage.setItem("wz_access", t.access_token);
    localStorage.setItem("wz_refresh", t.refresh_token);
  }
  function clearTokens() {
    localStorage.removeItem("wz_access");
    localStorage.removeItem("wz_refresh");
  }

  /* ---------- API helper ---------- */
  async function api(path, opts) {
    opts = opts || {};
    const headers = opts.headers || {};
    const token = getAccessToken();
    if (token) headers["Authorization"] = "Bearer " + token;
    if (!(opts.body instanceof FormData) && opts.body && typeof opts.body === "object") {
      headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(opts.body);
    }
    opts.headers = headers;

    let res = await fetch("/api/v1" + path, opts);

    // auto refresh on 401
    if (res.status === 401 && getRefreshToken()) {
      const refreshRes = await fetch("/api/v1/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: getRefreshToken() }),
      });
      if (refreshRes.ok) {
        const data = await refreshRes.json();
        saveTokens(data);
        headers["Authorization"] = "Bearer " + data.access_token;
        res = await fetch("/api/v1" + path, opts);
      } else {
        clearTokens();
        showS("auth");
        throw new Error("session_expired");
      }
    }
    return res;
  }

  /* ---------- Screen navigation ---------- */
  function showS(name) {
    Object.values(screens).forEach((s) => s.classList.remove("active"));
    if (screens[name]) screens[name].classList.add("active");
    if (name === "list") loadSouls();
    if (name === "chat" && currentSoul) loadHistory();
  }

  /* ---------- Particles ---------- */
  function initParticles() {
    const c = $("#particles");
    for (let i = 0; i < 18; i++) {
      const d = document.createElement("div");
      d.className = "pt";
      const sz = Math.random() * 3 + 1.5;
      Object.assign(d.style, {
        width: sz + "px",
        height: sz + "px",
        left: Math.random() * 100 + "%",
        background:
          Math.random() > 0.5
            ? "var(--amber)"
            : "rgba(196,133,58," + (0.15 + Math.random() * 0.2) + ")",
        animationDuration: 8 + Math.random() * 12 + "s",
        animationDelay: Math.random() * 10 + "s",
      });
      c.appendChild(d);
    }
  }

  /* ---------- Toast ---------- */
  let toastTimer;
  function toast(msg) {
    toastEl.textContent = msg;
    toastEl.classList.add("show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => toastEl.classList.remove("show"), 2800);
  }

  /* ---------- Loading overlay ---------- */
  function showOv(msg) {
    ovTxt.textContent = msg || "加载中…";
    ov.classList.add("show");
  }
  function hideOv() {
    ov.classList.remove("show");
  }

  /* ================================================================
   *  AUTH
   * ================================================================ */
  // Auth tabs
  document.querySelectorAll(".auth-tab").forEach((tab) => {
    tab.addEventListener("click", () => {
      document.querySelectorAll(".auth-tab").forEach((t) => t.classList.remove("active"));
      tab.classList.add("active");
      document.querySelectorAll(".auth-form").forEach((f) => f.classList.remove("active"));
      const target = tab.dataset.tab === "login" ? "#form-login" : "#form-register";
      $(target).classList.add("active");
    });
  });

  // Login
  $("#form-login").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = $("#login-err");
    errEl.textContent = "";
    const email = $("#login-email").value.trim();
    const pass = $("#login-pass").value;
    if (!email || !pass) {
      errEl.textContent = "请填写完整";
      return;
    }
    try {
      const res = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email, password: pass }),
      });
      const data = await res.json();
      if (!res.ok) {
        errEl.textContent = data.error || "登录失败";
        return;
      }
      saveTokens(data.tokens);
      toast("欢迎回来");
      showS("list");
    } catch (err) {
      errEl.textContent = "网络错误";
    }
  });

  // Register
  $("#form-register").addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = $("#reg-err");
    errEl.textContent = "";
    const email = $("#reg-email").value.trim();
    const nick = $("#reg-nick").value.trim();
    const pass = $("#reg-pass").value;
    if (!email || !nick || !pass) {
      errEl.textContent = "请填写完整";
      return;
    }
    if (pass.length < 8) {
      errEl.textContent = "密码至少8位";
      return;
    }
    try {
      const res = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email, nickname: nick, password: pass }),
      });
      const data = await res.json();
      if (!res.ok) {
        errEl.textContent = data.error || "注册失败";
        return;
      }
      saveTokens(data.tokens);
      toast("注册成功，欢迎你");
      showS("list");
    } catch (err) {
      errEl.textContent = "网络错误";
    }
  });

  /* ================================================================
   *  SOULS LIST
   * ================================================================ */
  async function loadSouls() {
    const list = $("#soul-list");
    list.innerHTML = "";
    try {
      const res = await api("/souls");
      if (!res.ok) throw new Error("load fail");
      const data = await res.json();
      const souls = data.souls || [];
      if (souls.length === 0) {
        list.innerHTML =
          '<div class="ph" style="padding:20px 0"><p>还没有数字灵魂<br>点击下方按钮，创建第一个</p></div>';
        return;
      }
      souls.forEach((s) => {
        const card = document.createElement("div");
        card.className = "card";
        card.innerHTML =
          '<div class="card-av">' +
          (s.name ? s.name[0] : "灵") +
          "</div>" +
          '<div class="card-info">' +
          "<h3>" + escHTML(s.name) + "</h3>" +
          "<p>" + escHTML(s.relation || "数字灵魂") + "</p>" +
          "</div>" +
          '<div class="card-actions">' +
          '<button class="card-btn card-del" title="删除">✕</button>' +
          "</div>";
        card.querySelector(".card-info").addEventListener("click", () => openChat(s));
        card.querySelector(".card-del").addEventListener("click", (e) => {
          e.stopPropagation();
          deleteSoul(s.id);
        });
        list.appendChild(card);
      });
    } catch (err) {
      if (err.message !== "session_expired") toast("加载失败");
    }
  }

  async function deleteSoul(id) {
    if (!confirm("确定要删除这个数字灵魂吗？")) return;
    try {
      const res = await api("/souls/" + id, { method: "DELETE" });
      if (res.ok || res.status === 204) {
        toast("已删除");
        loadSouls();
      } else {
        toast("删除失败");
      }
    } catch {
      toast("删除失败");
    }
  }

  /* ================================================================
   *  CREATE SOUL
   * ================================================================ */
  const stepsEl = document.querySelectorAll(".sdot");

  function updateStep(idx) {
    stepsEl.forEach((d, i) => d.classList.toggle("on", i <= idx));
  }

  $("#form-create").addEventListener("submit", async (e) => {
    e.preventDefault();
    const name = $("#f-name").value.trim();
    if (!name) {
      toast("请输入名字");
      return;
    }
    showOv("正在创建灵魂…");
    try {
      const res = await api("/souls", {
        method: "POST",
        body: {
          name: name,
          relation: $("#f-rel").value.trim(),
          personality: $("#f-pers").value.trim(),
          speech_style: $("#f-speech").value.trim(),
          memory: $("#f-mem").value.trim(),
        },
      });
      hideOv();
      if (!res.ok) {
        const data = await res.json();
        toast(data.error || "创建失败");
        return;
      }
      toast("灵魂已创建");
      $("#form-create").reset();
      updateStep(0);
      showS("list");
    } catch {
      hideOv();
      toast("创建失败");
    }
  });

  // Step indicator on input focus
  const fieldStep = { "f-name": 0, "f-rel": 0, "f-pers": 1, "f-speech": 1, "f-mem": 2 };
  Object.keys(fieldStep).forEach((id) => {
    const el = $("#" + id);
    if (el) el.addEventListener("focus", () => updateStep(fieldStep[id]));
  });

  /* ================================================================
   *  CHAT
   * ================================================================ */
  let currentSoul = null;
  let sending = false;

  function openChat(soul) {
    currentSoul = soul;
    $("#c-name").textContent = soul.name;
    $("#c-rel").textContent = soul.relation || "数字灵魂";
    $("#c-av").textContent = soul.name ? soul.name[0] : "灵";
    $("#msgs").innerHTML = "";
    showS("chat");
  }

  async function loadHistory() {
    if (!currentSoul) return;
    try {
      const res = await api("/souls/" + currentSoul.id + "/messages?limit=50");
      if (!res.ok) return;
      const data = await res.json();
      const msgs = data.messages || [];
      const container = $("#msgs");
      container.innerHTML = "";
      if (msgs.length === 0) {
        addMsg("assistant", "你好，" + (currentSoul.relation ? "我是" + currentSoul.name : "我在这里") + "。想对我说些什么？");
      } else {
        msgs.forEach((m) => addMsg(m.role, m.content));
      }
    } catch {
      // silently fail
    }
  }

  function addMsg(role, text) {
    const d = document.createElement("div");
    d.className = "msg " + (role === "user" ? "msg-u" : "msg-a");
    const bub = document.createElement("div");
    bub.className = "bub";
    bub.textContent = text;
    d.appendChild(bub);

    if (role === "assistant") {
      const speakBtn = document.createElement("button");
      speakBtn.className = "btn-speak";
      speakBtn.textContent = "🔊";
      speakBtn.title = "听TA说";
      speakBtn.addEventListener("click", () => speak(text));
      d.appendChild(speakBtn);
    }

    const container = $("#msgs");
    container.appendChild(d);
    container.scrollTop = container.scrollHeight;
    return bub;
  }

  function showTyping() {
    const d = document.createElement("div");
    d.className = "msg msg-a";
    d.id = "typing-indicator";
    const bub = document.createElement("div");
    bub.className = "bub typing";
    bub.innerHTML = "<span></span><span></span><span></span>";
    d.appendChild(bub);
    const container = $("#msgs");
    container.appendChild(d);
    container.scrollTop = container.scrollHeight;
    return d;
  }

  function removeTyping() {
    const t = $("#typing-indicator");
    if (t) t.remove();
  }

  async function sendMsg() {
    if (sending || !currentSoul) return;
    const input = $("#ci");
    const text = input.value.trim();
    if (!text) return;

    sending = true;
    input.value = "";
    resizeInput();
    addMsg("user", text);
    showTyping();

    try {
      const res = await api("/souls/" + currentSoul.id + "/chat", {
        method: "POST",
        body: { message: text },
      });
      removeTyping();
      if (!res.ok) {
        const data = await res.json();
        addMsg("assistant", data.error || "抱歉，我暂时无法回应…");
        sending = false;
        return;
      }
      const data = await res.json();
      typeMsg(data.reply || "…");
    } catch {
      removeTyping();
      addMsg("assistant", "网络波动，请稍后再试…");
    }
    sending = false;
  }

  /* Typing animation for assistant reply */
  function typeMsg(text) {
    const d = document.createElement("div");
    d.className = "msg msg-a";
    const bub = document.createElement("div");
    bub.className = "bub";
    d.appendChild(bub);

    const speakBtn = document.createElement("button");
    speakBtn.className = "btn-speak";
    speakBtn.textContent = "🔊";
    speakBtn.title = "听TA说";
    speakBtn.style.display = "none";
    speakBtn.addEventListener("click", () => speak(text));
    d.appendChild(speakBtn);

    const container = $("#msgs");
    container.appendChild(d);

    let idx = 0;
    const step = () => {
      if (idx < text.length) {
        bub.textContent += text[idx++];
        container.scrollTop = container.scrollHeight;
        setTimeout(step, 30 + Math.random() * 40);
      } else {
        speakBtn.style.display = "";
      }
    };
    step();
  }

  /* ---------- TTS ---------- */
  async function speak(text) {
    if (!currentSoul) return;
    try {
      toast("正在生成语音…");
      const res = await api("/souls/" + currentSoul.id + "/speak", {
        method: "POST",
        body: { text: text, voice: "male_1" },
      });
      if (!res.ok) {
        toast("语音生成失败");
        return;
      }
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      ap.src = url;
      ap.play().catch(() => {});
      ap.onended = () => URL.revokeObjectURL(url);
    } catch {
      toast("语音生成失败");
    }
  }

  /* ---------- Input resize ---------- */
  function resizeInput() {
    const ci = $("#ci");
    ci.style.height = "auto";
    ci.style.height = Math.min(ci.scrollHeight, 120) + "px";
  }

  function handleKey(e) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMsg();
    }
  }

  /* ---------- HTML escape ---------- */
  function escHTML(s) {
    const d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }

  /* ================================================================
   *  INIT & EVENT BINDINGS
   * ================================================================ */
  initParticles();

  // Home → Auth (or list if already logged in)
  $("#btn-start").addEventListener("click", () => {
    if (getAccessToken()) {
      showS("list");
    } else {
      showS("auth");
    }
  });

  // Auth back
  $("#btn-auth-back").addEventListener("click", () => showS("home"));

  // List page
  $("#btn-to-create").addEventListener("click", () => showS("create"));
  $("#btn-logout").addEventListener("click", () => {
    clearTokens();
    toast("已退出");
    showS("home");
  });

  // Create page
  $("#btn-create-cancel").addEventListener("click", () => {
    $("#form-create").reset();
    updateStep(0);
    showS("list");
  });

  // Chat page
  $("#btn-chat-back").addEventListener("click", () => {
    currentSoul = null;
    showS("list");
  });
  $("#btn-send").addEventListener("click", sendMsg);
  $("#ci").addEventListener("input", resizeInput);
  $("#ci").addEventListener("keydown", handleKey);

  // Auto-redirect if token exists
  if (getAccessToken()) {
    // stay on home, user clicks 开始
  }
})();

/* ===== app.js — 我在 · WoZai Digital Soul (Enhanced) ===== */
(function () {
  "use strict";

  /* ---------- DOM refs ---------- */
  const $ = (s) => document.querySelector(s);
  const screens = {
    home: $("#s-home"),
    auth: $("#s-auth"),
    list: $("#s-list"),
    profile: $("#s-profile"),
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

  /* ---------- Dark mode ---------- */
  function initTheme() {
    const saved = localStorage.getItem("wz_theme");
    if (saved === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    }
  }
  function toggleDark() {
    const isDark =
      document.documentElement.getAttribute("data-theme") === "dark";
    if (isDark) {
      document.documentElement.removeAttribute("data-theme");
      localStorage.setItem("wz_theme", "light");
    } else {
      document.documentElement.setAttribute("data-theme", "dark");
      localStorage.setItem("wz_theme", "dark");
    }
  }

  /* ---------- PWA Service Worker ---------- */
  function initPWA() {
    if ("serviceWorker" in navigator) {
      navigator.serviceWorker.register("/sw.js").catch(function () {});
    }
  }

  /* ---------- API helper ---------- */
  async function api(path, opts) {
    opts = opts || {};
    const headers = opts.headers || {};
    const token = getAccessToken();
    if (token) headers["Authorization"] = "Bearer " + token;
    if (
      !(opts.body instanceof FormData) &&
      opts.body &&
      typeof opts.body === "object"
    ) {
      headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(opts.body);
    }
    opts.headers = headers;

    var res = await fetch("/api/v1" + path, opts);

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
    if (name === "list") {
      loadSouls();
      loadUserStats();
    }
    if (name === "chat" && currentSoul) loadHistory();
    if (name === "profile") loadProfile();
  }

  /* ---------- Particles ---------- */
  function initParticles() {
    const c = $("#particles");
    for (var i = 0; i < 18; i++) {
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
  var toastTimer;
  function toast(msg) {
    toastEl.textContent = msg;
    toastEl.classList.add("show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () {
      toastEl.classList.remove("show");
    }, 2800);
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
  document.querySelectorAll(".auth-tab").forEach(function (tab) {
    tab.addEventListener("click", function () {
      document
        .querySelectorAll(".auth-tab")
        .forEach(function (t) { t.classList.remove("active"); });
      tab.classList.add("active");
      document
        .querySelectorAll(".auth-form")
        .forEach(function (f) { f.classList.remove("active"); });
      var target =
        tab.dataset.tab === "login" ? "#form-login" : "#form-register";
      $(target).classList.add("active");
    });
  });

  // Login
  $("#form-login").addEventListener("submit", async function (e) {
    e.preventDefault();
    var errEl = $("#login-err");
    errEl.textContent = "";
    var email = $("#login-email").value.trim();
    var pass = $("#login-pass").value;
    if (!email || !pass) {
      errEl.textContent = "请填写完整";
      return;
    }
    try {
      var res = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email, password: pass }),
      });
      var data = await res.json();
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
  $("#form-register").addEventListener("submit", async function (e) {
    e.preventDefault();
    var errEl = $("#reg-err");
    errEl.textContent = "";
    var email = $("#reg-email").value.trim();
    var nick = $("#reg-nick").value.trim();
    var pass = $("#reg-pass").value;
    if (!email || !nick || !pass) {
      errEl.textContent = "请填写完整";
      return;
    }
    if (pass.length < 8) {
      errEl.textContent = "密码至少8位";
      return;
    }
    try {
      var res = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email, nickname: nick, password: pass }),
      });
      var data = await res.json();
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
   *  USER PROFILE
   * ================================================================ */
  async function loadProfile() {
    try {
      var res = await api("/profile");
      if (!res.ok) return;
      var data = await res.json();
      var user = data.user;
      if (user) {
        $("#p-nick").value = user.nickname || "";
        $("#p-avatar").value = user.avatar || "";
        $("#p-bio").value = user.bio || "";
      }
    } catch (err) {
      // silently fail
    }
  }

  $("#form-profile").addEventListener("submit", async function (e) {
    e.preventDefault();
    try {
      var res = await api("/profile", {
        method: "PUT",
        body: {
          nickname: $("#p-nick").value.trim(),
          avatar: $("#p-avatar").value.trim(),
          bio: $("#p-bio").value.trim(),
        },
      });
      if (res.ok) {
        toast("资料已更新");
        showS("list");
      } else {
        toast("更新失败");
      }
    } catch (err) {
      toast("更新失败");
    }
  });

  /* ================================================================
   *  USER STATS
   * ================================================================ */
  async function loadUserStats() {
    try {
      var res = await api("/stats");
      if (!res.ok) return;
      var data = await res.json();
      var container = $("#user-stats");
      container.innerHTML =
        '<div class="stat-card"><div class="stat-num">' +
        (data.souls_count || 0) +
        '</div><div class="stat-label">灵魂</div></div>' +
        '<div class="stat-card"><div class="stat-num">' +
        (data.messages_count || 0) +
        '</div><div class="stat-label">对话</div></div>';
    } catch (err) {
      // silently fail
    }
  }

  /* ================================================================
   *  SOULS LIST
   * ================================================================ */
  async function loadSouls() {
    var list = $("#soul-list");
    list.innerHTML = "";
    try {
      var res = await api("/souls");
      if (!res.ok) throw new Error("load fail");
      var data = await res.json();
      var souls = data.souls || [];
      if (souls.length === 0) {
        list.innerHTML =
          '<div class="ph" style="padding:20px 0"><p>还没有数字灵魂<br>点击下方按钮，创建第一个</p></div>';
        return;
      }
      souls.forEach(function (s) {
        var card = document.createElement("div");
        card.className = "card";
        card.innerHTML =
          '<div class="card-av">' +
          (s.name ? s.name[0] : "灵") +
          "</div>" +
          '<div class="card-info">' +
          "<h3>" +
          escHTML(s.name) +
          "</h3>" +
          "<p>" +
          escHTML(s.relation || "数字灵魂") +
          "</p>" +
          "</div>" +
          '<div class="card-actions">' +
          '<button class="card-btn card-del" title="删除">✕</button>' +
          "</div>";
        card.querySelector(".card-info").addEventListener("click", function () {
          openChat(s);
        });
        card
          .querySelector(".card-del")
          .addEventListener("click", function (e) {
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
      var res = await api("/souls/" + id, { method: "DELETE" });
      if (res.ok || res.status === 204) {
        toast("已删除");
        loadSouls();
      } else {
        toast("删除失败");
      }
    } catch (err) {
      toast("删除失败");
    }
  }

  /* ================================================================
   *  CREATE SOUL
   * ================================================================ */
  var stepsEl = document.querySelectorAll(".sdot");

  function updateStep(idx) {
    stepsEl.forEach(function (d, i) {
      d.classList.toggle("on", i <= idx);
    });
  }

  $("#form-create").addEventListener("submit", async function (e) {
    e.preventDefault();
    var name = $("#f-name").value.trim();
    if (!name) {
      toast("请输入名字");
      return;
    }
    showOv("正在创建灵魂…");
    try {
      var res = await api("/souls", {
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
        var data = await res.json();
        toast(data.error || "创建失败");
        return;
      }
      toast("灵魂已创建");
      $("#form-create").reset();
      updateStep(0);
      showS("list");
    } catch (err) {
      hideOv();
      toast("创建失败");
    }
  });

  // Step indicator on input focus
  var fieldStep = { "f-name": 0, "f-rel": 0, "f-pers": 1, "f-speech": 1, "f-mem": 2 };
  Object.keys(fieldStep).forEach(function (id) {
    var el = $("#" + id);
    if (el)
      el.addEventListener("focus", function () {
        updateStep(fieldStep[id]);
      });
  });

  /* ================================================================
   *  CHAT
   * ================================================================ */
  var currentSoul = null;
  var sending = false;

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
      var res = await api(
        "/souls/" + currentSoul.id + "/messages?limit=50"
      );
      if (!res.ok) return;
      var data = await res.json();
      var msgs = data.messages || [];
      var container = $("#msgs");
      container.innerHTML = "";
      if (msgs.length === 0) {
        addMsg(
          "assistant",
          "你好，" +
            (currentSoul.relation
              ? "我是" + currentSoul.name
              : "我在这里") +
            "。想对我说些什么？"
        );
      } else {
        msgs.forEach(function (m) {
          addMsg(m.role, m.content);
        });
      }
    } catch (err) {
      // silently fail
    }
  }

  function addMsg(role, text) {
    var d = document.createElement("div");
    d.className = "msg " + (role === "user" ? "msg-u" : "msg-a");
    var bub = document.createElement("div");
    bub.className = "bub";
    bub.textContent = text;
    d.appendChild(bub);

    if (role === "assistant") {
      var speakBtn = document.createElement("button");
      speakBtn.className = "btn-speak";
      speakBtn.textContent = "🔊";
      speakBtn.title = "听TA说";
      speakBtn.addEventListener("click", function () {
        speak(text);
      });
      d.appendChild(speakBtn);
    }

    var container = $("#msgs");
    container.appendChild(d);
    container.scrollTop = container.scrollHeight;
    return bub;
  }

  function showTyping() {
    var d = document.createElement("div");
    d.className = "msg msg-a";
    d.id = "typing-indicator";
    var bub = document.createElement("div");
    bub.className = "bub typing";
    bub.innerHTML = "<span></span><span></span><span></span>";
    d.appendChild(bub);
    var container = $("#msgs");
    container.appendChild(d);
    container.scrollTop = container.scrollHeight;
    return d;
  }

  function removeTyping() {
    var t = $("#typing-indicator");
    if (t) t.remove();
  }

  async function sendMsg() {
    if (sending || !currentSoul) return;
    var input = $("#ci");
    var text = input.value.trim();
    if (!text) return;

    sending = true;
    input.value = "";
    resizeInput();
    addMsg("user", text);
    showTyping();

    try {
      var res = await api("/souls/" + currentSoul.id + "/chat", {
        method: "POST",
        body: { message: text },
      });
      removeTyping();
      if (!res.ok) {
        var data = await res.json();
        addMsg("assistant", data.error || "抱歉，我暂时无法回应…");
        sending = false;
        return;
      }
      var data = await res.json();
      typeMsg(data.reply || "…");
    } catch (err) {
      removeTyping();
      addMsg("assistant", "网络波动，请稍后再试…");
    }
    sending = false;
  }

  /* Typing animation for assistant reply */
  function typeMsg(text) {
    var d = document.createElement("div");
    d.className = "msg msg-a";
    var bub = document.createElement("div");
    bub.className = "bub";
    d.appendChild(bub);

    var speakBtn = document.createElement("button");
    speakBtn.className = "btn-speak";
    speakBtn.textContent = "🔊";
    speakBtn.title = "听TA说";
    speakBtn.style.display = "none";
    speakBtn.addEventListener("click", function () {
      speak(text);
    });
    d.appendChild(speakBtn);

    var container = $("#msgs");
    container.appendChild(d);

    var idx = 0;
    var step = function () {
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
      var res = await api("/souls/" + currentSoul.id + "/speak", {
        method: "POST",
        body: { text: text, voice: "male_1" },
      });
      if (!res.ok) {
        toast("语音生成失败");
        return;
      }
      var blob = await res.blob();
      var url = URL.createObjectURL(blob);
      ap.src = url;
      ap.play().catch(function () {});
      ap.onended = function () {
        URL.revokeObjectURL(url);
      };
    } catch (err) {
      toast("语音生成失败");
    }
  }

  /* ---------- Voice Input (Web Speech API) ---------- */
  var recognition = null;
  var isRecording = false;

  function initVoiceInput() {
    var SpeechRecognition =
      window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SpeechRecognition) {
      var micBtn = $("#btn-mic");
      if (micBtn) micBtn.style.display = "none";
      return;
    }

    recognition = new SpeechRecognition();
    recognition.lang = "zh-CN";
    recognition.interimResults = true;
    recognition.continuous = false;

    recognition.onresult = function (event) {
      var transcript = "";
      for (var i = event.resultIndex; i < event.results.length; i++) {
        transcript += event.results[i][0].transcript;
      }
      var ci = $("#ci");
      ci.value = transcript;
      resizeInput();
    };

    recognition.onend = function () {
      isRecording = false;
      var micBtn = $("#btn-mic");
      if (micBtn) micBtn.classList.remove("recording");
    };

    recognition.onerror = function () {
      isRecording = false;
      var micBtn = $("#btn-mic");
      if (micBtn) micBtn.classList.remove("recording");
    };
  }

  function toggleVoiceInput() {
    if (!recognition) {
      toast("你的浏览器不支持语音输入");
      return;
    }
    if (isRecording) {
      recognition.stop();
      isRecording = false;
      $("#btn-mic").classList.remove("recording");
    } else {
      recognition.start();
      isRecording = true;
      $("#btn-mic").classList.add("recording");
      toast("请说话…");
    }
  }

  /* ---------- Input resize ---------- */
  function resizeInput() {
    var ci = $("#ci");
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
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }

  /* ================================================================
   *  INIT & EVENT BINDINGS
   * ================================================================ */
  initTheme();
  initPWA();
  initParticles();
  initVoiceInput();

  // Home → Auth (or list if already logged in)
  $("#btn-start").addEventListener("click", function () {
    if (getAccessToken()) {
      showS("list");
    } else {
      showS("auth");
    }
  });

  // Auth back
  $("#btn-auth-back").addEventListener("click", function () {
    showS("home");
  });

  // List page
  $("#btn-to-create").addEventListener("click", function () {
    showS("create");
  });
  $("#btn-logout").addEventListener("click", function () {
    clearTokens();
    toast("已退出");
    showS("home");
  });

  // Profile
  $("#btn-profile").addEventListener("click", function () {
    showS("profile");
  });
  $("#btn-profile-back").addEventListener("click", function () {
    showS("list");
  });

  // Dark mode toggle
  $("#btn-dark").addEventListener("click", toggleDark);

  // Create page
  $("#btn-create-cancel").addEventListener("click", function () {
    $("#form-create").reset();
    updateStep(0);
    showS("list");
  });

  // Chat page
  $("#btn-chat-back").addEventListener("click", function () {
    currentSoul = null;
    showS("list");
  });
  $("#btn-send").addEventListener("click", sendMsg);
  $("#ci").addEventListener("input", resizeInput);
  $("#ci").addEventListener("keydown", handleKey);

  // Voice input
  var micBtn = $("#btn-mic");
  if (micBtn) micBtn.addEventListener("click", toggleVoiceInput);

  // Auto-redirect if token exists
  if (getAccessToken()) {
    // stay on home, user clicks 开始
  }
})();

/* ================================================================
   GCIT UI — toast notifications + custom confirm/alert dialogs
   ================================================================ */

// ── Toast ─────────────────────────────────────────────────────────

function _getToastContainer() {
    let el = document.getElementById("gcit-toasts");
    if (!el) {
        el = document.createElement("div");
        el.id = "gcit-toasts";
        (document.body || document.documentElement).appendChild(el);
    }
    return el;
}

function gcitToast(message, type = "success", duration = 3800) {
    const themes = {
        success: { bg: "#d8f3dc", border: "#52b788", color: "#174d2a", icon: "✓" },
        error:   { bg: "#fde8e8", border: "#e63946", color: "#7a1a1a", icon: "✕" },
        info:    { bg: "#fff3cd", border: "#e9c46a", color: "#6b4600", icon: "!" },
    };
    const t = themes[type] || themes.success;

    const container = _getToastContainer();
    const el = document.createElement("div");
    el.className = "gcit-toast";
    el.style.cssText = `background:${t.bg};border:1px solid ${t.border};color:${t.color};`;
    el.innerHTML = `<span class="gcit-toast-icon">${t.icon}</span><span>${message}</span>`;
    el.addEventListener("click", () => _dismissToast(el));
    container.appendChild(el);

    // force a reflow so the browser registers the initial hidden state,
    // then add the --in class to trigger the CSS transition
    void el.offsetHeight;
    el.classList.add("gcit-toast--in");

    const timer = setTimeout(() => _dismissToast(el), duration);
    el._gcitTimer = timer;
}

function _dismissToast(el) {
    clearTimeout(el._gcitTimer);
    el.classList.remove("gcit-toast--in");
    el.classList.add("gcit-toast--out");
    el.addEventListener("transitionend", () => el.remove(), { once: true });
}

// ── Confirm ────────────────────────────────────────────────────────

function gcitConfirm({ title = "Are you sure?", message = "", confirmText = "Confirm", cancelText = "Cancel", danger = false } = {}) {
    return new Promise(resolve => {
        const overlay = document.createElement("div");
        overlay.className = "gcit-overlay";

        const card = document.createElement("div");
        card.className = "gcit-dialog";

        const confirmBtnCls = danger ? "gcit-btn-danger" : "gcit-btn-confirm";

        card.innerHTML = `
            <p class="gcit-dialog-title">${title}</p>
            ${message ? `<p class="gcit-dialog-msg">${message}</p>` : ""}
            <div class="gcit-dialog-actions">
                <button class="gcit-btn-cancel" data-action="cancel">${cancelText}</button>
                <button class="${confirmBtnCls}" data-action="confirm">${confirmText}</button>
            </div>
        `;
        overlay.appendChild(card);
        document.body.appendChild(overlay);
        document.body.style.overflow = "hidden";

        // force reflow then animate in
        void overlay.offsetHeight;
        overlay.classList.add("gcit-overlay--in");
        card.classList.add("gcit-dialog--in");

        function close(result) {
            document.body.style.overflow = "";
            overlay.classList.remove("gcit-overlay--in");
            card.classList.remove("gcit-dialog--in");
            overlay.classList.add("gcit-overlay--out");
            setTimeout(() => overlay.remove(), 220);
            resolve(result);
        }

        card.querySelector("[data-action='confirm']").onclick = () => close(true);
        card.querySelector("[data-action='cancel']").onclick  = () => close(false);
        overlay.addEventListener("click", e => { if (e.target === overlay) close(false); });

        function onKey(e) {
            if (e.key === "Escape") { close(false); document.removeEventListener("keydown", onKey); }
            if (e.key === "Enter")  { close(true);  document.removeEventListener("keydown", onKey); }
        }
        document.addEventListener("keydown", onKey);

        setTimeout(() => card.querySelector("[data-action='confirm']")?.focus(), 60);
    });
}

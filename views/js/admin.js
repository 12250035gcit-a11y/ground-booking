/* admin.js — auth helpers only (login page + logout button) */

function adminLogin() {
    const data = {
        email:    document.getElementById("email").value.trim(),
        password: document.getElementById("password").value.trim(),
    };

    if (!data.email || !data.password) {
        showError("Please fill all fields");
        return;
    }

    fetch("/admin/login", {
        method: "POST",
        body: JSON.stringify(data),
        headers: { "Content-Type": "application/json" },
    })
    .then(res => res.json().then(result => ({ ok: res.ok, result })))
    .then(({ ok, result }) => {
        if (ok && result.message === "admin login success") {
            localStorage.setItem("admin", JSON.stringify(result.admin));
            window.location.href = "admin.html";
        } else {
            showError("Invalid admin credentials");
        }
    })
    .catch(() => showError("Server error"));
}

function adminLogout() {
    localStorage.removeItem("admin");
    window.location.href = "admin-login.html";
}

function showError(msg) {
    const el = document.getElementById("errorMsg");
    if (el) {
        el.textContent = msg;
        el.style.display = "block";
    } else if (typeof gcitToast === "function") {
        gcitToast(msg, "error");
    }
}

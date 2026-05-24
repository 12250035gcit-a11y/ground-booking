function login() {
    const data = {
        email: document.getElementById("email").value.trim(),
        password: document.getElementById("password").value.trim()
    };

    if (!data.email || !data.password) {
        showError("Please fill all fields");
        return;
    }

    fetch("/user/login", {
        method: "POST",
        body: JSON.stringify(data),
        headers: { "Content-Type": "application/json" }
    })
    .then(res => res.json().then(result => ({ ok: res.ok, status: res.status, result })))
    .then(({ ok, status, result }) => {
        if (ok && result.message === "login success") {
            localStorage.setItem("user", JSON.stringify(result.user));
            window.location.href = "home.html";
        } else if (status === 403) {
            showError("Your account is pending admin approval. Please wait.");
        } else {
            showError("Invalid email or password");
        }
    })
    .catch(() => showError("Server error. Please try again."));
}

function logout() {
    localStorage.removeItem("user");
    window.location.href = "index.html";
}

function showError(msg) {
    const el = document.getElementById("errorMsg");
    if (el) {
        el.textContent = msg;
        el.style.display = "block";
    } else {
        alert(msg);
    }
}

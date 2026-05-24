function adminLogin() {
    const data = {
        email: document.getElementById("email").value.trim(),
        password: document.getElementById("password").value.trim()
    };

    if (!data.email || !data.password) {
        showError("Please fill all fields");
        return;
    }

    fetch("/admin/login", {
        method: "POST",
        body: JSON.stringify(data),
        headers: { "Content-Type": "application/json" }
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
    } else {
        alert(msg);
    }
}

// ---- Admin Dashboard ----

async function loadUsers() {
    try {
        const res = await fetch("/admin/users");
        const users = await res.json();
        renderUsers(Array.isArray(users) ? users : []);
    } catch {
        console.error("Failed to load users");
    }
}

function renderUsers(users) {
    const tbody = document.getElementById("usersList");
    if (!tbody) return;
    tbody.innerHTML = "";

    if (!users.length) {
        tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;padding:20px;">No users found</td></tr>`;
        return;
    }

    users.forEach(u => {
        const row = document.createElement("tr");
        row.id = `user-row-${u.id}`;
        row.innerHTML = `
            <td>${u.id}</td>
            <td style="font-size:11px;color:var(--text-muted);font-family:monospace;">${u.student_id || "–"}</td>
            <td>${u.first_name} ${u.last_name}</td>
            <td>${u.email}</td>
            <td>${u.phone || "–"}</td>
            <td class="action-cell">
                <button class="btn-reject" onclick="deleteUser(${u.id})">Delete</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

async function deleteUser(id) {
    if (!confirm("Delete this user? This cannot be undone.")) return;
    try {
        const res = await fetch(`/admin/users/${id}`, { method: "DELETE" });
        if (res.ok) loadUsersWithStats();
        else alert("Failed to delete user");
    } catch {
        alert("Server error");
    }
}

async function loadAdminBookings() {
    try {
        const res = await fetch("/bookings");
        const bookings = await res.json();
        renderAdminBookings(Array.isArray(bookings) ? bookings : []);
    } catch {
        console.error("Failed to load bookings");
    }
}

function renderAdminBookings(bookings) {
    const tbody = document.getElementById("adminBookingsList");
    if (!tbody) return;
    tbody.innerHTML = "";

    if (!bookings.length) {
        tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;padding:20px;">No bookings</td></tr>`;
        return;
    }

    bookings.forEach(b => {
        const row = document.createElement("tr");
        row.innerHTML = `
            <td>${b.id}</td>
            <td>${b.student_id}</td>
            <td>${b.match_type || "-"}</td>
            <td>${b.date}</td>
            <td>${b.starting_time} – ${b.ending_time}</td>
            <td>
                <button class="btn-reject" onclick="deleteAdminBooking(${b.id})">Delete</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

async function deleteAdminBooking(id) {
    if (!confirm("Delete this booking?")) return;
    try {
        const res = await fetch(`/booking/${id}`, { method: "DELETE" });
        if (res.ok) loadAdminBookings();
    } catch {
        alert("Failed to delete booking");
    }
}

document.addEventListener("DOMContentLoaded", () => {
    if (document.getElementById("usersList")) loadUsers();
    if (document.getElementById("adminBookingsList")) loadAdminBookings();
});

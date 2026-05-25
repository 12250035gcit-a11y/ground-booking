const API = "";
const MAX_BOOKING_MINUTES = 90;

// Timeline range: 06:00 – 22:00
const TIMELINE_START = 6;
const TIMELINE_END   = 22;
const TIMELINE_HOURS = TIMELINE_END - TIMELINE_START;

const _user = JSON.parse(localStorage.getItem("user") || "{}");

document.addEventListener("DOMContentLoaded", () => {
    const sid = document.getElementById("studentId");
    if (sid && _user.email) sid.value = _user.email;

    setupDate();
    buildTimelineHours();
    prefillFromURL();
    loadTimeline();
    loadBookings();
});

function prefillFromURL() {
    const params = new URLSearchParams(window.location.search);
    const start  = params.get("start");
    const end    = params.get("end");
    const date   = params.get("date");

    if (date) {
        const dateEl = document.getElementById("bookDate");
        const tlDate = document.getElementById("timelineDate");
        if (dateEl) dateEl.value = date;
        if (tlDate) tlDate.value = date;
    }
    if (start) {
        const sel = document.getElementById("startTime");
        if (sel) sel.value = start;
    }
    if (end) {
        const sel = document.getElementById("endTime");
        if (sel) sel.value = end;
    }
}

function setupDate() {
    const today = new Date().toISOString().split("T")[0];
    const dateInput = document.getElementById("bookDate");
    const tlDate    = document.getElementById("timelineDate");
    if (dateInput) { dateInput.min = today; dateInput.value = today; }
    if (tlDate)    { tlDate.value = today; }
}

/* ===== TIMELINE ===== */

function buildTimelineHours() {
    const container = document.getElementById("timelineHours");
    if (!container) return;
    container.innerHTML = "";
    for (let h = TIMELINE_START; h < TIMELINE_END; h++) {
        const div = document.createElement("div");
        div.className = "timeline-hour";
        const label = h < 12 ? `${h}am` : h === 12 ? "12pm" : `${h-12}pm`;
        div.innerHTML = `<span class="timeline-hour-label">${label}</span>`;
        container.appendChild(div);
    }
}

function timeToFraction(timeStr) {
    const [h, m] = timeStr.split(":").map(Number);
    return (h + m / 60 - TIMELINE_START) / TIMELINE_HOURS;
}

// Uses /all endpoint so pending bookings are visible to everyone
async function loadTimeline() {
    const date = document.getElementById("timelineDate")?.value;
    if (!date) return;

    const bar   = document.getElementById("timelineBar");
    const chips = document.getElementById("bookedChips");
    if (!bar || !chips) return;

    bar.querySelectorAll(".timeline-booked-slot").forEach(el => el.remove());
    chips.innerHTML = '<span style="font-size:13px;color:var(--text-muted);">Loading...</span>';

    try {
        const res = await fetch(`${API}/bookings/date/${date}/all`);
        const data = await res.json();
        const bookings = Array.isArray(data) ? data : [];

        if (!bookings.length) {
            chips.innerHTML = '<span class="no-bookings-msg"><i class="fas fa-check-circle"></i> No bookings — ground is free all day!</span>';
            return;
        }

        chips.innerHTML = "";

        bookings.forEach(b => {
            const startFrac = timeToFraction(b.starting_time);
            const endFrac   = timeToFraction(b.ending_time);
            if (endFrac <= 0 || startFrac >= 1) return;

            const left  = Math.max(0, startFrac) * 100;
            const width = (Math.min(1, endFrac) - Math.max(0, startFrac)) * 100;
            const isPending = b.status === "pending";

            const slot = document.createElement("div");
            slot.className = "timeline-booked-slot" + (isPending ? " timeline-pending-slot" : "");
            slot.style.left  = left + "%";
            slot.style.width = width + "%";
            const statusLabel = isPending ? "Pending" : (b.status === "cancel_requested" ? "Cancel Req." : "Approved");
            slot.title = `${b.student_id} · ${b.match_type || "Booking"} · ${b.starting_time}–${b.ending_time} [${statusLabel}]`;
            slot.innerHTML = `<span class="timeline-booked-label">${b.starting_time}–${b.ending_time}</span>`;
            bar.appendChild(slot);

            const chip = document.createElement("div");
            chip.className = "booked-chip" + (isPending ? " pending-chip" : "");
            chip.innerHTML = `
                <div class="booked-chip-dot${isPending ? ' pending-dot' : ''}"></div>
                <strong>${b.starting_time}–${b.ending_time}</strong>&nbsp;·&nbsp;${b.student_id}&nbsp;(${b.match_type || "–"})
                <span style="font-size:10px;opacity:0.7;margin-left:3px;">[${statusLabel}]</span>
            `;
            chips.appendChild(chip);
        });

    } catch {
        chips.innerHTML = '<span style="color:var(--text-muted);font-size:13px;">Could not load availability.</span>';
    }
}

function syncTimeline() {
    const bookDate = document.getElementById("bookDate")?.value;
    const tlDate   = document.getElementById("timelineDate");
    if (tlDate && bookDate) { tlDate.value = bookDate; loadTimeline(); }
}

/* ===== MY BOOKINGS TABLE ===== */

async function loadBookings() {
    const tbody = document.getElementById("bookingsList");
    if (!tbody) return;
    try {
        const res = await fetch(`${API}/bookings`);
        if (!res.ok) throw new Error(`Server error ${res.status}`);
        const data = await res.json();
        const all = Array.isArray(data) ? data : [];
        // Show only the current user's bookings
        const mine = all.filter(b => b.student_id === _user.email);
        renderBookings(mine);
    } catch (err) {
        console.error("Load error:", err);
        renderBookings([]);
    }
}

function statusBadge(status) {
    const map = {
        pending:          { cls: "badge-pending",   label: '<i class="fas fa-clock"></i> Pending' },
        approved:         { cls: "badge-approved",  label: '<i class="fas fa-check-circle"></i> Approved' },
        rejected:         { cls: "badge-rejected",  label: '<i class="fas fa-times-circle"></i> Rejected' },
        cancel_requested: { cls: "badge-cancel",    label: '<i class="fas fa-hourglass-half"></i> Cancel Pending' },
        cancelled:        { cls: "badge-cancelled", label: '<i class="fas fa-ban"></i> Cancelled' },
    };
    const s = map[status] || { cls: "badge-pending", label: status };
    return `<span class="status-badge ${s.cls}">${s.label}</span>`;
}

function renderBookings(bookings) {
    const tbody = document.getElementById("bookingsList");
    if (!tbody) return;
    tbody.innerHTML = "";

    if (!bookings || bookings.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="5" style="text-align:center;padding:24px;color:var(--text-muted);">
                    No bookings yet. Submit your first request above!
                </td>
            </tr>
        `;
        return;
    }

    bookings.forEach(b => {
        let actionCell = `<span style="font-size:12px;color:var(--text-muted);">–</span>`;

        if (b.status === "pending" || b.status === "rejected") {
            actionCell = `<button class="delete-btn" onclick="deleteBooking(${b.id})">Delete</button>`;
        } else if (b.status === "approved") {
            actionCell = `<button class="cancel-req-btn" onclick="requestCancelBooking(${b.id})">Request Cancel</button>`;
        } else if (b.status === "cancel_requested") {
            actionCell = `<span style="font-size:12px;color:var(--text-muted);font-style:italic;">Awaiting admin</span>`;
        }

        const row = document.createElement("tr");
        row.innerHTML = `
            <td>${b.match_type || "–"}</td>
            <td>${b.date || "–"}</td>
            <td style="white-space:nowrap;">${b.starting_time || "–"} – ${b.ending_time || "–"}</td>
            <td>${statusBadge(b.status)}</td>
            <td>${actionCell}</td>
        `;
        tbody.appendChild(row);
    });
}

/* ===== SUBMIT BOOKING ===== */

async function submitBooking() {
    const data = {
        student_id:    document.getElementById("studentId").value.trim(),
        match_type:    document.getElementById("purpose").value,
        date:          document.getElementById("bookDate").value,
        starting_time: document.getElementById("startTime").value,
        ending_time:   document.getElementById("endTime").value,
        notes:         document.getElementById("notes").value
    };

    if (!data.student_id || !data.match_type || !data.date || !data.starting_time || !data.ending_time) {
        alert("Please fill all required fields");
        return;
    }

    if (data.starting_time >= data.ending_time) {
        alert("End time must be after start time");
        return;
    }

    // Enforce 1.5-hour max on the frontend
    const [sh, sm] = data.starting_time.split(":").map(Number);
    const [eh, em] = data.ending_time.split(":").map(Number);
    const durationMins = (eh * 60 + em) - (sh * 60 + sm);
    if (durationMins > MAX_BOOKING_MINUTES) {
        alert("Bookings cannot exceed 1.5 hours (90 minutes).");
        return;
    }

    try {
        const res = await fetch(`${API}/booking`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(data)
        });

        const result = await res.json().catch(() => ({}));

        if (!res.ok) {
            throw new Error(result.error || "Booking failed");
        }

        alert("✅ Booking request submitted!\n\nYour booking is pending admin approval. Check the status in 'My Bookings' on your profile.");
        resetForm();
        loadBookings();
        loadTimeline();

    } catch (err) {
        alert("❌ " + err.message);
    }
}

async function deleteBooking(id) {
    if (!confirm("Delete this booking request?")) return;
    try {
        const res = await fetch(`${API}/booking/${id}`, { method: "DELETE" });
        if (!res.ok) throw new Error("Delete failed");
        loadBookings();
        loadTimeline();
    } catch (err) {
        alert(err.message);
    }
}

async function requestCancelBooking(id) {
    if (!confirm("Request cancellation of this approved booking?\n\nAn admin will confirm and the slot will be freed.")) return;
    try {
        const res = await fetch(`${API}/booking/${id}/request-cancel`, { method: "PUT" });
        const result = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(result.error || "Request failed");
        alert("Cancellation request submitted. The admin will confirm shortly.");
        loadBookings();
        loadTimeline();
    } catch (err) {
        alert("❌ " + err.message);
    }
}

function resetForm() {
    document.getElementById("studentId").value = _user.email || "";
    document.getElementById("purpose").value = "";
    document.getElementById("startTime").value = "";
    document.getElementById("endTime").value = "";
    document.getElementById("notes").value = "";
    setupDate();
}

const API = "";

// Timeline range: 06:00 – 22:00
const TIMELINE_START = 6;   // 6am
const TIMELINE_END   = 22;  // 10pm
const TIMELINE_HOURS = TIMELINE_END - TIMELINE_START;

const _user = JSON.parse(localStorage.getItem("user") || "{}");

document.addEventListener("DOMContentLoaded", () => {
    const sid = document.getElementById("studentId");
    if (sid && _user.email) sid.value = _user.email;

    setupDate();
    buildTimelineHours();
    loadTimeline();
    loadBookings();
});

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

async function loadTimeline() {
    const date = document.getElementById("timelineDate")?.value;
    if (!date) return;

    const bar   = document.getElementById("timelineBar");
    const chips = document.getElementById("bookedChips");
    if (!bar || !chips) return;

    // Remove old overlays (keep .timeline-hours)
    bar.querySelectorAll(".timeline-booked-slot").forEach(el => el.remove());
    chips.innerHTML = '<span style="font-size:13px;color:var(--text-muted);">Loading...</span>';

    try {
        const res = await fetch(`${API}/bookings/date/${date}`);
        const data = await res.json();
        const bookings = Array.isArray(data) ? data : [];

        if (!bookings.length) {
            chips.innerHTML = '<span class="no-bookings-msg"><i class="fas fa-check-circle"></i> No approved bookings — ground is free all day!</span>';
            return;
        }

        chips.innerHTML = "";

        bookings.forEach(b => {
            const startFrac = timeToFraction(b.starting_time);
            const endFrac   = timeToFraction(b.ending_time);
            if (endFrac <= 0 || startFrac >= 1) return; // outside range

            const left  = Math.max(0, startFrac) * 100;
            const width = (Math.min(1, endFrac) - Math.max(0, startFrac)) * 100;

            const slot = document.createElement("div");
            slot.className = "timeline-booked-slot";
            slot.style.left  = left + "%";
            slot.style.width = width + "%";
            slot.title = `${b.student_id} · ${b.match_type || "Booking"} · ${b.starting_time}–${b.ending_time}`;
            slot.innerHTML = `<span class="timeline-booked-label">${b.starting_time}–${b.ending_time}</span>`;
            bar.appendChild(slot);

            // Chip
            const chip = document.createElement("div");
            chip.className = "booked-chip";
            chip.innerHTML = `
                <div class="booked-chip-dot"></div>
                <strong>${b.starting_time}–${b.ending_time}</strong>&nbsp;·&nbsp;${b.student_id}&nbsp;(${b.match_type || "–"})
            `;
            chips.appendChild(chip);
        });

    } catch (err) {
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
    try {
        const res = await fetch(`${API}/bookings`);
        if (!res.ok) throw new Error(`Server error ${res.status}`);
        const data = await res.json();
        renderBookings(Array.isArray(data) ? data : []);
    } catch (err) {
        console.error("Load error:", err);
        renderBookings([]);
    }
}

function statusBadge(status) {
    const map = {
        pending:  { cls: "badge-pending",  label: '<i class="fas fa-clock"></i> Pending' },
        approved: { cls: "badge-approved", label: '<i class="fas fa-check-circle"></i> Approved' },
        rejected: { cls: "badge-rejected", label: '<i class="fas fa-times-circle"></i> Rejected' },
    };
    const s = map[status] || { cls: "badge-pending", label: status };
    return `<span class="status-badge ${s.cls}">${s.label}</span>`;
}

function renderBookings(bookings) {
    const tbody = document.getElementById("bookingsList");
    tbody.innerHTML = "";

    if (!bookings || bookings.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" style="text-align:center;padding:24px;color:var(--text-muted);">
                    No bookings yet. Submit your first request!
                </td>
            </tr>
        `;
        return;
    }

    bookings.forEach(b => {
        const row = document.createElement("tr");
        row.innerHTML = `
            <td>${b.student_id || "–"}</td>
            <td>${b.match_type || "–"}</td>
            <td>${b.date || "–"}</td>
            <td>${b.starting_time || "–"} – ${b.ending_time || "–"}</td>
            <td>${statusBadge(b.status)}</td>
            <td>
                ${b.status === "pending" || b.status === "rejected"
                    ? `<button class="delete-btn" onclick="deleteBooking(${b.id})">Cancel</button>`
                    : `<span style="font-size:12px;color:var(--text-muted);">–</span>`
                }
            </td>
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

        alert("✅ Booking request submitted!\n\nYour booking is pending admin approval. You can check the status in 'My Bookings'.");
        resetForm();
        loadBookings();
        loadTimeline();

    } catch (err) {
        alert("❌ " + err.message);
    }
}

async function deleteBooking(id) {
    if (!confirm("Cancel this booking request?")) return;
    try {
        const res = await fetch(`${API}/booking/${id}`, { method: "DELETE" });
        if (!res.ok) throw new Error("Delete failed");
        loadBookings();
        loadTimeline();
    } catch (err) {
        alert(err.message);
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

// Auth guard — call on every protected page.
// Checks the session cookie via /user/me (HttpOnly, so JS can't read it directly).
// On success, syncs fresh user data to localStorage and calls the optional callback.
// On failure (401), redirects to login immediately.
async function requireAuth(onUser) {
    try {
        const res = await fetch("/user/me", { credentials: "same-origin" });
        if (!res.ok) {
            localStorage.removeItem("user");
            window.location.replace("index.html");
            return;
        }
        const user = await res.json();
        localStorage.setItem("user", JSON.stringify(user));
        if (typeof onUser === "function") onUser(user);
    } catch {
        localStorage.removeItem("user");
        window.location.replace("index.html");
    }
}

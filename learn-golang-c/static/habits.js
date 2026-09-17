// habits.js: progressive enhancement untuk toggle check-in (B.15 JSON
// response). checkbox tetap berfungsi kalau JS mati - browser cuma tidak
// dapat update tanpa reload. Endpoint diproteksi Basic Auth (B.18); fetch()
// tanpa header eksplisit tetap terautentikasi SELAMA browser sudah pernah
// login lewat request lain ke realm yang sama (mis. buka /notes/new
// sekali) - browser otomatis melampirkan ulang credential yang di-cache,
// tanpa perlu ditulis ulang manual di sini.
document.querySelectorAll(".habit-checkin").forEach(function (el) {
    el.addEventListener("change", function () {
        var id = el.dataset.id;
        fetch("/habits/" + id + "/checkin", { method: "POST" })
            .then(function (res) {
                if (!res.ok) throw new Error("checkin gagal: " + res.status);
                return res.json();
            })
            .then(function (data) {
                el.checked = data.done;
                var badge = el.closest(".habit-row").querySelector(".badge");
                if (badge) badge.textContent = "streak: " + data.streak + " hari";
            })
            .catch(function (err) {
                el.checked = !el.checked; // rollback kalau gagal (mis. belum login)
                alert(err.message);
            });
    });
});

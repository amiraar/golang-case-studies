// notes.js (B.16 Multiple File Upload: MultipartReader): kirim banyak file
// sekaligus ke POST /notes/{id}/attachments lewat fetch+FormData, bukan
// form-submit biasa, supaya browser tetap di halaman note setelah selesai
// (redirect 303 dari server tetap diikuti fetch, lalu halaman di-reload
// manual di sini supaya daftar lampiran ter-refresh).
document.addEventListener("DOMContentLoaded", function () {
    var form = document.getElementById("attachments-form");
    if (!form) return;

    form.addEventListener("submit", function (e) {
        e.preventDefault();

        var noteId = form.dataset.noteId;
        var input = form.querySelector('input[name="files"]');
        var files = input.files;
        if (!files || files.length === 0) return;

        var data = new FormData();
        for (var i = 0; i < files.length; i++) {
            data.append("files", files[i]);
        }

        fetch("/notes/" + noteId + "/attachments", {
            method: "POST",
            body: data,
            credentials: "same-origin", // ikut kirim Basic Auth yang sudah di-cache browser
        }).then(function () {
            window.location.reload();
        });
    });
});

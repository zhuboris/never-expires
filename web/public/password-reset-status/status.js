const Statuses = {
    SUCCESS: "success",
    FAILURE: "failure",
};

const StatusMessages = {
    [Statuses.SUCCESS]: "Your password has been successfully reset! We have sent a new one to your email.",
    [Statuses.FAILURE]: "This link is no longer valid. You can request a new one in the app.",
};

document.addEventListener("DOMContentLoaded", function() {
    const urlParams = new URLSearchParams(window.location.search);
    const status = urlParams.get("status");
    let messageElement = document.getElementById("statusMessage");

    messageElement.textContent = StatusMessages[status] || StatusMessages[Statuses.FAILURE];
});

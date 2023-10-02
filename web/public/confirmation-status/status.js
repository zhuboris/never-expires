const Statuses = {
    SUCCESS: "success",
    ALREADY_CONFIRMED: "already_confirmed",
    FAILURE: "failure",
};

const StatusMessages = {
    [Statuses.SUCCESS]: "Your email has been successfully confirmed!",
    [Statuses.ALREADY_CONFIRMED]: "Your email is already confirmed.",
    [Statuses.FAILURE]: "This confirmation link is no longer valid. You can request a new one in the app.",
};

document.addEventListener("DOMContentLoaded", function() {
    const urlParams = new URLSearchParams(window.location.search);
    const status = urlParams.get("status");
    let messageElement = document.getElementById("statusMessage");

    messageElement.textContent = StatusMessages[status] || StatusMessages[Statuses.FAILURE];
});

"use strict";
const form = document.getElementById("guest-form");
const nameInput = document.getElementById("name");
const surnameInput = document.getElementById("surname");
const studyDirectionInput = document.getElementById("studyDirection");
const visitedInput = document.getElementById("visitedEvent");
const addButton = document.getElementById("add-btn");
const checkButton = document.getElementById("check-btn");
const toastRoot = document.getElementById("toast-root");
function showToast(message, type) {
    const toast = document.createElement("div");
    toast.className = `toast ${type}`;
    toast.textContent = message;
    toastRoot.appendChild(toast);
    window.setTimeout(() => {
        toast.remove();
    }, 2800);
}
function readRequiredNames() {
    const name = nameInput.value.trim();
    const surname = surnameInput.value.trim();
    if (!name || !surname) {
        showToast("Имя и фамилия обязательны", "error");
        return null;
    }
    return { name, surname };
}
async function parseResponse(response) {
    if (response.ok) {
        return await response.json();
    }
    const errorText = await response.text();
    throw new Error(errorText || "Request failed");
}
function setBusy(isBusy) {
    addButton.disabled = isBusy;
    checkButton.disabled = isBusy;
}
async function addGuest() {
    const payloadNames = readRequiredNames();
    if (!payloadNames)
        return;
    const studyDirection = studyDirectionInput.value.trim();
    if (!studyDirection) {
        showToast("Направление обучения обязательно", "error");
        return;
    }
    setBusy(true);
    try {
        const person = await parseResponse(await fetch("/people", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                name: payloadNames.name,
                surname: payloadNames.surname,
                studyDirection,
                visitedEvent: visitedInput.checked,
            }),
        }));
        showToast(`Гость добавлен (#${person.orderNumber})`, "ok");
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        setBusy(false);
    }
}
async function checkGuest() {
    const payloadNames = readRequiredNames();
    if (!payloadNames)
        return;
    setBusy(true);
    try {
        const person = await parseResponse(await fetch("/people/check-in", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payloadNames),
        }));
        showToast(`OK: ${person.name} ${person.surname} отмечен`, "ok");
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        setBusy(false);
    }
}
addButton.addEventListener("click", () => {
    addGuest();
});
checkButton.addEventListener("click", () => {
    checkGuest();
});
form.addEventListener("submit", (event) => {
    event.preventDefault();
    addGuest();
});

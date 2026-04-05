"use strict";
const form = document.getElementById("guest-form");
const nameInput = document.getElementById("name");
const surnameInput = document.getElementById("surname");
const studyDirectionInput = document.getElementById("studyDirection");
const visitedInput = document.getElementById("visitedEvent");
const addButton = document.getElementById("add-btn");
const checkButton = document.getElementById("check-btn");
const deleteButton = document.getElementById("delete-btn");
const logoutButton = document.getElementById("logout-btn");
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
    const errorText = (await response.text()).trim();
    throw new Error(localizeError(errorText || "Request failed"));
}
function localizeError(message) {
    const map = {
        "person not found": "Человек не найден",
        "such person already passed": "Этот человек уже прошел",
        "person already exists": "Такой человек уже добавлен",
        "name and surname are required": "Нужно указать имя и фамилию",
        "name, surname and studyDirection are required": "Нужно указать имя, фамилию и направление обучения",
        "invalid json body": "Некорректные данные запроса",
        "method not allowed": "Метод запроса не поддерживается",
        unauthorized: "Сессия истекла. Войдите снова",
        forbidden: "Нет прав для этого действия",
        "internal server error": "Внутренняя ошибка сервера",
        "Request failed": "Не удалось выполнить запрос",
    };
    return map[message] ?? message;
}
function setBusy(isBusy) {
    addButton.disabled = isBusy;
    checkButton.disabled = isBusy;
    deleteButton.disabled = isBusy;
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
        showToast(`${person.name} ${person.surname} успешно отмечен`, "ok");
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        setBusy(false);
    }
}
async function deleteGuest() {
    const payloadNames = readRequiredNames();
    if (!payloadNames)
        return;
    setBusy(true);
    try {
        await parseResponse(await fetch("/people/delete", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payloadNames),
        }));
        showToast(`${payloadNames.name} ${payloadNames.surname} удален`, "ok");
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        setBusy(false);
    }
}
async function logout() {
    logoutButton.disabled = true;
    try {
        await fetch("/auth/logout", { method: "POST" });
    }
    finally {
        window.location.href = "/login";
    }
}
addButton.addEventListener("click", () => {
    addGuest();
});
checkButton.addEventListener("click", () => {
    checkGuest();
});
deleteButton.addEventListener("click", () => {
    deleteGuest();
});
logoutButton.addEventListener("click", () => {
    logout();
});
form.addEventListener("submit", (event) => {
    event.preventDefault();
    addGuest();
});

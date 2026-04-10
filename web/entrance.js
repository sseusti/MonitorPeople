"use strict";
const form = document.getElementById("check-form");
const nameInput = document.getElementById("name");
const surnameInput = document.getElementById("surname");
const nameSuggestions = document.getElementById("name-suggestions");
const surnameSuggestions = document.getElementById("surname-suggestions");
const checkButton = document.getElementById("check-btn");
const logoutButton = document.getElementById("logout-btn");
const toastRoot = document.getElementById("toast-root");
let nameSuggestTimer;
let surnameSuggestTimer;
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
function localizeError(message) {
    const map = {
        "person not found": "Человек не найден",
        "such person already passed": "Этот человек уже прошел",
        "name and surname are required": "Нужно указать имя и фамилию",
        unauthorized: "Сессия истекла. Войдите снова",
        forbidden: "Нет прав для этого действия",
        "internal server error": "Внутренняя ошибка сервера",
        "Request failed": "Не удалось выполнить запрос",
    };
    return map[message] ?? message;
}
async function parseResponse(response) {
    if (response.ok) {
        return await response.json();
    }
    const errorText = (await response.text()).trim();
    throw new Error(localizeError(errorText || "Request failed"));
}
async function fetchSuggestions(field, query) {
    const response = await fetch(`/people/suggest?field=${field}&q=${encodeURIComponent(query)}`);
    if (!response.ok) {
        return [];
    }
    return await response.json();
}
function renderSuggestions(datalist, values) {
    datalist.innerHTML = "";
    for (const value of values) {
        const option = document.createElement("option");
        option.value = value;
        datalist.appendChild(option);
    }
}
function clearNameAndSurnameFields() {
    nameInput.value = "";
    surnameInput.value = "";
    renderSuggestions(nameSuggestions, []);
    renderSuggestions(surnameSuggestions, []);
    nameInput.focus();
}
function scheduleNameSuggestions() {
    if (nameSuggestTimer !== undefined) {
        window.clearTimeout(nameSuggestTimer);
    }
    nameSuggestTimer = window.setTimeout(async () => {
        const query = nameInput.value.trim();
        if (query.length < 1) {
            renderSuggestions(nameSuggestions, []);
            return;
        }
        const values = await fetchSuggestions("name", query);
        renderSuggestions(nameSuggestions, values);
    }, 220);
}
function scheduleSurnameSuggestions() {
    if (surnameSuggestTimer !== undefined) {
        window.clearTimeout(surnameSuggestTimer);
    }
    surnameSuggestTimer = window.setTimeout(async () => {
        const query = surnameInput.value.trim();
        if (query.length < 1) {
            renderSuggestions(surnameSuggestions, []);
            return;
        }
        const values = await fetchSuggestions("surname", query);
        renderSuggestions(surnameSuggestions, values);
    }, 220);
}
async function checkGuest() {
    const payload = readRequiredNames();
    if (!payload)
        return;
    checkButton.disabled = true;
    try {
        const person = await parseResponse(await fetch("/people/check-in", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
        }));
        showToast(`${person.name} ${person.surname} успешно отмечен`, "ok");
        clearNameAndSurnameFields();
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        checkButton.disabled = false;
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
checkButton.addEventListener("click", () => {
    checkGuest();
});
logoutButton.addEventListener("click", () => {
    logout();
});
form.addEventListener("submit", (event) => {
    event.preventDefault();
    checkGuest();
});
nameInput.addEventListener("input", () => {
    scheduleNameSuggestions();
});
surnameInput.addEventListener("input", () => {
    scheduleSurnameSuggestions();
});

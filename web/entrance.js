"use strict";
const form = document.getElementById("check-form");
const nameInput = document.getElementById("name");
const surnameInput = document.getElementById("surname");
const nameSuggestions = document.getElementById("name-suggestions");
const surnameSuggestions = document.getElementById("surname-suggestions");
const container = document.querySelector("main.container");
let recentCheckinsList = document.getElementById("recent-checkins-list");
const checkButton = document.getElementById("check-btn");
const logoutButton = document.getElementById("logout-btn");
const toastRoot = document.getElementById("toast-root");
let nameSuggestTimer;
let surnameSuggestTimer;
const undoWindowMs = 60000;
const recentCheckins = [];
function ensureRecentCheckinsList() {
    if (recentCheckinsList) {
        return recentCheckinsList;
    }
    if (!container) {
        return null;
    }
    const section = document.createElement("section");
    section.className = "card recent-checkins-card";
    const title = document.createElement("h2");
    title.textContent = "Недавно отмеченные";
    const subtitle = document.createElement("p");
    subtitle.className = "subtitle-inline";
    subtitle.textContent = "Можно отменить в течение 1 минуты";
    recentCheckinsList = document.createElement("ul");
    recentCheckinsList.id = "recent-checkins-list";
    recentCheckinsList.className = "recent-checkins-list";
    section.appendChild(title);
    section.appendChild(subtitle);
    section.appendChild(recentCheckinsList);
    container.appendChild(section);
    return recentCheckinsList;
}
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
        "person is not checked in": "Гость уже отмечен как не пришедший",
        "undo window expired": "Окно отмены истекло (больше 1 минуты)",
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
function pruneRecentCheckins() {
    const now = Date.now();
    for (let i = recentCheckins.length - 1; i >= 0; i--) {
        if (now - recentCheckins[i].checkedAtMs > undoWindowMs) {
            recentCheckins.splice(i, 1);
        }
    }
}
function formatSecondsLeft(checkedAtMs) {
    const remainingMs = Math.max(0, undoWindowMs - (Date.now() - checkedAtMs));
    return Math.max(0, Math.ceil(remainingMs / 1000));
}
function renderRecentCheckins() {
    const list = ensureRecentCheckinsList();
    if (!list) {
        return;
    }
    pruneRecentCheckins();
    list.innerHTML = "";
    if (recentCheckins.length === 0) {
        const emptyItem = document.createElement("li");
        emptyItem.className = "recent-empty";
        emptyItem.textContent = "Пока нет недавних отметок";
        list.appendChild(emptyItem);
        return;
    }
    for (const checkIn of recentCheckins) {
        const item = document.createElement("li");
        const meta = document.createElement("div");
        meta.className = "checkin-meta";
        const nameLine = document.createElement("span");
        nameLine.className = "checkin-name";
        nameLine.textContent = `${checkIn.surname} ${checkIn.name}`;
        const timerLine = document.createElement("span");
        timerLine.className = "checkin-timer";
        timerLine.textContent = `Осталось: ${formatSecondsLeft(checkIn.checkedAtMs)} сек`;
        meta.appendChild(nameLine);
        meta.appendChild(timerLine);
        const undoButton = document.createElement("button");
        undoButton.type = "button";
        undoButton.className = "undo-btn";
        undoButton.textContent = "Отменить";
        undoButton.dataset.undoName = checkIn.name;
        undoButton.dataset.undoSurname = checkIn.surname;
        item.appendChild(meta);
        item.appendChild(undoButton);
        list.appendChild(item);
    }
}
function addRecentCheckin(person) {
    const checkedAtMs = Date.now();
    const deduped = recentCheckins.filter((item) => item.name !== person.name || item.surname !== person.surname);
    deduped.unshift({ name: person.name, surname: person.surname, checkedAtMs });
    recentCheckins.splice(0, recentCheckins.length, ...deduped.slice(0, 10));
    renderRecentCheckins();
}
async function undoCheckIn(name, surname, button) {
    button.disabled = true;
    try {
        const person = await parseResponse(await fetch("/people/check-in/undo", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, surname }),
        }));
        const nextItems = recentCheckins.filter((item) => item.name !== person.name || item.surname !== person.surname);
        recentCheckins.splice(0, recentCheckins.length, ...nextItems);
        renderRecentCheckins();
        showToast(`${person.name} ${person.surname}: статус "не пришел" восстановлен`, "ok");
    }
    catch (error) {
        showToast(error.message, "error");
        renderRecentCheckins();
    }
    finally {
        button.disabled = false;
    }
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
        addRecentCheckin(person);
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
const listElement = ensureRecentCheckinsList();
if (listElement) {
    listElement.addEventListener("click", (event) => {
        const target = event.target;
        const button = target.closest("button[data-undo-name][data-undo-surname]");
        if (!button) {
            return;
        }
        const name = button.dataset.undoName ?? "";
        const surname = button.dataset.undoSurname ?? "";
        if (!name || !surname) {
            return;
        }
        undoCheckIn(name, surname, button);
    });
}
window.setInterval(() => {
    renderRecentCheckins();
}, 1000);
renderRecentCheckins();

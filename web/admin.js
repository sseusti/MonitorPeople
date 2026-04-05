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
const refreshStatsButton = document.getElementById("refresh-stats-btn");
const filterVisited = document.getElementById("filter-visited");
const filterStudyDirection = document.getElementById("filter-studyDirection");
const programStatsList = document.getElementById("program-stats-list");
const guestsTbody = document.getElementById("guests-tbody");
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
function localizeError(message) {
    const map = {
        "person not found": "Человек не найден",
        "such person already passed": "Этот человек уже прошел",
        "person already exists": "Такой человек уже добавлен",
        "name and surname are required": "Нужно указать имя и фамилию",
        "name, surname and studyDirection are required": "Нужно указать имя, фамилию и направление обучения",
        "invalid study direction": "Выбрано некорректное направление",
        "visited must be true, false or all": "Некорректный фильтр посещения",
        "invalid json body": "Некорректные данные запроса",
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
function setBusy(isBusy) {
    addButton.disabled = isBusy;
    checkButton.disabled = isBusy;
    deleteButton.disabled = isBusy;
}
function getFilterQuery() {
    const params = new URLSearchParams();
    const visited = filterVisited.value.trim();
    const studyDirection = filterStudyDirection.value.trim();
    if (visited) {
        params.set("visited", visited);
    }
    if (studyDirection) {
        params.set("studyDirection", studyDirection);
    }
    const query = params.toString();
    return query ? `?${query}` : "";
}
function renderGuestsTable(people) {
    guestsTbody.innerHTML = "";
    if (people.length === 0) {
        const row = document.createElement("tr");
        row.innerHTML = `<td colspan="5">Ничего не найдено</td>`;
        guestsTbody.appendChild(row);
        return;
    }
    for (const person of people) {
        const row = document.createElement("tr");
        const status = person.visitedEvent ? "Пришел" : "Не пришел";
        row.innerHTML = `
      <td>${person.orderNumber}</td>
      <td>${person.name}</td>
      <td>${person.surname}</td>
      <td>${person.studyDirection}</td>
      <td>${status}</td>
    `;
        guestsTbody.appendChild(row);
    }
}
function renderProgramStats(stats) {
    programStatsList.innerHTML = "";
    if (stats.length === 0) {
        const item = document.createElement("li");
        item.textContent = "Нет данных по пришедшим гостям";
        programStatsList.appendChild(item);
        return;
    }
    for (const stat of stats) {
        const item = document.createElement("li");
        item.textContent = `${stat.studyDirection}: ${stat.count}`;
        programStatsList.appendChild(item);
    }
}
async function refreshStatistics(showSuccessToast = false) {
    refreshStatsButton.disabled = true;
    try {
        const query = getFilterQuery();
        const [people, stats] = await Promise.all([
            parseResponse(await fetch(`/people/list${query}`)),
            parseResponse(await fetch(`/people/stats/programs${query}`)),
        ]);
        renderGuestsTable(people);
        renderProgramStats(stats);
        if (showSuccessToast) {
            showToast("Статистика обновлена", "ok");
        }
    }
    catch (error) {
        showToast(error.message, "error");
    }
    finally {
        refreshStatsButton.disabled = false;
    }
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
        await refreshStatistics();
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
        await refreshStatistics();
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
        await refreshStatistics();
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
refreshStatsButton.addEventListener("click", () => {
    refreshStatistics(true);
});
filterVisited.addEventListener("change", () => {
    refreshStatistics();
});
filterStudyDirection.addEventListener("input", () => {
    refreshStatistics();
});
form.addEventListener("submit", (event) => {
    event.preventDefault();
    addGuest();
});
refreshStatistics();

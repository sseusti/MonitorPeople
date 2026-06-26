type Person = {
  name: string;
  surname: string;
  checkInTime?: string;
};

type ToastType = "ok" | "error";
type RecentCheckIn = {
  name: string;
  surname: string;
  checkedAtMs: number;
};

const form = document.getElementById("check-form") as HTMLFormElement;
const nameInput = document.getElementById("name") as HTMLInputElement;
const surnameInput = document.getElementById("surname") as HTMLInputElement;
const nameSuggestions = document.getElementById("name-suggestions") as HTMLDataListElement;
const surnameSuggestions = document.getElementById("surname-suggestions") as HTMLDataListElement;
const container = document.querySelector("main.container") as HTMLElement | null;
let recentCheckinsList = document.getElementById("recent-checkins-list") as HTMLUListElement | null;
const checkButton = document.getElementById("check-btn") as HTMLButtonElement;
const logoutButton = document.getElementById("logout-btn") as HTMLButtonElement;
const toastRoot = document.getElementById("toast-root") as HTMLDivElement;
let nameSuggestTimer: number | undefined;
let surnameSuggestTimer: number | undefined;
let nameSuggestRequestId = 0;
let surnameSuggestRequestId = 0;
const undoWindowMs = 60_000;
const recentCheckins: RecentCheckIn[] = [];

function ensureRecentCheckinsList(): HTMLUListElement | null {
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

function showToast(message: string, type: ToastType): void {
  const toast = document.createElement("div");
  toast.className = `toast ${type}`;
  toast.textContent = message;
  toastRoot.appendChild(toast);

  window.setTimeout(() => {
    toast.remove();
  }, 2800);
}

function readRequiredNames(): { name: string; surname: string } | null {
  const name = nameInput.value.trim();
  const surname = surnameInput.value.trim();
  if (!name || !surname) {
    showToast("Имя и фамилия обязательны", "error");
    return null;
  }
  return { name, surname };
}

function localizeError(message: string): string {
  const map: Record<string, string> = {
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

async function parseResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    return (await response.json()) as T;
  }
  const errorText = (await response.text()).trim();
  throw new Error(localizeError(errorText || "Request failed"));
}

async function fetchSuggestions(field: "name" | "surname", query: string): Promise<string[]> {
  const params = new URLSearchParams({ field, q: query });
  const response = await fetch(`/people/suggest?${params.toString()}`);
  if (!response.ok) {
    return [];
  }
  return (await response.json()) as string[];
}

function renderSuggestions(datalist: HTMLDataListElement, values: string[]): void {
  datalist.innerHTML = "";
  for (const value of values) {
    const option = document.createElement("option");
    option.value = value;
    datalist.appendChild(option);
  }
}

function clearNameAndSurnameFields(): void {
  nameInput.value = "";
  surnameInput.value = "";
  renderSuggestions(nameSuggestions, []);
  renderSuggestions(surnameSuggestions, []);
  nameInput.focus();
}

function pruneRecentCheckins(): void {
  const now = Date.now();
  for (let i = recentCheckins.length - 1; i >= 0; i--) {
    if (now - recentCheckins[i].checkedAtMs > undoWindowMs) {
      recentCheckins.splice(i, 1);
    }
  }
}

function formatSecondsLeft(checkedAtMs: number): number {
  const remainingMs = Math.max(0, undoWindowMs - (Date.now() - checkedAtMs));
  return Math.max(0, Math.ceil(remainingMs / 1000));
}

function renderRecentCheckins(): void {
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

function addRecentCheckin(person: Person): void {
  const checkedAtMs = Date.now();

  const deduped = recentCheckins.filter((item) => item.name !== person.name || item.surname !== person.surname);
  deduped.unshift({ name: person.name, surname: person.surname, checkedAtMs });
  recentCheckins.splice(0, recentCheckins.length, ...deduped.slice(0, 10));
  renderRecentCheckins();
}

async function undoCheckIn(name: string, surname: string, button: HTMLButtonElement): Promise<void> {
  button.disabled = true;
  try {
    const person = await parseResponse<Person>(
      await fetch("/people/check-in/undo", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, surname }),
      }),
    );

    const nextItems = recentCheckins.filter((item) => item.name !== person.name || item.surname !== person.surname);
    recentCheckins.splice(0, recentCheckins.length, ...nextItems);
    renderRecentCheckins();
    showToast(`${person.name} ${person.surname}: статус "не пришел" восстановлен`, "ok");
  } catch (error) {
    showToast((error as Error).message, "error");
    renderRecentCheckins();
  } finally {
    button.disabled = false;
  }
}

function scheduleNameSuggestions(): void {
  if (nameSuggestTimer !== undefined) {
    window.clearTimeout(nameSuggestTimer);
  }
  nameSuggestTimer = window.setTimeout(async () => {
    const query = nameInput.value.trim();
    if (query.length < 1) {
      renderSuggestions(nameSuggestions, []);
      return;
    }
    const requestId = ++nameSuggestRequestId;
    const values = await fetchSuggestions("name", query);
    if (requestId !== nameSuggestRequestId || query !== nameInput.value.trim()) {
      return;
    }
    renderSuggestions(nameSuggestions, values);
  }, 220);
}

function scheduleSurnameSuggestions(): void {
  if (surnameSuggestTimer !== undefined) {
    window.clearTimeout(surnameSuggestTimer);
  }
  surnameSuggestTimer = window.setTimeout(async () => {
    const query = surnameInput.value.trim();
    if (query.length < 1) {
      renderSuggestions(surnameSuggestions, []);
      return;
    }
    const requestId = ++surnameSuggestRequestId;
    const values = await fetchSuggestions("surname", query);
    if (requestId !== surnameSuggestRequestId || query !== surnameInput.value.trim()) {
      return;
    }
    renderSuggestions(surnameSuggestions, values);
  }, 220);
}

async function checkGuest(): Promise<void> {
  const payload = readRequiredNames();
  if (!payload) return;

  checkButton.disabled = true;
  try {
    const person = await parseResponse<Person>(
      await fetch("/people/check-in", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      }),
    );
    showToast(`${person.name} ${person.surname} успешно отмечен`, "ok");
    addRecentCheckin(person);
    clearNameAndSurnameFields();
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    checkButton.disabled = false;
  }
}

async function logout(): Promise<void> {
  logoutButton.disabled = true;
  try {
    await fetch("/auth/logout", { method: "POST" });
  } finally {
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
    const target = event.target as HTMLElement;
    const button = target.closest("button[data-undo-name][data-undo-surname]") as HTMLButtonElement | null;
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

type Person = {
  name: string;
  surname: string;
};

type ToastType = "ok" | "error";

const form = document.getElementById("check-form") as HTMLFormElement;
const nameInput = document.getElementById("name") as HTMLInputElement;
const surnameInput = document.getElementById("surname") as HTMLInputElement;
const nameSuggestions = document.getElementById("name-suggestions") as HTMLDataListElement;
const surnameSuggestions = document.getElementById("surname-suggestions") as HTMLDataListElement;
const checkButton = document.getElementById("check-btn") as HTMLButtonElement;
const logoutButton = document.getElementById("logout-btn") as HTMLButtonElement;
const toastRoot = document.getElementById("toast-root") as HTMLDivElement;
let nameSuggestTimer: number | undefined;
let surnameSuggestTimer: number | undefined;

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
  const response = await fetch(`/people/suggest?field=${field}&q=${encodeURIComponent(query)}`);
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
    const values = await fetchSuggestions("name", query);
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
    const values = await fetchSuggestions("surname", query);
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

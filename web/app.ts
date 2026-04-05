type Person = {
  orderNumber: number;
  name: string;
  surname: string;
  studyDirection: string;
  visitedEvent: boolean;
  checkInTime?: string;
};

type ToastType = "ok" | "error";

const form = document.getElementById("guest-form") as HTMLFormElement;
const nameInput = document.getElementById("name") as HTMLInputElement;
const surnameInput = document.getElementById("surname") as HTMLInputElement;
const studyDirectionInput = document.getElementById("studyDirection") as HTMLInputElement;
const visitedInput = document.getElementById("visitedEvent") as HTMLInputElement;
const addButton = document.getElementById("add-btn") as HTMLButtonElement;
const checkButton = document.getElementById("check-btn") as HTMLButtonElement;
const deleteButton = document.getElementById("delete-btn") as HTMLButtonElement;
const logoutButton = document.getElementById("logout-btn") as HTMLButtonElement;
const toastRoot = document.getElementById("toast-root") as HTMLDivElement;

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

async function parseResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    return (await response.json()) as T;
  }
  const errorText = (await response.text()).trim();
  throw new Error(localizeError(errorText || "Request failed"));
}

function localizeError(message: string): string {
  const map: Record<string, string> = {
    "person not found": "Человек не найден",
    "such person already passed": "Этот человек уже прошел",
    "person already exists": "Такой человек уже добавлен",
    "name and surname are required": "Нужно указать имя и фамилию",
    "name, surname and studyDirection are required": "Нужно указать имя, фамилию и направление обучения",
    "invalid json body": "Некорректные данные запроса",
    "method not allowed": "Метод запроса не поддерживается",
    "internal server error": "Внутренняя ошибка сервера",
    unauthorized: "Сессия истекла. Войдите в админку снова",
    "Request failed": "Не удалось выполнить запрос",
  };

  return map[message] ?? message;
}

function setBusy(isBusy: boolean): void {
  addButton.disabled = isBusy;
  checkButton.disabled = isBusy;
  deleteButton.disabled = isBusy;
}

async function addGuest(): Promise<void> {
  const payloadNames = readRequiredNames();
  if (!payloadNames) return;

  const studyDirection = studyDirectionInput.value.trim();
  if (!studyDirection) {
    showToast("Направление обучения обязательно", "error");
    return;
  }

  setBusy(true);
  try {
    const person = await parseResponse<Person>(
      await fetch("/people", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: payloadNames.name,
          surname: payloadNames.surname,
          studyDirection,
          visitedEvent: visitedInput.checked,
        }),
      }),
    );

    showToast(`Гость добавлен (#${person.orderNumber})`, "ok");
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    setBusy(false);
  }
}

async function checkGuest(): Promise<void> {
  const payloadNames = readRequiredNames();
  if (!payloadNames) return;

  setBusy(true);
  try {
    const person = await parseResponse<Person>(
      await fetch("/people/check-in", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadNames),
      }),
    );

    showToast(`${person.name} ${person.surname} успешно отмечен`, "ok");
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    setBusy(false);
  }
}

async function deleteGuest(): Promise<void> {
  const payloadNames = readRequiredNames();
  if (!payloadNames) return;

  setBusy(true);
  try {
    await parseResponse<{ ok: boolean }>(
      await fetch("/people/delete", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payloadNames),
      }),
    );

    showToast(`${payloadNames.name} ${payloadNames.surname} удален`, "ok");
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    setBusy(false);
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

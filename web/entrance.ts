type Person = {
  name: string;
  surname: string;
};

type ToastType = "ok" | "error";

const form = document.getElementById("check-form") as HTMLFormElement;
const nameInput = document.getElementById("name") as HTMLInputElement;
const surnameInput = document.getElementById("surname") as HTMLInputElement;
const checkButton = document.getElementById("check-btn") as HTMLButtonElement;
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

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
  const errorText = await response.text();
  throw new Error(errorText || "Request failed");
}

function setBusy(isBusy: boolean): void {
  addButton.disabled = isBusy;
  checkButton.disabled = isBusy;
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

    showToast(`OK: ${person.name} ${person.surname} отмечен`, "ok");
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
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

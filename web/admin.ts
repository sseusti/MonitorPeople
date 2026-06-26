type Person = {
  orderNumber: number;
  name: string;
  surname: string;
  studyDirection: string;
  visitedEvent: boolean;
  checkInTime?: string;
};

type ProgramStat = {
  studyDirection: string;
  count: number;
};

type ImportResult = {
  processed: number;
  imported: number;
  skippedDuplicates: number;
  skippedInvalid: number;
  errors?: string[];
};

type DeleteStudentsResult = {
  deleted: number;
};

type ToastType = "ok" | "error";

const form = document.getElementById("guest-form") as HTMLFormElement;
const nameInput = document.getElementById("name") as HTMLInputElement;
const surnameInput = document.getElementById("surname") as HTMLInputElement;
const nameSuggestions = document.getElementById("name-suggestions") as HTMLDataListElement;
const surnameSuggestions = document.getElementById("surname-suggestions") as HTMLDataListElement;
const programSuggestions = document.getElementById("program-suggestions") as HTMLDataListElement;
const studyDirectionInput = document.getElementById("studyDirection") as HTMLInputElement;
const visitedInput = document.getElementById("visitedEvent") as HTMLInputElement;
const addButton = document.getElementById("add-btn") as HTMLButtonElement;
const checkButton = document.getElementById("check-btn") as HTMLButtonElement;
const deleteButton = document.getElementById("delete-btn") as HTMLButtonElement;
const studentsFileInput = document.getElementById("students-file") as HTMLInputElement;
const teachersFileInput = document.getElementById("teachers-file") as HTMLInputElement;
const importStudentsButton = document.getElementById("import-students-btn") as HTMLButtonElement;
const importTeachersButton = document.getElementById("import-teachers-btn") as HTMLButtonElement;
const deleteStudentsButton = document.getElementById("delete-students-btn") as HTMLButtonElement;
const logoutButton = document.getElementById("logout-btn") as HTMLButtonElement;
const refreshStatsButton = document.getElementById("refresh-stats-btn") as HTMLButtonElement;
const filterVisited = document.getElementById("filter-visited") as HTMLSelectElement;
const filterStudyDirection = document.getElementById("filter-studyDirection") as HTMLInputElement;
const programStatsList = document.getElementById("program-stats-list") as HTMLUListElement;
const guestsTbody = document.getElementById("guests-tbody") as HTMLTableSectionElement;
const toastRoot = document.getElementById("toast-root") as HTMLDivElement;
let nameSuggestTimer: number | undefined;
let surnameSuggestTimer: number | undefined;
let programSuggestTimer: number | undefined;
let filterProgramSuggestTimer: number | undefined;
let nameSuggestRequestId = 0;
let surnameSuggestRequestId = 0;
let programSuggestRequestId = 0;
let filterProgramSuggestRequestId = 0;

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
    "person already exists": "Такой человек уже добавлен",
    "name and surname are required": "Нужно указать имя и фамилию",
    "name, surname and studyDirection are required": "Нужно указать имя, фамилию и направление обучения",
    "invalid study direction": "Выбрано некорректное направление",
    "visited must be true, false or all": "Некорректный фильтр посещения",
    "invalid json body": "Некорректные данные запроса",
    "invalid multipart form": "Некорректная форма загрузки",
    "file is required": "Выберите файл для загрузки",
    "unsupported import file type": "Поддерживаются только .xls и .xlsx",
    "required import columns not found": "Не найдены нужные колонки в файле",
    "invalid import file": "Не удалось прочитать файл",
    "no valid rows found": "В файле не найдено строк для импорта",
    "delete students failed": "Не удалось удалить студентов",
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

function setBusy(isBusy: boolean): void {
  addButton.disabled = isBusy;
  checkButton.disabled = isBusy;
  deleteButton.disabled = isBusy;
  importStudentsButton.disabled = isBusy;
  importTeachersButton.disabled = isBusy;
  deleteStudentsButton.disabled = isBusy;
}

async function fetchSuggestions(field: "name" | "surname" | "studyDirection", query: string): Promise<string[]> {
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

function scheduleProgramSuggestions(input: HTMLInputElement, isFilter: boolean): void {
  const currentTimer = isFilter ? filterProgramSuggestTimer : programSuggestTimer;
  if (currentTimer !== undefined) {
    window.clearTimeout(currentTimer);
  }

  const timer = window.setTimeout(async () => {
    const query = input.value.trim();
    if (query.length < 1) {
      renderSuggestions(programSuggestions, []);
      return;
    }

    const requestId = isFilter ? ++filterProgramSuggestRequestId : ++programSuggestRequestId;
    const values = await fetchSuggestions("studyDirection", query);
    const latestRequestId = isFilter ? filterProgramSuggestRequestId : programSuggestRequestId;
    if (requestId !== latestRequestId || query !== input.value.trim()) {
      return;
    }
    renderSuggestions(programSuggestions, values);
  }, 220);

  if (isFilter) {
    filterProgramSuggestTimer = timer;
  } else {
    programSuggestTimer = timer;
  }
}

function getFilterQuery(): string {
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

function renderGuestsTable(people: Person[]): void {
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

function renderProgramStats(stats: ProgramStat[]): void {
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

async function refreshStatistics(showSuccessToast = false): Promise<void> {
  refreshStatsButton.disabled = true;
  try {
    const query = getFilterQuery();
    const [people, stats] = await Promise.all([
      parseResponse<Person[]>(await fetch(`/people/list${query}`)),
      parseResponse<ProgramStat[]>(await fetch(`/people/stats/programs${query}`)),
    ]);
    renderGuestsTable(people);
    renderProgramStats(stats);
    if (showSuccessToast) {
      showToast("Статистика обновлена", "ok");
    }
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    refreshStatsButton.disabled = false;
  }
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
    await refreshStatistics();
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
    clearNameAndSurnameFields();
    await refreshStatistics();
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
    await refreshStatistics();
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    setBusy(false);
  }
}

async function importPeople(endpoint: string, fileInput: HTMLInputElement, label: string): Promise<void> {
  const file = fileInput.files?.[0];
  if (!file) {
    showToast("Выберите файл для загрузки", "error");
    return;
  }

  const formData = new FormData();
  formData.append("file", file);

  setBusy(true);
  try {
    const result = await parseResponse<ImportResult>(
      await fetch(endpoint, {
        method: "POST",
        body: formData,
      }),
    );
    fileInput.value = "";
    let message = `${label}: добавлено ${result.imported}, дубли ${result.skippedDuplicates}, пропущено ${result.skippedInvalid}`;
    if (result.errors && result.errors.length > 0) {
      message += `. Первая ошибка: ${result.errors[0]}`;
    }
    showToast(message, result.imported > 0 ? "ok" : "error");
    await refreshStatistics();
  } catch (error) {
    showToast((error as Error).message, "error");
  } finally {
    setBusy(false);
  }
}

async function deleteAllStudents(): Promise<void> {
  const confirmed = window.confirm(
    "Удалить всех студентов из базы? Преподаватели останутся. Это действие нельзя отменить.",
  );
  if (!confirmed) {
    return;
  }

  setBusy(true);
  try {
    const result = await parseResponse<DeleteStudentsResult>(
      await fetch("/people/students/delete-all", {
        method: "POST",
      }),
    );
    showToast(`Удалено студентов: ${result.deleted}`, "ok");
    await refreshStatistics();
  } catch (error) {
    showToast((error as Error).message || localizeError("delete students failed"), "error");
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

importStudentsButton.addEventListener("click", () => {
  importPeople("/people/import/students", studentsFileInput, "Выпускники");
});

importTeachersButton.addEventListener("click", () => {
  importPeople("/people/import/teachers", teachersFileInput, "Преподаватели");
});

deleteStudentsButton.addEventListener("click", () => {
  deleteAllStudents();
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
  scheduleProgramSuggestions(filterStudyDirection, true);
  refreshStatistics();
});

nameInput.addEventListener("input", () => {
  scheduleNameSuggestions();
});

surnameInput.addEventListener("input", () => {
  scheduleSurnameSuggestions();
});

studyDirectionInput.addEventListener("input", () => {
  scheduleProgramSuggestions(studyDirectionInput, false);
});

form.addEventListener("submit", (event) => {
  event.preventDefault();
  addGuest();
});

refreshStatistics();

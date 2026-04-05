"use strict";
const loginForm = document.getElementById("login-form");
const loginInput = document.getElementById("login");
const passwordInput = document.getElementById("password");
const loginButton = document.getElementById("login-btn");
const toastRoot = document.getElementById("toast-root");
function showToast(message) {
    const toast = document.createElement("div");
    toast.className = "toast";
    toast.textContent = message;
    toastRoot.appendChild(toast);
    window.setTimeout(() => {
        toast.remove();
    }, 2800);
}
function localizeError(message) {
    const map = {
        "invalid login or password": "Неверный логин или пароль",
        "invalid json body": "Некорректные данные запроса",
        "method not allowed": "Метод запроса не поддерживается",
        "internal server error": "Внутренняя ошибка сервера",
    };
    return map[message] ?? "Не удалось выполнить вход";
}
async function login() {
    const login = loginInput.value.trim();
    const password = passwordInput.value.trim();
    if (!login || !password) {
        showToast("Укажите логин и пароль");
        return;
    }
    loginButton.disabled = true;
    try {
        const response = await fetch("/auth/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ login, password }),
        });
        if (!response.ok) {
            const text = (await response.text()).trim();
            showToast(localizeError(text));
            return;
        }
        const payload = await response.json();
        window.location.href = payload.redirectPath ?? "/";
    }
    catch {
        showToast("Сервер недоступен");
    }
    finally {
        loginButton.disabled = false;
    }
}
loginForm.addEventListener("submit", (event) => {
    event.preventDefault();
    login();
});

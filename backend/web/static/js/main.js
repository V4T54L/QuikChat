import * as api from './api.js';
import * as store from './store.js';
import * as ui from './ui.js';
import * as ws from './ws.js';

/**
 * Handles the submission of login and signup forms.
 * @param {Event} event
 */
async function handleAuthFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const isLogin = form.id === 'login-form';
    const username = form.username.value;
    const password = form.password.value;

    try {
        if (isLogin) {
            await api.login(username, password);
        } else {
            await api.signup(username, password);
            // Automatically log in after successful signup
            await api.login(username, password);
        }
        window.location.href = '/chat';
    } catch (error) {
        ui.showError(form.id, error.message);
    }
}

/**
 * Handles the submission of the profile update form.
 * @param {Event} event
 */
async function handleProfileUpdateFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    // Remove empty fields from formData
    for (let [key, value] of [...formData.entries()]) {
        if (value === '' || (value instanceof File && value.size === 0)) {
            formData.delete(key);
        }
    }

    if (Array.from(formData.keys()).length === 0) {
        ui.showNotification('No changes to update.', 'info');
        return;
    }

    try {
        const updatedUser = await api.updateProfile(formData);
        store.setCurrentUser(updatedUser);
        ui.renderProfile(updatedUser);
        ui.showNotification('Profile updated successfully!', 'success');
        form.reset();
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

/**
 * Handles user logout.
 */
async function handleLogout() {
    await api.logout();
    store.clearTokens();
    window.location.href = '/';
}

/**
 * Initializes the authentication page (login/signup).
 */
function initAuthPage() {
    const loginForm = document.getElementById('login-form');
    const signupForm = document.getElementById('signup-form');
    const showSignupBtn = document.getElementById('show-signup-form');
    const showLoginBtn = document.getElementById('show-login-form');

    loginForm?.addEventListener('submit', handleAuthFormSubmit);
    signupForm?.addEventListener('submit', handleAuthFormSubmit);
    showSignupBtn?.addEventListener('click', ui.toggleAuthForms);
    showLoginBtn?.addEventListener('click', ui.toggleAuthForms);
}

/**
 * Initializes the main chat page.
 */
async function initChatPage() {
    const logoutButton = document.getElementById('logout-button');
    const profileUpdateForm = document.getElementById('profile-update-form');

    logoutButton?.addEventListener('click', handleLogout);
    profileUpdateForm?.addEventListener('submit', handleProfileUpdateFormSubmit);

    try {
        const user = await api.getMe();
        store.setCurrentUser(user);
        ui.renderProfile(user);
        // Connect to WebSocket after successfully fetching user profile
        ws.connect();
    } catch (error) {
        console.error('Failed to fetch user profile:', error);
        // If fetching fails (e.g., invalid token), redirect to login
        store.clearTokens();
        window.location.href = '/';
    }
}

// --- Main Execution ---
document.addEventListener('DOMContentLoaded', () => {
    const path = window.location.pathname;

    if (path === '/' || path === '/index.html') {
        initAuthPage();
    } else if (path === '/chat' || path === '/chat.html') {
        initChatPage();
    }
});


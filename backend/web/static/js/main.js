import * as api from './api.js';
import * as store from './store.js';
import * as ui from './ui.js';

function handleAuthFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const isLogin = form.id === 'login-form';
    const username = form.username.value;
    const password = form.password.value;

    const action = isLogin ? api.login(username, password) : api.signup(username, password);

    action
        .then(() => {
            if (isLogin) {
                window.location.href = '/chat';
            } else {
                alert('Signup successful! Please log in.');
                ui.toggleAuthForms();
                form.reset();
            }
        })
        .catch(err => {
            ui.showError(form.id, err.message);
        });
}

function handleProfileUpdateFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    // Do not include empty fields
    if (!formData.get('username')) formData.delete('username');
    if (!formData.get('password')) formData.delete('password');
    if (formData.get('profile_pic')?.size === 0) formData.delete('profile_pic');

    api.updateProfile(formData)
        .then(updatedUser => {
            store.setCurrentUser(updatedUser);
            ui.renderProfile(updatedUser);
            ui.showNotification('Profile updated successfully!', 'success');
            form.reset();
        })
        .catch(err => {
            ui.showNotification(`Update failed: ${err.message}`, 'error');
        });
}

function handleLogout() {
    api.logout();
    window.location.href = '/';
}

function initAuthPage() {
    document.getElementById('login-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('signup-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('show-signup-form')?.addEventListener('click', ui.toggleAuthForms);
    document.getElementById('show-login-form')?.addEventListener('click', ui.toggleAuthForms);
}

async function initChatPage() {
    document.getElementById('logout-button')?.addEventListener('click', handleLogout);
    document.getElementById('profile-update-form')?.addEventListener('submit', handleProfileUpdateFormSubmit);

    try {
        const user = await api.getMe();
        store.setCurrentUser(user);
        ui.renderProfile(user);
    } catch (error) {
        console.error('Failed to fetch user profile, redirecting to login.', error);
        store.clearTokens();
        window.location.href = '/';
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const path = window.location.pathname;
    if (path === '/chat') {
        if (!store.getAccessToken()) {
            window.location.href = '/';
            return;
        }
        initChatPage();
    } else {
        initAuthPage();
    }
});


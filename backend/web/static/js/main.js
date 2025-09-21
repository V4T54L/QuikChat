import * as api from './api.js';
import * as store from './store.js';
import * as ui from './ui.js';
import * as ws from './ws.js';

// --- Event Handlers ---

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
        }
        window.location.href = '/chat';
    } catch (error) {
        ui.showError(form.id, error.message);
    }
}

async function handleProfileUpdateFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    // Don't send empty values
    if (!formData.get('username').trim()) {
        formData.delete('username');
    }
    if (!formData.get('password').trim()) {
        formData.delete('password');
    }
    if (formData.get('profile_pic') && formData.get('profile_pic').size === 0) {
        formData.delete('profile_pic');
    }

    try {
        const updatedUser = await api.updateProfile(formData);
        store.setCurrentUser(updatedUser);
        ui.renderProfile(updatedUser);
        ui.showNotification('Profile updated successfully!', 'success');
        form.reset();
    } catch (error) {
        ui.showNotification(`Update failed: ${error.message}`, 'error');
    }
}

async function handleLogout() {
    await api.logout();
    window.location.href = '/';
}

async function handleAddFriendSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const usernameInput = form.querySelector('#add-friend-username');
    const username = usernameInput.value.trim();
    if (!username) return;

    try {
        await api.sendFriendRequest(username);
        ui.showNotification(`Friend request sent to ${username}`, 'success');
        usernameInput.value = '';
        ui.closeModal('add-friend-modal');
        // Optionally, refresh pending requests list
        const requests = await api.getPendingFriendRequests();
        store.setPendingRequests(requests);
        ui.renderPendingRequests(requests, store.getCurrentUser().id);
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

async function handleFriendAction(event) {
    const button = event.target.closest('button');
    if (!button) return;

    const action = button.dataset.action;
    const id = button.dataset.id;

    try {
        if (action === 'accept-friend') {
            await api.acceptFriendRequest(id);
            ui.showNotification('Friend request accepted!', 'success');
        } else if (action === 'reject-friend') {
            await api.rejectFriendRequest(id);
            ui.showNotification('Friend request rejected.', 'info');
        } else if (action === 'unfriend') {
            if (confirm('Are you sure you want to unfriend this user?')) {
                await api.unfriendUser(id);
                ui.showNotification('User unfriended.', 'info');
            } else {
                return;
            }
        }
        // Refresh both lists after any action
        await Promise.all([
            refreshFriendsList(),
            refreshPendingRequests()
        ]);
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

// --- Initialization ---

function initAuthPage() {
    document.getElementById('login-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('signup-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('show-signup-form')?.addEventListener('click', ui.toggleAuthForms);
    document.getElementById('show-login-form')?.addEventListener('click', ui.toggleAuthForms);
}

async function refreshFriendsList() {
    const friends = await api.listFriends();
    store.setFriends(friends);
    ui.renderFriendList(friends);
}

async function refreshPendingRequests() {
    const requests = await api.getPendingFriendRequests();
    store.setPendingRequests(requests);
    ui.renderPendingRequests(requests, store.getCurrentUser().id);
}

async function initChatPage() {
    document.getElementById('logout-button')?.addEventListener('click', handleLogout);
    document.getElementById('profile-update-form')?.addEventListener('submit', handleProfileUpdateFormSubmit);
    document.getElementById('add-friend-button')?.addEventListener('click', () => ui.openModal('add-friend-modal'));
    document.getElementById('add-friend-form')?.addEventListener('submit', handleAddFriendSubmit);
    document.getElementById('friends-container')?.addEventListener('click', handleFriendAction);

    try {
        const user = await api.getMe();
        store.setCurrentUser(user);
        ui.renderProfile(user);

        await Promise.all([
            refreshFriendsList(),
            refreshPendingRequests()
        ]);

        ws.connect();
    } catch (error) {
        console.error('Auth failed, redirecting to login.', error);
        api.logout(); // Clear any invalid tokens
        window.location.href = '/';
    }
}

// --- Main Execution ---

document.addEventListener('DOMContentLoaded', () => {
    const path = window.location.pathname;
    if (path === '/' || path === '/index.html') {
        initAuthPage();
    } else if (path.startsWith('/chat')) {
        initChatPage();
    }
});


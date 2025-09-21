import * as api from './api.js';
import * as store from './store.js';
import * as ui from './ui.js';
import * as ws from './ws.js';

// --- Event Handlers ---

async function handleAuthFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const isSignUp = form.id === 'signup-form';
    const username = form.username.value;
    const password = form.password.value;
    const errorEl = form.querySelector('.error-message');
    if (errorEl) errorEl.classList.add('hidden'); // Hide previous errors

    try {
        let data;
        if (isSignUp) {
            await api.signup(username, password);
            // After successful signup, automatically log in
            data = await api.login(username, password);
        } else {
            data = await api.login(username, password);
        }
        store.setTokens(data.access_token, data.refresh_token);
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
    if (!formData.get('username').trim()) formData.delete('username');
    if (!formData.get('password').trim()) formData.delete('password');
    if (formData.get('profile_pic') && formData.get('profile_pic').size === 0) formData.delete('profile_pic');

    try {
        const updatedUser = await api.updateProfile(formData);
        store.setCurrentUser(updatedUser);
        ui.renderProfile(updatedUser);
        ui.showNotification('Profile updated successfully!', 'success');
        form.reset(); // Clear form fields after successful update
    } catch (error) {
        ui.showNotification(`Profile update failed: ${error.message}`, 'error');
    }
}

function handleLogout() {
    api.logout(); // Clears tokens from store and local storage
    window.location.href = '/';
}

async function handleAddFriendSubmit(event) {
    event.preventDefault();
    const usernameInput = document.getElementById('add-friend-username');
    const username = usernameInput.value.trim();
    if (!username) return;

    try {
        await api.sendFriendRequest(username);
        ui.showNotification(`Friend request sent to ${username}`, 'success');
        ui.closeModal('add-friend-modal');
        usernameInput.value = ''; // Clear input field
        refreshPendingRequests();
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

async function handleFriendAction(event) {
    const button = event.target;
    try {
        if (button.classList.contains('accept-request-button')) {
            const requestId = button.dataset.requestId;
            await api.acceptFriendRequest(requestId);
            ui.showNotification('Friend request accepted!', 'success');
        } else if (button.classList.contains('reject-request-button')) {
            const requestId = button.dataset.requestId;
            await api.rejectFriendRequest(requestId);
            ui.showNotification('Friend request rejected.', 'info');
        } else if (button.classList.contains('unfriend-button')) {
            const friendId = button.dataset.friendId;
            if (confirm('Are you sure you want to unfriend this user?')) {
                await api.unfriendUser(friendId);
                ui.showNotification('User unfriended.', 'info');
            } else {
                return; // Do not proceed with refresh if cancelled
            }
        } else {
            return; // Not a recognized action button
        }
        refreshFriendsList();
        refreshPendingRequests();
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

async function handleCreateGroupSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    // Ensure handle starts with '#'
    const handleInput = form.querySelector('#create-group-handle');
    formData.set('handle', `#${handleInput.value.replace(/^#/, '')}`);

    try {
        const group = await api.createGroup(formData);
        ui.showNotification(`Group "${group.name}" created!`, 'success');
        ui.closeModal('create-group-modal');
        form.reset();
        refreshGroupsList();
    } catch (error) {
        ui.showNotification(`Error creating group: ${error.message}`, 'error');
    }
}

async function handleSearchGroup(event) {
    const query = event.target.value.trim();
    if (query.length < 2) {
        document.getElementById('search-group-results').innerHTML = '';
        return;
    }
    try {
        const groups = await api.searchGroups(query);
        ui.renderGroupSearchResults(groups, handleJoinGroupClick);
    } catch (error) {
        ui.showNotification(`Error searching groups: ${error.message}`, 'error');
    }
}

async function handleJoinGroupClick(event) {
    const handle = event.target.dataset.handle;
    try {
        await api.joinGroup(handle);
        ui.showNotification(`Successfully joined group #${handle}`, 'success');
        refreshGroupsList();
        ui.closeModal('search-group-modal');
    } catch (error) {
        ui.showNotification(`Error joining group: ${error.message}`, 'error');
    }
}

// --- Data Refresh Functions ---

export async function refreshFriendsList() {
    try {
        const friends = await api.listFriends();
        store.setFriends(friends);
        ui.renderFriendList(friends);
    } catch (error) {
        console.error('Failed to refresh friends list:', error);
        ui.showNotification('Failed to load friends.', 'error');
    }
}

export async function refreshPendingRequests() {
    try {
        const requests = await api.getPendingFriendRequests();
        store.setPendingRequests(requests);
        const currentUser = store.getCurrentUser();
        if (currentUser) {
            ui.renderPendingRequests(requests, currentUser.id);
        }
    } catch (error) {
        console.error('Failed to refresh pending requests:', error);
        ui.showNotification('Failed to load pending requests.', 'error');
    }
}

export async function refreshGroupsList() {
    try {
        const groups = await api.listMyGroups();
        store.setGroups(groups);
        ui.renderGroupList(groups);
    } catch (error) {
        console.error('Failed to refresh groups list:', error);
        ui.showNotification('Failed to load groups.', 'error');
    }
}

// --- Initialization ---

function initAuthPage() {
    document.getElementById('login-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('signup-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('show-signup-form')?.addEventListener('click', ui.toggleAuthForms);
    document.getElementById('show-login-form')?.addEventListener('click', ui.toggleAuthForms);
}

async function initChatPage() {
    // Bind static event listeners
    document.getElementById('logout-button').addEventListener('click', handleLogout);
    document.getElementById('profile-update-form').addEventListener('submit', handleProfileUpdateFormSubmit);
    document.getElementById('add-friend-button').addEventListener('click', () => ui.openModal('add-friend-modal'));
    document.getElementById('add-friend-form').addEventListener('submit', handleAddFriendSubmit);
    document.getElementById('create-group-button').addEventListener('click', () => ui.openModal('create-group-modal'));
    document.getElementById('create-group-form').addEventListener('submit', handleCreateGroupSubmit);
    document.getElementById('search-group-button').addEventListener('click', () => ui.openModal('search-group-modal'));
    document.getElementById('search-group-query').addEventListener('input', handleSearchGroup);

    // Bind dynamic event listeners for friend actions (delegation)
    document.getElementById('friends-container').addEventListener('click', handleFriendAction);
    document.getElementById('pending-requests-list').addEventListener('click', handleFriendAction); // Listen for accept/reject on pending requests

    // Initial data fetch
    try {
        const user = await api.getMe();
        store.setCurrentUser(user);
        ui.renderProfile(user);

        await Promise.all([
            refreshFriendsList(),
            refreshPendingRequests(),
            refreshGroupsList(),
        ]);

        // Connect to WebSocket
        ws.connect();

    } catch (error) {
        console.error('Failed to initialize chat page:', error);
        ui.showNotification('Failed to load chat data. Please log in again.', 'error');
        store.clearTokens(); // Clear any invalid tokens
        window.location.href = '/';
    }
}

// --- Main Execution ---

document.addEventListener('DOMContentLoaded', () => {
    const path = window.location.pathname;
    if (path === '/' || path === '/index.html') {
        if (store.getAccessToken()) {
            // If tokens exist, redirect to chat page
            window.location.href = '/chat';
        } else {
            initAuthPage();
        }
    } else if (path.startsWith('/chat')) {
        if (!store.getAccessToken()) {
            // If no tokens, redirect to login page
            window.location.href = '/';
        } else {
            initChatPage();
        }
    }
});

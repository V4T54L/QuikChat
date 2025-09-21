import * as api from './api.js';
import * as store from './store.js';
import * as ui from './ui.js';
import * as ws from './ws.js';

// --- AUTH PAGE ---
async function handleAuthFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const isSignUp = form.id === 'signup-form';
    const username = form.username.value;
    const password = form.password.value;

    try {
        const data = isSignUp
            ? await api.signup(username, password)
            : await api.login(username, password);

        if (isSignUp) {
            ui.showNotification('Signup successful! Please log in.', 'success');
            ui.toggleAuthForms();
        } else {
            store.setTokens(data.access_token, data.refresh_token);
            window.location.href = '/chat.html'; // Redirect to chat.html
        }
    } catch (error) {
        ui.showError(form.id, error.message);
    }
}

function initAuthPage() {
    document.getElementById('login-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('signup-form')?.addEventListener('submit', handleAuthFormSubmit);
    document.getElementById('show-signup-form')?.addEventListener('click', ui.toggleAuthForms);
    document.getElementById('show-login-form')?.addEventListener('click', ui.toggleAuthForms);
}

// --- CHAT PAGE ---
async function handleProfileUpdateFormSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);

    // Remove empty fields so they are not sent
    if (!formData.get('username').trim()) formData.delete('username');
    if (!formData.get('password').trim()) formData.delete('password');
    if (!formData.get('profile_pic').name) formData.delete('profile_pic'); // More robust check for file input

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

function handleLogout() {
    api.logout();
    window.location.href = '/';
}

async function handleAddFriendSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const username = form.username.value;
    try {
        await api.sendFriendRequest(username);
        ui.showNotification(`Friend request sent to ${username}`, 'success');
        form.reset();
        ui.closeModal('add-friend-modal');
        await refreshPendingRequests();
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

async function handleFriendAction(event) {
    const button = event.target.closest('button'); // Use closest for delegation
    if (!button) return;

    const action = button.dataset.action;
    const id = button.dataset.id;

    try {
        if (action === 'accept') {
            await api.acceptFriendRequest(id);
            ui.showNotification('Friend request accepted!', 'success');
        } else if (action === 'reject') {
            await api.rejectFriendRequest(id);
            ui.showNotification('Friend request rejected.', 'info');
        } else if (action === 'unfriend') {
            if (confirm('Are you sure you want to unfriend this user?')) {
                await api.unfriendUser(id);
                ui.showNotification('User unfriended.', 'info');
            }
        }
        await Promise.all([refreshFriendsList(), refreshPendingRequests()]);
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

async function handleCreateGroupSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    try {
        await api.createGroup(formData);
        ui.showNotification('Group created successfully!', 'success');
        form.reset();
        ui.closeModal('create-group-modal');
        await refreshGroupsList();
    } catch (error) {
        ui.showNotification(`Error: ${error.message}`, 'error');
    }
}

let searchTimeout;
async function handleSearchGroup(event) {
    clearTimeout(searchTimeout);
    const query = event.target.value;
    if (query.length < 2) {
        ui.renderGroupSearchResults([]);
        return;
    }
    searchTimeout = setTimeout(async () => {
        try {
            const groups = await api.searchGroups(query);
            ui.renderGroupSearchResults(groups, handleJoinGroupClick);
        } catch (error) {
            ui.showNotification(`Search failed: ${error.message}`, 'error');
        }
    }, 300);
}

async function handleJoinGroupClick(event) {
    const handle = event.target.dataset.handle;
    try {
        await api.joinGroup(handle);
        ui.showNotification(`Joined group ${handle}!`, 'success');
        await refreshGroupsList();
    } catch (error) {
        ui.showNotification(`Failed to join: ${error.message}`, 'error');
    }
}

async function handleMessageSubmit(event) {
    event.preventDefault();
    const form = event.target;
    const input = form.querySelector('input[name="message"]');
    const content = input.value.trim();
    const activeChat = store.getActiveChat();

    if (content && activeChat && activeChat.id) {
        ws.send('send_message', {
            conversation_id: activeChat.id,
            content: content,
        });
        input.value = '';
    }
}

export async function handleNewMessage(message) {
    const activeChat = store.getActiveChat();
    const currentUser = store.getCurrentUser();

    if (activeChat && activeChat.id === message.conversation_id) {
        ui.appendMessage(message, currentUser.id);
    } else {
        const senderName = message.sender?.username || 'Someone';
        ui.showNotification(`New message from ${senderName}`, 'info');
        // Potentially update a badge count on the conversation list item
    }
}

async function selectChat(event) {
    const target = event.currentTarget;
    const type = target.dataset.chatType;
    const id = target.dataset.chatId;
    const name = target.dataset.chatName;
    const pic = target.dataset.chatPic;

    let conversationId = id;
    if (type === 'friend') {
        const currentUser = store.getCurrentUser();
        // Ensure consistent conversation ID for direct messages
        conversationId = [currentUser.id, id].sort().join(':');
    }

    store.setActiveChat(type, conversationId, name);
    ui.renderChatWindow({ type, id, name, pic });
    ui.clearMessages();

    try {
        const messages = await api.getMessageHistory(conversationId);
        if (messages && messages.length > 0) {
            ui.prependMessages(messages, store.getCurrentUser().id);
        }
    } catch (error) {
        ui.showNotification(`Failed to load messages: ${error.message}`, 'error');
    }
}

export async function refreshFriendsList() {
    try {
        const friends = await api.listFriends();
        store.setFriends(friends || []); // Ensure array
        ui.renderFriendList(friends || [], selectChat); // Pass selectChat handler
    } catch (error) {
        console.error('Failed to refresh friends list:', error);
    }
}

export async function refreshPendingRequests() {
    try {
        const requests = await api.getPendingFriendRequests();
        store.setPendingRequests(requests || []); // Ensure array
        ui.renderPendingRequests(requests || [], store.getCurrentUser().id);
    } catch (error) {
        console.error('Failed to refresh pending requests:', error);
    }
}

export async function refreshGroupsList() {
    try {
        const groups = await api.listMyGroups();
        store.setGroups(groups || []); // Ensure array
        ui.renderGroupList(groups || [], selectChat); // Pass selectChat handler
    } catch (error) {
        console.error('Failed to refresh groups list:', error);
    }
}

async function initChatPage() {
    if (!store.getAccessToken()) {
        window.location.href = '/';
        return;
    }

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

        ws.connect();
    } catch (error) {
        console.error('Initialization failed:', error);
        handleLogout(); // If we can't get user data, log out
    }

    // Bind event listeners
    document.getElementById('profile-update-form').addEventListener('submit', handleProfileUpdateFormSubmit);
    document.getElementById('logout-button').addEventListener('click', handleLogout);
    document.getElementById('add-friend-form').addEventListener('submit', handleAddFriendSubmit);
    document.getElementById('create-group-form').addEventListener('submit', handleCreateGroupSubmit);
    document.getElementById('group-search-input').addEventListener('input', handleSearchGroup); // Updated ID
    document.getElementById('message-form').addEventListener('submit', handleMessageSubmit); // Added message form listener

    // Event delegation for dynamic lists
    document.getElementById('friends-list').addEventListener('click', handleFriendAction);
    document.getElementById('pending-requests-list').addEventListener('click', handleFriendAction);

    // Modal triggers
    document.getElementById('add-friend-btn').addEventListener('click', () => ui.openModal('add-friend-modal')); // Updated ID
    document.getElementById('create-group-btn').addEventListener('click', () => ui.openModal('create-group-modal')); // Updated ID
    document.getElementById('find-group-btn').addEventListener('click', () => ui.openModal('search-group-modal')); // Updated ID
}

// --- MAIN ---
document.addEventListener('DOMContentLoaded', () => {
    if (window.location.pathname.includes('chat.html')) { // Check for chat.html
        initChatPage();
    } else {
        initAuthPage();
    }
});

import { api } from './api.js';
import { getState, setState } from './state.js';
import { ui } from './ui.js';
import { ws } from './ws.js';

const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const messageForm = document.getElementById('message-form');
const logoutBtn = document.getElementById('logout-btn');
const loginTabBtn = document.getElementById('login-tab-btn');
const registerTabBtn = document.getElementById('register-tab-btn');

async function handleLogin(e) {
    e.preventDefault();
    const button = e.target.querySelector('button');
    ui.toggleButton(button, true);
    try {
        const formData = new FormData(e.target);
        const { username, password } = Object.fromEntries(formData.entries());
        const { accessToken, refreshToken } = await api.login(username, password);
        setState({ accessToken, refreshToken });
        await initializeApp();
    } catch (error) {
        ui.showAuthError(error.message);
    } finally {
        ui.toggleButton(button, false);
    }
}

async function handleRegister(e) {
    e.preventDefault();
    const button = e.target.querySelector('button');
    ui.toggleButton(button, true);
    try {
        const formData = new FormData(e.target);
        const { username, password } = Object.fromEntries(formData.entries());
        await api.register(username, password);
        alert('Registration successful! Please log in.');
        e.target.reset();
        ui.switchAuthTab('login');
    } catch (error) {
        ui.showAuthError(error.message);
    } finally {
        ui.toggleButton(button, false);
    }
}

async function handleLogout() {
    try {
        const { refreshToken } = getState();
        if (refreshToken) {
            await api.logout(refreshToken);
        }
    } catch (error) {
        console.error('Logout failed:', error);
    } finally {
        ws.disconnect();
        setState({
            currentUser: null,
            accessToken: null,
            refreshToken: null,
            friends: [],
            groups: [],
            activeChat: null,
        });
        ui.showView('auth');
    }
}

function handleSendMessage(e) {
    e.preventDefault();
    const { activeChat } = getState();
    const input = e.target.elements.message;
    const content = input.value.trim();

    if (content && activeChat) {
        const payload = {
            content,
            recipientId: activeChat.id,
        };
        ws.sendMessage({ type: 'message_sent', payload });
        ui.clearMessageInput();
    }
}

function setupWsListeners() {
    ws.onEvent('message_sent', (message) => {
        const { currentUser, activeChat } = getState();
        if (
            activeChat &&
            (message.recipientId === activeChat.id || message.senderId === activeChat.id)
        ) {
            ui.renderMessage(message, message.senderId === currentUser.id);
        }
        // TODO: Add notification for inactive chats
    });

    ws.onEvent('message_ack', (payload) => {
        console.log('Message acknowledged:', payload.messageId);
        // Can be used to update message status to "sent"
    });
}

async function initializeApp() {
    ui.showView('chat');
    const { accessToken } = getState();

    try {
        // Mock current user from token until we have a /me endpoint
        const tokenPayload = JSON.parse(atob(accessToken.split('.')[1]));
        const currentUser = { id: tokenPayload.user_id, username: 'You' }; // Username is a placeholder
        setState({ currentUser });

        ui.renderUserProfile(currentUser);

        ws.connect(accessToken);
        setupWsListeners();

        const friends = await api.getFriends();
        setState({ friends });
        ui.renderFriendList(friends);

        const groups = await api.getGroups();
        setState({ groups });
        ui.renderGroupList(groups);
    } catch (error) {
        console.error('Initialization failed:', error);
        await handleLogout();
    }
}

// Event Listeners
loginForm.addEventListener('submit', handleLogin);
registerForm.addEventListener('submit', handleRegister);
messageForm.addEventListener('submit', handleSendMessage);
logoutBtn.addEventListener('click', handleLogout);
loginTabBtn.addEventListener('click', () => ui.switchAuthTab('login'));
registerTabBtn.addEventListener('click', () => ui.switchAuthTab('register'));

// Initial check (e.g., for a stored refresh token) could go here
// For now, we start at the auth view.
ui.showView('auth');


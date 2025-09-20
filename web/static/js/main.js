import { api } from './api.js';
import { getState, setState } from './state.js';
import { ui } from './ui.js';
import { ws }_ from './ws.js';

document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    const messageForm = document.getElementById('message-form');
    const logoutBtn = document.getElementById('logout-btn');
    const loginTabBtn = document.getElementById('login-tab-btn');
    const registerTabBtn = document.getElementById('register-tab-btn');

    const handleLogin = async (e) => {
        e.preventDefault();
        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;
        try {
            const data = await api.login(username, password);
            setState({ 
                accessToken: data.access_token, 
                refreshToken: data.refresh_token,
            });
            awaitinitializeApp();
        } catch (error) {
            ui.showAuthError(error.message);
        }
    };

    const handleRegister = async (e) => {
        e.preventDefault();
        const username = document.getElementById('register-username').value;
        const password = document.getElementById('register-password').value;
        try {
            const user = await api.register(username, password);
            alert(`Registration successful for ${user.username}! Please log in.`);
            ui.switchAuthTab('login');
            document.getElementById('login-username').value = username;
            document.getElementById('login-password').focus();
        } catch (error) {
            ui.showAuthError(error.message);
        }
    };

    const handleLogout = async () => {
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
    };

    const handleSendMessage = (e) => {
        e.preventDefault();
        const input = document.getElementById('message-input');
        const content = input.value.trim();
        const { activeChat } = getState();

        if (content && activeChat) {
            ws.sendMessage({
                type: 'message_sent',
                payload: {
                    content,
                    recipientId: activeChat.id,
                },
            });
            input.value = '';
        }
    };

    const initializeApp = async () => {
        ui.showView('chat');
        
        // Mock current user data - a real app would have a /users/me endpoint
        const { accessToken } = getState();
        const decodedToken = JSON.parse(atob(accessToken.split('.')[1]));
        const currentUser = { id: decodedToken.user_id, username: 'You' }; // Username is a placeholder
        setState({ currentUser });

        ui.renderUserProfile(currentUser);

        ws.connect(accessToken);

        try {
            const [friends, groups] = await Promise.all([
                api.getFriends(),
                api.getGroups(), // Mocked
            ]);
            setState({ friends, groups });
            ui.renderFriendList(friends);
            ui.renderGroupList(groups);
        } catch (error) {
            console.error('Failed to fetch initial data:', error);
            handleLogout();
        }
    };

    // Event Listeners
    loginForm.addEventListener('submit', handleLogin);
    registerForm.addEventListener('submit', handleRegister);
    messageForm.addEventListener('submit', handleSendMessage);
    logoutBtn.addEventListener('click', handleLogout);
    loginTabBtn.addEventListener('click', () => ui.switchAuthTab('login'));
    registerTabBtn.addEventListener('click', () => ui.switchAuthTab('register'));

    // WebSocket Event Handlers
    ws.onEvent('message_sent', (message) => {
        const { activeChat, currentUser } = getState();
        if (activeChat && (message.recipientId === activeChat.id || message.senderId === activeChat.id)) {
            ui.renderMessage(message, message.senderId === currentUser.id);
        }
    });

    ws.onEvent('message_ack', (ack) => {
        console.log('Message acknowledged:', ack);
        // Here you could update the UI to show a "sent" checkmark
    });

    // Initial UI setup
    ui.showView('auth');
});


import { setState } from './state.js';
import { userProfileTemplate, friendListTemplate, groupListTemplate, messageTemplate } from './templates.js';

const authView = document.getElementById('auth-view');
const chatView = document.getElementById('chat-view');
const authError = document.getElementById('auth-error');
const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const loginTabBtn = document.getElementById('login-tab-btn');
const registerTabBtn = document.getElementById('register-tab-btn');
const userProfileContainer = document.getElementById('user-profile-container');
const friendsListContainer = document.getElementById('friends-list-container');
const groupsListContainer = document.getElementById('groups-list-container');
const messagesContainer = document.getElementById('messages-container');
const chatTitle = document.getElementById('chat-title');
const messageInput = document.getElementById('message-input');
const messageFormBtn = document.querySelector('#message-form button');

const setActiveChat = (chat) => {
    setState({ activeChat: chat });
    chatTitle.textContent = chat.name;
    messagesContainer.innerHTML = ''; // Clear previous messages
    messageInput.disabled = false;
    messageFormBtn.disabled = false;
    messageInput.focus();
    // In a real app, you would fetch message history here
};

export const ui = {
    showView: (view) => {
        authView.classList.toggle('hidden', view === 'chat');
        chatView.classList.toggle('hidden', view === 'auth');
    },
    showAuthError: (message) => {
        authError.textContent = message;
    },
    switchAuthTab: (tab) => {
        const isLogin = tab === 'login';
        loginForm.classList.toggle('hidden', !isLogin);
        registerForm.classList.toggle('hidden', isLogin);
        loginTabBtn.classList.toggle('text-highlight', isLogin);
        loginTabBtn.classList.toggle('border-highlight', isLogin);
        loginTabBtn.classList.toggle('text-text-dim', !isLogin);
        registerTabBtn.classList.toggle('text-highlight', !isLogin);
        registerTabBtn.classList.toggle('border-highlight', !isLogin);
        registerTabBtn.classList.toggle('text-text-dim', isLogin);
        authError.textContent = '';
    },
    renderUserProfile: (user) => {
        userProfileContainer.innerHTML = userProfileTemplate(user);
    },
    renderFriendList: (friends) => {
        friendsListContainer.innerHTML = friendListTemplate(friends);
        friendsListContainer.querySelectorAll('li').forEach(item => {
            item.addEventListener('click', () => {
                setActiveChat({
                    id: item.dataset.id,
                    name: item.dataset.username,
                    type: 'user',
                });
            });
        });
    },
    renderGroupList: (groups) => {
        groupsListContainer.innerHTML = groupListTemplate(groups);
        groupsListContainer.querySelectorAll('li').forEach(item => {
            item.addEventListener('click', () => {
                setActiveChat({
                    id: item.dataset.id,
                    name: item.dataset.name,
                    type: 'group',
                });
            });
        });
    },
    renderMessage: (message, isOwnMessage) => {
        messagesContainer.insertAdjacentHTML('beforeend', messageTemplate(message, isOwnMessage));
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    },
};


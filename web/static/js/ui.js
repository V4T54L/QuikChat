import { setState } from './state.js';
import {
    userProfileTemplate,
    friendListTemplate,
    groupListTemplate,
    messageTemplate,
} from './templates.js';

// DOM Elements
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
const messageForm = document.getElementById('message-form');
const messageFormBtn = messageForm.querySelector('button');

/** @param {import('./types.js').ActiveChat} chat */
function setActiveChat(chat) {
    setState({ activeChat: chat });
    chatTitle.textContent = chat.name;
    messagesContainer.innerHTML = ''; // Clear previous messages
    messageInput.disabled = false;
    messageFormBtn.disabled = false;
    messageInput.focus();
    // TODO: Fetch message history for this chat
}

export const ui = {
    /** @param {'auth' | 'chat'} view */
    showView(view) {
        authView.classList.toggle('hidden', view === 'chat');
        chatView.classList.toggle('hidden', view === 'auth');
    },

    /** @param {string} message */
    showAuthError(message) {
        authError.textContent = message;
        authError.classList.toggle('hidden', !message);
    },

    /** @param {'login' | 'register'} tab */
    switchAuthTab(tab) {
        const isLogin = tab === 'login';
        loginForm.classList.toggle('hidden', !isLogin);
        registerForm.classList.toggle('hidden', isLogin);

        loginTabBtn.classList.toggle('border-accent', isLogin);
        loginTabBtn.classList.toggle('text-white', isLogin);
        loginTabBtn.classList.toggle('border-transparent', !isLogin);
        loginTabBtn.classList.toggle('text-text-dim', !isLogin);

        registerTabBtn.classList.toggle('border-accent', !isLogin);
        registerTabBtn.classList.toggle('text-white', !isLogin);
        registerTabBtn.classList.toggle('border-transparent', isLogin);
        registerTabBtn.classList.toggle('text-text-dim', isLogin);

        ui.showAuthError(''); // Clear error on tab switch
    },

    /**
     * @param {HTMLButtonElement} button
     * @param {boolean} disabled
     */
    toggleButton(button, disabled) {
        if (button) {
            button.disabled = disabled;
            button.classList.toggle('opacity-50', disabled);
            button.classList.toggle('cursor-not-allowed', disabled);
        }
    },

    clearMessageInput() {
        messageInput.value = '';
    },

    /** @param {import('./types.js').User} user */
    renderUserProfile(user) {
        userProfileContainer.innerHTML = userProfileTemplate(user);
    },

    /** @param {import('./types.js').User[]} friends */
    renderFriendList(friends) {
        friendsListContainer.innerHTML = friendListTemplate(friends);
        friendsListContainer.querySelectorAll('li').forEach((item) => {
            item.addEventListener('click', () => {
                setActiveChat({
                    id: item.dataset.id,
                    name: item.dataset.username,
                    type: 'user',
                });
            });
        });
    },

    /** @param {import('./types.js').Group[]} groups */
    renderGroupList(groups) {
        groupsListContainer.innerHTML = groupListTemplate(groups);
        groupsListContainer.querySelectorAll('li').forEach((item) => {
            item.addEventListener('click', () => {
                setActiveChat({
                    id: item.dataset.id,
                    name: item.dataset.name,
                    type: 'group',
                });
            });
        });
    },

    /**
     * @param {import('./types.js').Message} message
     * @param {boolean} isOwnMessage
     */
    renderMessage(message, isOwnMessage) {
        const messageEl = document.createElement('div');
        messageEl.innerHTML = messageTemplate(message, isOwnMessage);
        messagesContainer.appendChild(messageEl);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    },
};

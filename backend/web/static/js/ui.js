/**
 * @typedef {import('./api.js').User} User
 * @typedef {import('./store.js').FriendRequest} FriendRequest
 * @typedef {import('./api.js').Group} Group
 * @typedef {import('./api.js').Message} Message // Added Message typedef
 */

// --- AUTH ---
export function toggleAuthForms() {
    document.getElementById('login-container').classList.toggle('hidden');
    document.getElementById('signup-container').classList.toggle('hidden');
}

/**
 * @param {string} formId
 * @param {string} message
 */
export function showError(formId, message) {
    const errorEl = document.querySelector(`#${formId} .error-message`);
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.classList.remove('hidden');
    }
}

// --- NOTIFICATIONS ---
/**
 * @param {string} message
 * @param {'info'|'success'|'error'} type
 */
export function showNotification(message, type = 'info') {
    const container = document.getElementById('notification-container');
    if (!container) return; // Added null check
    const notification = document.createElement('div');
    const colors = {
        info: 'bg-blue-500',
        success: 'bg-green-500',
        error: 'bg-red-500',
    };
    notification.className = `text-white p-3 rounded-lg shadow-lg mb-2 ${colors[type] || colors.info}`; // Updated styling
    notification.textContent = message;
    container.appendChild(notification);
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

// --- MODALS ---
/**
 * @param {string} modalId
 */
export function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (!modal) return; // Added null check

    const overlay = modal.querySelector('.modal-overlay');
    const closeButtons = modal.querySelectorAll('.modal-close'); // Changed class name

    const closeModalHandler = () => closeModal(modalId);

    modal.classList.remove('hidden');
    document.body.classList.add('overflow-hidden'); // Added overflow-hidden to body

    overlay.onclick = closeModalHandler;
    closeButtons.forEach(btn => btn.onclick = closeModalHandler);
    document.onkeydown = (evt) => { // Updated keydown handler
        evt = evt || window.event;
        if (evt.key === "Escape") {
            closeModalHandler();
        }
    };

    const firstInput = modal.querySelector('input, textarea, select'); // Focus on first input
    if (firstInput) {
        firstInput.focus();
    }
}

/**
 * @param {string} modalId
 */
export function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('hidden');
        document.body.classList.remove('overflow-hidden'); // Removed overflow-hidden from body
        // Clean up event listeners
        const overlay = modal.querySelector('.modal-overlay');
        if (overlay) overlay.onclick = null; // Added null check
        modal.querySelectorAll('.modal-close').forEach(btn => btn.onclick = null);
        document.onkeydown = null;
    }
}

// --- CHAT PAGE RENDERING ---
/**
 * @param {User} user
 */
export function renderProfile(user) {
    const container = document.getElementById('profile-container');
    if (!container || !user) return;
    container.innerHTML = `
        <img src="${user.profile_pic_url || 'https://placehold.co/40'}" alt="My Profile" class="w-10 h-10 rounded-full mr-3">
        <div>
            <p class="font-semibold text-white">${user.username}</p>
            <p class="text-xs text-gray-400">Joined: ${new Date(user.created_at).toLocaleDateString()}</p>
        </div>
    `;
}

/**
 * @param {User[]} friends
 * @param {(event: Event) => void} selectChatHandler
 */
export function renderFriendList(friends, selectChatHandler) { // Added selectChatHandler
    const list = document.getElementById('friends-list');
    if (!list) return;
    if (friends.length === 0) { // Added empty state
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No friends yet.</li>';
        return;
    }
    list.innerHTML = friends.map(friend => `
        <li class="flex items-center justify-between p-2 hover:bg-gray-700 rounded-md cursor-pointer"
            data-chat-type="friend" data-chat-id="${friend.id}" data-chat-name="${friend.username}" data-chat-pic="${friend.profile_pic_url || 'https://placehold.co/40'}">
            <div class="flex items-center flex-grow">
                <img src="${friend.profile_pic_url || 'https://placehold.co/40'}" alt="${friend.username}" class="w-8 h-8 rounded-full mr-3">
                <span class="text-gray-300">${friend.username}</span>
            </div>
            <button data-action="unfriend" data-id="${friend.id}" class="text-red-500 hover:text-red-400 text-xs p-1">Unfriend</button>
        </li>
    `).join('');
    list.querySelectorAll('li').forEach(item => item.addEventListener('click', selectChatHandler)); // Added event listener
}

/**
 * @param {Group[]} groups
 * @param {(event: Event) => void} selectChatHandler
 */
export function renderGroupList(groups, selectChatHandler) { // Added selectChatHandler
    const list = document.getElementById('groups-list');
    if (!list) return;
    if (groups.length === 0) { // Added empty state
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No groups yet.</li>';
        return;
    }
    list.innerHTML = groups.map(group => `
        <li class="flex items-center p-2 hover:bg-gray-700 rounded-md cursor-pointer"
            data-chat-type="group" data-chat-id="${group.id}" data-chat-name="${group.name}" data-chat-pic="${group.photo_url || 'https://placehold.co/40'}">
            <img src="${group.photo_url || 'https://placehold.co/40'}" alt="${group.name}" class="w-8 h-8 rounded-full mr-3">
            <div>
                <p class="text-gray-300">${group.name}</p>
                <p class="text-xs text-gray-500">${group.handle}</p>
            </div>
        </li>
    `).join('');
    list.querySelectorAll('li').forEach(item => item.addEventListener('click', selectChatHandler)); // Added event listener
}

/**
 * @param {FriendRequest[]} requests
 * @param {string} currentUserID
 */
export function renderPendingRequests(requests, currentUserID) {
    const list = document.getElementById('pending-requests-list');
    if (!list) return;
    if (requests.length === 0) { // Added empty state
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No pending requests.</li>';
        return;
    }
    list.innerHTML = requests.map(req => {
        const isIncoming = req.receiver.id === currentUserID; // Changed to req.receiver.id
        const user = isIncoming ? req.sender : req.receiver; // Changed to req.sender/receiver
        return `
            <li class="flex items-center justify-between p-2 rounded-md bg-gray-700 mb-2">
                <div class="flex items-center">
                    <img src="${user.profile_pic_url || 'https://via.placeholder.com/32'}" alt="${user.username}" class="w-8 h-8 rounded-full mr-3">
                    <div>
                        <p class="text-gray-300">${user.username}</p>
                        <p class="text-xs ${isIncoming ? 'text-green-400' : 'text-yellow-400'}">${isIncoming ? 'Incoming' : 'Outgoing'}</p>
                    </div>
                </div>
                ${isIncoming ? `
                <div>
                    <button data-action="accept" data-id="${req.id}" class="bg-green-500 hover:bg-green-600 text-white text-xs px-2 py-1 rounded">Accept</button>
                    <button data-action="reject" data-id="${req.id}" class="bg-red-500 hover:bg-red-600 text-white text-xs px-2 py-1 rounded ml-1">Reject</button>
                </div>
                ` : ''}
            </li>
        `;
    }).join('');
}

/**
 * @param {Group[]} groups
 * @param {(event: Event) => void} joinHandler
 */
export function renderGroupSearchResults(groups, joinHandler) {
    const resultsContainer = document.getElementById('group-search-results'); // Updated ID
    if (!resultsContainer) return;
    if (groups.length === 0) {
        resultsContainer.innerHTML = '<p class="text-gray-400 text-center">No groups found.</p>'; // Updated styling
        return;
    }
    resultsContainer.innerHTML = groups.map(group => `
        <div class="flex items-center justify-between p-2 border-b border-gray-600">
            <div class="flex items-center">
                <img src="${group.photo_url || 'https://placehold.co/40'}" alt="${group.name}" class="w-10 h-10 rounded-full mr-3">
                <div>
                    <p class="font-semibold text-white">${group.name}</p>
                    <p class="text-sm text-gray-400">${group.handle}</p>
                </div>
            </div>
            <button data-handle="${group.handle}" class="join-group-btn bg-blue-500 hover:bg-blue-600 text-white px-3 py-1 rounded">Join</button>
        </div>
    `).join('');
    resultsContainer.querySelectorAll('.join-group-btn').forEach(btn => btn.addEventListener('click', joinHandler)); // Updated class name
}

// --- MESSAGING UI ---
/**
 * @param {object} chat
 * @param {'friend'|'group'} chat.type
 * @param {string} chat.id
 * @param {string} chat.name
 * @param {string} chat.pic
 */
export function renderChatWindow(chat) { // New function for rendering chat window
    const container = document.getElementById('chat-container');
    const header = document.getElementById('chat-header');
    const welcomeScreen = document.getElementById('welcome-screen'); // Added welcome screen

    if (!container || !header || !welcomeScreen) return;

    if (!chat || !chat.id) {
        container.classList.add('hidden');
        welcomeScreen.classList.remove('hidden');
        return;
    }

    container.classList.remove('hidden');
    welcomeScreen.classList.add('hidden');
    header.innerHTML = `
        <img src="${chat.pic || 'https://placehold.co/40'}" alt="${chat.name}" class="w-10 h-10 rounded-full mr-4">
        <h2 class="text-xl font-semibold text-white">${chat.name}</h2>
    `;
}

/**
 * @param {Message} message
 * @param {string} currentUserID
 */
function renderMessage(message, currentUserID) { // New function for rendering a single message
    const isOwnMessage = message.sender_id === currentUserID;
    const senderName = isOwnMessage ? 'You' : message.sender?.username || '...';
    const time = new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

    return `
        <div class="flex items-start mb-4 ${isOwnMessage ? 'justify-end' : ''}">
            ${!isOwnMessage ? `<img src="${message.sender?.profile_pic_url || 'https://placehold.co/40'}" class="w-8 h-8 rounded-full mr-3">` : ''}
            <div class="flex flex-col ${isOwnMessage ? 'items-end' : 'items-start'}">
                <div class="${isOwnMessage ? 'bg-blue-600' : 'bg-gray-700'} text-white p-3 rounded-lg max-w-xs lg:max-w-md">
                    <p class="text-sm">${message.content}</p>
                </div>
                <div class="text-xs text-gray-500 mt-1">
                    ${!isOwnMessage ? `<span class="font-semibold">${senderName}</span> at ` : ''}
                    ${time}
                </div>
            </div>
        </div>
    `;
}

/**
 * @param {Message} message
 * @param {string} currentUserID
 */
export function appendMessage(message, currentUserID) { // New function to append messages
    const container = document.getElementById('messages-container');
    if (!container) return;
    container.innerHTML += renderMessage(message, currentUserID);
    container.scrollTop = container.scrollHeight;
}

/**
 * @param {Message[]} messages
 * @param {string} currentUserID
 */
export function prependMessages(messages, currentUserID) { // New function to prepend messages (for history)
    const container = document.getElementById('messages-container');
    if (!container) return;
    const oldScrollHeight = container.scrollHeight;
    container.innerHTML = messages.map(msg => renderMessage(msg, currentUserID)).join('') + container.innerHTML;
    container.scrollTop = container.scrollHeight - oldScrollHeight;
}

export function clearMessages() { // New function to clear messages
    const container = document.getElementById('messages-container');
    if (container) {
        container.innerHTML = '';
    }
}

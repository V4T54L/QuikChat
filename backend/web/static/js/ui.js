/**
 * @typedef {import('./api.js').User} User
 */
/**
 * @typedef {import('./store.js').FriendRequest} FriendRequest
 */
/**
 * @typedef {import('./api.js').Group} Group
 */

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

/**
 * @param {User} user
 */
export function renderProfile(user) {
    const container = document.getElementById('profile-container');
    if (!container || !user) return;
    const profilePic = user.profile_pic_url ? `/uploads/${user.profile_pic_url}` : 'https://via.placeholder.com/40';
    container.innerHTML = `
        <div class="flex items-center">
            <img src="${profilePic}" alt="Profile" class="w-10 h-10 rounded-full mr-3 object-cover">
            <div>
                <p class="font-semibold">${user.username}</p>
                <p class="text-sm text-gray-500">Joined: ${new Date(user.created_at).toLocaleDateString()}</p>
            </div>
        </div>
    `;
}

/**
 * @param {string} message
 * @param {'info'|'success'|'error'} type
 */
export function showNotification(message, type = 'info') {
    const container = document.getElementById('notification-container');
    if (!container) return;

    const colors = {
        info: 'bg-blue-500',
        success: 'bg-green-500',
        error: 'bg-red-500',
    };

    const notification = document.createElement('div');
    notification.className = `text-white px-4 py-2 rounded-md shadow-lg mb-2 ${colors[type] || colors.info}`;
    notification.textContent = message;

    container.appendChild(notification);

    setTimeout(() => {
        notification.remove();
    }, 3000);
}

/**
 * @param {User[]} friends
 */
export function renderFriendList(friends) {
    const list = document.getElementById('friends-list');
    if (!list) return;
    if (friends.length === 0) {
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No friends yet.</li>';
        return;
    }
    list.innerHTML = friends.map(friend => `
        <li class="flex justify-between items-center p-2 hover:bg-gray-100 rounded cursor-pointer" data-user-id="${friend.id}">
            <div class="flex items-center">
                <img src="${friend.profile_pic_url ? `/uploads/${friend.profile_pic_url}` : 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                <span>${friend.username}</span>
            </div>
            <button class="unfriend-button text-red-500 hover:text-red-700 text-xs" data-friend-id="${friend.id}">Unfriend</button>
        </li>
    `).join('');
}

/**
 * @param {Group[]} groups
 */
export function renderGroupList(groups) {
    const list = document.getElementById('groups-list');
    if (!list) return;
    if (groups.length === 0) {
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No groups yet.</li>';
        return;
    }
    list.innerHTML = groups.map(group => `
        <li class="flex justify-between items-center p-2 hover:bg-gray-100 rounded cursor-pointer" data-group-id="${group.id}">
            <div class="flex items-center">
                <img src="${group.photo_url ? `/uploads/${group.photo_url}` : 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                <span>${group.name} (${group.handle})</span>
            </div>
        </li>
    `).join('');
}

/**
 * @param {FriendRequest[]} requests
 * @param {string} currentUserID
 */
export function renderPendingRequests(requests, currentUserID) {
    const list = document.getElementById('pending-requests-list');
    if (!list) return;
    if (requests.length === 0) {
        list.innerHTML = '<li class="text-gray-500 text-sm p-2">No pending requests.</li>';
        return;
    }
    list.innerHTML = requests.map(req => {
        const isIncoming = req.receiver_id === currentUserID;
        const user = isIncoming ? req.Sender : req.Receiver; // Use capitalized Sender/Receiver as per WS event
        const actionButtons = isIncoming ? `
            <div>
                <button class="accept-request-button text-green-500 hover:text-green-700 text-xs mr-2" data-request-id="${req.id}">Accept</button>
                <button class="reject-request-button text-red-500 hover:text-red-700 text-xs" data-request-id="${req.id}">Reject</button>
            </div>
        ` : '<span class="text-xs text-gray-400">Sent</span>';

        return `
            <li class="flex justify-between items-center p-2 hover:bg-gray-100 rounded">
                <div class="flex items-center">
                    <img src="${user.profile_pic_url ? `/uploads/${user.profile_pic_url}` : 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                    <span>${user.username}</span>
                </div>
                ${actionButtons}
            </li>
        `;
    }).join('');
}

/**
 * @param {Group[]} groups
 * @param {(event: Event) => void} joinHandler
 */
export function renderGroupSearchResults(groups, joinHandler) {
    const resultsContainer = document.getElementById('search-group-results');
    if (!resultsContainer) return;
    if (groups.length === 0) {
        resultsContainer.innerHTML = '<p class="text-gray-500 p-2">No groups found.</p>';
        return;
    }
    resultsContainer.innerHTML = groups.map(group => `
        <div class="flex justify-between items-center p-2 border-b border-gray-200 last:border-b-0">
            <div class="flex items-center">
                <img src="${group.photo_url ? `/uploads/${group.photo_url}` : 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                <div>
                    <p class="font-semibold">${group.name}</p>
                    <p class="text-sm text-gray-600">${group.handle}</p>
                </div>
            </div>
            <button class="join-group-button px-3 py-1 bg-blue-500 text-white text-sm rounded hover:bg-blue-600" data-handle="${group.handle.substring(1)}">Join</button>
        </div>
    `).join('');

    document.querySelectorAll('.join-group-button').forEach(button => {
        button.addEventListener('click', joinHandler);
    });
}

/**
 * @param {string} modalId
 */
export function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('hidden');
        modal.classList.add('flex'); // Use flex to center content

        // Focus on the first input
        const firstInput = modal.querySelector('input[type="text"]');
        if (firstInput) {
            firstInput.focus();
        }

        // Define handlers to be able to remove them later
        const closeModalHandler = () => {
            closeModal(modalId);
            modal.removeEventListener('click', overlayClickHandler);
            document.removeEventListener('keydown', escapeKeyHandler);
        };

        const overlayClickHandler = (event) => {
            if (event.target === modal) {
                closeModalHandler();
            }
        };

        const escapeKeyHandler = (event) => {
            if (event.key === 'Escape') {
                closeModalHandler();
            }
        };

        modal.querySelectorAll('.modal-close-button').forEach(btn => {
            btn.onclick = closeModalHandler;
        });
        modal.addEventListener('click', overlayClickHandler);
        document.addEventListener('keydown', escapeKeyHandler);
    }
}

/**
 * @param {string} modalId
 */
export function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('hidden');
        modal.classList.remove('flex');
    }
}


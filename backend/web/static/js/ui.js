/**
 * @typedef {import('./api.js').User} User
 */
/**
 * @typedef {import('./store.js').FriendRequest} FriendRequest
 */

export function toggleAuthForms() {
    document.getElementById('login-container').classList.toggle('hidden');
    document.getElementById('signup-container').classList.toggle('hidden');
}

export function showError(formId, message) {
    const errorElement = document.querySelector(`#${formId} .error-message`);
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.classList.remove('hidden');
    }
}

/**
 * @param {User} user
 */
export function renderProfile(user) {
    const container = document.getElementById('profile-container');
    if (!container) return;
    const defaultPic = 'https://via.placeholder.com/100';
    container.innerHTML = `
        <img src="${user.profile_pic_url || defaultPic}" alt="Profile Picture" class="w-24 h-24 rounded-full mx-auto mb-4 object-cover">
        <h2 class="text-2xl font-bold text-center">${user.username}</h2>
        <p class="text-gray-500 text-center text-sm">Joined: ${new Date(user.created_at).toLocaleDateString()}</p>
    `;
}

export function showNotification(message, type = 'info') {
    const container = document.getElementById('notification-container');
    if (!container) return;

    const colors = {
        info: 'bg-blue-500',
        success: 'bg-green-500',
        error: 'bg-red-500',
    };

    const notification = document.createElement('div');
    notification.className = `p-4 rounded-lg text-white shadow-lg mb-2 ${colors[type] || colors.info}`;
    notification.textContent = message;

    container.appendChild(notification);

    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transition = 'opacity 0.5s ease';
        setTimeout(() => notification.remove(), 500);
    }, 3000);
}

/**
 * @param {User[]} friends
 */
export function renderFriendList(friends) {
    const list = document.getElementById('friends-list');
    if (!list) return;
    if (friends.length === 0) {
        list.innerHTML = '<p class="text-gray-400 px-4">No friends yet.</p>';
        return;
    }
    list.innerHTML = friends.map(friend => `
        <li class="flex items-center justify-between p-2 hover:bg-gray-700 rounded-md">
            <div class="flex items-center">
                <img src="${friend.profile_pic_url || 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                <span>${friend.username}</span>
            </div>
            <button data-action="unfriend" data-id="${friend.id}" class="text-red-500 hover:text-red-400 text-xs">Unfriend</button>
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
        list.innerHTML = '<p class="text-gray-400 px-4">No pending requests.</p>';
        return;
    }
    list.innerHTML = requests.map(req => {
        const isIncoming = req.receiver_id === currentUserID;
        const otherUser = isIncoming ? req.sender : req.receiver;
        return `
            <li class="flex items-center justify-between p-2 hover:bg-gray-700 rounded-md">
                <div class="flex items-center">
                    <img src="${otherUser.profile_pic_url || 'https://via.placeholder.com/40'}" class="w-8 h-8 rounded-full mr-3 object-cover">
                    <div>
                        <p>${otherUser.username}</p>
                        <p class="text-xs text-gray-400">${isIncoming ? 'Wants to be your friend' : 'Request sent'}</p>
                    </div>
                </div>
                ${isIncoming ? `
                <div class="flex items-center space-x-2">
                    <button data-action="accept-friend" data-id="${req.id}" class="text-green-500 hover:text-green-400">✓</button>
                    <button data-action="reject-friend" data-id="${req.id}" class="text-red-500 hover:text-red-400">✗</button>
                </div>
                ` : ''}
            </li>
        `;
    }).join('');
}

export function openModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.remove('hidden');
        modal.querySelector('input[type="text"]')?.focus();
        // Add event listener to close modal on escape key
        const closeOnEscape = (e) => {
            if (e.key === 'Escape') {
                closeModal(modalId);
                document.removeEventListener('keydown', closeOnEscape);
            }
        };
        document.addEventListener('keydown', closeOnEscape);
        // Add event listener to close modal on overlay click
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                closeModal(modalId);
            }
        });
    }
}

export function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.classList.add('hidden');
    }
}


/**
 * @typedef {import('./api.js').User} User
 */

export function toggleAuthForms() {
    document.getElementById('login-container').classList.toggle('hidden');
    document.getElementById('signup-container').classList.toggle('hidden');
}

export function showError(formId, message) {
    const errorEl = document.querySelector(`#${formId} .error-message`);
    if (errorEl) {
        errorEl.textContent = message;
        errorEl.classList.remove('hidden');
    }
}

/** @param {User} user */
export function renderProfile(user) {
    const profileContainer = document.getElementById('profile-container');
    if (!profileContainer) return;

    const profilePicUrl = user.profile_pic_url ? user.profile_pic_url : 'https://via.placeholder.com/100';

    profileContainer.innerHTML = `
        <img src="${profilePicUrl}" alt="Profile Picture" class="w-24 h-24 rounded-full mx-auto mb-4 object-cover">
        <h2 class="text-2xl font-bold text-center">${user.username}</h2>
        <p class="text-gray-500 text-center">Joined: ${new Date(user.created_at).toLocaleDateString()}</p>
    `;
}

export function showNotification(message, type = 'info') {
    const container = document.getElementById('notification-container');
    if (!container) return;

    const color = type === 'success' ? 'bg-green-500' : type === 'error' ? 'bg-red-500' : 'bg-blue-500';

    const notification = document.createElement('div');
    notification.className = `fixed bottom-5 right-5 p-4 rounded-lg text-white shadow-lg transition-opacity duration-300 ${color}`;
    notification.textContent = message;

    container.appendChild(notification);

    setTimeout(() => {
        notification.style.opacity = '0';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}


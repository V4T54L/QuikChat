/**
 * @param {import('./types.js').User} user
 */
export const userProfileTemplate = (user) => `
    <div class="flex items-center space-x-4">
        <img src="${user.profilePicUrl || 'https://via.placeholder.com/40'}" alt="Profile" class="w-10 h-10 rounded-full">
        <div>
            <h3 class="font-bold">${user.username}</h3>
            <p class="text-sm text-text-dim">Online</p>
        </div>
    </div>
`;

/**
 * @param {import('./types.js').User[]} friends
 */
export const friendListTemplate = (friends) => `
    <h4 class="mb-2 text-sm font-semibold tracking-wider uppercase text-text-dim">Friends</h4>
    <ul>
        ${friends.map(friend => `
            <li data-id="${friend.id}" data-username="${friend.username}" class="flex items-center p-2 space-x-3 rounded-md cursor-pointer hover:bg-accent">
                <img src="${friend.profilePicUrl || 'https://via.placeholder.com/32'}" alt="${friend.username}" class="w-8 h-8 rounded-full">
                <span>${friend.username}</span>
            </li>
        `).join('')}
    </ul>
`;

/**
 * @param {import('./types.js').Group[]} groups
 */
export const groupListTemplate = (groups) => `
    <h4 class="mt-4 mb-2 text-sm font-semibold tracking-wider uppercase text-text-dim">Groups</h4>
    <ul>
        ${groups.map(group => `
            <li data-id="${group.id}" data-name="${group.name}" class="flex items-center p-2 space-x-3 rounded-md cursor-pointer hover:bg-accent">
                <img src="${group.photoUrl || 'https://via.placeholder.com/32'}" alt="${group.name}" class="w-8 h-8 rounded-full">
                <span>${group.name}</span>
            </li>
        `).join('')}
    </ul>
`;

/**
 * @param {import('./types.js').Message} message
 * @param {boolean} isOwnMessage
 */
export const messageTemplate = (message, isOwnMessage) => `
    <div class="flex ${isOwnMessage ? 'justify-end' : 'justify-start'} mb-4">
        <div class="max-w-xs px-4 py-2 rounded-lg ${isOwnMessage ? 'bg-highlight text-primary' : 'bg-accent'}">
            <p class="text-sm">${message.content}</p>
            <p class="mt-1 text-xs text-right ${isOwnMessage ? 'text-teal-900' : 'text-text-dim'}">${new Date(message.timestamp).toLocaleTimeString()}</p>
        </div>
    </div>
`;


/**
 * @typedef {import('./api.js').User} User
 * @typedef {import('./api.js').Group} Group
 */

/**
 * @typedef {object} FriendRequest
 * @property {string} id
 * @property {User} sender // Changed to User object
 * @property {User} receiver // Changed to User object
 * @property {string} status // Added status
 * @property {string} created_at
 */

/**
 * @typedef {object} ActiveChat
 * @property {'friend'|'group'|null} type
 * @property {string|null} id - The conversation ID
 * @property {string|null} name - The display name // Added name
 */

const state = {
    /** @type {User | null} */
    currentUser: null,
    accessToken: localStorage.getItem('accessToken'),
    refreshToken: localStorage.getItem('refreshToken'),
    /** @type {User[]} */
    friends: [],
    /** @type {FriendRequest[]} */
    pendingRequests: [],
    /** @type {Group[]} */
    groups: [],
    /** @type {ActiveChat} */
    activeChat: { type: null, id: null, name: null }, // Initialized name
};

/** @param {User} user */
export function setCurrentUser(user) {
    state.currentUser = user;
}

export function getCurrentUser() {
    return state.currentUser;
}

export function setTokens(access, refresh) {
    state.accessToken = access;
    state.refreshToken = refresh;
    localStorage.setItem('accessToken', access);
    localStorage.setItem('refreshToken', refresh);
}

export function getAccessToken() {
    return state.accessToken;
}

export function getRefreshToken() {
    return state.refreshToken;
}

export function clearTokens() {
    state.currentUser = null;
    state.accessToken = null;
    state.refreshToken = null;
    state.friends = [];
    state.pendingRequests = [];
    state.groups = [];
    state.activeChat = { type: null, id: null, name: null }; // Cleared name
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
}

export function setFriends(friends) {
    state.friends = friends;
}

export function getFriends() {
    return state.friends;
}

export function setPendingRequests(requests) {
    state.pendingRequests = requests;
}

export function getPendingRequests() {
    return state.pendingRequests;
}

/** @param {Group[]} groups */
export function setGroups(groups) {
    state.groups = groups;
}

/** @returns {Group[]} */
export function getGroups() {
    return state.groups;
}

/**
 * @param {'friend'|'group'|null} type
 * @param {string|null} id
 * @param {string|null} name
 */
export function setActiveChat(type, id, name) { // Added setActiveChat
    state.activeChat = { type, id, name };
}

/** @returns {ActiveChat} */
export function getActiveChat() { // Added getActiveChat
    return state.activeChat;
}

/**
 * @typedef {object} User
 * @property {string} id
 * @property {string} username
 * @property {string} profile_pic_url
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {object} FriendRequest
 * @property {string} id
 * @property {string} sender_id
 * @property {string} receiver_id
 * @property {User} Sender // Note: Capitalized 'Sender' as per attempted content's WS event payload
 * @property {User} Receiver // Note: Capitalized 'Receiver' as per attempted content's WS event payload
 */

/**
 * @typedef {object} Group
 * @property {string} id
 * @property {string} handle
 * @property {string} name
 * @property {string} photo_url
 * @property {string} owner_id
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {object} ActiveChat
 * @property {('friend'|'group'|null)} type
 * @property {string|null} id
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
    groups: [], // Added groups
    /** @type {ActiveChat} */
    activeChat: { type: null, id: null }, // Added activeChat
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
    state.groups = []; // Clear groups on logout
    state.activeChat = { type: null, id: null }; // Clear active chat on logout
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


/**
 * @typedef {object} User
 * @property {string} id
 * @property {string} username
 * @property {string} profilePicUrl
 */

/**
 * @typedef {object} Group
 * @property {string} id
 * @property {string} name
 * @property {string} handle
 * @property {string} photoUrl
 */

/**
 * @typedef {object} Message
 * @property {string} id
 * @property {string} content
 * @property {string} senderId
 * @property {string} recipientId
 * @property {string} timestamp
 */

/**
 * @typedef {object} ActiveChat
 * @property {string} id
 * @property {string} name
 * @property {'user' | 'group'} type
 */

/**
 * @typedef {object} AppState
 * @property {User | null} currentUser
 * @property {string | null} accessToken
 * @property {string | null} refreshToken
 * @property {User[]} friends
 * @property {Group[]} groups
 * @property {ActiveChat | null} activeChat
 */

 export default {};


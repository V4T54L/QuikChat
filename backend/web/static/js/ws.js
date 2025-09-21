import { getAccessToken } from './store.js';
import { showNotification } from './ui.js';
import { refreshFriendsList, refreshPendingRequests, refreshGroupsList } from './main.js';

let socket = null;
let reconnectInterval = 1000; // Initial reconnect interval in ms

/**
 * Handles incoming WebSocket events and dispatches actions.
 * @param {object} event - The parsed WebSocket event.
 * @param {string} event.type - The type of the event.
 * @param {object} event.payload - The payload of the event.
 */
function handleIncomingEvent(event) {
    console.log('WebSocket event received:', event);
    switch (event.type) {
        case 'new_message':
            // TODO: Handle new message
            showNotification(`New message from ${event.payload.sender_id}`);
            break;
        case 'friend_request_received':
            showNotification(`New friend request from ${event.payload.Sender.username}`);
            refreshPendingRequests();
            break;
        case 'friend_request_accepted':
            showNotification(`Your friend request to ${event.payload.Sender.username} was accepted.`);
            refreshFriendsList();
            refreshPendingRequests();
            break;
        case 'friend_request_rejected':
            showNotification(`Your friend request to ${event.payload.Sender.username} was rejected.`);
            refreshPendingRequests();
            break;
        case 'unfriended':
            showNotification(`You were unfriended by a user.`);
            refreshFriendsList();
            break;
        case 'group_joined':
            showNotification(`You joined the group: ${event.payload.name}`);
            refreshGroupsList();
            break;
        case 'group_created':
            showNotification(`Group "${event.payload.name}" created!`);
            refreshGroupsList();
            break;
        case 'group_left':
            showNotification(`You left a group.`);
            refreshGroupsList();
            break;
        case 'group_member_added':
            showNotification(`${event.payload.member.username} was added to group ${event.payload.group.name}.`);
            // Potentially refresh group details if active, or just groups list
            refreshGroupsList();
            break;
        case 'group_member_removed':
            showNotification(`${event.payload.member.username} was removed from group ${event.payload.group.name}.`);
            refreshGroupsList();
            break;
        case 'group_ownership_transferred':
            showNotification(`Ownership of group ${event.payload.group.name} transferred to ${event.payload.new_owner.username}.`);
            refreshGroupsList();
            break;
        case 'group_updated':
            showNotification(`Group ${event.payload.name} was updated.`);
            refreshGroupsList();
            break;
        default:
            console.warn('Unhandled event type:', event.type);
    }
}

export function connect() {
    const token = getAccessToken();
    if (!token) {
        console.error('No access token found for WebSocket connection.');
        return;
    }

    // Prevent multiple connections
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
        console.log('WebSocket is already connected or connecting.');
        return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/ws`;

    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
        console.log('WebSocket connected');
        reconnectInterval = 1000; // Reset reconnect interval on successful connection
        // Send auth token
        socket.send(JSON.stringify({ type: 'auth', payload: { token } }));
        showNotification('Connected to real-time service.', 'success');
    };

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            handleIncomingEvent(data);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    socket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.reason);
        showNotification('Real-time connection lost. Attempting to reconnect...', 'error');
        setTimeout(connect, reconnectInterval);
        reconnectInterval = Math.min(reconnectInterval * 2, 30000); // Exponential backoff, max 30 seconds
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        showNotification('A real-time connection error occurred.', 'error');
        socket.close(); // Close to trigger onclose and reconnect logic
    };
}

/**
 * Sends a message over the WebSocket connection.
 * @param {string} messageType - The type of the message.
 * @param {object} payload - The payload of the message.
 */
export function send(messageType, payload) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: messageType, payload }));
    } else {
        console.error('WebSocket is not connected.');
        showNotification('Cannot send message: Not connected to real-time service.', 'error');
    }
}


let socket = null;
const eventListeners = new Map();

const handleMessage = (event) => {
    try {
        const data = JSON.parse(event.data);
        if (data.type && eventListeners.has(data.type)) {
            eventListeners.get(data.type).forEach(callback => callback(data.payload));
        } else {
            console.warn('Unhandled WebSocket event type:', data.type);
        }
    } catch (error) {
        console.error('Error parsing WebSocket message:', error);
    }
};

export const ws = {
    connect: (token) => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            console.log('WebSocket is already connected.');
            return;
        }
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${protocol}//${window.location.host}/api/v1/ws`;
        
        socket = new WebSocket(url);

        socket.onopen = () => {
            console.log('WebSocket connected.');
            // Send auth token immediately after connection
            socket.send(JSON.stringify({ type: 'auth', payload: { token } }));
        };

        socket.onmessage = handleMessage;

        socket.onclose = () => {
            console.log('WebSocket disconnected.');
            socket = null;
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },

    disconnect: () => {
        if (socket) {
            socket.close();
        }
    },

    sendMessage: (message) => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(message));
        } else {
            console.error('WebSocket is not connected.');
        }
    },

    onEvent: (eventType, callback) => {
        if (!eventListeners.has(eventType)) {
            eventListeners.set(eventType, []);
        }
        eventListeners.get(eventType).push(callback);
    },
};
```

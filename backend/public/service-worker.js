// this eventually belongs in a CDN
self.addEventListener('push', function(event) {
    // convert string to JSON
    const data = event.data.json();
    const title = data.title;

    console.log('Push received: ', data);
    
    const options = {
        body: data.body,
        icon: data.icon,
        badge: data.badge,
        data : {
            url : data.url
        },
    };
    event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', function(event) {
    if (event.notification.data?.url) {
        clients.openWindow(event.notification.data.url);
    }
}, false);


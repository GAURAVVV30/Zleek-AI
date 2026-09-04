export const getNotifications = () => {
  try {
    const raw = localStorage.getItem('platform_notifications');
    if (!raw) return [];
    return JSON.parse(raw);
  } catch (e) {
    return [];
  }
};

export const addNotification = (title, message, type = 'info') => {
  const current = getNotifications();
  const newNotif = {
    id: `notif_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
    title,
    message,
    type,
    read: false,
    createdAt: 'Just now',
    timestamp: Date.now(),
  };

  const updated = [newNotif, ...current].slice(0, 30);
  localStorage.setItem('platform_notifications', JSON.stringify(updated));
  window.dispatchEvent(new Event('platform_notifications_updated'));
  return newNotif;
};

export const markAllNotificationsRead = () => {
  const current = getNotifications();
  const updated = current.map((n) => ({ ...n, read: true }));
  localStorage.setItem('platform_notifications', JSON.stringify(updated));
  window.dispatchEvent(new Event('platform_notifications_updated'));
};

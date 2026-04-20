import { Bell } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useMarkAsRead, useNotifications } from "@/hooks/useNotification";
import { Notification } from "@/types/notification";
import { format } from "date-fns";

export default function NotificationBell() {
  const [open, setOpen] = useState(false);
  const { data: notifications = [] } = useNotifications();
  const markAsRead = useMarkAsRead();

  const unread = notifications.filter((n) => !n.Read);

  const handleClick = (notif: Notification) => {
    if (!notif.Read) {
      markAsRead.mutate(notif.ID);
    }
  };

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        className="relative p-2 rounded-lg hover:bg-muted transition-colors"
      >
        <Bell size={18} className="text-muted-foreground" />
        {unread.length > 0 && (
          <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-blue-500" />
        )}
      </button>

      {open && (
        <>
          {/* backdrop */}
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />

          <div className="absolute right-0 top-10 z-20 w-80 bg-background border border-border rounded-xl shadow-lg overflow-hidden">
            <div className="px-4 py-3 border-b border-border flex items-center justify-between">
              <span className="text-sm font-medium">Notifications</span>
              {unread.length > 0 && (
                <span className="text-xs text-muted-foreground">
                  {unread.length} unread
                </span>
              )}
            </div>

            {notifications.length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-muted-foreground">
                No notifications yet
              </div>
            ) : (
              <ul className="max-h-80 overflow-y-auto divide-y divide-border">
                {notifications.map((notif) => (
                  <li
                    key={notif.ID}
                    onClick={() => handleClick(notif)}
                    className={cn(
                      "px-4 py-3 text-sm cursor-pointer hover:bg-muted transition-colors",
                      !notif.Read && "bg-blue-500/5",
                    )}
                  >
                    <div className="flex items-start gap-2">
                      <div
                        className={cn(
                          "mt-1 w-2 h-2 rounded-full shrink-0",
                          notif.Type === "kyc_approved"
                            ? "bg-green-500"
                            : "bg-red-500",
                        )}
                      />
                      <div>
                        <p className="text-foreground">{notif.Message}</p>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          {format(new Date(notif.created_at), "PPP p")}
                        </p>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </>
      )}
    </div>
  );
}

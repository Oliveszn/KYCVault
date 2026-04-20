import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { WSMessage } from "@/types/notification";
import { notifKeys } from "./useNotification";

const WS_URL = import.meta.env.VITE_WS_URL ?? "ws://localhost:8000/ws";

export const useWebSocket = () => {
  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const connect = () => {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("WebSocket connected");
      };

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data);

          if (msg.type === "notification") {
            const notif = msg.data;

            if (notif.Type === "kyc_approved") {
              toast.success(notif.Message);
            } else if (notif.Type === "kyc_rejected") {
              toast.error(notif.Message);
            }

            // update notifications cache
            queryClient.invalidateQueries({ queryKey: notifKeys.all });

            // sync session state since status changed
            if (notif.SessionId) {
              queryClient.invalidateQueries({
                queryKey: ["kyc", "session", notif.SessionId],
              });
              queryClient.invalidateQueries({ queryKey: ["kyc", "active"] });
              queryClient.invalidateQueries({ queryKey: ["kyc", "history"] });
            }
          }
        } catch (err) {
          console.error("WebSocket message parse error", err);
        }
      };

      ws.onclose = () => {
        // reconnect after 3s on unexpected close
        setTimeout(connect, 3000);
      };

      ws.onerror = (err) => {
        console.error("WebSocket error", err);
        ws.close();
      };
    };

    connect();

    return () => {
      wsRef.current?.close();
    };
  }, [queryClient]);
};

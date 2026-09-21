import { useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { useCallback, useEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { internalApi } from "../../../services/api";
import { useAuthStore } from "../../../stores/auth";

const SessionExpiryHandler = () => {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const clearSession = useAuthStore((state) => state.actions.clearSession);
  const verifySession = useAuthStore((state) => state.actions.verifySession);
  const hasVerifiedSession = useRef(false);

  const handleExpiredSession = useCallback(() => {
    queryClient.clear();
    clearSession();
    navigate("/login", { replace: true });
  }, [clearSession, navigate, queryClient]);

  useEffect(() => {
    const interceptorID = internalApi.interceptors.response.use(
      (response) => response,
      (error: unknown) => {
        if (axios.isAxiosError(error) && error.response?.status === 401) {
          handleExpiredSession();
        }

        return Promise.reject(error);
      },
    );

    if (!hasVerifiedSession.current) {
      hasVerifiedSession.current = true;
      void verifySession();
    }

    return () => internalApi.interceptors.response.eject(interceptorID);
  }, [handleExpiredSession, verifySession]);

  return null;
};

export default SessionExpiryHandler;

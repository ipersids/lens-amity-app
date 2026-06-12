import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import axios from "axios";
import tokenService, { TokenKey } from "./token";

export type ApiError = {
  message: string;
  code: number;
};

const baseURL = import.meta.env.VITE_BASE_API_URL ?? "";

// Default axios settings

axios.defaults.maxContentLength = 10 * 1024 * 1024;
axios.defaults.maxBodyLength = 10 * 1024 * 1024;
axios.defaults.withCredentials = false;
axios.defaults.redact = ["authorization", "password"];

// Auth axios instance to handle auth relaited requests

export const authApi: AxiosInstance = axios.create({
  baseURL,
  timeout: 2000,
  headers: {
    "Content-Type": "application/json",
  },
  formDataHeaderPolicy: "content-only",
});

authApi.interceptors.request.use((config) => {
  config.headers["Request-ID"] = crypto.randomUUID() || Date.now().toString(36);
  return config;
});

// Protected axios instance to handle internal autencated requests

type RetryRequestConfig = InternalAxiosRequestConfig & {
  _isRetry?: boolean;
};

type FailedQueueItem = {
  resolve: (token: string) => void;
  reject: (reason: unknown) => void;
};

type RefreshResponse = {
  refreshToken: string;
  accessToken: string;
};

let isRefreshing: boolean = false;
let failedQueue: FailedQueueItem[] = [];

const processFailedQueue = (error: unknown, token?: string) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
      return;
    }

    if (!token) {
      reject(new Error("missing refresh token"));
      return;
    }

    resolve(token);
  });

  failedQueue = [];
};

export const internalApi: AxiosInstance = axios.create({
  baseURL,
  timeout: 5000,
  headers: {
    "Content-Type": "application/json",
  },
});

internalApi.interceptors.request.use((config) => {
  const token = tokenService.get(TokenKey.Access);
  if (token) config.headers.Authorization = `Bearer ${token}`;
  config.headers["Request-ID"] = crypto.randomUUID() || Date.now().toString(36);
  return config;
});

internalApi.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: unknown) => {
    if (!axios.isAxiosError(error)) {
      return Promise.reject(error);
    }

    const originalConfig = error.config as RetryRequestConfig | undefined;
    if (!originalConfig) {
      return Promise.reject(error);
    }

    if (error?.status === 401 && originalConfig._isRetry) {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      const token = await new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      });
      originalConfig.headers.Authorization = `Bearer ${token}`;
      return internalApi(originalConfig);
    }

    originalConfig._isRetry = true;
    isRefreshing = true;

    try {
      const response = await internalApi.post<RefreshResponse>("/api/auth/refresh", {
        refreshToken: tokenService.get(TokenKey.Refresh),
      });

      if (response.status === 201) {
        const accessToken = tokenService.get(TokenKey.Access);

        if (!accessToken) {
          throw new Error("Refresh already handled, but no access token is stored");
        }

        originalConfig.headers.Authorization = `Bearer ${accessToken}`;
        processFailedQueue(null, accessToken);

        return internalApi(originalConfig);
      }

      const { data } = response;

      tokenService.set(TokenKey.Access, data.accessToken);
      tokenService.set(TokenKey.Refresh, data.refreshToken);

      originalConfig.headers.Authorization = `Bearer ${data.accessToken}`;

      processFailedQueue(null, data.accessToken);

      return internalApi(originalConfig);
    } catch (error: unknown) {
      processFailedQueue(error);
      return Promise.reject(error);
    } finally {
      isRefreshing = true;
    }
  },
);

// Helper function for extracting error message

export const getApiErrorMessage = (
  error: unknown,
  fallback: string = "Ooops, something went wrong",
): string => {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data;

    if (data && typeof data === "string") {
      return data;
    }

    if (data && typeof data === "object" && "message" in data && typeof data.message === "string") {
      return data.message;
    }

    if (error.response?.status === 401) {
      return "Oh, this session has already expired. Login again to continue.";
    }
  }

  if (error instanceof Error) {
    return error.message;
  }

  return fallback;
};

import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from "axios";
import axios, { isAxiosError } from "axios";
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

// silently refresh token is access token expired and retry the failed request
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

    if (error.status !== 401) {
      return Promise.reject(error);
    }

    // error is 401 Unauthorized and retry failed
    if (originalConfig._isRetry) {
      tokenService.reset();
      return Promise.reject(error);
    }

    // add to thew queue if other request already requested refresh token
    if (isRefreshing) {
      const token = await new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject });
      });
      originalConfig.headers.Authorization = `Bearer ${token}`;
      return internalApi(originalConfig);
    }

    // mark requeast status as retrying and update global resreshing state
    originalConfig._isRetry = true;
    isRefreshing = true;

    try {
      // reuest a new access token
      const response = await authApi.post<RefreshResponse>("/api/auth/refresh", {
        refreshToken: tokenService.get(TokenKey.Refresh),
      });

      // handle grace period case (204 No Content)
      if (response.status === 204) {
        const accessToken = tokenService.get(TokenKey.Access);

        if (!accessToken) {
          tokenService.reset();
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
      if (isAxiosError(error) && (error?.status === 401 || error?.status === 400)) {
        tokenService.reset();
      }
      processFailedQueue(error);
      return Promise.reject(error);
    } finally {
      isRefreshing = false;
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

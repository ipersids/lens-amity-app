import type { AxiosInstance } from "axios";
import axios from "axios";

export type ApiError = {
  error: {
    code: string;
    message: string;
  };
};

const isApiError = (value: unknown): value is ApiError => {
  if (!value || typeof value !== "object" || !("error" in value)) return false;

  const body = value.error;
  return (
    !!body &&
    typeof body === "object" &&
    "code" in body &&
    typeof body.code === "string" &&
    "message" in body &&
    typeof body.message === "string"
  );
};

const baseURL = import.meta.env.VITE_BASE_API_URL ?? "";

// Default axios settings

axios.defaults.maxContentLength = 10 * 1024 * 1024;
axios.defaults.maxBodyLength = 10 * 1024 * 1024;
axios.defaults.withCredentials = true;
axios.defaults.redact = ["authorization", "password"];

// Auth axios instance to handle auth relaited requests

export const authApi: AxiosInstance = axios.create({
  baseURL,
  timeout: 2000,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
  formDataHeaderPolicy: "content-only",
});

authApi.interceptors.request.use((config) => {
  config.headers["Request-ID"] = crypto.randomUUID() || Date.now().toString(36);
  return config;
});

// Protected axios instance for cookie-authenticated requests.

export const internalApi: AxiosInstance = axios.create({
  baseURL,
  timeout: 5000,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

internalApi.interceptors.request.use((config) => {
  config.headers["Request-ID"] = crypto.randomUUID() || Date.now().toString(36);
  return config;
});

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

    if (isApiError(data)) {
      if (data.error.code === "invalid_credentials") {
        return "Invalid username or password.";
      }

      if (data.error.message) {
        return data.error.message;
      }
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

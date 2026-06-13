import { create } from "zustand";
import { persist } from "zustand/middleware";
import { getApiErrorMessage } from "../services/api";
import type { LoginItem, SignupItem } from "../services/auth";
import authService from "../services/auth";
import tokenService, { TokenKey } from "../services/token";

interface CurrentUser {
  username: string;
  displayName: string;
}

interface AuthActions {
  signup: (input: SignupItem) => Promise<void>;
  login: (input: LoginItem) => Promise<void>;
  logout: () => Promise<void>;
  logoutAll: () => Promise<void>;
  syncSession: () => void;
}

interface AuthState {
  user: CurrentUser | null;
  isLoading: boolean;
  actions: AuthActions;
}

const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isLoading: false,

      actions: {
        signup: async (input) => {
          if (get().user || get().isLoading) return;

          set(() => ({ isLoading: true }));

          try {
            await authService.signup({ ...input });
          } catch (err: unknown) {
            throw new Error(getApiErrorMessage(err));
          } finally {
            set(() => ({ isLoading: false }));
          }
        },

        login: async (input) => {
          if (get().user || get().isLoading) return;

          set(() => ({ isLoading: true }));

          try {
            const data = await authService.login({ ...input });

            tokenService.set(TokenKey.Access, data.accessToken);
            tokenService.set(TokenKey.Refresh, data.refreshToken);

            set(() => ({
              user: {
                username: data.username,
                displayName: data.displayName,
              },
            }));
          } catch (err: unknown) {
            throw new Error(getApiErrorMessage(err));
          } finally {
            set(() => ({ isLoading: false }));
          }
        },

        logout: async () => {
          if (!get().user || get().isLoading) return;

          set(() => ({ isLoading: true }));

          try {
            const refreshToken = tokenService.get(TokenKey.Refresh);

            if (refreshToken) {
              await authService.logout({ refreshToken });
            }
          } finally {
            set(() => ({
              user: null,
              isLoading: false,
            }));

            tokenService.reset();
          }
        },

        logoutAll: async () => {
          if (!get().user || get().isLoading) return;

          set(() => ({ isLoading: true }));

          try {
            await authService.logoutAll();
          } catch (err: unknown) {
            throw new Error(getApiErrorMessage(err));
          } finally {
            set(() => ({
              user: null,
              isLoading: false,
            }));

            tokenService.reset();
          }
        },
        syncSession: () => {
          if (!tokenService.hasAuthTokens()) {
            set(() => ({ user: null }));
          }
        },
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        user: state.user,
      }),
    },
  ),
);

export const useUser = () => useAuthStore((state) => state.user);
export const useLoading = () => useAuthStore((state) => state.isLoading);
export const useSyncSession = () => useAuthStore((state) => state.actions.syncSession);
export const useSignup = () => useAuthStore((state) => state.actions.signup);
export const useLogin = () => useAuthStore((state) => state.actions.login);
export const useLogout = () => useAuthStore((state) => state.actions.logout);
export const useLogoutAll = () => useAuthStore((state) => state.actions.logoutAll);

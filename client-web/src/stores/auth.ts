import { create } from "zustand";
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
}

interface AuthState {
  user: CurrentUser | null;
  accessToken: string | null;
  isLoading: boolean;
  error: string | null;
  actions: AuthActions;
}

const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  isLoading: false,
  error: null,
  actions: {
    signup: async (input) => {
      if (get().user) return;

      set(() => ({ isLoading: true, error: null }));

      try {
        await authService.signup({ ...input });
      } catch (err: unknown) {
        set(() => ({ error: getApiErrorMessage(err) }));
      } finally {
        set(() => ({ isLoading: false }));
      }
    },

    login: async (input) => {
      if (get().user) return;

      set(() => ({ isLoading: true, error: null }));

      try {
        const data = await authService.login({ ...input });

        set(() => ({
          user: { username: data.username, displayName: data.displayName },
          accessToken: data.accessToken,
        }));

        tokenService.set(TokenKey.Access, data.accessToken);
        tokenService.set(TokenKey.Refresh, data.refreshToken);
      } catch (err: unknown) {
        set(() => ({ error: getApiErrorMessage(err) }));
      } finally {
        set(() => ({ isLoading: false }));
      }
    },

    logout: async () => {
      if (!get().user) return;

      set(() => ({ isLoading: true, error: null }));

      try {
        const refreshToken = tokenService.get(TokenKey.Refresh);

        if (refreshToken) {
          await authService.logout({ refreshToken });
        }
      } catch (err: unknown) {
        set(() => ({ error: getApiErrorMessage(err) }));
      } finally {
        tokenService.remove(TokenKey.Access);
        tokenService.remove(TokenKey.Refresh);

        set(() => ({
          user: null,
          accessToken: null,
          isLoading: false,
        }));
      }
    },

    logoutAll: async () => {
      if (!get().user) return;

      set(() => ({ isLoading: true, error: null }));

      try {
        await authService.logoutAll();
      } catch (err: unknown) {
        set(() => ({ error: getApiErrorMessage(err) }));
      } finally {
        tokenService.remove(TokenKey.Access);
        tokenService.remove(TokenKey.Refresh);

        set(() => ({
          user: null,
          accessToken: null,
          isLoading: false,
        }));
      }
    },
  },
}));

export default useAuthStore;

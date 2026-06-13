import { authApi, internalApi } from "./api";

// export type AuthResponse = SignupResponse | LoginResponse | RefreshResponse;
const baseAuthURI = "/api/auth";

export type SignupItem = {
  username: string;
  displayName?: string;
  password: string;
};

export type LoginItem = {
  username: string;
  password: string;
};

export type LogoutItem = {
  refreshToken: string;
};

type SignupResponse = {
  username: string;
  displayName: string;
};

type LoginResponse = {
  accessToken: string;
  refreshToken: string;
  username: string;
  displayName: string;
};

const signup = async (credentials: SignupItem): Promise<SignupResponse> => {
  const { data } = await authApi.post<SignupResponse>(`${baseAuthURI}/signup`, {
    ...credentials,
    displayName: credentials.displayName ?? credentials.username,
  });
  return data;
};

const login = async (credentials: LoginItem): Promise<LoginResponse> => {
  const { data } = await authApi.post<LoginResponse>(`${baseAuthURI}/login`, { ...credentials });
  return data;
};

const logout = async (credentials: LogoutItem) => {
  await internalApi.post<void>(`${baseAuthURI}/logout`, { ...credentials });
};

const logoutAll = async () => {
  await internalApi.post<void>(`${baseAuthURI}/logout-all`);
};

const authService = { signup, login, logout, logoutAll };

export default authService;

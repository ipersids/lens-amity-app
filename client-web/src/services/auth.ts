import { authApi, internalApi } from "./api";

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

type SignupResponse = {
  username: string;
  displayName: string;
};

type LoginResponse = {
  username: string;
  displayName: string;
};

type SessionResponse = {
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

const logout = async () => {
  await authApi.post<void>(`${baseAuthURI}/logout`);
};

const logoutAll = async () => {
  await internalApi.post<void>(`${baseAuthURI}/logout-all`);
};

const session = async () => {
  const { data } = await internalApi.get<SessionResponse>(`${baseAuthURI}/session`);
  return data;
};

const authService = { signup, login, logout, logoutAll, session };

export default authService;

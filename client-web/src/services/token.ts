export const TokenKey = {
  Access: "access_token",
  Refresh: "refresh_token",
} as const;

type TokenKey = (typeof TokenKey)[keyof typeof TokenKey];

const get = (key: TokenKey): string | null => {
  return localStorage.getItem(key);
};

const set = (key: TokenKey, value: string) => {
  localStorage.setItem(key, value);
};

const remove = (key: TokenKey) => {
  localStorage.removeItem(key);
};

const tokenService = { get, set, remove };

export default tokenService;

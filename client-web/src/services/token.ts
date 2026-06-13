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

const reset = () => {
  localStorage.removeItem(TokenKey.Access);
  localStorage.removeItem(TokenKey.Refresh);
};

const hasAuthTokens = (): boolean => {
  return (
    localStorage.getItem(TokenKey.Access) !== null &&
    localStorage.getItem(TokenKey.Refresh) !== null
  );
};

const tokenService = { get, set, remove, reset, hasAuthTokens };

export default tokenService;

export const ProfilePhotosQueryKey = (username: string) => ["photos", username];
export const ProfileQueryKey = (username: string) => ["profile", username];

export const USERNAME_REGEX = /^[A-Za-z0-9_-]+$/;
export const MIN_USERNAME_LENGTH = 3;
export const MAX_USERNAME_LENGTH = 32;
export const MIN_PASSWORD_LENGTH = 15;

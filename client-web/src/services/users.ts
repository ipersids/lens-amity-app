import { internalApi } from "./api";

const baseUsersURL = "/api/users";

export type Visibility = "public" | "private";

export type UserProfileResponse = {
  username: string;
  displayName: string;
  photoCount: number;
  canEdit: boolean;
  canViewPhotos: boolean;
  visibility: Visibility;
  joinedAt: string;
  avatar?: {
    url: string;
    method: string;
    header: Record<string, string[]>;
  };
  about?: string;
};

const getUserProfile = async (username: string): Promise<UserProfileResponse> => {
  const { data } = await internalApi.get<UserProfileResponse>(`${baseUsersURL}/${username}`);

  return data;
};

export type Photo = {
  photoID: string;
  title: string;
  description: string;
  date: string;
  request: {
    url: string;
    method: string;
    header: Record<string, string[]>;
    isReady: boolean;
  };
};

export type UserPhotosResponse = {
  canEdit: boolean;
  canViewPhotos: boolean;
  photoCount: number;
  items: Photo[];
  nextCursor?: string;
};

const getUserPhotos = async (
  username: string,
  limit?: number,
  cursor?: string,
): Promise<UserPhotosResponse> => {
  const { data } = await internalApi.get<UserPhotosResponse>(`${baseUsersURL}/${username}/photos`, {
    params: { limit, cursor },
  });

  return data;
};

export type UpdateProfileProps = {
  displayName: string;
  about?: string;
  visibility: Visibility;
};

const updateUserProfile = async ({
  displayName,
  about,
  visibility,
}: UpdateProfileProps): Promise<void> => {
  await internalApi.put<void>(`${baseUsersURL}/me/profile`, { displayName, about, visibility });
};

export type updateUsernameResponse = {
  updatedUsername: string;
};

const updateUsername = async (newUsername: string): Promise<updateUsernameResponse> => {
  const { data } = await internalApi.put<updateUsernameResponse>(`${baseUsersURL}/me/username`, {
    newUsername,
  });

  return data;
};

export type updatePasswordProps = {
  oldPassword: string;
  newPassword: string;
  revokeAll: boolean;
};

const updatePassword = async ({
  oldPassword,
  newPassword,
  revokeAll,
}: updatePasswordProps): Promise<void> => {
  await internalApi.put<void>(`${baseUsersURL}/me/password`, {
    oldPassword,
    newPassword,
    revokeAll,
  });
};

const usersService = {
  getUserProfile,
  getUserPhotos,
  updateUserProfile,
  updateUsername,
  updatePassword,
};

export default usersService;

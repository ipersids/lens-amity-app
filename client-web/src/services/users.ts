import { internalApi } from "./api";

const baseUsersURL = "/api/users";

type UserProfileResponse = {
  username: string;
  displayName: string;
  photoCount: number;
  canEdit: boolean;
  canViewPhotos: boolean;
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
  id: string;
  title: string;
  description: string;
  date: string;
  request: {
    url: string;
    method: string;
    header: string;
    isReady: boolean;
  };
};

type UserPhotosResponse = {
  canEdit: boolean;
  canViewPhotos: boolean;
  photoCount: number;
  items: [];
  Photo;
  nextCursor: string;
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

const usersService = { getUserProfile, getUserPhotos };

export default usersService;

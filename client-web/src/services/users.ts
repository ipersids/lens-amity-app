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

const usersService = { getUserProfile };

export default usersService;

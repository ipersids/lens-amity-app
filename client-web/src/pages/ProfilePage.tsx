import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";
import {
  ProfileHeader,
  ProfileNotAvailable,
  ProfilePhotos,
  ProfileShell,
} from "../features/profile";
import usersService from "../services/users";

const ProfilePage = () => {
  const { username } = useParams<{ username: string }>();
  const {
    isPending,
    isError,
    data: user,
  } = useQuery({
    queryKey: ["profile", username],
    queryFn: () => usersService.getUserProfile(username ?? ""),
    enabled: !!username,
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  });

  if (!username || isError) {
    return <ProfileNotAvailable />;
  }

  if (isPending || !user) {
    return null;
  }

  return (
    <ProfileShell>
      <ProfileHeader
        profile={{
          username: user.username,
          displayName: user.displayName,
          photoCount: user.photoCount,
          about: user.about,
          avatarURL: user.avatar?.url,
          canEdit: user.canEdit,
          canViewPhotos: user.canViewPhotos,
        }}
      />

      <ProfilePhotos username={username} canViewPhotos={user.canViewPhotos} />
    </ProfileShell>
  );
};

export default ProfilePage;

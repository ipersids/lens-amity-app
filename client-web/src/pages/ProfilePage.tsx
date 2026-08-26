import { useEffect, useState } from "react";
import { useParams } from "react-router";
import { profilePhotosList } from "../features/photos";
import {
  ProfileHeader,
  ProfileNotAvailable,
  ProfilePhotos,
  ProfileShell,
} from "../features/profile";

const ProfilePage = () => {
  const { username } = useParams<{ username: string }>();
  const [profileUsername, setProfileUsername] = useState<string | null>(username ?? null);

  useEffect(() => {
    if (!username) return;
    setProfileUsername(username);
  }, [username]);

  if (!profileUsername) {
    return <ProfileNotAvailable />;
  }

  const displayName = profileUsername;
  const about =
    "Collecting small moments, soft light, favorite places, and photos worth coming back to.";
  const avatarURL = "https://picsum.photos/seed/face/900/900";
  const canEdit = false;
  const canViewPhotos = true;

  return (
    <ProfileShell>
      <ProfileHeader
        profile={{
          username: profileUsername,
          displayName: displayName,
          photoCount: profilePhotosList.length,
          about: about,
          avatarURL: avatarURL,
          canEdit: canEdit,
          canViewPhotos: canViewPhotos,
        }}
      />

      <ProfilePhotos username={profileUsername} canViewPhotos={canViewPhotos} />
    </ProfileShell>
  );
};

export default ProfilePage;

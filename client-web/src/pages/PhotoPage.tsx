import { useParams } from "react-router";
import { PhotoInfo, PhotoShell } from "../features/photos";
import { ProfileNotAvailable, useProfile } from "../features/profile";

const PhotoPage = () => {
  const { username } = useParams<{ username: string }>();
  const { photoID } = useParams<{ photoID: string }>();
  const { isPending, isError, data: user } = useProfile(username);

  if (isPending || !user) return null;

  if (!username || !photoID || isError || !user.canViewPhotos) {
    return <ProfileNotAvailable />;
  }

  return (
    <PhotoShell>
      <PhotoInfo profile={user} photoID={photoID} />
    </PhotoShell>
  );
};

export default PhotoPage;

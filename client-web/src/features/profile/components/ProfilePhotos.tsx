import { PhotosFeed } from "../../photos";

type ProfilePhotosProps = {
  username: string;
  canViewPhotos: boolean;
};

const ProfilePhotos = ({ username, canViewPhotos }: ProfilePhotosProps) => {
  if (!canViewPhotos) {
    return <p>Can't see photos</p>;
  }

  return <PhotosFeed username={username} />;
};

export default ProfilePhotos;

type UserParams = {
  username: string;
  displayName: string;
  canEdit: boolean;
  canViewPhotos: boolean;
  avatarURL?: string;
};

type PhotoInfoParams = {
  profile: UserParams;
  photoID: string;
};

const PhotoInfo = ({ profile, photoID }: PhotoInfoParams) => {
  return <p>{`hello ${profile.displayName} (${photoID})`}</p>;
};

export default PhotoInfo;

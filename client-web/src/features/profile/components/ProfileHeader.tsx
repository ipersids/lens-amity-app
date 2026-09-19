import Avatar from "./Avatar";

type Profile = {
  username: string;
  displayName: string;
  photoCount: number;
  canEdit: boolean;
  canViewPhotos: boolean;
  avatarURL?: string;
  about?: string;
};

type ProfileHeaderProps = {
  profile: Profile;
};

const ProfileHeader = ({ profile }: ProfileHeaderProps) => {
  return (
    <header className="profile-header">
      <Avatar avatarURL={profile.avatarURL} displayName={profile.displayName} />

      <div className="profile-info">
        <div>
          <h1>{profile.displayName}</h1>
          <p className="profile-username">@{profile.username}</p>
        </div>

        <p className="profile-about">{profile.about}</p>

        <dl className="profile-count" aria-label="Profile stats">
          <div>
            <dd>{profile.photoCount}</dd>
            <dt>photos</dt>
          </div>
        </dl>
      </div>
    </header>
  );
};

export default ProfileHeader;

type AvatarProps = {
  avatarURL?: string;
  displayName: string;
  size?: "profile" | "settings";
};

const Avatar = ({ avatarURL, displayName, size = "profile" }: AvatarProps) => {
  return (
    <div
      className={size === "settings" ? "profile-avatar profile-avatar--settings" : "profile-avatar"}
    >
      {avatarURL ? (
        <img src={avatarURL} alt={`${displayName} avatar`} />
      ) : (
        <span role="img" aria-label={`${displayName} avatar`}>
          {displayName.charAt(0).toUpperCase()}
        </span>
      )}
    </div>
  );
};

export default Avatar;

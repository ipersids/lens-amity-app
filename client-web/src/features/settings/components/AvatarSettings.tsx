import { Avatar } from "../../profile";
import AvatarChange from "./AvatarChange";
import AvatarDelete from "./AvatarDelete";

type AvatarSettingsProps = {
  url?: string;
  displayName: string;
};

const AvatarSettings = ({ url, displayName }: AvatarSettingsProps) => {
  return (
    <div className="settings-photo-field">
      <span className="settings-field-title">Photo</span>
      <div className="settings-photo-row">
        <Avatar avatarURL={url} displayName={displayName} size="settings" />
        <div className="settings-avatar-actions">
          <AvatarChange />
          <AvatarDelete />
        </div>
      </div>
    </div>
  );
};

export default AvatarSettings;

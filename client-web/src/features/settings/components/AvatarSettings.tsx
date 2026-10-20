import { useEffect, useRef, useState } from "react";
import { Avatar } from "../../profile";
import AvatarChange from "./AvatarChange";
import AvatarDelete from "./AvatarDelete";

type AvatarSettingsProps = {
  url?: string;
  displayName: string;
};

const AvatarSettings = ({ url, displayName }: AvatarSettingsProps) => {
  const [avatarURL, setAvatarURL] = useState(url);
  const localAvatarURL = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (localAvatarURL.current && localAvatarURL.current !== url) {
      URL.revokeObjectURL(localAvatarURL.current);
      localAvatarURL.current = undefined;
    }

    setAvatarURL(url);
  }, [url]);

  useEffect(
    () => () => {
      if (localAvatarURL.current) URL.revokeObjectURL(localAvatarURL.current);
    },
    [],
  );

  const handleAvatarSave = (file: File) => {
    if (localAvatarURL.current) URL.revokeObjectURL(localAvatarURL.current);

    const nextURL = URL.createObjectURL(file);
    localAvatarURL.current = nextURL;
    setAvatarURL(nextURL);
  };

  return (
    <div className="settings-photo-field">
      <span className="settings-field-title">Photo</span>
      <div className="settings-photo-row">
        <Avatar avatarURL={avatarURL} displayName={displayName} size="settings" />
        <div className="settings-avatar-actions">
          <AvatarChange onSave={handleAvatarSave} />
          <AvatarDelete />
        </div>
      </div>
    </div>
  );
};

export default AvatarSettings;

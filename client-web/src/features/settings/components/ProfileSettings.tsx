import { useState } from "react";
import { Navigate } from "react-router";
import { useUser } from "../../../stores/auth";
import { Avatar, useProfile } from "../../profile";

// type Visibility = "public" | "private";

const AvatarUpdate = ({ url, displayName }: { url?: string; displayName: string }) => {
  return (
    <div className="settings-photo-field">
      <span className="settings-field-title">Photo</span>
      <div className="settings-photo-row">
        <Avatar avatarURL={url} displayName={displayName} size="settings" />
        <button className="settings-avatar-button" type="button" disabled>
          Change
        </button>
      </div>
    </div>
  );
};

type ProfileUpdateProps = {
  username: string;
  displayName: string;
  about?: string;
};

const ProfileUpdate = ({ username, displayName, about }: ProfileUpdateProps) => {
  const [updatedDisplayName, setUpdatedDisplayName] = useState<string | undefined>(displayName);
  const [updatedAbout, setUpdatedAbout] = useState<string | undefined>(about);

  const isUpdated =
    (updatedDisplayName !== displayName &&
      !(updatedDisplayName === undefined && displayName === username)) ||
    updatedAbout !== about;

  return (
    <>
      <div className="settings-field">
        <label>
          Display name
          <input
            type="text"
            value={updatedDisplayName}
            placeholder={username}
            aria-describedby="profile-display-name-note"
            onChange={(e) => {
              e.preventDefault();
              setUpdatedDisplayName(e.target.value === "" ? undefined : e.target.value);
            }}
          />
        </label>
        <p className="settings-field-note" id="profile-display-name-note">
          By default, your display name is the same as your username.
        </p>
      </div>

      <label>
        About
        <textarea
          value={updatedAbout}
          rows={4}
          maxLength={300}
          onChange={(e) => {
            e.preventDefault();
            setUpdatedAbout(e.target.value === "" ? undefined : e.target.value);
          }}
        />
      </label>

      {/*<fieldset className="settings-visibility-field">
        <legend className="settings-field-title">Profile visibility</legend>
        <div className="settings-visibility-switcher">
          <button type="button" className="active" aria-pressed="true" disabled>
            Public
          </button>
          <button type="button" aria-pressed="false" disabled>
            Private
          </button>
        </div>
        <div className="settings-visibility-help">
          <p>
            <strong>Public:</strong> Visible to all logged-in users.
          </p>
          <p>
            <strong>Private:</strong> Visible only to you.
          </p>
        </div>
      </fieldset>*/}

      <button className="settings-save-button" type="button" disabled={!isUpdated}>
        Update profile
      </button>
    </>
  );
};

const ProfileSettings = () => {
  const currentUser = useUser();
  const { isPending, isError, data: profile } = useProfile(currentUser?.username);

  if (!currentUser) {
    return <Navigate to="/" replace />;
  }

  if (isError) {
    return <p>Error</p>;
  }

  if (isPending || !profile) {
    return null;
  }

  return (
    <form className="settings-form">
      <h2>Profile</h2>
      <p className="settings-intro">
        Information you add here is visible to anyone who can view your profile.
      </p>
      <AvatarUpdate url={profile.avatar?.url} displayName={profile.displayName} />
      <ProfileUpdate
        username={profile.username}
        displayName={profile.displayName}
        about={profile.about}
      />
    </form>
  );
};

export default ProfileSettings;

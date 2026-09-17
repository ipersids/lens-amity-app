import { PhotoIcon } from "@heroicons/react/24/solid";
import { Navigate } from "react-router";
import { useUser } from "../../../stores/auth";
import { useProfile } from "../../profile";

const Avatar = () => {
  return (
    <div className="settings-photo-field">
      <span className="settings-field-title">Photo</span>
      <div className="settings-photo-row">
        <span className="settings-avatar-preview">
          <PhotoIcon aria-hidden="true" />
        </span>
        <button className="settings-avatar-button" type="button" disabled>
          Change
        </button>
      </div>
    </div>
  );
};

// type Visibility = "public" | "private";

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

      <Avatar />

      <label>
        Display name
        <input type="text" value={profile.displayName} placeholder={"Display name"} />
      </label>

      <label>
        About
        <textarea value={profile.about} placeholder={"A few words about you"} rows={4} />
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

      <button className="settings-save-button" type="button" disabled>
        Save profile
      </button>
    </form>
  );
};

export default ProfileSettings;

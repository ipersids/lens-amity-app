import { PhotoIcon } from "@heroicons/react/24/solid";
import { useQuery } from "@tanstack/react-query";
import { Navigate } from "react-router";
import usersService from "../../../services/users";
import { useUser } from "../../../stores/auth";

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
  const {
    isPending,
    isError,
    data: profile,
  } = useQuery({
    queryKey: ["profile", currentUser?.username],
    queryFn: () => usersService.getUserProfile(currentUser?.username ?? ""),
    enabled: !!currentUser && !!currentUser.username,
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  });

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
        <input type="text" placeholder={profile.displayName ?? "Display name"} disabled />
      </label>

      <label>
        About
        <textarea placeholder={profile.about ?? "A few words about you"} rows={4} disabled />
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

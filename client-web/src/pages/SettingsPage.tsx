import { PhotoIcon } from "@heroicons/react/24/outline";

const SettingsPage = () => {
  return (
    <section className="settings-page">
      <form className="settings-form">
        <h1>Settings</h1>
        <p className="settings-intro">
          Keep your personal details private. Information you add here is visible to anyone who can
          view your profile.
        </p>

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
          <input type="file" accept="image/jpeg, image/png, image/webp" disabled />
        </div>

        <label>
          Username
          <input type="text" placeholder="username" disabled />
        </label>

        <label>
          Display name
          <input type="text" placeholder="Display name" disabled />
        </label>

        <label>
          About
          <textarea placeholder="A few words about you" rows={4} disabled />
        </label>

        <label>
          Profile visibility
          <select disabled defaultValue="public">
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
        </label>

        <button type="button" disabled>
          Save settings
        </button>
      </form>
    </section>
  );
};

export default SettingsPage;

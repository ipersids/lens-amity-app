const AccountSettings = () => {
  return (
    <div className="settings-form">
      <h2>Account</h2>

      <div className="settings-account-action">
        <div>
          <h3>Username</h3>
          <p>Change the username shown on your profile.</p>
        </div>
        <button type="button" disabled>
          Change username
        </button>
      </div>

      <div className="settings-account-action">
        <div>
          <h3>Change password</h3>
          <p>Update the password you use to sign in.</p>
        </div>
        <button type="button" disabled>
          Change password
        </button>
      </div>

      <div className="settings-account-action settings-danger-zone">
        <div>
          <h3>Delete account</h3>
          <p>Permanently remove your account and profile.</p>
        </div>
        <button type="button" disabled>
          Delete account
        </button>
      </div>

      <button className="settings-save-button" type="button" disabled>
        Save account
      </button>
    </div>
  );
};

export default AccountSettings;

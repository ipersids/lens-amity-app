const AccountDelete = () => {
  return (
    <div className="settings-account-action settings-danger-zone">
      <div>
        <h3>Delete account</h3>
        <p>Permanently remove your account and profile.</p>
      </div>
      <button type="button" disabled>
        Delete account
      </button>
    </div>
  );
};

export default AccountDelete;

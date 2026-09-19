import { Navigate } from "react-router";
import { useUser } from "../../../stores/auth";
import AccountDelete from "./AccountDelete";
import PasswordChange from "./PasswordChange";
import UsernameChange from "./UsernameChange";

const AccountSettings = () => {
  const currentUser = useUser();

  if (!currentUser) {
    return <Navigate to="/" replace />;
  }

  return (
    <div className="settings-form">
      <h2>Account</h2>

      <UsernameChange username={currentUser.username} />
      <PasswordChange />
      <AccountDelete />
    </div>
  );
};

export default AccountSettings;

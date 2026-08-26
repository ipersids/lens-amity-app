import { Navigate } from "react-router";
import { useUser } from "../stores/auth";

const MyProfilePageRedirect = () => {
  const user = useUser();

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  return <Navigate to={`/users/${user.username}`} replace />;
};

export default MyProfilePageRedirect;

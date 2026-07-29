import { Navigate, Outlet } from "react-router";
import { useUser } from "../stores/auth";

const ProtectedPage = () => {
  const user = useUser();

  if (!user) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
};

export default ProtectedPage;

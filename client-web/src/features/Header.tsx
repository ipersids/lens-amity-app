import { NavLink, useNavigate } from "react-router";
import { useLoading, useLogout, useUser } from "../stores/auth";

const Header = () => {
  const user = useUser();
  const logout = useLogout();
  const isLoading = useLoading();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/", { replace: true });
  };

  return (
    <header>
      <NavLink to="/">Lensamity</NavLink>

      {user ? (
        <nav style={{ display: "flex", gap: "10px" }} aria-label="Account">
          <span>{user.displayName}</span>
          <button disabled={isLoading} onClick={handleLogout} type="button">
            Log out
          </button>
        </nav>
      ) : (
        <nav style={{ display: "flex", gap: "10px" }} aria-label="Auth">
          <NavLink to="/login" end>
            Log in
          </NavLink>
          <NavLink to="/signup" end>
            Sign up
          </NavLink>
        </nav>
      )}
    </header>
  );
};

export default Header;

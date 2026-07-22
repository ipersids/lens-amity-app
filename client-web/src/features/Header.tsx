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
    <header className="site-header">
      <NavLink className="brand-link" to="/">
        Lensamity
      </NavLink>

      {user ? (
        <nav className="site-nav" aria-label="Account">
          <span className="nav-user">{user.displayName}</span>
          <button className="nav-button" disabled={isLoading} onClick={handleLogout} type="button">
            Log out
          </button>
        </nav>
      ) : (
        <nav className="site-nav" aria-label="Auth">
          <NavLink to="/login" end>
            Log in
          </NavLink>
          <NavLink className="nav-button" to="/signup" end>
            Sign up
          </NavLink>
        </nav>
      )}
    </header>
  );
};

export default Header;

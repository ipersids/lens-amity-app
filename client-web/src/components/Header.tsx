import { NavLink } from "react-router";

const Header = () => {
  return (
    <header>
      <NavLink to="/">Lensamity</NavLink>
      <NavLink to="/login" end>
        Log in
      </NavLink>
      <NavLink to="/signup" end>
        Sign up
      </NavLink>
    </header>
  );
};

export default Header;

import { NavLink, Outlet } from "react-router";

const HeaderComponent = () => {
  return (
    <header className="min-w-2xl h-14 border-b px-6 flex justify-between items-center">
      <NavLink to="/" className="uppercase">
        Lensamity
      </NavLink>
      <nav className="flex gap-5">
        <NavLink to="/login" end>
          Log in
        </NavLink>
        <NavLink to="/signup" end>
          Sign up
        </NavLink>
      </nav>
    </header>
  );
};

const Header = () => {
  return (
    <div>
      <HeaderComponent />
      <Outlet />
    </div>
  );
};

export default Header;

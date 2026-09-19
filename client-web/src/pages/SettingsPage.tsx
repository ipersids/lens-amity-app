import { NavLink, Outlet } from "react-router";

const SettingsPage = () => {
  return (
    <section className="settings-page">
      <nav className="settings-nav" aria-label="Settings">
        <h1>Settings</h1>
        <NavLink to="profile" className={({ isActive }) => (isActive ? "active" : undefined)}>
          Profile
        </NavLink>
        <NavLink to="account" className={({ isActive }) => (isActive ? "active" : undefined)}>
          Account
        </NavLink>
      </nav>
      <Outlet />
    </section>
  );
};

export default SettingsPage;

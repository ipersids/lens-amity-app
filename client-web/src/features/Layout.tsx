import { Outlet } from "react-router";
import Footer from "./Footer";
import Header from "./Header";

const Layout = () => {
  return (
    <section className="layout">
      <Header />
      <main>
        <Outlet />
      </main>
      <Footer />
    </section>
  );
};

export default Layout;

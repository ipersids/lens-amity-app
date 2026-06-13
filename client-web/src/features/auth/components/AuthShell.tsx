import type { ReactNode } from "react";
import { Link } from "react-router";

type AuthShellProps = {
  title: string;
  switchTo: string;
  switchText: string;
  children: ReactNode;
};

const AuthShell = ({ title, switchTo, switchText, children }: AuthShellProps) => {
  return (
    <section className="auth-page">
      <div className="auth-form">
        <h1>{title}</h1>

        {children}

        <Link to={switchTo}>{switchText}</Link>
      </div>
    </section>
  );
};

export default AuthShell;

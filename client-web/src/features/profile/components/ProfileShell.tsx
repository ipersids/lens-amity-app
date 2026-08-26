import type { ReactNode } from "react";

const ProfileShell = ({ children }: { children: ReactNode }) => {
  return <section className="profile">{children}</section>;
};

export default ProfileShell;

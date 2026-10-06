import type { ReactNode } from "react";

const PhotoShell = ({ children }: { children: ReactNode }) => {
  return <section className="profile">{children}</section>;
};

export default PhotoShell;

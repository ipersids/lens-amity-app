import type { ReactNode } from "react";

const PhotoShell = ({ children }: { children: ReactNode }) => {
  return <section className="photo-page">{children}</section>;
};

export default PhotoShell;

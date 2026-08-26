import PhotosFeed from "./components/PhotosFeed";

export { PhotosFeed };

type Photo = {
  title: string;
  date: string;
  imageUrl: string;
};

export const profilePhotosList: Photo[] = [
  {
    title: "City light",
    date: "2026-08-24",
    imageUrl: "https://picsum.photos/seed/city-light/900/900",
  },
  {
    title: "Slow morning",
    date: "2026-08-22",
    imageUrl: "https://picsum.photos/seed/slow-morning/900/900",
  },
  {
    title: "Gallery wall",
    date: "2026-08-20",
    imageUrl: "https://picsum.photos/seed/gallery-wall/900/900",
  },
  {
    title: "Coffee table",
    date: "2026-08-18",
    imageUrl: "https://picsum.photos/seed/coffee-table/900/900",
  },
  {
    title: "Window seat",
    date: "2026-08-16",
    imageUrl: "https://picsum.photos/seed/window-seat/900/900",
  },
  {
    title: "Weekend walk",
    date: "2026-08-14",
    imageUrl: "https://picsum.photos/seed/weekend-walk/900/900",
  },
  {
    title: "Studio corner",
    date: "2026-08-12",
    imageUrl: "https://picsum.photos/seed/studio-corner/900/900",
  },
  {
    title: "Evening sky",
    date: "2026-08-10",
    imageUrl: "https://picsum.photos/seed/evening-sky/900/900",
  },
  {
    title: "Quiet detail",
    date: "2026-08-08",
    imageUrl: "https://picsum.photos/seed/quiet-detail/900/900",
  },
];

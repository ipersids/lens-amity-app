import { useUser } from "../stores/auth";

const profilePhotos = [
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

const ProfilePage = () => {
  const user = useUser();
  const displayName = user?.displayName || "Lensamity Creator";
  const username = user?.username || "creator";

  return (
    <section className="profile">
      <header className="profile-header">
        <div className="profile-avatar" role="img" aria-label={`${displayName} avatar`}>
          {displayName.charAt(0).toUpperCase()}
        </div>

        <div className="profile-info">
          <div>
            <h1>{displayName}</h1>
            <p className="profile-username">@{username}</p>
          </div>

          <p className="profile-about">
            Collecting small moments, soft light, favorite places, and photos worth coming
            back to.
          </p>

          <dl className="profile-count" aria-label="Profile stats">
            <div>
              <dd>{profilePhotos.length}</dd>
              <dt>photos</dt>
            </div>
          </dl>
        </div>
      </header>

      <ul className="profile-photos" aria-label="Profile photos">
        {profilePhotos.map((photo) => (
          <li className="profile-photo" key={photo.title}>
            <a href={photo.imageUrl} aria-label={`${photo.title}, ${photo.date}`}>
              <img src={photo.imageUrl} alt={photo.title} />
              <span className="profile-photo-name">{photo.title}</span>
              <time dateTime={photo.date}>{photo.date}</time>
            </a>
          </li>
        ))}
      </ul>
    </section>
  );
};

export default ProfilePage;

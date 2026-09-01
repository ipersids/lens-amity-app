import { profilePhotosList } from "../index";

const PhotosFeed = ({ username }: { username: string }) => {
  void username;
  return (
    <ul className="profile-photos" aria-label="Profile photos">
      {profilePhotosList.map((photo) => (
        <li className="profile-photo" key={photo.title}>
          <a href={photo.imageUrl} aria-label={`${photo.title}, ${photo.date}`}>
            <img src={photo.imageUrl} alt={photo.title} />
            <span className="profile-photo-name">{photo.title}</span>
            <time dateTime={photo.date}>{photo.date}</time>
          </a>
        </li>
      ))}
    </ul>
  );
};

export default PhotosFeed;
